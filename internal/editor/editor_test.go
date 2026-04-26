package editor

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewEditor(t *testing.T) {
	m := New()
	if m.buf == nil {
		t.Error("expected non-nil buffer")
	}
	if m.cursor == nil {
		t.Error("expected non-nil cursor")
	}
	if m.undoStack == nil {
		t.Error("expected non-nil undo stack")
	}
	if m.clipboard == nil {
		t.Error("expected non-nil clipboard")
	}
	if m.mode != ModeWelcome {
		t.Errorf("expected ModeWelcome, got %v", m.mode)
	}
}

func TestInsertText(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	if m.buf.String() != "hi" {
		t.Errorf("expected 'hi', got %q", m.buf.String())
	}
	if !m.modified {
		t.Error("expected modified to be true after insert")
	}
}

func TestBackspace(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})

	if m.buf.String() != "a" {
		t.Errorf("expected 'a', got %q", m.buf.String())
	}
}

func TestBackspaceAtStart(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.buf.String() != "" {
		t.Errorf("expected empty, got %q", m.buf.String())
	}
}

func TestEnterKey(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if m.buf.String() != "a\nb" {
		t.Errorf("expected 'a\\nb', got %q", m.buf.String())
	}
	if m.cursor.Line != 1 {
		t.Errorf("expected cursor line 1, got %d", m.cursor.Line)
	}
}

func TestUndo(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlZ})
	if m.buf.String() != "h" {
		t.Errorf("expected 'h' after undo, got %q", m.buf.String())
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlZ})
	if m.buf.String() != "" {
		t.Errorf("expected empty after 2 undos, got %q", m.buf.String())
	}
}

func TestRedo(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlZ})

	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlY})
	if m.buf.String() != "a" {
		t.Errorf("expected 'a' after redo, got %q", m.buf.String())
	}
}

func TestCopyPaste(t *testing.T) {
	m := New()
	// Type "hello"
	for _, r := range "hello" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Select all
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	// Copy
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Paste
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlV})

	if strings.Count(m.buf.String(), "hello") != 2 {
		t.Errorf("expected 'hellohello', got %q", m.buf.String())
	}
}

func TestCopyLine(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	// Copy current line (no selection copies line)
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.clipboard.HasText() {
		t.Error("expected clipboard to have text")
	}
}

func TestQuitUnsavedShowsConfirm(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlQ})

	if m.mode != ModeConfirm {
		t.Errorf("expected ModeConfirm, got %v", m.mode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestQuitSavedNoConfirm(t *testing.T) {
	m := New()
	// Switch to normal mode first
	m.mode = ModeNormal
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlQ})

	if m.mode != ModeNormal {
		t.Error("expected no confirm for unmodified buffer")
	}
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

func TestConfirmDiscard(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlQ})

	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Error("expected quit cmd after discard")
	}
}

func TestConfirmCancel(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlQ})

	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if m.mode != ModeNormal {
		t.Errorf("expected ModeNormal after cancel, got %v", m.mode)
	}
}

func TestSearchMode(t *testing.T) {
	m := New()
	for _, r := range "hello world" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Enter search mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlF})
	if m.mode != ModeSearch {
		t.Errorf("expected ModeSearch, got %v", m.mode)
	}
}

func TestSearchExecute(t *testing.T) {
	m := New()
	for _, r := range "hello world hello" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Enter search mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlF})
	// Type search query (in search mode)
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	// Press Enter to search
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if len(m.searchMatches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(m.searchMatches))
	}
}
