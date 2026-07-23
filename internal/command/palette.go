package command

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaletteModel is the Bubble Tea model for the command palette.
type PaletteModel struct {
	registry *Registry
	query    string
	results  []Action
	selected int
	width    int
	handler  Handler

	style         lipgloss.Style
	inputStyle    lipgloss.Style
	itemStyle     lipgloss.Style
	activeStyle   lipgloss.Style
	shortcutStyle lipgloss.Style
}

// NewPalette creates a new command palette.
func NewPalette(registry *Registry, handler Handler) *PaletteModel {
	p := &PaletteModel{
		registry:      registry,
		handler:       handler,
		style:         lipgloss.NewStyle().Background(lipgloss.Color("236")).Padding(1, 2),
		inputStyle:    lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")).Padding(0, 1),
		itemStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		activeStyle:   lipgloss.NewStyle().Background(lipgloss.Color("214")).Foreground(lipgloss.Color("0")).Padding(0, 1),
		shortcutStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
	}
	p.filter()
	return p
}

// Init implements tea.Model.
func (p *PaletteModel) Init() tea.Cmd {
	return nil
}

// Update handles palette input.
func (p *PaletteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return p, tea.Quit // Exit palette

		case "enter":
			if p.selected >= 0 && p.selected < len(p.results) {
				return p, p.handler()
			}

		case "up":
			p.selected--
			if p.selected < 0 {
				p.selected = len(p.results) - 1
			}

		case "down":
			p.selected++
			if p.selected >= len(p.results) {
				p.selected = 0
			}

		case "backspace":
			if len(p.query) > 0 {
				p.query = p.query[:len(p.query)-1]
				p.filter()
				p.selected = 0
			}

		default:
			if len(msg.Runes) > 0 {
				p.query += string(msg.Runes)
				p.filter()
				p.selected = 0
			}
		}
	}
	return p, nil
}

// View renders the palette.
func (p *PaletteModel) View() string {
	maxItems := 10
	start := p.selected - 5
	if start < 0 {
		start = 0
	}
	if start > len(p.results)-maxItems {
		start = len(p.results) - maxItems
	}
	if start < 0 {
		start = 0
	}

	end := start + maxItems
	if end > len(p.results) {
		end = len(p.results)
	}

	var items []string
	items = append(items, p.inputStyle.Render("> "+p.query))

	if len(p.query) == 0 {
		items = append(items, p.shortcutStyle.Render("  Type to search commands..."))
	}

	for i := start; i < end; i++ {
		item := p.results[i]
		line := fmt.Sprintf("  %-30s", item.Label)
		if item.Shortcut != "" {
			line += " " + p.shortcutStyle.Render(item.Shortcut)
		}
		if i == p.selected {
			line = p.activeStyle.Render(strings.TrimRight(line, " "))
		} else {
			line = p.itemStyle.Render(line)
		}
		items = append(items, line)
	}

	if len(p.results) == 0 && len(p.query) > 0 {
		items = append(items, p.shortcutStyle.Render("  No commands found"))
	}

	content := strings.Join(items, "\n")
	return p.style.Render(content)
}

func (p *PaletteModel) filter() {
	p.results = p.registry.Filter(p.query)
}

// Execute runs the palette as a Bubble Tea program (modal overlay).
func Execute(registry *Registry, actionHandler Handler) tea.Cmd {
	return func() tea.Msg {
		p := NewPalette(registry, actionHandler)
		program := tea.NewProgram(p)
		_, err := program.Run()
		if err != nil {
			// Return the error somehow
		}
		return nil
	}
}
