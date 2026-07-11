package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// --- File Picker ---

func (m *Model) renderFilePicker() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Open: %s > %s", m.filePicker.Dir, m.filePicker.Query))
	lines = append(lines, "")

	start := m.filePicker.Sel - 10
	if start < 0 {
		start = 0
	}
	end := start + 20
	if end > len(m.filePicker.Files) {
		end = len(m.filePicker.Files)
	}

	for i := start; i < end; i++ {
		line := "  " + m.filePicker.Files[i]
		if i == m.filePicker.Sel {
			line = m.paletteActive.Render(line)
		}
		lines = append(lines, line)
	}

	if len(m.filePicker.Files) == 0 {
		lines = append(lines, "  (diretorio vazio)")
	}

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}
