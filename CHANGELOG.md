# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.2] — 2026-05-01

### Added — Phase 10 & 11: Quality, Stability, Production

- **English translation** — All UI strings, palette commands, status bar, confirm dialogs, welcome screen, and error messages translated to English. README, QUICKSTART, and CHEATSHEET in English.
- **Error handling** — `internal/editor/errors.go` with `showError()` (status bar overlay in red), `logError()` (writes to `~/.cmdit/cmdit.log` with stack trace), and `safeRun()` (panic recovery wrapper).
- **Security** — Path traversal prevention via `filepath.Clean()` in `openFile()`. Escape sequence sanitization test. Filename validation blocks `../` and special characters.
- **Installers** — `install.sh` (Linux/macOS) and `install.ps1` (Windows PowerShell). One-liner install: `curl -sSL .../install.sh | bash` or `irm .../install.ps1 | iex`.
- **Documentation** — `QUICKSTART.md` ("5 minutes to master cmdit"), `CHEATSHEET.md` (complete keyboard shortcut reference).
- **Benchmarks** — 5 benchmarks: `BenchmarkBufferInsert`, `BenchmarkBufferInsertWithCursor`, `BenchmarkRenderSmallFile`, `BenchmarkRenderLargeFile`, `BenchmarkSearchInLargeFile`.
- **Edge case tests** — Binary file handling, empty file, single line 10K chars, undo/redo to empty, multi-cursor dedup, vim mode normal behavior.
- **Welcome screen** — Version display updated to `v0.4.2 — Text editor for humans`.

### Changed

- **Coverage** — Editor: 34% → 46%. Tabs: 22% → 26%. Overall: 36% → 40%.
- **Total tests:** 86+ (was 75). All `go vet` clean.
- **Config tests** — Now use temporary directories (no longer depend on real `~/.cmdit/config.json`).
- **Roadmap** — All 11 phases marked complete.

---

## [0.4.1] — 2026-05-01

### Changed

- **Tab navigation** — `F7` (previous tab) and `F8` (next tab) replace `Ctrl+Tab`/`Ctrl+V`. `Ctrl+V` restored as Paste. Reason: `Ctrl+Tab` captured by terminal emulators, `Ctrl+V` conflicted with paste.
- **First save** — `Ctrl+S` on an unnamed file now opens the "Save As" prompt instead of defaulting to `untitled.txt`.
- **Welcome screen** — Version bumped to `v0.4.1`.

### Added

- **Auto-save toggle** — `F9` toggles auto-save on/off. `[AutoSave]` indicator in status bar. `auto_save_enabled` field in `config.json` (default: true).
- **Config tests** — `AutoSaveEnabled` field tested in roundtrip, defaults, and save/load.

---

## [0.4.0] — 2026-05-01

### Added — Fase 9: Built-in Features

- **Config System** — `~/.cmdit/config.json` with fields: `auto_close_enabled`, `vim_mode`, `format_on_save`, `word_wrap`, `theme`, `keybindings`. Persists all user preferences between sessions.
- **Auto-Close Brackets** — Automatically closes `() [] {} "" '' \`\`` pairs. Smart-skip: typing the closing character skips over the auto-inserted one. Toggle with `F4`.
- **Vim Mode Toggle** — `F5` enables Vim-style modal editing: Normal (`hjkl`, `w`, `b`, `0`, `$`, `gg`, `G`, `x`, `dd`, `yy`, `p`, `P`, `u`, `/`, `v`), Insert (`i`/`I`/`a`/`A`/`o`/`O`), Visual (`y`/`d`), Command (`:w`/`:q`/`:wq`/`:q!`). Count prefix support (`3dd`, `2x`). Built-in, no plugins.
- **Theme Switching** — `F6` cycles through 5 themes in real time: dark, light, monokai, dracula, solarized-dark. Current theme shown in status bar and persisted.
- **Word Wrap** — `Alt+Z` toggles line wrapping at the viewport width. Soft-wrap with smart word-boundary breaks.
- **Format on Save** — Automatically formats code on save using external formatters: `gofmt` (Go), `black` (Python), `rustfmt` (Rust). Silently falls back if formatter is not installed.
- **Custom Keybindings** — Remap any key to any action via `config.json` keybindings map. Overrides default key dispatch. Configurable without editing source code.

### Changed

- **Plan v2.0** — Development plan restructured: Phase 9 (built-in features) replaces Lua plugins. New Phase 10 (quality & stability). Phase 11 (production) renumbered from old Phase 10.
- **Total tests:** 73 (was 65)

---

## [0.3.1] — 2026-04-30

### Fixed

- **tea.Quit propagation** — When closing the last tab via confirm dialog (`Ctrl+W` → `S`/`D`), the `tea.Quit` command was lost because `delegateToEditor` returned `nil` instead of the editor's command. Fixed by propagating `cmd` from the editor after tab closure.
- **Ctrl+Tab key mapping** — Changed from `ctrl+tab` to also accept `ctrl+v` (BubbleTea v1.3.10 limitation — both map to the same key code on Windows). Removed non-functional `ctrl+shift+tab` handler (same issue).
- **gofmt compliance** — All 9 source files reformatted with `gofmt -w .` for consistent line endings across platforms. CI Lint job now passes.

### Added

- **6 flow tests** — New `internal/tabs/flow_test.go` covering all tab close scenarios: save+close, discard+close, cancel, new tab, clean tab close, Ctrl+Tab cycling. Total tests: 65 (was 59).

## [0.3.0] — 2026-04-30

### Added — Phase 8: Power Features

- **Tabs** — `Ctrl+T` opens a new tab, `Ctrl+W` closes the current tab (with confirmation if modified), `Ctrl+Tab` cycles tabs, `Ctrl+1-9` jumps to a specific tab. Tab bar shows filename and modified indicator (●). New `internal/tabs/` package with `TabManager` implementing `tea.Model`.
- **Multi-cursor editing** — `Ctrl+D` adds the next occurrence of the word under the cursor as an extra cursor, `Escape` clears all extra cursors. Typing, backspace, delete, enter, and tab work across all cursors simultaneously. Status bar shows cursor count.
- **Indent guides** — Subtle vertical lines (│) at each 4-space indent level, rendered in the left margin. Helps visualize code block hierarchy.
- **Splits** — `Ctrl+\` creates a horizontal split with a new `TabManager` on the right. Click to focus a pane, `Ctrl+\` toggles focus. Active border highlighted in orange. New `SplitContainer` in `internal/tabs/` supports single-pane and split modes.
- **LSP client** — New `internal/lsp/` package with JSON-RPC 2.0 client over stdio. Automatically starts `gopls` for Go files (extensible to Python, Rust, TypeScript). Sends `textDocument/didChange` on every edit. Receives `publishDiagnostics` and shows error (✗) and warning (⚠) counts in the status bar.

### Changed

- **Refactored editor** — `editor.go` split into 4 files: `editor.go` (Model, core), `view.go` (rendering), `keys.go` (key dispatch), `actions.go` (commands). Reduced from 1613 lines to ~650 + 440 + 700 + 260.
- **Architecture** — `main.go` now creates a `SplitContainer` wrapping a `TabManager` wrapping `editor.Model`. Container hierarchy: `SplitContainer` → `TabManager` → `editor.Model`.

## [0.2.0] — 2026-04-26

### Added

- **Rename file** — `F2` renames the current file with inline input bar. Also accessible via command palette (`file.rename`). Auto-saves before rename if file is dirty. Cross-platform validation for invalid characters and empty names.

### Fixed

- **Command palette shortcut** — Changed from `Ctrl+Shift+P` to `Ctrl+P`. The original shortcut was undetectable because the terminal ASCII protocol cannot distinguish `Ctrl+Shift+letter` from `Ctrl+letter` (both send the same control byte). `Ctrl+P` is the VS Code / Sublime Text standard.
- **Save As shortcut** — Changed from `Ctrl+Shift+S` to `F3`. Same limitation as above. Function keys work universally via escape sequences. Also fixed palette handler for `file.save-as` action.
- **Paleta de comandos** — Added missing `file.save-as` handler in `executeAction()` so Save As works when selected via command palette.

## [0.1.0] — 2026-04-26

### Added

- **Modeless TUI editor** — No modes, no learning curve. Type immediately.
- **Gap buffer** — Efficient in-memory text storage with O(1) insertions
- **Cursor navigation** — Arrow keys, Home/End, Ctrl+Arrow for word jumps
- **Mouse support** — Click to position cursor, drag to select, scroll to navigate
- **Undo/Redo** — Unlimited history (Ctrl+Z / Ctrl+Y)
- **Clipboard** — Copy, cut, paste (Ctrl+C / Ctrl+X / Ctrl+V)
- **Syntax highlighting** — 50+ languages via Chroma. Auto-detects language.
- **5 color themes** — Dark, light, monokai, dracula, solarized
- **Command palette** — Ctrl+P for fuzzy search of all commands
- **File operations** — Open (Ctrl+O), Save (Ctrl+S), Save As (F3)
- **File picker** — Navigate directories to open files
- **Welcome screen** — Shows recent files and shortcuts on startup
- **Search** — Ctrl+F to search within the file
- **Replace** — Ctrl+H to search and replace
- **Auto-save** — Automatically saves every 30 seconds
- **Confirmation dialog** — Warns before quitting with unsaved changes
- **Cross-platform** — Windows, Linux, macOS. Single binary, zero dependencies.
- **Go vet** — All code passes `go vet ./...`
- **59 tests** — Unit tests across 5 packages with race detection

### Architecture

- **Language:** Go 1.23+
- **TUI Framework:** Bubble Tea + Lip Gloss (charm.sh)
- **Syntax:** Chroma v2
- **Structure:** Elm-like (Model/Update/View)
- **Buffer:** Gap buffer v1 (Rope planned for v2)

[0.3.1]: https://github.com/alexlivre/cmdit/releases/tag/v0.3.1
[0.3.0]: https://github.com/alexlivre/cmdit/releases/tag/v0.3.0
[0.2.0]: https://github.com/alexlivre/cmdit/releases/tag/v0.2.0
[0.1.0]: https://github.com/alexlivre/cmdit/releases/tag/v0.1.0
