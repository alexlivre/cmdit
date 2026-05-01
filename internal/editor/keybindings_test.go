package editor

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCustomKeybinding(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	dir := t.TempDir()
	m.filename = dir + "/test.txt"
	m.buf.InsertString("custom save test")

	// Set custom binding: Ctrl+J = save
	m.config.Keybindings = map[string]string{
		"ctrl+j": "file.save",
	}

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

func TestCustomKeybindingNoMatch(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.config.Keybindings = map[string]string{
		"ctrl+j": "file.save",
	}
	m.buf.InsertString("test")

	// Ctrl+K should pass through normally
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlK})

	if m.buf.String() != "test" {
		t.Error("buffer should be unchanged when no custom binding matches")
	}
}
