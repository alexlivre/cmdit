package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestThemeSwitchF6(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	initialTheme := m.config.Theme

	m.handleKey(tea.KeyMsg{Type: tea.KeyF6})

	if m.config.Theme == initialTheme {
		t.Error("theme should have changed after F6")
	}
	if m.highlighter.Theme() != m.config.Theme {
		t.Errorf("highlighter theme '%s' != config theme '%s'", m.highlighter.Theme(), m.config.Theme)
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

	if len(visited) != 5 {
		t.Errorf("expected 5 themes, visited %d: %v", len(visited), visited)
	}
}
