# DiskSizeScan

A fast, minimal disk-usage analyzer in the spirit of TreeSize — built with
[Wails](https://wails.io) (Go) + Svelte. Scan a drive or folder, then browse a
size-sorted tree with inline usage bars.

![overview](docs/screenshot.png)

## Features

- ⚡ **Parallel scan engine** — bounded goroutine pool over `os.ReadDir`, no
  admin rights required, works on any drive / filesystem / network share.
- 💽 **Start-up drive dashboard** — each drive's used / total capacity at a
  glance; click a card to scan it.
- 🌳 **Lazy tree UI** — the full tree is held once in Go; the frontend only
  pulls a node's children (size-sorted) when you expand it, so even
  million-file drives stay responsive.
- 📊 Per-item **usage bar + percentage of the whole drive**, color-graded by
  size — consistent with the dashboard.
- 🖱️ **Right-click → scan this folder** to re-root the scan on any subfolder.
- 🛑 **Cancelable** scans with live progress (files / bytes / current path).
- 🪟 **Cross-platform** — Windows, macOS, Linux.

## How the scan works

The scanner (`scanner/scanner.go`) recurses directories concurrently with a
bounded worker pool sized to the CPU count. The key trick for staying both fast
and deadlock-free under a bounded pool:

```go
select {
case sem <- struct{}{}: // a worker slot is free → scan the subdir on a goroutine
    go scanDir(child)
default:                 // pool is saturated → scan it inline, right here
    scanDir(child)
}
```

Each `Node` is written by exactly one goroutine (the one scanning its parent),
and a parent only reads child results after a `WaitGroup` barrier — so the hot
path needs no locks. IDs are assigned in a single pass *after* the scan.

> **Note:** this is the portable parallel-walk approach. It does **not** read
> the NTFS MFT directly (the WizTree technique), which would be faster on
> Windows but requires admin rights and is NTFS-only.

## Prerequisites

- [Go](https://go.dev/dl/) 1.22+
- [Node.js](https://nodejs.org/) 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation):
  `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

Run `wails doctor` to confirm your toolchain is ready.

## Run / build

```bash
# one-time: pull Go deps
go mod tidy

# live-reload dev app
wails dev

# production binary -> build/bin/
wails build
```

If `go mod tidy` reports the pinned Wails version is unavailable, refresh it
with `go get github.com/wailsapp/wails/v2@latest && go mod tidy`.

## Usage

1. On launch you get a **drive dashboard** — click a drive to scan it, or type
   any path in the box and press **Scan** / Enter.
2. Expand folders to drill in; rows are sorted largest-first and bars show each
   item's share of the whole drive.
3. **Right-click** a folder → *이 폴더 스캔* to re-scan from there; click **↗**
   (or right-click → open) to reveal it in your OS file manager.

## Project layout

```
main.go               Wails app entry + asset embed
app.go                Bound API: Scan / GetChildren / Cancel / ListDrives / ListDriveInfo / OpenPath
capacity_windows.go   Drive total/free via GetDiskFreeSpaceEx
capacity_unix.go      Drive total/free via statfs (macOS/Linux)
scanner/scanner.go    Parallel directory walker + size tree
frontend/src/
  App.svelte             Controls, status, dashboard/tree host, context menu
  lib/DriveDashboard.svelte  Start-up per-drive usage cards
  lib/TreeNode.svelte    Recursive tree row (lazy load, drive-relative bar)
  lib/contextmenu.js     Shared right-click menu store
  wails.js               Wrappers over the injected window.go / window.runtime
```

## Limitations

- Symlinks / reparse points are skipped (avoids cycles and double counting).
- Very large directories are capped at 500 visible children per expand; the
  remainder is shown as a "… N more items hidden" line.
- Every file is kept as a node for completeness, so peak memory scales with the
  total file count (~hundreds of MB for multi-million-file drives).

## License

[MIT](LICENSE)
