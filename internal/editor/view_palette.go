package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// --- Palette ---

func (m *Model) renderPalette() string {
	maxItems := 10
	start := m.palette.Sel - 5
	if start < 0 {
		start = 0
	}
	end := start + maxItems
	if end > len(m.palette.Results) {
		end = len(m.palette.Results)
	}

	var items []string
	items = append(items, m.paletteInput.Render("> "+m.palette.Query))

	if len(m.palette.Query) == 0 {
		items = append(items, m.paletteShortcut.Render("  Type to search commands..."))
	}
	if len(m.palette.Results) == 0 && len(m.palette.Query) > 0 {
		items = append(items, m.paletteShortcut.Render("  No commands found"))
	}

	for i := start; i < end; i++ {
		a := m.palette.Results[i]
		line := fmt.Sprintf("  %-30s", a.Label)
		if a.Shortcut != "" {
			line += " " + m.paletteShortcut.Render(a.Shortcut)
		}
		if i == m.palette.Sel {
			items = append(items, m.paletteActive.Render(strings.TrimRight(line, " ")))
		} else {
			items = append(items, line)
		}
	}

	content := strings.Join(items, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}
