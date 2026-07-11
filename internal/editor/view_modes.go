package editor

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// --- Confirm dialog ---

func (m *Model) renderConfirm() string {
	var msg string
	if m.confirmAction == ConfirmQuit {
		msg = "File modified! Save before quitting?"
	} else {
		msg = "File modified! Save before closing?"
	}
	btns := "\n\n" + m.confirmBtnStyle.Render("[S] Save") + "  [D] Discard  [C] Cancel"
	dialog := m.confirmStyle.Render(msg + btns)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// --- Save As ---

func (m *Model) renderSaveAs() string {
	var lines []string
	lines = append(lines, "Save as:")
	lines = append(lines, "")
	lines = append(lines, "> "+m.saveAsQuery)
	lines = append(lines, "")
	lines = append(lines, "Enter to confirm, Esc to cancel")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}
