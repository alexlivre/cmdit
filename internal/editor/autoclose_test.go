package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAutoCloseParens(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})

	text := m.buf.String()
	if text != "()" {
		t.Errorf("expected '()', got '%s'", text)
	}
	pos := m.buf.GapPosition()
	if pos != 1 {
		t.Errorf("expected cursor at 1, got %d", pos)
	}
}

func TestAutoCloseBrackets(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	if m.buf.String() != "[]" {
		t.Errorf("expected '[]', got '%s'", m.buf.String())
	}
}

func TestAutoCloseBraces(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'{'}})
	if m.buf.String() != "{}" {
		t.Errorf("expected '{}', got '%s'", m.buf.String())
	}
}

func TestAutoCloseSmartSkip(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{')'}})

	if m.buf.String() != "()" {
		t.Errorf("expected '()', got '%s'", m.buf.String())
	}
	if m.buf.GapPosition() != 2 {
		t.Errorf("expected cursor at 2, got %d", m.buf.GapPosition())
	}
}

func TestAutoCloseToggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.config.AutoCloseEnabled = false
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})

	if m.buf.String() != "(" {
		t.Errorf("expected '(', got '%s'", m.buf.String())
	}
}

func TestAutoCloseF4Toggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	if !m.config.AutoCloseEnabled {
		t.Error("should be enabled by default")
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyF4})
	if m.config.AutoCloseEnabled {
		t.Error("should be disabled after F4")
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyF4})
	if !m.config.AutoCloseEnabled {
		t.Error("should be re-enabled after second F4")
	}
}

func TestAutoCloseQuotes(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'"'}})
	if m.buf.String() != "\"\"" {
		t.Errorf("expected '\"\"', got '%s'", m.buf.String())
	}
}

func TestAutoCloseWithText(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})
	for _, ch := range "hello" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	if m.buf.String() != "(hello)" {
		t.Errorf("expected '(hello)', got '%s'", m.buf.String())
	}
}
