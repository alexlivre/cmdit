// Package tabs implements tab management for the cmdit editor.
package tabs

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexb/cmdit/internal/editor"
)

// Tab wraps an editor model with display metadata.
type Tab struct {
	Editor *editor.Model
	Name   string // display name (filename or "[novo]")
}

// TabManager manages multiple editor tabs and implements tea.Model.
type TabManager struct {
	tabs      []*Tab
	activeIdx int
	width     int
	height    int

	// Tab bar styles
	tabBarStyle      lipgloss.Style
	tabActiveStyle   lipgloss.Style
	tabInactiveStyle lipgloss.Style
	tabModifiedDot   lipgloss.Style
}

// New creates a new TabManager with one empty tab.
func New() *TabManager {
	tm := &TabManager{
		tabs:             make([]*Tab, 0),
		activeIdx:        0,
		tabBarStyle:      lipgloss.NewStyle().Background(lipgloss.Color("237")),
		tabActiveStyle:   lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")).Padding(0, 1),
		tabInactiveStyle: lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("246")).Padding(0, 1),
		tabModifiedDot:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
	}

	// Start with one empty tab
	tm.addTab(editor.New())
	tm.activeIdx = 0

	return tm
}

// NewWithFile creates a TabManager and opens the given file in the first tab.
func NewWithFile(path string) (*TabManager, error) {
	tm := &TabManager{
		tabs:             make([]*Tab, 0),
		activeIdx:        0,
		tabBarStyle:      lipgloss.NewStyle().Background(lipgloss.Color("237")),
		tabActiveStyle:   lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")).Padding(0, 1),
		tabInactiveStyle: lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("246")).Padding(0, 1),
		tabModifiedDot:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
	}

	m, err := editor.NewWithFile(path)
	if err != nil {
		return tm, err
	}

	tm.addTab(m)
	tm.activeIdx = 0
	return tm, nil
}

// addTab appends a new tab to the manager.
func (tm *TabManager) addTab(m *editor.Model) {
	name := filepath.Base(m.Filename())
	if name == "" || name == "." {
		name = "[new]"
	}
	tm.tabs = append(tm.tabs, &Tab{
		Editor: m,
		Name:   name,
	})
}

// activeTab returns the currently active tab.
func (tm *TabManager) activeTab() *Tab {
	if tm.activeIdx < 0 || tm.activeIdx >= len(tm.tabs) {
		return nil
	}
	return tm.tabs[tm.activeIdx]
}

// Init implements tea.Model.
func (tm *TabManager) Init() tea.Cmd {
	tab := tm.activeTab()
	if tab != nil {
		return tab.Editor.Init()
	}
	return nil
}

// Update implements tea.Model.
func (tm *TabManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tm.width = msg.Width
		tm.height = msg.Height

		// Resize all editors (subtract 1 for tab bar)
		editorHeight := msg.Height - 1
		for _, t := range tm.tabs {
			t.Editor.Update(tea.WindowSizeMsg{
				Width:  msg.Width,
				Height: editorHeight,
			})
		}
		return tm, nil

	case tea.KeyMsg:
		return tm.handleKey(msg)
	}

	// Delegate to active editor
	return tm.delegateToEditor(msg)
}

// delegateToEditor passes a message to the active editor and handles close requests.
func (tm *TabManager) delegateToEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	tab := tm.activeTab()
	if tab == nil {
		return tm, nil
	}

	newEditor, cmd := tab.Editor.Update(msg)
	if newEditor != nil {
		if ed, ok := newEditor.(*editor.Model); ok {
			tm.tabs[tm.activeIdx].Editor = ed
			tm.tabs[tm.activeIdx].Name = tm.tabName(ed)

			// Check if the editor requests the tab to be closed (e.g., after confirm dialog)
			if ed.CloseRequested() {
				ed.SaveRecentFiles()
				tm.closeTab(tm.activeIdx)
				// Propagate any command from the editor (e.g., tea.Quit if last tab)
				return tm, cmd
			}
		}
	}

	return tm, cmd
}

// View implements tea.Model.
func (tm *TabManager) View() string {
	tabBar := tm.renderTabBar()

	tab := tm.activeTab()
	var content string
	if tab != nil {
		content = tab.Editor.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)
}

// renderTabBar draws the tab bar at the top.
func (tm *TabManager) renderTabBar() string {
	if len(tm.tabs) == 0 {
		return ""
	}

	var parts []string
	maxWidth := tm.width

	for i, t := range tm.tabs {
		name := t.Name
		// Truncate very long names
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		// Add modified indicator
		modifiedDot := ""
		if t.Editor.Modified() {
			modifiedDot = tm.tabModifiedDot.Render("●") + " "
		}

		display := modifiedDot + name

		// Add close hint
		if i == tm.activeIdx {
			parts = append(parts, tm.tabActiveStyle.Render(display+" ×"))
		} else {
			parts = append(parts, tm.tabInactiveStyle.Render(display))
		}
	}

	bar := strings.Join(parts, " ")
	// Pad to full width
	if len(bar) < maxWidth {
		bar += strings.Repeat(" ", maxWidth-len(bar))
	}

	return tm.tabBarStyle.Render(bar)
}

// handleKey processes tab management keys before delegating to the editor.
func (tm *TabManager) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// If the editor is in a modal mode (confirm, search, palette, etc.),
	// pass ALL keys through to the editor without interception.
	tab := tm.activeTab()
	if tab != nil && tab.Editor.CurrentMode() != editor.ModeNormal &&
		tab.Editor.CurrentMode() != editor.ModeWelcome {
		return tm.delegateToEditor(msg)
	}

	// Tab management keys
	switch {
	case key == "ctrl+t":
		tm.newTab()
		return tm, tm.activeTab().Editor.Init()

	case key == "ctrl+w":
		return tm.closeCurrentTab()

	case key == "ctrl+q":
		// If only tab and modified, let editor handle confirmation
		if len(tm.tabs) == 1 && tab != nil && tab.Editor.Modified() {
			return tm.delegateToEditor(msg)
		}
		// Save all recent files and quit
		for _, t := range tm.tabs {
			t.Editor.SaveRecentFiles()
		}
		return tm, tea.Quit

	case key == "f7":
		tm.prevTab()
		return tm, nil

	case key == "f8":
		tm.nextTab()
		return tm, nil

	// Ctrl+1 through Ctrl+9: go to specific tab
	case len(key) == 6 && strings.HasPrefix(key, "ctrl+") && key[5] >= '1' && key[5] <= '9':
		idx := int(key[5] - '1')
		if idx < len(tm.tabs) {
			tm.activeIdx = idx
		}
		return tm, nil
	}

	// Delegate remaining keys to active editor
	return tm.delegateToEditor(msg)
}

// newTab creates a new empty tab and switches to it.
func (tm *TabManager) newTab() {
	m := editor.New()
	tm.addTab(m)
	tm.activeIdx = len(tm.tabs) - 1
}

// closeCurrentTab closes the active tab, with confirmation if modified.
func (tm *TabManager) closeCurrentTab() (tea.Model, tea.Cmd) {
	if len(tm.tabs) == 0 {
		return tm, tea.Quit
	}

	tab := tm.tabs[tm.activeIdx]

	// If modified, put editor into confirm-close-tab mode
	if tab.Editor.Modified() {
		tab.Editor.ConfirmCloseTabMode()
		return tm, nil
	}

	tm.tabs[tm.activeIdx].Editor.SaveRecentFiles()
	return tm.closeTabHelper(tm.activeIdx)
}

// closeTabHelper closes the tab at the given index.
func (tm *TabManager) closeTabHelper(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(tm.tabs) {
		return tm, nil
	}

	// Save recent files for the closing tab
	tm.tabs[idx].Editor.SaveRecentFiles()

	// Remove the tab
	tm.tabs = append(tm.tabs[:idx], tm.tabs[idx+1:]...)

	// Adjust active index
	if len(tm.tabs) == 0 {
		// No more tabs → signal quit
		return tm, tea.Quit
	}
	if tm.activeIdx >= len(tm.tabs) {
		tm.activeIdx = len(tm.tabs) - 1
	}
	return tm, nil
}

func (tm *TabManager) closeTab(idx int) {
	tm.closeTabHelper(idx)
}

// nextTab switches to the next tab (wraps around).
func (tm *TabManager) nextTab() {
	if len(tm.tabs) == 0 {
		return
	}
	tm.activeIdx = (tm.activeIdx + 1) % len(tm.tabs)
}

// prevTab switches to the previous tab (wraps around).
func (tm *TabManager) prevTab() {
	if len(tm.tabs) == 0 {
		return
	}
	tm.activeIdx--
	if tm.activeIdx < 0 {
		tm.activeIdx = len(tm.tabs) - 1
	}
}

// tabName returns the display name for a tab.
func (tm *TabManager) tabName(m *editor.Model) string {
	name := filepath.Base(m.Filename())
	if name == "" || name == "." {
		return "[novo]"
	}
	return name
}

// ActiveEditor returns the currently active editor (for external access).
func (tm *TabManager) ActiveEditor() *editor.Model {
	tab := tm.activeTab()
	if tab == nil {
		return nil
	}
	return tab.Editor
}
