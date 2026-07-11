package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// --- Welcome Screen ---

func (m *Model) renderWelcome() string {
	logo := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render("  cmdit v0.4.2")
	subtitle := "Text editor for humans"

	var lines []string
	lines = append(lines, logo)
	lines = append(lines, "")
	lines = append(lines, subtitle)
	lines = append(lines, "")

	if len(m.recentFiles) > 0 {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Recent files:"))
		for i, f := range m.recentFiles {
			if i >= 5 {
				break
			}
			lines = append(lines, fmt.Sprintf("  %d. %s", i+1, f))
		}
		lines = append(lines, "")
	}

	lines = append(lines, "Ctrl+O  Open file")
	lines = append(lines, "Ctrl+P  Command palette")
	lines = append(lines, "Ctrl+Q  Quit")
	lines = append(lines, "")
	lines = append(lines, "Start typing to create a new file...")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
