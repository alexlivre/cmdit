# cmdit — Terminal Text Editor

## Info
- **Repository:** https://github.com/alexlivre/cmdit
- **Location:** `C:\code\cmdit`
- **Status:** v0.1.0 released, active development
- **Language:** Go 1.23+ with Bubble Tea framework

## Architecture

| Layer | Technology |
|-------|------------|
| TUI | Bubble Tea + Lip Gloss (charm.sh) |
| Syntax | Chroma v2 |
| Text storage | Gap buffer v1 (Rope v2 planned) |
| Package structure | `cmd/`, `internal/{buffer,clipboard,command,editor,fileio,highlight,renderer}` |

## Features Implemented (v0.1.0)
- Modeless editing (no Vim modes)
- Mouse support: click, drag, scroll
- Keyboard: Ctrl+S/Z/C/V/F/H/O/Q
- Command palette: Ctrl+Shift+P
- Syntax highlighting: 50+ languages, 5 themes
- File operations: open, save, save-as
- File picker, welcome screen
- Auto-save every 30s
- Undo/Redo unlimited
- Search & Replace

## Features NOT Yet Implemented
- ❌ Rename file (no current implementation)
- ❌ Multiple cursors
- ❌ Tabs / Splits
- ❌ LSP client
- ❌ Treesitter
- ❌ Lua plugins

## Key Files
- `cmd/cmdit/main.go` — entry point
- `internal/editor/editor.go` — main Bubble Tea model (1480+ lines)
- `internal/buffer/buffer.go` — gap buffer implementation
- `internal/highlight/syntax.go` — Chroma integration

## Testing
- 50 tests across 5 packages
- All passing
- Run: `go test ./...`
- Coverage: buffer (17), clipboard (4), editor (14), fileio (4), renderer (9)

## Build
- `make build` → `bin/cmdit.exe` (8.8 MB)
- Cross-compile: `make build-all`
- CI/CD: GitHub Actions (test matrix 3×3, release on tags)