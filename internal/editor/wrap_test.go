package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWordWrapToggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	if m.config.WordWrap {
		t.Error("WordWrap should be false by default")
	}

	// Press Alt+Z
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}, Alt: true})

	if !m.config.WordWrap {
		t.Error("WordWrap should be true after Alt+Z")
	}

	// Press again
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
		{"short", "hello", 10, []string{"hello"}},
		{"wrap at space", "hello world foo bar", 12, []string{"hello world ", "foo bar"}},
		{"no space break", "abcdefghijklmnop", 5, []string{"abcde", "fghij", "klmno", "p"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.width)
			if len(result) != len(tt.expect) {
				t.Fatalf("%d lines, expected %d: %v", len(result), len(tt.expect), result)
			}
			for i := range result {
				if result[i] != tt.expect[i] {
					t.Errorf("line %d: '%s' != '%s'", i, result[i], tt.expect[i])
				}
			}
		})
	}
}
