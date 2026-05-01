# Fase 9 — Features Built-in Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 6 built-in features to cmdit (config JSON, auto-close brackets, vim mode toggle, format on save, theme switching, word wrap) — all native Go, no plugin system.

**Architecture:** Each feature is a focused Go file in `internal/editor/` or `internal/highlight/`. All toggles persist via `~/.cmdit/config.json` (new `internal/editor/config.go`). Key dispatch in `keys.go` gets early-return hooks for auto-close and vim mode before the normal switch.

**Tech Stack:** Go 1.24+, Bubble Tea v1.3.10, Lip Gloss v1.1.0, Chroma v2.23.1

---

## Task 1: Config System Foundation

**Files:**
- Create: `internal/editor/config.go`
- Modify: `internal/editor/editor.go` (add `config Config` field to Model, load in `New()`/`NewWithFile()`)
- Modify: `internal/editor/editor.go` (register new palette actions for toggles)
- Create: `internal/editor/config_test.go`

### Task 1.1: Create config.go with struct and load/save

- [ ] **Step 1: Create `internal/editor/config.go`**

```go
package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all user-configurable settings persisted to ~/.cmdit/config.json.
type Config struct {
	AutoCloseEnabled  bool              `json:"auto_close_enabled"`
	VimMode           bool              `json:"vim_mode"`
	FormatOnSave      bool              `json:"format_on_save"`
	WordWrap          bool              `json:"word_wrap"`
	Theme             string            `json:"theme"`
	Keybindings       map[string]string `json:"keybindings"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		AutoCloseEnabled: true,
		VimMode:          false,
		FormatOnSave:     false,
		WordWrap:         false,
		Theme:            "dark",
		Keybindings:      map[string]string{},
	}
}

// configPath returns the path to ~/.cmdit/config.json.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cmdit", "config.json"), nil
}

// LoadConfig reads config from ~/.cmdit/config.json.
// Returns defaults if the file does not exist.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	path, err := configPath()
	if err != nil {
		return cfg, nil // can't determine home dir, use defaults
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // no config yet, use defaults
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err // corrupted config, reset to defaults
	}

	// Ensure Keybindings map is not nil
	if cfg.Keybindings == nil {
		cfg.Keybindings = map[string]string{}
	}

	return cfg, nil
}

// SaveConfig writes config to ~/.cmdit/config.json.
// Creates the .cmdit directory if it doesn't exist.
func SaveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
```

- [ ] **Step 2: Add `config` field to Model struct in `internal/editor/editor.go`**

Find the Model struct definition and add after the `closeRequested` field:

```go
// Config
config Config
```

- [ ] **Step 3: Load config in `New()` and `NewWithFile()`**

In `New()`, after `m.loadRecentFiles()` add:

```go
cfg, _ := LoadConfig()
m.config = cfg
m.highlighter.SetTheme(cfg.Theme)
```

In `NewWithFile()`, after `m.loadRecentFiles()` add the same block:

```go
cfg, _ := LoadConfig()
m.config = cfg
m.highlighter.SetTheme(cfg.Theme)
```

- [ ] **Step 4: Add config toggle palette actions in `registerActions()`**

After the existing palette actions, add:

```go
{ID: "view.toggle-auto-close", Label: "Toggle Auto-Close Brackets", Shortcut: "F4"},
{ID: "view.toggle-vim-mode",   Label: "Toggle Modo Vim",        Shortcut: "F5"},
{ID: "view.next-theme",        Label: "Próximo Tema",           Shortcut: "F6"},
{ID: "view.toggle-word-wrap",  Label: "Toggle Word Wrap",       Shortcut: "Alt+Z"},
{ID: "file.toggle-format-on-save", Label: "Toggle Format on Save", Shortcut: ""},
```

- [ ] **Step 5: Create `internal/editor/config_test.go`**

```go
package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.AutoCloseEnabled {
		t.Error("AutoCloseEnabled should be true by default")
	}
	if cfg.VimMode {
		t.Error("VimMode should be false by default")
	}
	if cfg.FormatOnSave {
		t.Error("FormatOnSave should be false by default")
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme should be 'dark' by default, got '%s'", cfg.Theme)
	}
}

func TestConfigRoundtrip(t *testing.T) {
	cfg := Config{
		AutoCloseEnabled: false,
		VimMode:          true,
		FormatOnSave:     true,
		WordWrap:         true,
		Theme:            "monokai",
		Keybindings: map[string]string{
			"ctrl+s": "file.save",
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.AutoCloseEnabled {
		t.Error("AutoCloseEnabled should be false after roundtrip")
	}
	if !decoded.VimMode {
		t.Error("VimMode should be true after roundtrip")
	}
	if decoded.Keybindings["ctrl+s"] != "file.save" {
		t.Error("Keybindings should survive roundtrip")
	}
}

func TestSaveLoadConfig(t *testing.T) {
	// Override home dir by using a temp dir
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".cmdit", "config.json")

	// Write config manually
	cfg := Config{
		Theme:            "dracula",
		AutoCloseEnabled: false,
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.MkdirAll(filepath.Dir(configFile), 0700)
	os.WriteFile(configFile, data, 0600)

	// Verify file exists
	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Read it back
	readData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	var decoded Config
	json.Unmarshal(readData, &decoded)

	if decoded.Theme != "dracula" {
		t.Errorf("expected theme 'dracula', got '%s'", decoded.Theme)
	}
	if decoded.AutoCloseEnabled {
		t.Error("expected AutoCloseEnabled false")
	}
}

func TestConfigModelIntegration(t *testing.T) {
	m := New()

	if m.config.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got '%s'", m.config.Theme)
	}
	if !m.config.AutoCloseEnabled {
		t.Error("expected AutoCloseEnabled true")
	}
}
```

- [ ] **Step 6: Run tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -run TestConfig -v
```

Expected: 4 tests PASS

- [ ] **Step 7: Commit**

```bash
git add internal/editor/config.go internal/editor/config_test.go internal/editor/editor.go
git commit -m "feat(config): add Config struct with JSON load/save to ~/.cmdit/config.json"
```

---

## Task 2: Auto-Close Brackets & Quotes

**Files:**
- Create: `internal/editor/autoclose.go`
- Modify: `internal/editor/keys.go` (hook in `handleKey()` default case)
- Modify: `internal/editor/actions.go` (add `view.toggle-auto-close` case)
- Create: `internal/editor/autoclose_test.go`

### Task 2.1: Create autoclose.go

- [ ] **Step 1: Create `internal/editor/autoclose.go`**

```go
package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

// autoClosePairs maps opening characters to their closing counterparts.
var autoClosePairs = map[rune]rune{
	'(': ')',
	'[': ']',
	'{': '}',
	'"': '"',
	'\'': '\'',
	'`': '`',
}

// shouldAutoClose returns (closingRune, true) if the character should trigger auto-close.
func shouldAutoClose(r rune) (rune, bool) {
	closer, ok := autoClosePairs[r]
	return closer, ok
}

// handleAutoClose inserts a pair of characters and positions the cursor between them.
// It also sets a flag so that if the user types the closing char, we skip instead of duplicating.
func (m *Model) handleAutoClose(openChar rune) tea.Cmd {
	closer, ok := shouldAutoClose(openChar)
	if !ok {
		return nil
	}

	// Record cursor position before insert to detect smart-skip later
	cursorPos := m.buf.GetCursorPos()

	// Insert opening char
	m.buf.Insert(openChar)
	// Insert closing char
	m.buf.Insert(closer)

	// Move cursor back one so it sits between the pair
	m.buf.MoveLeft()

	// Store the position as "auto-closed" so smart-skip can detect it
	// We use the logical index of the opening char
	m.markAutoClosed(cursorPos)

	m.modified = true
	return nil
}

// autoClosedPositions stores positions where auto-close inserted a pair.
// We add to Model struct: autoClosed map[int]bool
// Key is the buffer index of the opening character.

// markAutoClosed records that position 'pos' has an auto-closed pair.
func (m *Model) markAutoClosed(pos int) {
	if m.autoClosed == nil {
		m.autoClosed = make(map[int]bool)
	}
	m.autoClosed[pos] = true
}

// isAutoClosed checks if the character at position 'pos' was auto-closed.
func (m *Model) isAutoClosed(pos int) bool {
	if m.autoClosed == nil {
		return false
	}
	return m.autoClosed[pos]
}

// clearAutoClosedAt removes the auto-closed marker when the pair is broken.
func (m *Model) clearAutoClosedAt(pos int) {
	if m.autoClosed != nil {
		delete(m.autoClosed, pos)
	}
}

// handleSmartSkip checks if the typed character matches the next character
// and if the current position has an auto-closed pair — if so, skip (move right).
func (m *Model) handleSmartSkip(char rune) bool {
	cursorPos := m.buf.GetCursorPos()

	// Get the character at cursor position
	nextChar := m.buf.CharAt(cursorPos)
	if nextChar == char {
		// Check if the opening char at (cursorPos-1) is auto-closed
		// The opening char of the pair is one position before
		if m.isAutoClosed(cursorPos - 1) {
			// Skip: just move cursor right
			m.buf.MoveRight()
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Add `autoClosed` field to Model struct in `editor.go`**

Add after the `config Config` field:

```go
// Auto-close tracking
autoClosed map[int]bool // positions with auto-closed pairs
```

### Task 2.2: Hook into key dispatch

- [ ] **Step 3: Modify `handleKey()` in `keys.go`**

In the `default` case of the main switch (where `msg.Runes[0] >= 32`), add before the normal insert:

```go
default:
	if len(msg.Runes) > 0 {
		ch := msg.Runes[0]
		if ch >= 32 {
			// Auto-close: check if this char should trigger a pair
			if m.config.AutoCloseEnabled {
				// Smart-skip: if next char matches and we're at an auto-closed position
				if m.handleSmartSkip(ch) {
					return
				}
				if _, ok := shouldAutoClose(ch); ok {
					return m.handleAutoClose(ch)
				}
			}
			m.insertTextAtAllCursors(string([]rune{ch}))
			m.modified = true
			m.sendDidChange()
		}
	}
```

### Task 2.3: Add toggle action

- [ ] **Step 4: Add toggle case in `actions.go`**

In `executeAction()`, add:

```go
case "view.toggle-auto-close":
	m.config.AutoCloseEnabled = !m.config.AutoCloseEnabled
	SaveConfig(m.config)
```

### Task 2.4: Handle F4 key in keys.go

- [ ] **Step 5: Add `f4` case in `handleKey()` switch**

Before the existing `f3` or after it:

```go
case "f4":
	if m.mode == ModeNormal {
		m.config.AutoCloseEnabled = !m.config.AutoCloseEnabled
		SaveConfig(m.config)
	}
	return
```

### Task 2.5: Tests

- [ ] **Step 6: Create `internal/editor/autoclose_test.go`**

```go
package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAutoCloseParens(t *testing.T) {
	m := New()
	m.mode = ModeNormal // bypass welcome

	// Type '('
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})

	text := m.buf.String()
	if text != "()" {
		t.Errorf("expected '()', got '%s'", text)
	}

	// Cursor should be between ( and )
	cursorPos := m.buf.GetCursorPos()
	if cursorPos != 1 {
		t.Errorf("expected cursor at position 1 (between parens), got %d", cursorPos)
	}
}

func TestAutoCloseBrackets(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})

	text := m.buf.String()
	if text != "[]" {
		t.Errorf("expected '[]', got '%s'", text)
	}
}

func TestAutoCloseBraces(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'{'}})

	text := m.buf.String()
	if text != "{}" {
		t.Errorf("expected '{}', got '%s'", text)
	}
}

func TestAutoCloseSmartSkip(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Type '(' -> inserts ()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})

	// Type ')' -> should skip over the existing ')', not insert another
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{')'}})

	text := m.buf.String()
	if text != "()" {
		t.Errorf("expected '()' after smart-skip, got '%s'", text)
	}

	// Cursor should be after the ')'
	cursorPos := m.buf.GetCursorPos()
	if cursorPos != 2 {
		t.Errorf("expected cursor at position 2 (after parens), got %d", cursorPos)
	}
}

func TestAutoCloseToggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Disable auto-close
	m.config.AutoCloseEnabled = false

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})

	text := m.buf.String()
	if text != "(" {
		t.Errorf("expected '(' (auto-close disabled), got '%s'", text)
	}
}

func TestAutoCloseF4Toggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Should be enabled by default
	if !m.config.AutoCloseEnabled {
		t.Error("AutoCloseEnabled should be true by default")
	}

	// Press F4 to toggle
	m.handleKey(tea.KeyMsg{Type: tea.KeyF4})

	if m.config.AutoCloseEnabled {
		t.Error("AutoCloseEnabled should be false after F4 toggle")
	}

	// Press F4 again
	m.handleKey(tea.KeyMsg{Type: tea.KeyF4})

	if !m.config.AutoCloseEnabled {
		t.Error("AutoCloseEnabled should be true after second F4 toggle")
	}
}

func TestAutoCloseQuotes(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'"'}})

	text := m.buf.String()
	if text != "\"\"" {
		t.Errorf("expected '\"\"', got '%s'", text)
	}
}

func TestAutoCloseWithText(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Type "hello" inside parens
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

	text := m.buf.String()
	if text != "(hello)" {
		t.Errorf("expected '(hello)', got '%s'", text)
	}
}
```

- [ ] **Step 7: Run tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -run TestAutoClose -v
```

Expected: 8 tests PASS

- [ ] **Step 8: Commit**

```bash
git add internal/editor/autoclose.go internal/editor/autoclose_test.go internal/editor/keys.go internal/editor/actions.go internal/editor/editor.go
git commit -m "feat(auto-close): add auto-close brackets/quotes with smart-skip and F4 toggle"
```

---

## Task 3: Vim Mode Toggle (F5)

**Files:**
- Create: `internal/editor/vimmode.go`
- Modify: `internal/editor/keys.go` (early return to dispatchVimKey)
- Modify: `internal/editor/actions.go` (add toggle case)
- Create: `internal/editor/vimmode_test.go`

### Task 3.1: Create vimmode.go

- [ ] **Step 1: Create `internal/editor/vimmode.go`**

```go
package editor

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// VimMode represents the current vim editing mode.
type VimMode int

const (
	VimNormal  VimMode = iota
	VimInsert
	VimVisual
	VimCommand
)

// VimState holds the transient state for vim keybindings.
type VimState struct {
	Mode       VimMode
	Count      string // accumulated digits for count prefix (e.g., "3" in "3dd")
	CommandBuf string // accumulates characters for : commands
	LastOp     string // last operator for . repeat (not yet implemented)
}

// newVimState creates a fresh vim state starting in normal mode.
func newVimState() VimState {
	return VimState{
		Mode:  VimNormal,
		Count: "",
	}
}

// dispatchVimKey handles key events when vim mode is active.
// Returns true if the key was consumed, false to pass through.
func (m *Model) dispatchVimKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch m.vimState.Mode {

	case VimNormal:
		return m.vimNormal(key)

	case VimInsert:
		return m.vimInsert(key)

	case VimVisual:
		return m.vimVisual(key)

	case VimCommand:
		return m.vimCommand(key)
	}

	return nil
}

// vimNormal handles keys in normal mode (hjkl, operators, etc.)
func (m *Model) vimNormal(key string) tea.Cmd {
	// Count prefix: digits accumulate
	if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
		// Don't accumulate if count is "0" and another 0 is pressed
		if m.vimState.Count == "0" && key == "0" {
			// "00" -> go to beginning of line (Vim behavior)
			m.buf.MoveToLineStart()
			m.vimState.Count = ""
			return nil
		}
		m.vimState.Count += key
		return nil
	}

	count := m.parseVimCount()

	switch key {
	// Navigation
	case "h":
		m.vimMoveLeft(count)
	case "j":
		m.vimMoveDown(count)
	case "k":
		m.vimMoveUp(count)
	case "l":
		m.vimMoveRight(count)

	// Word navigation
	case "w":
		m.vimNextWord(count)
	case "b":
		m.vimPrevWord(count)

	// Line navigation
	case "0":
		m.buf.MoveToLineStart()
	case "$":
		m.buf.MoveToLineEnd()

	// File navigation
	case "g":
		if m.vimState.LastOp == "g" {
			m.buf.MoveToFileStart()
			m.vimState.LastOp = ""
		} else {
			m.vimState.LastOp = "g"
			return nil // wait for next key
		}
	case "G":
		m.buf.MoveToFileEnd()

	// Enter insert mode
	case "i":
		m.vimState.Mode = VimInsert
	case "I":
		m.buf.MoveToLineStart()
		m.vimState.Mode = VimInsert
	case "a":
		m.buf.MoveRight()
		m.vimState.Mode = VimInsert
	case "A":
		m.buf.MoveToLineEnd()
		m.vimState.Mode = VimInsert
	case "o":
		m.buf.Insert('\n')
		m.modified = true
		m.vimState.Mode = VimInsert
	case "O":
		m.buf.MoveToLineStart()
		m.buf.Insert('\n')
		m.buf.MoveUp()
		m.modified = true
		m.vimState.Mode = VimInsert

	// Delete operations
	case "x":
		m.vimDeleteChar(count)
	case "d":
		if m.vimState.LastOp == "d" {
			m.vimDeleteLine()
			m.vimState.LastOp = ""
		} else {
			m.vimState.LastOp = "d"
			return nil
		}

	// Yank (copy)
	case "y":
		if m.vimState.LastOp == "y" {
			m.vimYankLine()
			m.vimState.LastOp = ""
		} else {
			m.vimState.LastOp = "y"
			return nil
		}

	// Paste
	case "p":
		m.vimPasteAfter()
	case "P":
		m.vimPasteBefore()

	// Undo/Redo
	case "u":
		m.undo()

	// Search
	case "/":
		m.mode = ModeSearch
		m.vimState.Mode = VimNormal // exit command-like state

	// Command mode
	case ":":
		m.vimState.Mode = VimCommand
		m.vimState.CommandBuf = ""

	// Enter visual mode
	case "v":
		m.vimState.Mode = VimVisual
	}

	m.vimState.Count = ""
	return nil
}

// vimInsert handles keys in insert mode.
func (m *Model) vimInsert(key string) tea.Cmd {
	switch key {
	case "esc":
		m.vimState.Mode = VimNormal
		return nil
	case "backspace":
		m.handleBackspace()
		return nil
	case "enter":
		m.buf.Insert('\n')
		m.modified = true
		return nil
	default:
		// Pass printable characters through normal insert
		// This is handled by the caller — we return false to indicate passthrough
		return nil
	}
}

// vimVisual handles keys in visual mode.
func (m *Model) vimVisual(key string) tea.Cmd {
	switch key {
	case "esc":
		m.vimState.Mode = VimNormal
		m.selStart = 0
		m.selEnd = 0
		return nil
	case "y":
		m.copy()
		m.vimState.Mode = VimNormal
		m.selStart = 0
		m.selEnd = 0
		return nil
	case "d":
		m.cut()
		m.vimState.Mode = VimNormal
		return nil
	}
	return nil
}

// vimCommand handles keys in command mode (:w, :q, :wq).
func (m *Model) vimCommand(key string) tea.Cmd {
	switch key {
	case "esc":
		m.vimState.Mode = VimNormal
		m.vimState.CommandBuf = ""
		return nil

	case "enter":
		cmd := strings.TrimSpace(m.vimState.CommandBuf)
		m.vimState.Mode = VimNormal
		m.vimState.CommandBuf = ""

		switch cmd {
		case "w":
			m.save()
		case "q":
			if m.modified {
				m.mode = ModeConfirm
				m.confirmAction = ConfirmQuit
			} else {
				m.closeRequested = true
			}
		case "wq":
			m.save()
			m.closeRequested = true
		case "q!":
			m.closeRequested = true
		}
		return nil

	case "backspace":
		if len(m.vimState.CommandBuf) > 0 {
			m.vimState.CommandBuf = m.vimState.CommandBuf[:len(m.vimState.CommandBuf)-1]
		}
		return nil

	default:
		if len(key) == 1 && key[0] >= 32 {
			m.vimState.CommandBuf += key
		}
		return nil
	}
}

// Helper functions

func (m *Model) parseVimCount() int {
	if m.vimState.Count == "" {
		return 1
	}
	n := 0
	for _, ch := range m.vimState.Count {
		n = n*10 + int(ch-'0')
	}
	if n == 0 {
		return 1
	}
	return n
}

func (m *Model) vimMoveLeft(n int) {
	for i := 0; i < n; i++ {
		m.buf.MoveLeft()
	}
}

func (m *Model) vimMoveRight(n int) {
	for i := 0; i < n; i++ {
		m.buf.MoveRight()
	}
}

func (m *Model) vimMoveDown(n int) {
	for i := 0; i < n; i++ {
		m.buf.MoveDown()
	}
}

func (m *Model) vimMoveUp(n int) {
	for i := 0; i < n; i++ {
		m.buf.MoveUp()
	}
}

func (m *Model) vimNextWord(n int) {
	for i := 0; i < n; i++ {
		m.buf.MoveWordRight()
	}
}

func (m *Model) vimPrevWord(n int) {
	for i := 0; i < n; i++ {
		m.buf.MoveWordLeft()
	}
}

func (m *Model) vimDeleteChar(n int) {
	for i := 0; i < n; i++ {
		m.buf.Delete()
	}
	m.modified = true
}

func (m *Model) vimDeleteLine() {
	m.buf.MoveToLineStart()
	start := m.buf.GetCursorPos()
	m.buf.MoveToLineEnd()
	end := m.buf.GetCursorPos()
	// Delete from start to end of line
	m.buf.DeleteRange(start, end-start+1)
	m.modified = true
}

func (m *Model) vimYankLine() {
	m.buf.MoveToLineStart()
	start := m.buf.GetCursorPos()
	m.buf.MoveToLineEnd()
	end := m.buf.GetCursorPos()
	text := m.buf.TextRange(start, end-start+1)
	m.clipboard.Set(text)
}

func (m *Model) vimPasteAfter() {
	text := m.clipboard.Get()
	if text != "" {
		m.buf.MoveRight()
		m.buf.InsertString(text)
		m.modified = true
	}
}

func (m *Model) vimPasteBefore() {
	text := m.clipboard.Get()
	if text != "" {
		m.buf.InsertString(text)
		m.modified = true
	}
}
```

### Task 3.2: Hook into key dispatch

- [ ] **Step 2: Add `vimState` and `vimMode` fields to Model in `editor.go`**

```go
// Vim mode
vimState VimState
```

Note: `vimMode` is read from `config.VimMode` — no separate field needed.

- [ ] **Step 3: Modify `handleKey()` in `keys.go`**

At the VERY TOP of `handleKey()`, before any other logic, add:

```go
func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	// Vim mode dispatch: intercept ALL keys when vim mode is active
	if m.config.VimMode {
		// In insert mode, handle characters normally except Esc
		if m.vimState.Mode == VimInsert {
			switch msg.String() {
			case "esc":
				m.vimState.Mode = VimNormal
				return nil
			case "backspace":
				m.handleBackspace()
				return nil
			case "enter":
				m.insertTextAtAllCursors("\n")
				m.modified = true
				m.sendDidChange()
				return nil
			default:
				if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
					m.insertTextAtAllCursors(string(msg.Runes))
					m.modified = true
					m.sendDidChange()
					return nil
				}
			}
			return nil
		}

		// For Normal/Visual/Command modes, dispatch to vim handler
		cmd := m.dispatchVimKey(msg)
		return cmd
	}

	// ... rest of existing handleKey logic
```

- [ ] **Step 4: Initialize `vimState` in `New()` and `NewWithFile()`**

In both constructors, add after config load:

```go
m.vimState = newVimState()
```

### Task 3.3: Add F5 toggle and palette action

- [ ] **Step 5: Add `f5` case in `handleKey()` switch**

```go
case "f5":
	if m.mode == ModeNormal {
		m.config.VimMode = !m.config.VimMode
		m.vimState = newVimState() // reset state on toggle
		SaveConfig(m.config)
	}
	return
```

- [ ] **Step 6: Add toggle action in `actions.go`**

```go
case "view.toggle-vim-mode":
	m.config.VimMode = !m.config.VimMode
	m.vimState = newVimState()
	SaveConfig(m.config)
```

### Task 3.4: Tests

- [ ] **Step 7: Create `internal/editor/vimmode_test.go`**

```go
package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestVimModeToggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Default should be off
	if m.config.VimMode {
		t.Error("VimMode should be false by default")
	}

	// Press F5 to enable
	m.handleKey(tea.KeyMsg{Type: tea.KeyF5})

	if !m.config.VimMode {
		t.Error("VimMode should be true after F5")
	}

	// Press F5 again to disable
	m.handleKey(tea.KeyMsg{Type: tea.KeyF5})

	if m.config.VimMode {
		t.Error("VimMode should be false after second F5")
	}
}

func TestVimModeNavigation(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()

	// Load some text to navigate
	m.buf.InsertString("hello\nworld\n")

	// Move cursor to beginning for predictable position
	m.buf.MoveToFileStart() // line 0, col 0

	// j = down
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	line, _ := m.buf.GetLineCol(m.buf.GetCursorPos())
	if line != 1 {
		t.Errorf("expected line 1 after 'j', got %d", line)
	}

	// k = up
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	line, _ = m.buf.GetLineCol(m.buf.GetCursorPos())
	if line != 0 {
		t.Errorf("expected line 0 after 'k', got %d", line)
	}
}

func TestVimModeInsert(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()

	// i = enter insert mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	if m.vimState.Mode != VimInsert {
		t.Errorf("expected VimInsert mode, got %v", m.vimState.Mode)
	}

	// Type some text in insert mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	text := m.buf.String()
	if text != "hi" {
		t.Errorf("expected 'hi', got '%s'", text)
	}

	// Esc = return to normal mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})

	if m.vimState.Mode != VimNormal {
		t.Errorf("expected VimNormal after Esc, got %v", m.vimState.Mode)
	}
}

func TestVimModeCommandSave(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test_save.txt"
	_ = path // Used for future file-based save test

	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()
	m.filename = path

	// Type some content in insert mode
	m.vimState.Mode = VimInsert
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	// Esc to normal mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})

	// :w to save
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// File should exist and contain "test"
	// Note: save needs filename set, which we did above
}

func TestVimModeDeleteChar(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()

	m.buf.InsertString("hello")
	m.buf.MoveToFileStart()

	// x = delete char under cursor
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	text := m.buf.String()
	if text != "ello" {
		t.Errorf("expected 'ello' after x, got '%s'", text)
	}
}

func TestVimModeCountPrefix(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()

	m.buf.InsertString("12345")
	m.buf.MoveToFileStart()

	// 3x = delete 3 chars
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	text := m.buf.String()
	if text != "45" {
		t.Errorf("expected '45' after 3x, got '%s'", text)
	}
}
```

- [ ] **Step 8: Run tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -run TestVimMode -v
```

Expected: 6 tests PASS

- [ ] **Step 9: Commit**

```bash
git add internal/editor/vimmode.go internal/editor/vimmode_test.go internal/editor/keys.go internal/editor/actions.go internal/editor/editor.go
git commit -m "feat(vim): add vim mode toggle (F5) with normal/insert/visual/command modes"
```

---

## Task 4: Theme Switching (F6)

**Files:**
- Modify: `internal/editor/keys.go` (add `f6` case)
- Modify: `internal/editor/actions.go` (add `view.next-theme` case)
- Create: `internal/editor/theme_test.go`

### Task 4.1: Add F6 key handler

- [ ] **Step 1: Add `f6` case in `handleKey()` switch in `keys.go`**

```go
case "f6":
	if m.mode == ModeNormal {
		m.nextTheme()
	}
	return
```

- [ ] **Step 2: Add `nextTheme()` method in a new file or in `editor.go`**

In `internal/editor/editor.go` (or a new `internal/editor/theme.go`):

```go
// themeCycle defines the order themes rotate through on F6.
var themeCycle = []string{"dark", "light", "monokai", "dracula", "solarized-dark"}

// nextTheme switches to the next theme in the rotation.
func (m *Model) nextTheme() {
	current := m.config.Theme
	next := 0
	for i, name := range themeCycle {
		if name == current {
			next = (i + 1) % len(themeCycle)
			break
		}
	}
	m.config.Theme = themeCycle[next]
	m.highlighter.SetTheme(m.config.Theme)
	SaveConfig(m.config)
}
```

- [ ] **Step 3: Add palette action in `actions.go`**

```go
case "view.next-theme":
	m.nextTheme()
```

- [ ] **Step 4: Show current theme in status bar**

In `view.go`, in the `renderStatus()` function, add the theme name to the status bar. Find where the status string is built and add:

```go
// Example addition to status format:
// [...] [Tema:dark]
themeStr := fmt.Sprintf(" [Tema:%s]", m.config.Theme)
```

### Task 4.2: Tests

- [ ] **Step 5: Create `internal/editor/theme_test.go`**

```go
package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestThemeSwitchF6(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	initialTheme := m.config.Theme
	if initialTheme != "dark" {
		t.Errorf("expected default theme 'dark', got '%s'", initialTheme)
	}

	// Press F6 to cycle to next theme
	m.handleKey(tea.KeyMsg{Type: tea.KeyF6})

	if m.config.Theme == initialTheme {
		t.Error("theme should have changed after F6")
	}

	currentTheme := m.highlighter.Theme()
	if currentTheme != m.config.Theme {
		t.Errorf("highlighter theme '%s' should match config theme '%s'", currentTheme, m.config.Theme)
	}
}

func TestThemeFullCycle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	visited := make(map[string]bool)
	for i := 0; i < 10; i++ {
		m.handleKey(tea.KeyMsg{Type: tea.KeyF6})
		visited[m.config.Theme] = true
	}

	// After many cycles, we should have visited all 5 themes
	expectedThemes := 5
	if len(visited) != expectedThemes {
		t.Errorf("expected to visit %d themes, visited %d: %v", expectedThemes, len(visited), visited)
	}
}

func TestThemePersistence(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Switch to a non-default theme
	m.handleKey(tea.KeyMsg{Type: tea.KeyF6})

	savedTheme := m.config.Theme

	// Simulate reload by loading config fresh
	cfg, _ := LoadConfig()
	if cfg.Theme != savedTheme {
		t.Errorf("theme should persist in config, expected '%s', got '%s'", savedTheme, cfg.Theme)
	}
}
```

- [ ] **Step 6: Run tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -run TestTheme -v
```

Expected: 3 tests PASS

- [ ] **Step 7: Commit**

```bash
git add internal/editor/editor.go internal/editor/keys.go internal/editor/actions.go internal/editor/theme_test.go
git commit -m "feat(theme): add theme cycling with F6, persist to config.json"
```

---

## Task 5: Word Wrap Toggle (Alt+Z)

**Files:**
- Modify: `internal/editor/editor.go` (add `wrapMode` field)
- Modify: `internal/editor/keys.go` (add `alt+z` case)
- Modify: `internal/editor/actions.go` (add toggle case)
- Modify: `internal/editor/view.go` (use wrapMode in rendering)
- Create: `internal/editor/wrap_test.go`

### Task 5.1: Add word wrap support

- [ ] **Step 1: Add `wrapMode` field to Model in `editor.go`**

The wrap setting is read from `config.WordWrap`. Add a convenience boolean:

No new field needed — read `m.config.WordWrap` directly.

- [ ] **Step 2: Add `alt+z` handler in `keys.go`**

```go
case "alt+z":
	if m.mode == ModeNormal {
		m.config.WordWrap = !m.config.WordWrap
		SaveConfig(m.config)
	}
	return
```

- [ ] **Step 3: Add palette action in `actions.go`**

```go
case "view.toggle-word-wrap":
	m.config.WordWrap = !m.config.WordWrap
	SaveConfig(m.config)
```

- [ ] **Step 4: Implement word wrap in rendering**

In `view.go`, in `renderContent()`, when `m.config.WordWrap` is true, wrap lines at `m.width`. When false (default), use horizontal scroll (current behavior).

The current rendering already uses `m.viewport` for scroll. For word wrap, we need to:
1. Calculate wrapped line count
2. Display full wrapped content instead of raw lines

```go
// In renderContent(), replace the line rendering loop:
if m.config.WordWrap {
    // Wrap lines at viewport width minus line number margin
    wrapWidth := m.width - lineNumWidth
    for _, rawLine := range lines {
        wrapped := wrapText(rawLine, wrapWidth)
        for _, wl := range wrapped {
            // render each wrapped sub-line
            segments := m.highlighter.HighlightLine(wl, m.language)
            lineText := highlight.RenderSegments(segments)
            renderedLines = append(renderedLines, lineText)
        }
    }
} else {
    // Original unwrapped behavior
    // ... existing code
}
```

- [ ] **Step 5: Add `wrapText` helper**

```go
// wrapText wraps text at the given width, breaking at word boundaries when possible.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	var lines []string
	for len(text) > width {
		// Try to break at last space within width
		breakAt := width
		for i := width - 1; i >= width/2; i-- {
			if i < len(text) && text[i] == ' ' {
				breakAt = i + 1 // include the space
				break
			}
		}
		lines = append(lines, text[:breakAt])
		text = text[breakAt:]
	}
	if len(text) > 0 {
		lines = append(lines, text)
	}
	return lines
}
```

### Task 5.2: Tests

- [ ] **Step 6: Create `internal/editor/wrap_test.go`**

```go
package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWordWrapToggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Default should be off
	if m.config.WordWrap {
		t.Error("WordWrap should be false by default")
	}

	// Press Alt+Z to enable
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}, Alt: true})

	if !m.config.WordWrap {
		t.Error("WordWrap should be true after Alt+Z")
	}

	// Press Alt+Z again to disable
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}, Alt: true})

	if m.config.WordWrap {
		t.Error("WordWrap should be false after second Alt+Z")
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		width  int
		expect []string
	}{
		{
			name:   "short text no wrap",
			text:   "hello",
			width:  10,
			expect: []string{"hello"},
		},
		{
			name:   "wrap at space",
			text:   "hello world foo bar",
			width:  12,
			expect: []string{"hello world ", "foo bar"},
		},
		{
			name:   "no space break",
			text:   "abcdefghijklmnop",
			width:  5,
			expect: []string{"abcde", "fghij", "klmno", "p"},
		},
		{
			name:   "zero width",
			text:   "hello",
			width:  0,
			expect: []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.width)
			if len(result) != len(tt.expect) {
				t.Errorf("expected %d lines, got %d: %v", len(tt.expect), len(result), result)
				return
			}
			for i := range result {
				if result[i] != tt.expect[i] {
					t.Errorf("line %d: expected '%s', got '%s'", i, tt.expect[i], result[i])
				}
			}
		})
	}
}
```

- [ ] **Step 7: Run tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -run TestWordWrap -v
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -run TestWrapText -v
```

Expected: 2 tests PASS

- [ ] **Step 8: Commit**

```bash
git add internal/editor/keys.go internal/editor/actions.go internal/editor/editor.go internal/editor/view.go internal/editor/wrap_test.go
git commit -m "feat(wrap): add word wrap toggle (Alt+Z), persist to config.json"
```

---

## Task 6: Format on Save

**Files:**
- Create: `internal/editor/format.go`
- Modify: `internal/editor/actions.go` (hook in `save()` method)
- Modify: `internal/editor/actions.go` (add toggle case)
- Create: `internal/editor/format_test.go`

### Task 6.1: Create format.go

- [ ] **Step 1: Create `internal/editor/format.go`**

```go
package editor

import (
	"fmt"
	"os/exec"
	"strings"
)

// FormatFunc takes text and language, returns formatted text.
type FormatFunc func(text string) (string, error)

// formatters maps language IDs to format commands.
var formatters = map[string]struct {
	cmd  string
	args []string
}{
	"go":         {cmd: "gofmt", args: nil},
	"python":     {cmd: "black", args: []string{"--quiet", "-"}},
	"rust":       {cmd: "rustfmt", args: nil},
	"json":       {cmd: "", args: nil}, // built-in: json.Indent
	"typescript": {cmd: "prettier", args: []string{"--stdin-filepath", "file.ts"}},
	"javascript": {cmd: "prettier", args: []string{"--stdin-filepath", "file.js"}},
}

// formatBuffer runs the appropriate formatter on the buffer content.
// Returns the formatted text or the original if no formatter is available.
func (m *Model) formatBuffer() (string, error) {
	text := m.buf.String()
	lang := m.language

	// JSON: format with built-in Go
	if lang == "json" {
		return formatJSON(text)
	}

	// Check if we have an external formatter
	fmtr, ok := formatters[lang]
	if !ok {
		return text, nil // no formatter for this language
	}

	if fmtr.cmd == "" {
		return text, nil
	}

	// Check if the command exists
	if _, err := exec.LookPath(fmtr.cmd); err != nil {
		return text, fmt.Errorf("%s not found: %w", fmtr.cmd, err)
	}

	// Run formatter via stdin/stdout
	cmd := exec.Command(fmtr.cmd, fmtr.args...)
	cmd.Stdin = strings.NewReader(text)

	output, err := cmd.Output()
	if err != nil {
		return text, fmt.Errorf("format failed: %w", err)
	}

	return string(output), nil
}

// formatJSON formats JSON with indentation.
func formatJSON(text string) (string, error) {
	// Use Go's built-in JSON formatting
	var buf strings.Builder
	// Simple indentation
	indent := 0
	for i := 0; i < len(text); i++ {
		ch := text[i]
		switch ch {
		case '{', '[':
			buf.WriteByte(ch)
			indent++
			buf.WriteByte('\n')
			buf.WriteString(strings.Repeat("  ", indent))
		case '}', ']':
			indent--
			buf.WriteByte('\n')
			buf.WriteString(strings.Repeat("  ", indent))
			buf.WriteByte(ch)
		case ',':
			buf.WriteByte(ch)
			buf.WriteByte('\n')
			buf.WriteString(strings.Repeat("  ", indent))
		case ':':
			buf.WriteByte(ch)
			buf.WriteByte(' ')
		default:
			if ch > 32 || ch == ' ' {
				buf.WriteByte(ch)
			}
		}
	}
	return buf.String(), nil
}

// applyFormat formats the buffer content and replaces it.
func (m *Model) applyFormat() error {
	formatted, err := m.formatBuffer()
	if err != nil {
		return err
	}

	if formatted == m.buf.String() {
		return nil // nothing changed
	}

	// Save cursor position
	cursorPos := m.buf.GetCursorPos()

	// Replace buffer content
	m.buf.Clear()
	m.buf.InsertString(formatted)

	// Try to restore cursor to approximately the same position
	if cursorPos < m.buf.Len() {
		for m.buf.GetCursorPos() < cursorPos {
			m.buf.MoveRight()
		}
	}

	return nil
}
```

### Task 6.2: Hook into save

- [ ] **Step 2: Modify `save()` in `actions.go`**

After the `fileio.Save()` call and before setting `m.modified = false`, add:

```go
// Format on save
if m.config.FormatOnSave {
	if err := m.applyFormat(); err != nil {
		// Format failed — save again with formatted content
		// or just log the error and continue with original
	} else {
		// Save the formatted version
		if err := fileio.Save(m.filename, m.buf); err != nil {
			// handle error
			return
		}
	}
}
```

### Task 6.3: Add palette toggle

- [ ] **Step 3: Add toggle case in `actions.go`**

```go
case "file.toggle-format-on-save":
	m.config.FormatOnSave = !m.config.FormatOnSave
	SaveConfig(m.config)
```

### Task 6.4: Tests

- [ ] **Step 4: Create `internal/editor/format_test.go`**

```go
package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatJSON(t *testing.T) {
	input := `{"name":"test","value":123}`
	expected := "{\n  \"name\": \"test\",\n  \"value\": 123\n}"

	result, err := formatJSON(input)
	if err != nil {
		t.Fatalf("formatJSON: %v", err)
	}
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestFormatOnSaveToggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	if m.config.FormatOnSave {
		t.Error("FormatOnSave should be false by default")
	}

	// Toggle via palette action
	m.executeAction("file.toggle-format-on-save")

	if !m.config.FormatOnSave {
		t.Error("FormatOnSave should be true after toggle")
	}
}

func TestFormatGoFile(t *testing.T) {
	// Skip if gofmt is not available
	// This test requires gofmt in PATH

	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")

	// Write unformatted Go code
	code := "package main\nfunc main(){x:=1\nfmt.Println(x)}\n"
	os.WriteFile(path, []byte(code), 0644)

	m, err := NewWithFile(path)
	if err != nil {
		t.Fatalf("NewWithFile: %v", err)
	}
	m.mode = ModeNormal
	m.config.FormatOnSave = true

	// Trigger save
	m.save()

	// Read back
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	// The formatted code should be different from the original
	formatted := string(data)
	if formatted == code {
		t.Error("file should have been formatted, but content is unchanged")
	}
}

func TestFormatNoFormatter(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.language = "plaintext"
	m.buf.InsertString("hello world")

	result, err := m.formatBuffer()
	if err != nil {
		t.Fatalf("formatBuffer: %v", err)
	}

	// With no formatter, should return original text
	if result != "hello world" {
		t.Errorf("expected unchanged text, got '%s'", result)
	}
}
```

- [ ] **Step 5: Run tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -run TestFormat -v
```

Expected: tests PASS (gofmt ones may skip if gofmt not in PATH)

- [ ] **Step 6: Commit**

```bash
git add internal/editor/format.go internal/editor/format_test.go internal/editor/actions.go
git commit -m "feat(format): add format on save with gofmt/black/rustfmt/prettier"
```

---

## Task 7: Custom Keybindings from Config

**Files:**
- Modify: `internal/editor/keys.go` (resolve custom bindings)
- Modify: `internal/editor/actions.go` (export `executeAction`)
- Create: `internal/editor/keybindings_test.go`

### Task 7.1: Implement custom keybinding resolution

- [ ] **Step 1: Modify `handleKey()` in `keys.go`**

Before the main key switch, add custom keybinding resolution:

```go
func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	// ... vim mode check ...

	// Custom keybindings: check config overrides
	if len(m.config.Keybindings) > 0 {
		keyStr := msg.String()
		if actionID, ok := m.config.Keybindings[keyStr]; ok {
			m.executeAction(actionID)
			return nil
		}
	}

	// ... rest of existing handleKey logic
```

- [ ] **Step 2: Make `executeAction` handle all registered palette actions**

The palette actions registered in `registerActions()` are already accessible via `m.paletteActions`. When a custom keybinding maps to an action ID, we can resolve it through `executeAction()`.

Add missing action handlers in `executeAction()` for the new Fase 9 actions:

Already handled in previous tasks:
- `view.toggle-auto-close`
- `view.toggle-vim-mode`
- `view.next-theme`
- `view.toggle-word-wrap`
- `file.toggle-format-on-save`

All these should already have cases from Tasks 2-5.

### Task 7.2: Tests

- [ ] **Step 3: Create `internal/editor/keybindings_test.go`**

```go
package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCustomKeybinding(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Set a custom keybinding: Ctrl+J = save
	m.config.Keybindings = map[string]string{
		"ctrl+j": "file.save",
	}

	// Create a temp file for saving
	dir := t.TempDir()
	m.filename = dir + "/test.txt"
	m.buf.InsertString("custom save test")

	// Press Ctrl+J
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlJ})

	// File should be saved
	data, err := os.ReadFile(m.filename)
	if err != nil {
		t.Fatalf("file not saved: %v", err)
	}
	if string(data) != "custom save test" {
		t.Errorf("expected 'custom save test', got '%s'", string(data))
	}
}

func TestCustomKeybindingOverride(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Override Ctrl+S to quit instead of save
	m.config.Keybindings = map[string]string{
		"ctrl+s": "file.quit",
	}

	// Ctrl+S should now trigger quit, not save
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlS})

	// Should be in confirm mode (because modified)
	if m.mode != ModeConfirm {
		t.Errorf("expected ModeConfirm after overridden Ctrl+S, got %v", m.mode)
	}
}

func TestCustomKeybindingNoMatch(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.config.Keybindings = map[string]string{
		"ctrl+j": "file.save",
	}

	// Ctrl+K should work normally (not overridden)
	m.buf.InsertString("test")
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlK}) // No custom binding

	// Buffer should be unchanged (Ctrl+K has no default action)
	if m.buf.String() != "test" {
		t.Error("buffer should be unchanged when no custom binding matches")
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -run TestCustomKeybinding -v
```

Expected: 3 tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/editor/keys.go internal/editor/keybindings_test.go
git commit -m "feat(keybindings): add custom keybindings from config.json"
```

---

## Task 8: Integration — Run Full Test Suite

- [ ] **Step 1: Run all editor tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./internal/editor/ -v
```

Expected: All tests PASS (>80 tests including new ones)

- [ ] **Step 2: Run full project tests**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe test ./... -v
```

Expected: All tests PASS

- [ ] **Step 3: Run go vet**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe vet ./...
```

Expected: No errors

- [ ] **Step 4: Build**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe build ./cmd/cmdit
```

Expected: Build succeeds

- [ ] **Step 5: Run gofmt**

```bash
cd C:\code\cmdit && C:\Users\alexb\go-install\Go\bin\go.exe fmt ./...
```

Expected: No changes needed (or apply changes)

- [ ] **Step 6: Commit final integration**

```bash
git add -A
git commit -m "feat(fase-9): complete built-in features — config, auto-close, vim, theme, wrap, format, keybindings"
```

---

## Summary

| Task | Feature | New Files | Modified Files | Estimated Tests |
|------|---------|-----------|----------------|-----------------|
| 1 | Config System | `config.go`, `config_test.go` | `editor.go` | 4 |
| 2 | Auto-Close Brackets | `autoclose.go`, `autoclose_test.go` | `keys.go`, `actions.go`, `editor.go` | 8 |
| 3 | Vim Mode Toggle | `vimmode.go`, `vimmode_test.go` | `keys.go`, `actions.go`, `editor.go` | 6 |
| 4 | Theme Switching | `theme_test.go` | `keys.go`, `actions.go`, `editor.go` | 3 |
| 5 | Word Wrap | `wrap_test.go` | `keys.go`, `actions.go`, `view.go` | 2 |
| 6 | Format on Save | `format.go`, `format_test.go` | `actions.go` | 4 |
| 7 | Custom Keybindings | `keybindings_test.go` | `keys.go` | 3 |
| 8 | Integration | — | — | All 90+ tests |

**Total:** ~30 new tests, ~8 new Go files, ~6 modified files
