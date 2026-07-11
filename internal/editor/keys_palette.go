package editor

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexb/cmdit/internal/command"
)

// --- Palette key handler ---

func (m *Model) enterPalette() {
	m.mode = ModePalette
	m.palette.Query = ""
	m.paletteResult()
	m.palette.Sel = 0
}

func (m *Model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil

	case "enter":
		if m.palette.Sel >= 0 && m.palette.Sel < len(m.palette.Results) {
			action := m.palette.Results[m.palette.Sel]
			m.mode = ModeNormal
			m.executeAction(action.ID)
		}
		return m, nil

	case "up":
		m.palette.Sel--
		if m.palette.Sel < 0 {
			m.palette.Sel = len(m.palette.Results) - 1
		}

	case "down":
		m.palette.Sel++
		if m.palette.Sel >= len(m.palette.Results) {
			m.palette.Sel = 0
		}

	case "backspace":
		if len(m.palette.Query) > 0 {
			m.palette.Query = m.palette.Query[:len(m.palette.Query)-1]
			m.paletteResult()
			m.palette.Sel = 0
		}

	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.palette.Query += string(msg.Runes)
			m.paletteResult()
			m.palette.Sel = 0
		}
	}
	return m, nil
}

func (m *Model) paletteResult() {
	if m.palette.Query == "" {
		m.palette.Results = m.palette.Actions
		return
	}
	q := strings.ToLower(m.palette.Query)
	var filtered []command.Action
	for _, a := range m.palette.Actions {
		if strings.Contains(strings.ToLower(a.Label), q) || strings.Contains(strings.ToLower(a.ID), q) {
			filtered = append(filtered, a)
		}
	}
	m.palette.Results = filtered
}
