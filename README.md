# cmdit — Text editor for humans in the terminal

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![Tests](https://img.shields.io/badge/tests-175%2B%20passing-brightgreen)]()
[![Version](https://img.shields.io/badge/version-v0.4.2-blue)](https://github.com/alexlivre/cmdit/releases)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Platforms](https://img.shields.io/badge/platforms-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)]()

**cmdit** is a terminal text editor made for people — not machines.
Open, type, exit. Works like Notepad. **Zero learning curve.**

Designed to run both on your desktop and via SSH on servers.
A single binary, zero dependencies.

---

## Why does it exist?

| Editor | Problem |
|--------|---------|
| **Vim** | Modes, `:wq`, `hjkl`, months-long learning curve |
| **Nano** | Too limited, no mouse, no undo |
| **Emacs** | `Ctrl+X Ctrl+C` to exit |
| **Helix** | Confusing `selection→action` paradigm |
| **Micro** | Abandoned, no native LSP |

**cmdit** solves: powerful like Vim, simple like Nano, modern like Helix.

---

## ✨ Features

- ✅ **Modeless** — type like in any editor. No Normal/Insert modes.
- ✅ **Mouse** — click to position cursor, drag to select, scroll to navigate.
- ✅ **Familiar shortcuts** — Ctrl+S, Ctrl+Z, Ctrl+C, Ctrl+V. CUA standard.
- ✅ **Unlimited Undo/Redo** — Ctrl+Z / Ctrl+Y.
- ✅ **Command Palette** — Ctrl+P accesses all commands with fuzzy search.
- ✅ **Syntax Highlighting** — Automatic language detection. dark/light/monokai/dracula/solarized themes.
- ✅ **Find and replace** — Ctrl+F / Ctrl+H with highlight of all occurrences.
- ✅ **Integrated file picker** — Ctrl+O to open files browsing directories.
- ✅ **Rename file** — F2 renames the current file with inline input.
- ✅ **Auto-save** — Automatically saves every 30 seconds. `F9` toggles on/off. `[AutoSave]` indicator on the status bar.
- ✅ **Welcome screen** — Shows recent files when opening without a file.
- ✅ **Confirmation dialog** — Warns when quitting with unsaved changes.
- ✅ **Tabs** — Ctrl+T new tab, Ctrl+W close, F8 next, F7 previous, Ctrl+1-9 jump. ● indicator for modified files.
- ✅ **Multiple cursors** — Ctrl+D adds cursor at the next occurrence of the word, Escape clears. Simultaneous editing across all cursors.
- ✅ **Splits** — Ctrl+\ splits the screen horizontally. Click to focus pane, Ctrl+\ toggles focus.
- ✅ **Indent guides** — Subtle vertical lines at indent levels.
- ✅ **LSP client** — Auto-starts gopls for Go. Error and warning diagnostics on the status bar. Extensible to Python, Rust, TypeScript.
- ✅ **Auto-close brackets** — Automatically closes `() [] {} "" '' \`\``. Smart-skip: typing the closing bracket skips the existing character.
- ✅ **Built-in Vim mode** — `F5` activates Vim mode (Normal/Insert/Visual/Command) with `hjkl`, `dd`, `yy`, `p`, `u`, `:w`, `:q`, `:wq`. Built-in, no plugins.
- ✅ **Theme switching** — `F6` cycles through 5 themes in real time: dark, light, monokai, dracula, solarized-dark.
- ✅ **Word wrap** — `Alt+Z` toggles line wrapping in the viewport.
- ✅ **Format on save** — Automatically formats code on save (gofmt, black, rustfmt).
- ✅ **JSON configuration** — `~/.cmdit/config.json` persists preferences across sessions.
- ✅ **Customizable keybindings** — Remap any key via `config.json`.
- ✅ **Single binary** — One file. Copy and run. Zero dependencies.
- ✅ **Cross-platform** — Windows, Linux, macOS. Desktop and SSH server.

### 📋 Planned (v5)

- 🔲 Treesitter for precise syntax highlighting
- 🔲 Rope data structure for large files
- 🔲 Auto-complete popup (LSP)

---

## ⚡ Installation

### One-liner (recommended)

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/alexlivre/cmdit/main/install.ps1 | iex
```

**Linux / macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/alexlivre/cmdit/main/install.sh | bash
```

### Pre-compiled

```bash
# Download the binary for your platform at:
# https://github.com/alexlivre/cmdit/releases

# Linux/macOS
chmod +x cmdit
sudo mv cmdit /usr/local/bin/

# Windows
# Move cmdit.exe to a folder in PATH
```

### Go Install

```bash
go install github.com/alexlivre/cmdit/cmd/cmdit@latest
```

### Build from source

```bash
git clone https://github.com/alexlivre/cmdit.git
cd cmdit
go build -o cmdit ./cmd/cmdit
```

---

## 🚀 Usage

```bash
# Open a file
cmdit README.md

# Create new file
cmdit new-file.txt

# Open without file (welcome screen)
cmdit
```

---

## ⌨️ Shortcut guide

### File

| Shortcut | Action |
|---------|--------|
| `Ctrl+S` | Save |
| `Ctrl+O` | Open file |
| `F3` | Save as |
| `F2` | Rename file |
| `F9` | Toggle auto-save |
| `Ctrl+Q` | Quit |

### Tabs and Splits

| Shortcut | Action |
|---------|--------|
| `Ctrl+T` | New tab |
| `Ctrl+W` | Close tab |
| `F8` | Next tab |
| `F7` | Previous tab |
| `Ctrl+1-9` | Go to tab N |
| `Ctrl+\` | Create/toggle split |

### Editing

| Shortcut | Action |
|---------|--------|
| `Ctrl+Z` | Undo |
| `Ctrl+Y` | Redo |
| `Ctrl+C` | Copy (selection or current line) |
| `Ctrl+X` | Cut |
| `Ctrl+V` | Paste |
| `Ctrl+A` | Select all |
| `Backspace` | Delete character to the left |
| `Delete` | Delete character to the right |
| `Tab` | Insert 4 spaces |
| `Ctrl+D` | Add cursor (next occurrence) |
| `Escape` | Clear extra cursors |

### Navigation

| Shortcut | Action |
|---------|--------|
| `↑ ↓ ← →` | Move cursor |
| `Ctrl+← / Ctrl+→` | Jump words |
| `Home / End` | Line start / end |
| `Ctrl+Home / Ctrl+End` | File start / end |
| `PageUp / PageDown` | Scroll page |
| `Mouse click` | Position cursor |
| `Mouse scroll` | Scroll |

### Search

| Shortcut | Action |
|---------|--------|
| `Ctrl+F` | Find |
| `Ctrl+H` | Find and replace |

### Commands

| Shortcut | Action |
|---------|--------|
| `Ctrl+P` | Command palette |
| `Esc` | Close palette / cancel |

---

## 🆚 Competitor comparison

|  | **cmdit** | Vim | Nano | Helix | Micro |
|--|-----------|-----|------|-------|-------|
| Modeless | ✅ | ❌ | ✅ | ⚠️ | ✅ |
| Mouse | ✅ | ❌ | ✅ | ❌ | ✅ |
| Ctrl+C/V/Z | ✅ | ❌ | ❌ | ❌ | ✅ |
| Command Palette | ✅ | ❌ | ❌ | ✅ | ✅ |
| Undo/Redo | ✅ | ✅ | ❌ | ✅ | ✅ |
| Syntax Highlight | ✅ | ✅ | ✅ | ✅ | ✅ |
| File Picker | ✅ | ❌ | ❌ | ❌ | ❌ |
| Auto-save | ✅ | ❌ | ❌ | ❌ | ❌ |
| Tabs | ✅ | ✅ | ❌ | ❌ | ✅ |
| Splits | ✅ | ✅ | ❌ | ✅ | ✅ |
| Multi-cursor | ✅ | ❌ | ❌ | ✅ | ✅ |
| Native LSP | ✅ | ❌ | ❌ | ✅ | ❌ |
| Single Binary | ✅ | ✅ | ✅ | ✅ | ✅ |
| Learning curve | 🟢 seconds | 🔴 months | 🟢 seconds | 🟡 days | 🟢 minutes |

---

## 🏗️ Architecture

```
cmdit/
├── cmd/cmdit/main.go            # Entry point (SplitContainer → TabManager → editor.Model)
├── internal/
│   ├── buffer/                  # Gap buffer + cursor + undo stack + line cache
│   ├── clipboard/               # Internal clipboard
│   ├── command/                 # Command palette + action registry
│   ├── editor/                  # Main Bubble Tea model
│   │   ├── editor.go            # Model, Init, Update, helpers
│   │   ├── state.go             # State structs (Search, Palette, FilePicker, etc.)
│   │   ├── interfaces.go        # Interfaces (FileLoader, SyntaxHighlighter, etc.)
│   │   ├── view.go              # Main view + rendering
│   │   ├── view_welcome.go      # Welcome screen rendering
│   │   ├── view_palette.go      # Command palette rendering
│   │   ├── view_filepicker.go   # File picker rendering
│   │   ├── view_modes.go        # Other mode renderings (save-as, confirm, etc.)
│   │   ├── keys.go              # Main key dispatch
│   │   ├── keys_search.go       # Search/replace key handling
│   │   ├── keys_palette.go      # Command palette key handling
│   │   ├── keys_filepicker.go   # File picker key handling
│   │   ├── keys_modes.go        # Other mode key handling
│   │   ├── actions.go           # Commands (save, undo, clipboard, etc.)
│   │   └── lsp_integration.go   # LSP integration
│   ├── fileio/                  # File load/save/rename (with size limits)
│   ├── highlight/               # Syntax highlighting via Chroma
│   ├── lsp/                     # LSP client (JSON-RPC 2.0)
│   ├── renderer/                # Viewport (scroll, resize)
│   └── tabs/                    # Tabs and splits
│       ├── tabmanager.go        # TabManager (editor.Model container)
│       └── split.go             # SplitContainer (pane layout)
├── bin/                         # Compiled binaries
├── go.mod / go.sum              # Go dependencies
├── Makefile                     # Build, test, cross-compile
└── .opencode/plans/             # Implementation plans
```

### Technology stack

| Layer | Technology | Reason |
|-------|------------|--------|
| Language | **Go 1.24+** | Single binary, cross-compile, predictable performance |
| TUI | **Bubble Tea + Lip Gloss** | Elm-like architecture, composable components |
| Syntax | **Chroma** | 500+ languages, no CGo, fast |
| Text structure | **Gap Buffer** | O(1) for localized edits, simple |

### Design decisions

- **Modeless-first**: the editor is always in insertion mode. Commands are accessed via CUA shortcuts or palette.
- **Progressive disclosure**: advanced features are gradually discovered via command palette.
- **Single binary**: native static compilation in Go. Trivial distribution (`scp cmdit server:/usr/local/bin/`).

---

## 🧪 Tests

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover

# Current result
# 175+ tests passing across 8 packages
# buffer (86.5%) | clipboard (100%) | command (100%) | editor (45.5%)
# fileio (90.9%) | highlight (59.3%) | renderer (84.8%) | tabs (26.2%)
# go vet ./... ✅ no warnings
# 5 benchmarks passing
```

---

## 📦 Build

```bash
# Local build
make build          # → bin/cmdit.exe

# Cross-compile (all platforms)
make build-all      # → bin/cmdit-linux-amd64
                    #   bin/cmdit-windows-amd64.exe
                    #   bin/cmdit-darwin-amd64
                    #   bin/cmdit-darwin-arm64

# Tests + lint
make test
make lint
```

---

## 🗺️ Roadmap

| Phase | Status | Description |
|------|--------|-------------|
| 0 | ✅ | Scaffold: Go + Bubble Tea |
| 1 | ✅ | Basic editor: type, save, exit |
| 2 | ✅ | Navigation: arrows, mouse, scroll, viewport |
| 3 | ✅ | Power editing: undo, clipboard, search |
| 4 | ✅ | Command palette |
| 5 | ✅ | Syntax highlighting via Chroma |
| 6 | ✅ | File picker, welcome screen, recent files |
| 7 | ✅ | Auto-save, cross-compile, polish |
| 8 | ✅ | Multiple cursors, tabs, splits, indent guides, LSP |
| 9 | ✅ | Built-in features: auto-close, vim mode, themes, format, keybindings |
| 10 | ✅ | Quality & stability: errors, security, coverage, benchmarks |
| 11 | ✅ | Production: installers, quickstart, cheatsheet |
| 12 | ✅ | Performance optimization: line cache, MoveGapTo O(n), search optimizations |
| 13 | ✅ | Code quality: state structs, file organization, interfaces, error handling |

---

## 🤝 Contributing

Contributions are welcome!

1. Fork the repository
2. Create a branch: `git checkout -b feature/name`
3. Commit: `git commit -m "feat: description"`
4. Push: `git push origin feature/name`
5. Open a Pull Request

**Commit conventions:**
`feat:` new feature
`fix:` bug fix
`refactor:` refactoring
`test:` tests
`docs:` documentation
`chore:` maintenance

---

## 📄 License

MIT © 2026 Alexb

---

**cmdit** — *text editor for humans.*