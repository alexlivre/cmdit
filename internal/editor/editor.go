// Package editor implements the main Bubble Tea model for the cmdit editor.
package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexb/cmdit/internal/buffer"
	"github.com/alexb/cmdit/internal/clipboard"
	"github.com/alexb/cmdit/internal/command"
	"github.com/alexb/cmdit/internal/fileio"
	"github.com/alexb/cmdit/internal/highlight"
	"github.com/alexb/cmdit/internal/renderer"
)

// Mode represents the current editor mode.
type Mode int

const (
	ModeNormal   Mode = iota
	ModeConfirm
	ModeSearch
	ModeReplace
	ModePalette
	ModeFilePicker
	ModeSaveAs
	ModeWelcome
)

type ConfirmAction int

const (
	ConfirmQuit ConfirmAction = iota
)

// Model is the main Bubble Tea model for the editor.
type Model struct {
	buf        *buffer.Buffer
	cursor     *buffer.Cursor
	viewport   *renderer.Viewport
	undoStack  *buffer.UndoStack
	clipboard  *clipboard.Clipboard
	filename   string
	modified   bool
	mode       Mode
	language   string

	highlighter *highlight.Highlighter

	confirmAction ConfirmAction

	// Selection (logical indices)
	selStart int
	selEnd   int

	// Search state
	searchQuery   string
	searchMatches []int
	searchCurrent int
	replaceQuery  string
	lastSearch    string

	// Palette state
	paletteActions []command.Action
	paletteQuery   string
	paletteResults []command.Action
	paletteSel     int

	// File picker state
	filePickerDir    string
	filePickerFiles  []string
	filePickerSel    int
	filePickerQuery  string
	saveAsQuery      string

	// Recent files
	recentFiles []string

	width  int
	height int

	// Styles
	statusStyle     lipgloss.Style
	statusModified  lipgloss.Style
	statusNormal    lipgloss.Style
	confirmStyle    lipgloss.Style
	confirmBtnStyle lipgloss.Style
	lineNumStyle    lipgloss.Style
	searchStyle     lipgloss.Style
	matchStyle      lipgloss.Style
	currentMatch    lipgloss.Style
	selectionStyle  lipgloss.Style
	paletteStyle    lipgloss.Style
	paletteInput    lipgloss.Style
	paletteActive   lipgloss.Style
	paletteShortcut lipgloss.Style
}

// New creates a new editor model.
func New() *Model {
	m := &Model{
		buf:             buffer.NewBuffer(),
		cursor:          buffer.NewCursor(),
		viewport:        renderer.NewViewport(80, 24),
		undoStack:       buffer.NewUndoStack(),
		clipboard:       clipboard.New(),
		filename:        "",
		modified:        false,
		mode:            ModeWelcome,
		language:        "",
		highlighter:     highlight.NewHighlighter(highlight.ThemeDark),
		selStart:        -1,
		selEnd:          -1,
		statusStyle:     lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252")).Padding(0, 1),
		statusModified:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		statusNormal:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		confirmStyle:    lipgloss.NewStyle().Background(lipgloss.Color("240")).Padding(1, 2),
		confirmBtnStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		lineNumStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		searchStyle:     lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("252")).Padding(0, 1),
		matchStyle:      lipgloss.NewStyle().Background(lipgloss.Color("58")),
		currentMatch:    lipgloss.NewStyle().Background(lipgloss.Color("214")).Foreground(lipgloss.Color("0")),
		selectionStyle:  lipgloss.NewStyle().Background(lipgloss.Color("240")),
		paletteStyle:    lipgloss.NewStyle().Background(lipgloss.Color("236")).Padding(1, 2),
		paletteInput:    lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")).Padding(0, 1),
		paletteActive:   lipgloss.NewStyle().Background(lipgloss.Color("214")).Foreground(lipgloss.Color("0")).Padding(0, 1),
		paletteShortcut: lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
	}
	m.registerActions()
	m.loadRecentFiles()
	return m
}

// NewWithFile creates a new editor model and loads a file if path is provided.
func NewWithFile(path string) (*Model, error) {
	m := New()
	if path == "" {
		return m, nil
	}

	b, err := fileio.Load(path)
	if err != nil {
		m.filename = path
		return m, nil
	}

	m.buf = b
	m.filename = path
	m.language = highlight.DetectLanguage(path)
	m.mode = ModeNormal
	m.loadRecentFiles()
	return m, nil
}

// registerActions populates the command palette actions.
func (m *Model) registerActions() {
	m.paletteActions = []command.Action{
		{ID: "file.save", Label: "Salvar", Shortcut: "Ctrl+S"},
		{ID: "file.save-as", Label: "Salvar como", Shortcut: "Ctrl+Shift+S"},
		{ID: "file.open", Label: "Abrir arquivo", Shortcut: "Ctrl+O"},
		{ID: "file.quit", Label: "Sair", Shortcut: "Ctrl+Q"},
		{ID: "edit.undo", Label: "Desfazer", Shortcut: "Ctrl+Z"},
		{ID: "edit.redo", Label: "Refazer", Shortcut: "Ctrl+Y"},
		{ID: "edit.cut", Label: "Recortar", Shortcut: "Ctrl+X"},
		{ID: "edit.copy", Label: "Copiar", Shortcut: "Ctrl+C"},
		{ID: "edit.paste", Label: "Colar", Shortcut: "Ctrl+V"},
		{ID: "edit.select-all", Label: "Selecionar tudo", Shortcut: "Ctrl+A"},
		{ID: "search.find", Label: "Buscar", Shortcut: "Ctrl+F"},
		{ID: "search.replace", Label: "Substituir", Shortcut: "Ctrl+H"},
		{ID: "view.go-line", Label: "Ir para linha", Shortcut: "Ctrl+G"},
	}
}

func (m *Model) Init() tea.Cmd {
	return autoSaveTick()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case autoSaveMsg:
		if m.modified && m.filename != "" && m.mode == ModeNormal {
			m.save()
		}
		return m, autoSaveTick()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Resize(msg.Width, msg.Height-1)
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) View() string {
	if m.mode == ModeWelcome {
		return m.renderWelcome()
	}
	if m.mode == ModePalette {
		return m.renderPalette()
	}
	if m.mode == ModeFilePicker {
		return m.renderFilePicker()
	}
	if m.mode == ModeSaveAs {
		return m.renderSaveAs()
	}
	if m.mode == ModeConfirm {
		return m.renderConfirm()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderContent(),
		m.renderStatus(),
	)
}

// renderContent renders the text content area with line numbers and search bar.
func (m *Model) renderContent() string {
	var parts []string

	// Search/replace bar
	if m.mode == ModeSearch || m.mode == ModeReplace {
		parts = append(parts, m.renderSearchBar())
	}

	// Main content
	lines := m.buf.Lines()
	lineNumWidth := m.lineNumberWidth()
	contentHeight := m.viewport.Height()
	if m.mode == ModeSearch || m.mode == ModeReplace {
		contentHeight -= 1
	}

	var visibleLines []string
	for i := m.viewport.ScrollY(); i < len(lines) && i < m.viewport.ScrollY()+contentHeight; i++ {
		lineNum := fmt.Sprintf("%*d ", lineNumWidth, i+1)
		styledLineNum := m.lineNumStyle.Render(lineNum)

		// Apply syntax highlighting
		segments := m.highlighter.HighlightLine(lines[i], m.language)
		lineText := highlight.RenderSegments(segments)

		// Apply search highlighting on top
		if len(m.searchMatches) > 0 {
			lineText = m.applySearchHighlight(lineText, i)
		}

		// Horizontal scroll
		if m.viewport.ScrollX() > 0 && m.viewport.ScrollX() < len(lineText) {
			lineText = lineText[m.viewport.ScrollX():]
		}
		if len(lineText) > m.viewport.Width()-lineNumWidth-1 {
			lineText = lineText[:m.viewport.Width()-lineNumWidth-1]
		}

		visibleLines = append(visibleLines, styledLineNum+lineText)
	}

	content := strings.Join(visibleLines, "\n")
	contentStyle := lipgloss.NewStyle().
		Width(m.viewport.Width()).
		Height(contentHeight)

	parts = append(parts, contentStyle.Render(content))
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *Model) lineNumberWidth() int {
	count := m.buf.LineCount()
	if count < 10 {
		return 2
	}
	if count < 100 {
		return 3
	}
	if count < 1000 {
		return 4
	}
	return 5
}

// --- Key handling ---

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == ModeWelcome {
		return m.handleWelcomeKey(msg)
	}
	if m.mode == ModePalette {
		return m.handlePaletteKey(msg)
	}
	if m.mode == ModeFilePicker {
		return m.handleFilePickerKey(msg)
	}
	if m.mode == ModeSaveAs {
		return m.handleSaveAsKey(msg)
	}
	// In search/replace mode, handle input differently
	if m.mode == ModeSearch || m.mode == ModeReplace {
		return m.handleSearchKey(msg)
	}

	if m.mode == ModeConfirm {
		return m.handleConfirmKey(msg)
	}

	switch msg.String() {
	case "ctrl+shift+p":
		m.enterPalette()
		return m, nil

	case "ctrl+o":
		m.enterFilePicker()
		return m, nil

	case "ctrl+shift+s":
		m.enterSaveAs()
		return m, nil

	case "ctrl+q":
		if m.modified {
			m.mode = ModeConfirm
			m.confirmAction = ConfirmQuit
			return m, nil
		}
		return m, tea.Quit

	case "ctrl+s":
		m.save()
		return m, nil

	// Undo/Redo
	case "ctrl+z":
		m.undo()
		return m, nil
	case "ctrl+y":
		m.redo()
		return m, nil

	// Clipboard
	case "ctrl+c":
		m.copy()
		return m, nil
	case "ctrl+x":
		m.cut()
		return m, nil
	case "ctrl+v":
		m.paste()
		return m, nil

	// Select all
	case "ctrl+a":
		m.selStart = 0
		m.selEnd = m.buf.Len()
		return m, nil

	// Search/Replace
	case "ctrl+f":
		m.mode = ModeSearch
		m.searchQuery = ""
		m.searchMatches = nil
		m.searchCurrent = 0
		return m, nil

	case "ctrl+h":
		m.mode = ModeReplace
		m.searchQuery = ""
		m.replaceQuery = ""
		m.searchMatches = nil
		m.searchCurrent = 0
		return m, nil

	case "backspace":
		m.clearSelection()
		if m.buf.GapPosition() == 0 {
			return m, nil
		}
		// Record undo operation
		r := m.buf.RuneAt(m.buf.GapPosition() - 1)
		m.undoStack.Push(buffer.Operation{
			Type: "insert",
			Pos:  m.buf.GapPosition() - 1,
			Text: string(r),
		})
		if m.buf.Backspace() {
			m.cursor.Col--
			if m.cursor.Col < 0 {
				m.cursor.Col = 0
			}
			m.modified = true
			m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		}
		return m, nil

	case "delete":
		m.clearSelection()
		if m.buf.GapPosition() >= m.buf.Len() {
			return m, nil
		}
		r := m.buf.RuneAt(m.buf.GapPosition())
		m.undoStack.Push(buffer.Operation{
			Type: "insert",
			Pos:  m.buf.GapPosition(),
			Text: string(r),
		})
		if m.buf.DeleteForward() {
			m.modified = true
		}
		return m, nil

	case "enter":
		m.clearSelection()
		m.insertText("\n")
		m.cursor.Line++
		m.cursor.Col = 0
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		return m, nil

	case "tab":
		m.clearSelection()
		m.insertText("    ")
		m.cursor.Col += 4
		return m, nil

	// Navigation
	case "up":
		m.clearSelection()
		m.moveCursorUp()
		return m, nil
	case "down":
		m.clearSelection()
		m.moveCursorDown()
		return m, nil
	case "left":
		m.clearSelection()
		m.moveCursorLeft()
		return m, nil
	case "right":
		m.clearSelection()
		m.moveCursorRight()
		return m, nil
	case "home":
		m.clearSelection()
		m.cursor.Col = 0
		m.syncGapToCursor()
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		return m, nil
	case "end":
		m.clearSelection()
		lineText := m.currentLineText()
		m.cursor.Col = len(lineText)
		m.syncGapToCursor()
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		return m, nil
	case "ctrl+home":
		m.clearSelection()
		m.cursor.SetPos(0, 0)
		m.syncGapToCursor()
		m.viewport.EnsureVisible(0, 0)
		return m, nil
	case "ctrl+end":
		m.clearSelection()
		lastLine := m.buf.LineCount() - 1
		lastLineText := m.lineText(lastLine)
		m.cursor.SetPos(lastLine, len(lastLineText))
		m.syncGapToCursor()
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		return m, nil
	case "ctrl+left":
		m.clearSelection()
		m.moveCursorWordLeft()
		return m, nil
	case "ctrl+right":
		m.clearSelection()
		m.moveCursorWordRight()
		return m, nil

	default:
		m.clearSelection()
		if len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				if r >= 32 {
					m.insertText(string(r))
					m.cursor.Col++
				}
			}
			m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		}
		return m, nil
	}
}

// --- Search input handling ---

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.searchMatches = nil
		return m, nil

	case "enter":
		m.doSearch()
		if len(m.searchMatches) > 0 {
			m.navigateToMatch(0)
		}
		if m.mode == ModeReplace {
			m.doReplace()
		}
		return m, nil

	case "tab":
		if m.mode == ModeReplace {
			// Toggle between search and replace fields
			// For now, just keep the mode
		}
		return m, nil

	case "backspace":
		if m.mode == ModeReplace && m.replaceQuery != "" {
			m.replaceQuery = m.replaceQuery[:len(m.replaceQuery)-1]
		} else if m.searchQuery != "" {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
		return m, nil

	default:
		if len(msg.Runes) > 0 {
			if m.mode == ModeReplace && m.searchQuery != "" && m.replaceQuery == "" && msg.Runes[0] == '\t' {
				// That was tab, already handled
			} else if m.mode == ModeReplace && msg.Runes[0] >= 32 {
				// Find which field is active (simplified: if searchQuery done, type replace)
				// For simplicity, just append to replace or search
				if len(m.searchMatches) > 0 {
					m.replaceQuery += string(msg.Runes)
				} else {
					m.searchQuery += string(msg.Runes)
				}
			} else if msg.Runes[0] >= 32 {
				m.searchQuery += string(msg.Runes)
			}
		}
		return m, nil
	}
}

func (m *Model) doSearch() {
	if m.searchQuery == "" {
		return
	}
	m.lastSearch = m.searchQuery

	content := m.buf.String()
	m.searchMatches = nil

	query := strings.ToLower(m.searchQuery)
	contentLower := strings.ToLower(content)

	for i := 0; i <= len(content)-len(query); i++ {
		if contentLower[i:i+len(query)] == query {
			m.searchMatches = append(m.searchMatches, i)
		}
	}
	m.searchCurrent = 0
}

func (m *Model) doReplace() {
	if len(m.searchMatches) == 0 || m.replaceQuery == "" {
		return
	}

	// Replace current match
	pos := m.searchMatches[m.searchCurrent]
	// Remove selection, insert replacement
	m.moveGapTo(pos)
	for i := 0; i < len(m.searchQuery); i++ {
		m.buf.DeleteForward()
	}
	m.buf.InsertString(m.replaceQuery)
	m.modified = true

	// Update match positions
	m.doSearch()
}

// --- Undo/Redo ---

func (m *Model) undo() {
	op, ok := m.undoStack.Undo()
	if !ok {
		return
	}

	m.moveGapTo(op.Pos)
	switch op.Type {
	case "insert":
		m.buf.InsertString(op.Text)
	case "delete":
		for i := 0; i < len(op.Text); i++ {
			m.buf.DeleteForward()
		}
	}
	m.modified = true
	m.cursor.Line, m.cursor.Col = m.buf.LineCol(m.buf.GapPosition())
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
}

func (m *Model) redo() {
	op, ok := m.undoStack.Redo()
	if !ok {
		return
	}

	m.moveGapTo(op.Pos)
	// Redo applies the INVERSE of the stored operation (re-doing the original action).
	switch op.Type {
	case "insert":
		// Reverse of stored "insert" = delete
		for i := 0; i < len(op.Text); i++ {
			m.buf.DeleteForward()
		}
	case "delete":
		// Reverse of stored "delete" = insert
		m.buf.InsertString(op.Text)
	}
	m.modified = true
	m.cursor.Line, m.cursor.Col = m.buf.LineCol(m.buf.GapPosition())
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
}

// --- Clipboard ---

func (m *Model) copy() {
	if m.hasSelection() {
		text := m.getSelectedText()
		m.clipboard.Copy(text)
	} else {
		// Copy current line
		text := m.currentLineText()
		m.clipboard.Copy(text)
	}
	m.selStart = -1
	m.selEnd = -1
}

func (m *Model) cut() {
	if m.hasSelection() {
		m.copy()
		m.deleteSelection()
	} else {
		// Cut current line
		m.clipboard.Copy(m.currentLineText())
		// Delete the line
		m.moveGapTo(m.buf.LineStart(m.cursor.Line))
		lineEnd := m.buf.LineStart(m.cursor.Line + 1)
		for i := m.buf.LineStart(m.cursor.Line); i < lineEnd && m.buf.GapPosition() < m.buf.Len(); i++ {
			m.buf.DeleteForward()
		}
		m.modified = true
	}
}

func (m *Model) paste() {
	if m.clipboard.HasText() {
		text := m.clipboard.Paste()
		m.undoStack.Push(buffer.Operation{
			Type: "delete",
			Pos:  m.buf.GapPosition(),
			Text: text,
		})
		m.buf.InsertString(text)
		m.cursor.Col += len(text)
		m.modified = true
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
	}
}

// --- Selection ---

func (m *Model) hasSelection() bool {
	return m.selStart >= 0 && m.selEnd >= 0 && m.selStart != m.selEnd
}

func (m *Model) getSelectedText() string {
	if !m.hasSelection() {
		return ""
	}
	start := m.selStart
	end := m.selEnd
	if start > end {
		start, end = end, start
	}
	var sb strings.Builder
	for i := start; i < end; i++ {
		sb.WriteRune(m.buf.RuneAt(i))
	}
	return sb.String()
}

func (m *Model) clearSelection() {
	m.selStart = -1
	m.selEnd = -1
}

func (m *Model) deleteSelection() {
	if !m.hasSelection() {
		return
	}
	text := m.getSelectedText()
	m.undoStack.Push(buffer.Operation{
		Type: "insert",
		Pos:  m.selStart,
		Text: text,
	})

	start := m.selStart
	end := m.selEnd
	if start > end {
		start, end = end, start
	}

	m.moveGapTo(start)
	for i := 0; i < end-start; i++ {
		m.buf.DeleteForward()
	}
	m.selStart = -1
	m.selEnd = -1
	m.modified = true
}

// --- Helpers ---

func (m *Model) insertText(text string) {
	m.undoStack.Push(buffer.Operation{
		Type: "delete",
		Pos:  m.buf.GapPosition(),
		Text: text,
	})
	m.buf.InsertString(text)
	m.modified = true
}

func (m *Model) moveGapTo(pos int) {
	current := m.buf.GapPosition()
	if pos > current {
		for i := current; i < pos; i++ {
			m.buf.MoveGapRight()
		}
	} else {
		for i := current; i > pos; i-- {
			m.buf.MoveGapLeft()
		}
	}
}

// --- Mouse ---

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
		m.viewport.ScrollUp(3)
		return m, nil

	case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
		m.viewport.ScrollDown(3)
		return m, nil

	case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
		m.clearSelection()
		lineNumWidth := m.lineNumberWidth()
		clickLine := m.viewport.ScrollY() + msg.Y
		clickCol := msg.X - lineNumWidth - 1 + m.viewport.ScrollX()
		if clickCol < 0 {
			clickCol = 0
		}
		m.cursor.SetPos(clickLine, clickCol)
		m.clampCursor()
		m.syncGapToCursor()
		return m, nil
	}
	return m, nil
}

// --- Confirmation ---

func (m *Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s", "S":
		if m.confirmAction == ConfirmQuit {
			m.save()
			return m, tea.Quit
		}
	case "d", "D", "n", "N":
		if m.confirmAction == ConfirmQuit {
			return m, tea.Quit
		}
	case "c", "C", "esc":
		m.mode = ModeNormal
		return m, nil
	}
	return m, nil
}

// --- Cursor movement ---

func (m *Model) moveCursorUp() {
	if m.cursor.Line > 0 {
		m.cursor.Line--
		m.clampCursorCol()
		m.syncGapToCursor()
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
	}
}

func (m *Model) moveCursorDown() {
	if m.cursor.Line < m.buf.LineCount()-1 {
		m.cursor.Line++
		m.clampCursorCol()
		m.syncGapToCursor()
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
	}
}

func (m *Model) moveCursorLeft() {
	if m.cursor.Col > 0 {
		m.cursor.Col--
	} else if m.cursor.Line > 0 {
		m.cursor.Line--
		prevLineText := m.lineText(m.cursor.Line)
		m.cursor.Col = len(prevLineText)
	}
	m.syncGapToCursor()
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
}

func (m *Model) moveCursorRight() {
	lineText := m.currentLineText()
	if m.cursor.Col < len(lineText) {
		m.cursor.Col++
	} else if m.cursor.Line < m.buf.LineCount()-1 {
		m.cursor.Line++
		m.cursor.Col = 0
	}
	m.syncGapToCursor()
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
}

func (m *Model) moveCursorWordLeft() {
	lineText := m.currentLineText()
	if m.cursor.Col == 0 {
		if m.cursor.Line > 0 {
			m.cursor.Line--
			prevLineText := m.lineText(m.cursor.Line)
			m.cursor.Col = len(prevLineText)
		}
		m.syncGapToCursor()
		return
	}
	for m.cursor.Col > 0 && isSpace(rune(lineText[m.cursor.Col-1])) {
		m.cursor.Col--
	}
	for m.cursor.Col > 0 && !isSpace(rune(lineText[m.cursor.Col-1])) {
		m.cursor.Col--
	}
	m.syncGapToCursor()
}

func (m *Model) moveCursorWordRight() {
	lineText := m.currentLineText()
	if m.cursor.Col >= len(lineText) {
		if m.cursor.Line < m.buf.LineCount()-1 {
			m.cursor.Line++
			m.cursor.Col = 0
		}
		m.syncGapToCursor()
		return
	}
	for m.cursor.Col < len(lineText) && !isSpace(rune(lineText[m.cursor.Col])) {
		m.cursor.Col++
	}
	for m.cursor.Col < len(lineText) && isSpace(rune(lineText[m.cursor.Col])) {
		m.cursor.Col++
	}
	m.syncGapToCursor()
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func (m *Model) clampCursor() {
	maxLine := m.buf.LineCount() - 1
	if m.cursor.Line > maxLine {
		m.cursor.Line = maxLine
	}
	if m.cursor.Line < 0 {
		m.cursor.Line = 0
	}
	m.clampCursorCol()
}

func (m *Model) clampCursorCol() {
	lineText := m.currentLineText()
	if m.cursor.Col > len(lineText) {
		m.cursor.Col = len(lineText)
	}
	if m.cursor.Col < 0 {
		m.cursor.Col = 0
	}
}

func (m *Model) currentLineText() string {
	return m.lineText(m.cursor.Line)
}

func (m *Model) lineText(line int) string {
	lines := m.buf.Lines()
	if line < 0 || line >= len(lines) {
		return ""
	}
	return lines[line]
}

func (m *Model) syncGapToCursor() {
	targetIndex := m.buf.LineStart(m.cursor.Line) + m.cursor.Col
	m.moveGapTo(targetIndex)
}

// --- Search highlighting ---

func (m *Model) applySearchHighlight(lineText string, lineNum int) string {
	if len(m.searchMatches) == 0 {
		return lineText
	}

	lineStart := m.buf.LineStart(lineNum)
	if lineStart < 0 {
		return lineText
	}

	var result strings.Builder
	pos := 0
	queryLen := len(m.searchQuery)

	for i := 0; i <= len(lineText)-queryLen; i++ {
		logicalPos := lineStart + i
		isMatch := false
		matchIdx := -1

		for j, matchPos := range m.searchMatches {
			if matchPos == logicalPos {
				isMatch = true
				matchIdx = j
				break
			}
		}

		if isMatch {
			result.WriteString(lineText[pos:i])

			matchText := lineText[i : i+queryLen]
			if matchIdx == m.searchCurrent {
				result.WriteString(m.currentMatch.Render(matchText))
			} else {
				result.WriteString(m.matchStyle.Render(matchText))
			}
			pos = i + queryLen
			i += queryLen - 1
		}
	}
	result.WriteString(lineText[pos:])
	return result.String()
}

func (m *Model) navigateToMatch(index int) {
	if index < 0 || index >= len(m.searchMatches) {
		return
	}
	m.searchCurrent = index
	pos := m.searchMatches[index]
	line, col := m.buf.LineCol(pos)
	m.cursor.SetPos(line, col)
	m.syncGapToCursor()
	m.viewport.EnsureVisible(line, col)
}

// --- Search bar rendering ---

func (m *Model) renderSearchBar() string {
	if m.mode == ModeReplace {
		s := fmt.Sprintf("Buscar: %s  Substituir: %s", m.searchQuery, m.replaceQuery)
		if len(m.searchMatches) > 0 {
			s += fmt.Sprintf("  [%d/%d]", m.searchCurrent+1, len(m.searchMatches))
		}
		return m.searchStyle.Render(s)
	}
	s := fmt.Sprintf("Buscar: %s", m.searchQuery)
	if len(m.searchMatches) > 0 {
		s += fmt.Sprintf("  [%d/%d]", m.searchCurrent+1, len(m.searchMatches))
	}
	return m.searchStyle.Render(s)
}

// --- File operations ---

func (m *Model) save() {
	if m.filename == "" {
		m.filename = "untitled.txt"
	}
	if err := fileio.Save(m.filename, m.buf); err == nil {
		m.modified = false
	}
}

// --- Status bar ---

func (m *Model) renderStatus() string {
	fname := m.filename
	if fname == "" {
		fname = "[novo]"
	}
	modified := ""
	if m.modified {
		modified = m.statusModified.Render(" ●")
	}
	lang := m.language
	if lang == "" {
		lang = "texto"
	}
	pos := fmt.Sprintf("L:%d C:%d [%s]", m.cursor.Line+1, m.cursor.Col+1, lang)
	return m.statusStyle.Render(fname + modified + "  " + pos + "  Ctrl+S Salvar  Ctrl+Q Sair")
}

// --- Confirm dialog ---

func (m *Model) renderConfirm() string {
	msg := "Arquivo modificado! Deseja salvar antes de sair?"
	btns := "\n\n" + m.confirmBtnStyle.Render("[S] Salvar") + "  [D] Descartar  [C] Cancelar"
	dialog := m.confirmStyle.Render(msg + btns)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// --- Palette ---

func (m *Model) enterPalette() {
	m.mode = ModePalette
	m.paletteQuery = ""
	m.paletteResult()
	m.paletteSel = 0
}

func (m *Model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil

	case "enter":
		if m.paletteSel >= 0 && m.paletteSel < len(m.paletteResults) {
			action := m.paletteResults[m.paletteSel]
			m.mode = ModeNormal
			m.executeAction(action.ID)
		}
		return m, nil

	case "up":
		m.paletteSel--
		if m.paletteSel < 0 {
			m.paletteSel = len(m.paletteResults) - 1
		}

	case "down":
		m.paletteSel++
		if m.paletteSel >= len(m.paletteResults) {
			m.paletteSel = 0
		}

	case "backspace":
		if len(m.paletteQuery) > 0 {
			m.paletteQuery = m.paletteQuery[:len(m.paletteQuery)-1]
			m.paletteResult()
			m.paletteSel = 0
		}

	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.paletteQuery += string(msg.Runes)
			m.paletteResult()
			m.paletteSel = 0
		}
	}
	return m, nil
}

func (m *Model) paletteResult() {
	if m.paletteQuery == "" {
		m.paletteResults = m.paletteActions
		return
	}
	q := strings.ToLower(m.paletteQuery)
	var filtered []command.Action
	for _, a := range m.paletteActions {
		if strings.Contains(strings.ToLower(a.Label), q) || strings.Contains(strings.ToLower(a.ID), q) {
			filtered = append(filtered, a)
		}
	}
	m.paletteResults = filtered
}

func (m *Model) executeAction(id string) {
	switch id {
	case "file.save":
		m.save()
	case "file.quit":
		if m.modified {
			m.mode = ModeConfirm
			m.confirmAction = ConfirmQuit
		} else {
			// Can't quit from here - handled by key handler
		}
	case "edit.undo":
		m.undo()
	case "edit.redo":
		m.redo()
	case "edit.cut":
		m.cut()
	case "edit.copy":
		m.copy()
	case "edit.paste":
		m.paste()
	case "edit.select-all":
		m.selStart = 0
		m.selEnd = m.buf.Len()
	case "search.find":
		m.mode = ModeSearch
		m.searchQuery = ""
		m.searchMatches = nil
	case "search.replace":
		m.mode = ModeReplace
		m.searchQuery = ""
		m.replaceQuery = ""
		m.searchMatches = nil
	}
}

func (m *Model) renderPalette() string {
	maxItems := 10
	start := m.paletteSel - 5
	if start < 0 {
		start = 0
	}
	end := start + maxItems
	if end > len(m.paletteResults) {
		end = len(m.paletteResults)
	}

	var items []string
	items = append(items, m.paletteInput.Render("> "+m.paletteQuery))

	if len(m.paletteQuery) == 0 {
		items = append(items, m.paletteShortcut.Render("  Digite para buscar comandos..."))
	}
	if len(m.paletteResults) == 0 && len(m.paletteQuery) > 0 {
		items = append(items, m.paletteShortcut.Render("  Nenhum comando encontrado"))
	}

	for i := start; i < end; i++ {
		a := m.paletteResults[i]
		line := fmt.Sprintf("  %-30s", a.Label)
		if a.Shortcut != "" {
			line += " " + m.paletteShortcut.Render(a.Shortcut)
		}
		if i == m.paletteSel {
			items = append(items, m.paletteActive.Render(strings.TrimRight(line, " ")))
		} else {
			items = append(items, line)
		}
	}

	content := strings.Join(items, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}

// --- Welcome Screen ---

func (m *Model) renderWelcome() string {
	logo := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render("  cmdit v0.1.0")
	subtitle := "Editor de texto para humanos"

	var lines []string
	lines = append(lines, logo)
	lines = append(lines, "")
	lines = append(lines, subtitle)
	lines = append(lines, "")

	if len(m.recentFiles) > 0 {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Arquivos recentes:"))
		for i, f := range m.recentFiles {
			if i >= 5 {
				break
			}
			lines = append(lines, fmt.Sprintf("  %d. %s", i+1, f))
		}
		lines = append(lines, "")
	}

	lines = append(lines, "Ctrl+O  Abrir arquivo")
	lines = append(lines, "Ctrl+Shift+P  Paleta de comandos")
	lines = append(lines, "Ctrl+Q  Sair")
	lines = append(lines, "")
	lines = append(lines, "Comece a digitar para criar um novo arquivo...")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) handleWelcomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+q":
		return m, tea.Quit
	case "ctrl+o":
		m.enterFilePicker()
		return m, nil
	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.mode = ModeNormal
			m.handleKey(msg) // Process the character
			return m, nil
		}
	}
	return m, nil
}

// --- File Picker ---

func (m *Model) enterFilePicker() {
	m.mode = ModeFilePicker
	m.filePickerDir = "."
	m.filePickerQuery = ""
	m.loadDirectory()
	m.filePickerSel = 0
}

func (m *Model) loadDirectory() {
	entries, err := os.ReadDir(m.filePickerDir)
	if err != nil {
		m.filePickerFiles = nil
		return
	}
	m.filePickerFiles = nil
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		if m.filePickerQuery == "" || strings.Contains(strings.ToLower(name), strings.ToLower(m.filePickerQuery)) {
			m.filePickerFiles = append(m.filePickerFiles, name)
		}
	}
}

func (m *Model) handleFilePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		if m.buf.Len() == 0 && m.filename == "" {
			m.mode = ModeWelcome
		}
		return m, nil

	case "enter":
		if m.filePickerSel >= 0 && m.filePickerSel < len(m.filePickerFiles) {
			name := m.filePickerFiles[m.filePickerSel]
			fullPath := filepath.Join(m.filePickerDir, strings.TrimSuffix(name, "/"))
			info, err := os.Stat(fullPath)
			if err == nil && info.IsDir() {
				m.filePickerDir = fullPath
				m.filePickerSel = 0
				m.loadDirectory()
				return m, nil
			}
			m.openFile(fullPath)
		}
		return m, nil

	case "up":
		m.filePickerSel--
		if m.filePickerSel < 0 {
			m.filePickerSel = len(m.filePickerFiles) - 1
		}

	case "down":
		m.filePickerSel++
		if m.filePickerSel >= len(m.filePickerFiles) {
			m.filePickerSel = 0
		}

	case "backspace":
		if len(m.filePickerQuery) > 0 {
			m.filePickerQuery = m.filePickerQuery[:len(m.filePickerQuery)-1]
			m.loadDirectory()
			m.filePickerSel = 0
		}

	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.filePickerQuery += string(msg.Runes)
			m.loadDirectory()
			m.filePickerSel = 0
		}
	}
	return m, nil
}

func (m *Model) openFile(path string) {
	b, err := fileio.Load(path)
	if err != nil {
		return
	}
	m.buf = b
	m.filename = path
	m.language = highlight.DetectLanguage(path)
	m.modified = false
	m.mode = ModeNormal
	m.cursor.SetPos(0, 0)
	m.syncGapToCursor()
	m.addRecentFile(path)
}

func (m *Model) renderFilePicker() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Abrir: %s > %s", m.filePickerDir, m.filePickerQuery))
	lines = append(lines, "")

	start := m.filePickerSel - 10
	if start < 0 {
		start = 0
	}
	end := start + 20
	if end > len(m.filePickerFiles) {
		end = len(m.filePickerFiles)
	}

	for i := start; i < end; i++ {
		line := "  " + m.filePickerFiles[i]
		if i == m.filePickerSel {
			line = m.paletteActive.Render(line)
		}
		lines = append(lines, line)
	}

	if len(m.filePickerFiles) == 0 {
		lines = append(lines, "  (diretório vazio)")
	}

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}

// --- Save As ---

func (m *Model) enterSaveAs() {
	m.mode = ModeSaveAs
	m.saveAsQuery = m.filename
}

func (m *Model) handleSaveAsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil

	case "enter":
		if m.saveAsQuery != "" {
			m.filename = m.saveAsQuery
			m.language = highlight.DetectLanguage(m.filename)
			m.save()
			m.addRecentFile(m.filename)
			m.mode = ModeNormal
		}
		return m, nil

	case "backspace":
		if len(m.saveAsQuery) > 0 {
			m.saveAsQuery = m.saveAsQuery[:len(m.saveAsQuery)-1]
		}

	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.saveAsQuery += string(msg.Runes)
		}
	}
	return m, nil
}

func (m *Model) renderSaveAs() string {
	var lines []string
	lines = append(lines, "Salvar como:")
	lines = append(lines, "")
	lines = append(lines, "> "+m.saveAsQuery)
	lines = append(lines, "")
	lines = append(lines, "Enter para confirmar, Esc para cancelar")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}

// --- Recent Files ---

func (m *Model) addRecentFile(path string) {
	// Remove if already exists
	for i, f := range m.recentFiles {
		if f == path {
			m.recentFiles = append(m.recentFiles[:i], m.recentFiles[i+1:]...)
			break
		}
	}
	// Add to front
	m.recentFiles = append([]string{path}, m.recentFiles...)
	// Limit to 10
	if len(m.recentFiles) > 10 {
		m.recentFiles = m.recentFiles[:10]
	}
	// Also save the current filename in OpenFile/SaveAs for recovery
}

func (m *Model) loadRecentFiles() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path := filepath.Join(home, ".cmdit", "recent.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	// Simple format: one path per line
	m.recentFiles = strings.Split(strings.TrimSpace(string(data)), "\n")
}

func (m *Model) saveRecentFiles() {
	if len(m.recentFiles) == 0 {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".cmdit")
	os.MkdirAll(dir, 0700)
	path := filepath.Join(dir, "recent.json")
	os.WriteFile(path, []byte(strings.Join(m.recentFiles, "\n")), 0600)
}

// --- Auto-save ---

type autoSaveMsg struct{}

func autoSaveTick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return autoSaveMsg{}
	})
}
