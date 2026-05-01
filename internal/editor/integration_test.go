package editor

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestVimModeQuitUnsaved tests :q with unsaved changes (should show confirm dialog)
func TestVimModeQuitUnsaved(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()
	m.filename = filepath.Join(t.TempDir(), "test.txt")

	// Type some text via vim insert mode
	m.vimState.Mode = VimInsert
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape}) // back to normal

	// :q should show confirm dialog since modified
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if m.mode != ModeConfirm {
		t.Errorf("expected ModeConfirm for unsaved :q, got %v", m.mode)
	}
}

// TestVimModeQuitSaved tests :q with no changes
func TestVimModeQuitSaved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	m, err := NewWithFile(path)
	if err != nil {
		t.Fatalf("NewWithFile: %v", err)
	}
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()

	// :q clean file should close
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.closeRequested {
		t.Error("expected closeRequested after :q on clean file")
	}
}

// TestAutoCloseDisabledInVimMode tests that auto-close still works in vim insert mode
func TestAutoCloseInVimInsertMode(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.config.AutoCloseEnabled = true
	m.vimState = newVimState()

	// Enter insert mode
	// In vim insert mode, chars go through the normal flow
	m.vimState.Mode = VimInsert

	// Type '(' — should auto-close
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})

	if m.buf.String() != "()" {
		t.Errorf("expected '()' in vim insert mode, got '%s'", m.buf.String())
	}
}

// TestF5VimToggleFromVimMode tests F5 pressing while in vim mode
func TestF5VimToggleFromVimMode(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()

	// F5 while in vim mode should toggle it off
	m.handleKey(tea.KeyMsg{Type: tea.KeyF5})

	if m.config.VimMode {
		t.Error("F5 should toggle VimMode off")
	}
}

// TestThemeChangeDoesntCrash tests cycling all themes
func TestThemeChangeDoesntCrash(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	themes := []string{"dark", "light", "monokai", "dracula", "solarized-dark"}
	for _, expected := range themes {
		m.executeAction("view.next-theme")
		if m.config.Theme != expected && m.config.Theme != themes[0] {
			// After all 5, it wraps back to dark
			continue
		}
		// Just verify it doesn't panic
		_ = m.highlighter.Theme()
	}
}

// TestConfigPersistence tests that SaveConfig writes and LoadConfig reads correctly
func TestConfigPersistence(t *testing.T) {
	dir := t.TempDir()
	// Override home dir via environment
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOME", dir)

	origDir, _ := os.UserHomeDir()
	_ = origDir

	cfg := DefaultConfig()
	cfg.Theme = "dracula"
	cfg.VimMode = true
	cfg.AutoCloseEnabled = false
	cfg.FormatOnSave = true
	cfg.WordWrap = true
	cfg.Keybindings = map[string]string{"ctrl+j": "file.save"}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if loaded.Theme != "dracula" {
		t.Errorf("theme: expected 'dracula', got '%s'", loaded.Theme)
	}
	if !loaded.VimMode {
		t.Error("VimMode should be true")
	}
	if loaded.AutoCloseEnabled {
		t.Error("AutoCloseEnabled should be false")
	}
	if !loaded.FormatOnSave {
		t.Error("FormatOnSave should be true")
	}
	if !loaded.WordWrap {
		t.Error("WordWrap should be true")
	}
	if loaded.Keybindings["ctrl+j"] != "file.save" {
		t.Error("Keybindings not preserved")
	}
}

// TestFormatDisabledStillSaves tests that save still works with format disabled
func TestFormatDisabledStillSaves(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	m, err := NewWithFile(path)
	if err != nil {
		t.Fatalf("NewWithFile: %v", err)
	}
	m.mode = ModeNormal
	m.config.FormatOnSave = false
	m.buf.InsertString("hello world")
	m.save()

	// File should exist with content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not saved: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(data))
	}
}
