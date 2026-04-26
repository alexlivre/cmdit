# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[0.1.0]: https://github.com/alexlivre/cmdit/releases/tag/v0.1.0
