// Package editor implements the main Bubble Tea model for the cmdit editor.
package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexb/cmdit/internal/buffer"
	"github.com/alexb/cmdit/internal/clipboard"
	"github.com/alexb/cmdit/internal/command"
	"github.com/alexb/cmdit/internal/fileio"
	"github.com/alexb/cmdit/internal/highlight"
	"github.com/alexb/cmdit/internal/lsp"
	"github.com/alexb/cmdit/internal/renderer"
)

// Mode represents the current editor mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeConfirm
	ModeSearch
	ModeReplace
	ModePalette
	ModeFilePicker
	ModeSaveAs
	ModeRename
	ModeWelcome
)

// ConfirmAction identifies which confirmation dialog is active.
type ConfirmAction int

const (
	ConfirmQuit ConfirmAction = iota
	ConfirmCloseTab
)

// CloseRequested returns true when the editor wants its container to close this tab.
func (m *Model) CloseRequested() bool {
	return m.closeRequested
}

// ConfirmCloseTabMode puts the editor in confirmation mode for closing a tab.
func (m *Model) ConfirmCloseTabMode() {
	m.mode = ModeConfirm
	m.confirmAction = ConfirmCloseTab
}

// EditorCursor holds a cursor with its resolved gap position for multi-cursor editing.
type EditorCursor struct {
	Line   int
	Col    int
	GapPos int
}

// Model is the main Bubble Tea model for the editor.
type Model struct {
	buf       *buffer.Buffer
	cursor    *buffer.Cursor
	viewport  *renderer.Viewport
	undoStack *buffer.UndoStack
	clipboard *clipboard.Clipboard
	filename  string
	modified  bool
	mode      Mode
	language  string

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
	filePickerDir   string
	filePickerFiles []string
	filePickerSel   int
	filePickerQuery string
	saveAsQuery     string

	// Recent files
	recentFiles []string

	// Rename state
	renameInput string
	renameError string

	// Multi-cursor state
	extraCursors []EditorCursor

	// Close request from container (tab/split manager)
	closeRequested bool

	// Config
	config Config

	// LSP integration
	lspClient        *lsp.Client
	diagnostics      map[int][]lsp.Diagnostic // line → diagnostics
	lspVersion       int
	lspDiagnosticsMu sync.Mutex

	width  int
	height int

	// Styles
	statusStyle      lipgloss.Style
	statusModified   lipgloss.Style
	statusNormal     lipgloss.Style
	confirmStyle     lipgloss.Style
	confirmBtnStyle  lipgloss.Style
	lineNumStyle     lipgloss.Style
	searchStyle      lipgloss.Style
	matchStyle       lipgloss.Style
	currentMatch     lipgloss.Style
	selectionStyle   lipgloss.Style
	paletteStyle     lipgloss.Style
	paletteInput     lipgloss.Style
	paletteActive    lipgloss.Style
	paletteShortcut  lipgloss.Style
	cursorExtraStyle lipgloss.Style
	indentGuideStyle lipgloss.Style
}

// New creates a new editor model.
func New() *Model {
	m := &Model{
		buf:              buffer.NewBuffer(),
		cursor:           buffer.NewCursor(),
		viewport:         renderer.NewViewport(80, 24),
		undoStack:        buffer.NewUndoStack(),
		clipboard:        clipboard.New(),
		filename:         "",
		modified:         false,
		mode:             ModeWelcome,
		language:         "",
		highlighter:      highlight.NewHighlighter(highlight.ThemeDark),
		selStart:         -1,
		selEnd:           -1,
		statusStyle:      lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252")).Padding(0, 1),
		statusModified:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		statusNormal:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		confirmStyle:     lipgloss.NewStyle().Background(lipgloss.Color("240")).Padding(1, 2),
		confirmBtnStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		lineNumStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		searchStyle:      lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("252")).Padding(0, 1),
		matchStyle:       lipgloss.NewStyle().Background(lipgloss.Color("58")),
		currentMatch:     lipgloss.NewStyle().Background(lipgloss.Color("214")).Foreground(lipgloss.Color("0")),
		selectionStyle:   lipgloss.NewStyle().Background(lipgloss.Color("240")),
		paletteStyle:     lipgloss.NewStyle().Background(lipgloss.Color("236")).Padding(1, 2),
		paletteInput:     lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")).Padding(0, 1),
		paletteActive:    lipgloss.NewStyle().Background(lipgloss.Color("214")).Foreground(lipgloss.Color("0")).Padding(0, 1),
		paletteShortcut:  lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		cursorExtraStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		indentGuideStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("236")),
	}
	m.registerActions()
	m.loadRecentFiles()

	cfg, _ := LoadConfig()
	m.config = cfg
	m.highlighter.SetTheme(cfg.Theme)

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

	return m, nil
}

// registerActions populates the command palette actions.
func (m *Model) registerActions() {
	m.paletteActions = []command.Action{
		{ID: "file.save", Label: "Salvar", Shortcut: "Ctrl+S"},
		{ID: "file.save-as", Label: "Salvar como", Shortcut: "F3"},
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
		{ID: "file.rename", Label: "Renomear arquivo", Shortcut: "F2"},
		{ID: "view.toggle-auto-close", Label: "Toggle Auto-Close Brackets", Shortcut: "F4"},
		{ID: "view.toggle-vim-mode", Label: "Toggle Modo Vim", Shortcut: "F5"},
		{ID: "view.next-theme", Label: "Proximo Tema", Shortcut: "F6"},
		{ID: "view.toggle-word-wrap", Label: "Toggle Word Wrap", Shortcut: "Alt+Z"},
		{ID: "file.toggle-format-on-save", Label: "Toggle Format on Save", Shortcut: ""},
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

// --- Accessors for tab/split containers ---

// Modified returns whether the buffer has unsaved changes.
func (m *Model) Modified() bool { return m.modified }

// Filename returns the current file path.
func (m *Model) Filename() string { return m.filename }

// Mode returns the current editor mode.
func (m *Model) CurrentMode() Mode { return m.mode }

// Buffer returns the underlying buffer (for external access).
func (m *Model) Buffer() *buffer.Buffer { return m.buf }

// SetFilename sets the filename (used by tab renaming).
func (m *Model) SetFilename(name string) {
	m.filename = name
	m.language = highlight.DetectLanguage(name)
}

// ResetMode forces the editor back to Normal mode (used when switching tabs).
func (m *Model) ResetMode() { m.mode = ModeNormal }

// Save exports the save operation for the tab container.
func (m *Model) Save() {
	if m.filename == "" {
		m.filename = "untitled.txt"
	}
	if err := fileio.Save(m.filename, m.buf); err == nil {
		m.modified = false
	}
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
	m.sendDidChange()
}

func (m *Model) insertTextAtAllCursors(text string) {
	if len(m.extraCursors) == 0 {
		m.insertText(text)
		return
	}

	all := m.allCursors()
	// Process from end to start to preserve positions
	for i := len(all) - 1; i >= 0; i-- {
		m.moveGapTo(all[i].GapPos)
		m.undoStack.Push(buffer.Operation{
			Type: "delete",
			Pos:  all[i].GapPos,
			Text: text,
		})
		m.buf.InsertString(text)
	}

	// Update cursor and extra cursors
	m.cursor.Col += len(text)
	for i := range m.extraCursors {
		m.extraCursors[i].Col += len(text)
	}
	m.modified = true
	m.sendDidChange()
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

func (m *Model) syncGapToCursor() {
	targetIndex := m.buf.LineStart(m.cursor.Line) + m.cursor.Col
	m.moveGapTo(targetIndex)
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

// --- Auto-save ---

type autoSaveMsg struct{}

func autoSaveTick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return autoSaveMsg{}
	})
}

// --- Recent Files ---

func (m *Model) addRecentFile(path string) {
	for i, f := range m.recentFiles {
		if f == path {
			m.recentFiles = append(m.recentFiles[:i], m.recentFiles[i+1:]...)
			break
		}
	}
	m.recentFiles = append([]string{path}, m.recentFiles...)
	if len(m.recentFiles) > 10 {
		m.recentFiles = m.recentFiles[:10]
	}
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
	m.recentFiles = strings.Split(strings.TrimSpace(string(data)), "\n")
}

func (m *Model) SaveRecentFiles() {
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

func (m *Model) saveRecentFiles() {
	m.SaveRecentFiles()
}

// validateFileName checks for invalid characters in a file name.
func validateFileName(name string) error {
	invalid := []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}
	for _, c := range invalid {
		if strings.ContainsRune(name, c) {
			return fmt.Errorf("caractere invalido: %c", c)
		}
	}
	return nil
}

// --- Multi-cursor helpers ---

// allCursors returns all active cursors (primary + extras) sorted by position.
func (m *Model) allCursors() []EditorCursor {
	primary := EditorCursor{
		Line:   m.cursor.Line,
		Col:    m.cursor.Col,
		GapPos: m.buf.GapPosition(),
	}
	all := make([]EditorCursor, 0, len(m.extraCursors)+1)
	all = append(all, primary)
	all = append(all, m.extraCursors...)
	return all
}

// wordAtCursor returns the word under the primary cursor.
func (m *Model) wordAtCursor() string {
	line := m.currentLineText()
	if len(line) == 0 {
		return ""
	}

	start := m.cursor.Col
	for start > 0 && isWordChar(rune(line[start-1])) {
		start--
	}
	end := m.cursor.Col
	for end < len(line) && isWordChar(rune(line[end])) {
		end++
	}

	if start == end {
		return ""
	}
	return line[start:end]
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// addNextOccurrence finds the next occurrence of the word at cursor and adds it.
func (m *Model) addNextOccurrence() {
	word := m.wordAtCursor()
	if word == "" {
		return
	}

	// Build list of existing positions (primary + extras)
	existing := make(map[int]bool)
	existing[m.buf.GapPosition()] = true
	for _, c := range m.extraCursors {
		existing[c.GapPos] = true
	}

	// Start search from after the last cursor position
	lastPos := m.buf.GapPosition()
	for _, c := range m.extraCursors {
		if c.GapPos > lastPos {
			lastPos = c.GapPos
		}
	}

	content := m.buf.String()
	searchStart := lastPos + 1

	idx := strings.Index(content[searchStart:], word)
	if idx == -1 {
		// Wrap around: search from beginning
		idx = strings.Index(content, word)
	}
	if idx == -1 {
		return
	}

	// Calculate actual position
	actualPos := idx
	if actualPos < len(content) && actualPos >= searchStart {
		// already correct
	} else if idx >= 0 && idx < searchStart-lastPos+searchStart {
		actualPos = idx
	} else {
		actualPos = searchStart + idx
	}

	// Clamp
	if actualPos >= len(content) {
		actualPos = idx
	}
	if actualPos >= len(content) {
		return
	}

	// Don't add duplicates
	if existing[actualPos] {
		return
	}

	line, col := m.buf.LineCol(actualPos)
	m.extraCursors = append(m.extraCursors, EditorCursor{
		Line:   line,
		Col:    col,
		GapPos: actualPos,
	})
}

// clearExtraCursors removes all extra cursors.
func (m *Model) clearExtraCursors() {
	m.extraCursors = nil
}
