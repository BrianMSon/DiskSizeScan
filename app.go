package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"disksizescan/scanner"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// childCap bounds how many children GetChildren returns at once so a directory
// with hundreds of thousands of entries can't stall the UI. The remainder is
// reported via Total/Truncated.
const childCap = 500

// NodeDTO is the serializable view of a tree node sent to the frontend.
type NodeDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	IsDir       bool   `json:"isDir"`
	Files       int64  `json:"files"`
	HasChildren bool   `json:"hasChildren"`
}

// ScanResult summarizes a completed scan. Drive* fields describe the volume
// that holds the scanned path, so the UI can show sizes relative to it.
type ScanResult struct {
	Root       NodeDTO `json:"root"`
	TotalSize  int64   `json:"totalSize"`
	TotalFiles int64   `json:"totalFiles"`
	DurationMs int64   `json:"durationMs"`
	DriveTotal int64   `json:"driveTotal"`
	DriveUsed  int64   `json:"driveUsed"`
	DriveFree  int64   `json:"driveFree"`
}

// DriveInfo is a drive's capacity, used for the start-up dashboard.
type DriveInfo struct {
	Path  string `json:"path"`
	Total int64  `json:"total"`
	Used  int64  `json:"used"`
	Free  int64  `json:"free"`
}

// ChildrenResult is the (possibly truncated) list of a node's children.
type ChildrenResult struct {
	Items     []NodeDTO `json:"items"`
	Total     int       `json:"total"`
	Truncated bool      `json:"truncated"`
}

// Progress is emitted on the "scan:progress" event while a scan runs.
type Progress struct {
	Files int64  `json:"files"`
	Bytes int64  `json:"bytes"`
	Path  string `json:"path"`
}

// App is the Wails-bound backend.
type App struct {
	ctx    context.Context
	mu     sync.Mutex
	nodes  []*scanner.Node
	cancel context.CancelFunc
}

// NewApp creates the App.
func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// Scan walks path, stores the resulting tree, and returns a summary. Progress
// is streamed via the "scan:progress" event roughly every 100ms.
func (a *App) Scan(path string) (ScanResult, error) {
	scanCtx, cancel := context.WithCancel(a.ctx)
	a.mu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	a.cancel = cancel
	a.nodes = nil
	a.mu.Unlock()

	sc := scanner.New()

	done := make(chan struct{})
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				wruntime.EventsEmit(a.ctx, "scan:progress", Progress{
					Files: sc.Files(), Bytes: sc.Bytes(), Path: sc.CurrentPath(),
				})
			}
		}
	}()

	start := time.Now()
	root, err := sc.Scan(scanCtx, path)
	close(done)
	if err != nil {
		return ScanResult{}, err
	}

	nodes := scanner.Index(root)
	a.mu.Lock()
	a.nodes = nodes
	a.mu.Unlock()

	// Final progress tick so the UI lands on exact totals.
	wruntime.EventsEmit(a.ctx, "scan:progress", Progress{
		Files: sc.Files(), Bytes: sc.Bytes(), Path: "",
	})

	total, free, _ := driveCapacity(path)

	return ScanResult{
		Root:       toDTO(root),
		TotalSize:  root.Size,
		TotalFiles: root.Files,
		DurationMs: time.Since(start).Milliseconds(),
		DriveTotal: int64(total),
		DriveUsed:  int64(total - free),
		DriveFree:  int64(free),
	}, nil
}

// GetChildren returns the children of the node with the given ID, sorted by
// sortBy ("size" | "usage" | "name" | "files") in the requested direction. When
// foldersOnly is set, files are excluded. Sorting/filtering run over the full
// child list before the childCap is applied, so the returned page is correct
// for the chosen order even when truncated.
func (a *App) GetChildren(id int, sortBy string, asc bool, foldersOnly bool) ChildrenResult {
	a.mu.Lock()
	nodes := a.nodes
	a.mu.Unlock()
	if id < 0 || id >= len(nodes) {
		return ChildrenResult{}
	}

	n := nodes[id]

	// Copy into a local slice (folders only when requested) so concurrent
	// calls never mutate the shared slice.
	var kids []*scanner.Node
	if foldersOnly {
		for _, c := range n.Children {
			if c.IsDir {
				kids = append(kids, c)
			}
		}
	} else {
		kids = make([]*scanner.Node, len(n.Children))
		copy(kids, n.Children)
	}
	sortNodes(kids, sortBy, asc)

	total := len(kids)
	limit := total
	truncated := false
	if limit > childCap {
		limit, truncated = childCap, true
	}

	items := make([]NodeDTO, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, toDTO(kids[i]))
	}
	return ChildrenResult{Items: items, Total: total, Truncated: truncated}
}

func sortNodes(nodes []*scanner.Node, sortBy string, asc bool) {
	var less func(a, b *scanner.Node) bool
	switch sortBy {
	case "name":
		less = func(a, b *scanner.Node) bool {
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
	case "files":
		less = func(a, b *scanner.Node) bool { return a.Files < b.Files }
	default: // "size" and "usage" (usage % is proportional to size)
		less = func(a, b *scanner.Node) bool { return a.Size < b.Size }
	}
	sort.Slice(nodes, func(i, j int) bool {
		if asc {
			return less(nodes[i], nodes[j])
		}
		return less(nodes[j], nodes[i])
	})
}

// Cancel stops the in-progress scan, if any.
func (a *App) Cancel() {
	a.mu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	a.mu.Unlock()
}

// ListDrives returns scannable roots (drive letters on Windows, "/" elsewhere).
func (a *App) ListDrives() []string {
	if runtime.GOOS == "windows" {
		var drives []string
		for c := 'A'; c <= 'Z'; c++ {
			p := string(c) + ":\\"
			if _, err := os.Stat(p); err == nil {
				drives = append(drives, p)
			}
		}
		return drives
	}
	return []string{"/"}
}

// ListDriveInfo returns capacity (total/used/free) for each scannable root.
// Used by the start-up dashboard; a drive whose capacity can't be read comes
// back with Total == 0.
func (a *App) ListDriveInfo() []DriveInfo {
	roots := a.ListDrives()
	out := make([]DriveInfo, 0, len(roots))
	for _, r := range roots {
		total, free, err := driveCapacity(r)
		if err != nil {
			out = append(out, DriveInfo{Path: r})
			continue
		}
		out = append(out, DriveInfo{
			Path:  r,
			Total: int64(total),
			Used:  int64(total - free),
			Free:  int64(free),
		})
	}
	return out
}

// OpenPath reveals a file or folder in the OS file manager.
func (a *App) OpenPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func toDTO(n *scanner.Node) NodeDTO {
	return NodeDTO{
		ID:          n.ID,
		Name:        n.Name,
		Path:        n.Path(),
		Size:        n.Size,
		IsDir:       n.IsDir,
		Files:       n.Files,
		HasChildren: len(n.Children) > 0,
	}
}
