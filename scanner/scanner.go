// Package scanner walks a directory tree in parallel and builds an in-memory
// size tree. The hot path is lock-free: each Node is written by exactly one
// goroutine (the one scanning its parent directory), and parents read child
// results only after a WaitGroup barrier.
package scanner

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

// Node is a single file or directory in the scanned tree.
//
// Exported fields are read by the App layer after a scan completes (single
// threaded, so no synchronization is needed there). `parent` and `sorted`
// stay internal to this package.
type Node struct {
	ID       int     // assigned after the scan, used by the frontend as a handle
	Name     string  // base name; the root keeps its full path here
	Size     int64   // total bytes in this subtree
	IsDir    bool    //
	Files    int64   // total file count in this subtree
	ModTime  int64   // last-modified time of this entry (unix seconds)
	Children []*Node //
	parent   *Node
}

// Path reconstructs the absolute path by walking up to the root.
func (n *Node) Path() string {
	if n.parent == nil {
		return n.Name
	}
	return filepath.Join(n.parent.Path(), n.Name)
}

// Scanner holds the bounded worker pool and live progress counters.
type Scanner struct {
	sem     chan struct{}
	files   atomic.Int64
	bytes   atomic.Int64
	curPath atomic.Value // string
	ctx     context.Context
}

// New creates a scanner. Concurrency scales with CPU count; disk I/O benefits
// from over-subscription on SSDs, and the inline fallback keeps HDDs sane.
func New() *Scanner {
	limit := max(runtime.NumCPU()*4, 8)
	return &Scanner{sem: make(chan struct{}, limit)}
}

// Files / Bytes / CurrentPath expose live progress for the UI.
func (s *Scanner) Files() int64 { return s.files.Load() }
func (s *Scanner) Bytes() int64 { return s.bytes.Load() }
func (s *Scanner) CurrentPath() string {
	if v := s.curPath.Load(); v != nil {
		return v.(string)
	}
	return ""
}

// Scan walks root and returns the populated tree. The returned tree is always
// usable even if ctx was canceled mid-scan (it reflects whatever was visited).
func (s *Scanner) Scan(ctx context.Context, root string) (*Node, error) {
	s.ctx = ctx
	s.files.Store(0)
	s.bytes.Store(0)

	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}

	node := &Node{Name: root, IsDir: info.IsDir(), ModTime: info.ModTime().Unix()}
	if node.IsDir {
		s.scanDir(node, root)
	} else {
		node.Size = info.Size()
		node.Files = 1
	}
	return node, nil
}

// scanDir reads one directory, accounts its files, and recurses into subdirs.
// Subdirectories run on the worker pool when a slot is free, otherwise inline
// in the current goroutine — that inline fallback is what makes recursive
// fan-out deadlock-free under a bounded pool.
func (s *Scanner) scanDir(node *Node, path string) {
	if s.ctx.Err() != nil {
		return
	}
	s.curPath.Store(path)

	entries, err := os.ReadDir(path)
	if err != nil {
		return // unreadable (permissions, vanished dir, ...) — skip silently
	}

	var wg sync.WaitGroup
	for _, e := range entries {
		if s.ctx.Err() != nil {
			break
		}
		// Skip symlinks/reparse points to avoid cycles and double counting.
		if e.Type()&fs.ModeSymlink != 0 {
			continue
		}

		if e.IsDir() {
			child := &Node{Name: e.Name(), IsDir: true, parent: node}
			if di, err := e.Info(); err == nil {
				child.ModTime = di.ModTime().Unix()
			}
			node.Children = append(node.Children, child)
			childPath := filepath.Join(path, e.Name())

			select {
			case s.sem <- struct{}{}:
				wg.Add(1)
				go func(c *Node, p string) {
					defer wg.Done()
					defer func() { <-s.sem }()
					s.scanDir(c, p)
				}(child, childPath)
			default:
				s.scanDir(child, childPath) // no slot free → run here
			}
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}
		size := info.Size()
		node.Children = append(node.Children, &Node{
			Name: e.Name(), Size: size, Files: 1, ModTime: info.ModTime().Unix(), parent: node,
		})
		node.Size += size
		node.Files++
		s.files.Add(1)
		s.bytes.Add(size)
	}

	wg.Wait()
	for _, c := range node.Children {
		if c.IsDir {
			node.Size += c.Size
			node.Files += c.Files
		}
	}
}

// Index assigns sequential IDs via a single-threaded DFS and returns a slice
// where slice[id] == node. Called once after a scan completes.
func Index(root *Node) []*Node {
	list := make([]*Node, 0, 1024)
	var walk func(n *Node)
	walk = func(n *Node) {
		n.ID = len(list)
		list = append(list, n)
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return list
}
