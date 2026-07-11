package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

// --- Search input handling ---

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.search.Matches = nil
		return m, nil

	case "enter":
		m.doSearch()
		if len(m.search.Matches) > 0 {
			m.navigateToMatch(0)
		}
		if m.mode == ModeReplace {
			m.doReplace()
		}
		return m, nil

	case "tab":
		if m.mode == ModeReplace {
			// For now, tab stays in replace mode
		}
		return m, nil

	case "backspace":
		if m.mode == ModeReplace && m.search.Replace != "" {
			m.search.Replace = m.search.Replace[:len(m.search.Replace)-1]
		} else if m.search.Query != "" {
			m.search.Query = m.search.Query[:len(m.search.Query)-1]
		}
		return m, nil

	default:
		if len(msg.Runes) > 0 {
			if m.mode == ModeReplace && msg.Runes[0] >= 32 {
				if len(m.search.Matches) > 0 {
					m.search.Replace += string(msg.Runes)
				} else {
					m.search.Query += string(msg.Runes)
				}
			} else if msg.Runes[0] >= 32 {
				m.search.Query += string(msg.Runes)
			}
		}
		return m, nil
	}
}
