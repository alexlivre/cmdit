// Package editor implements the main Bubble Tea model for the cmdit editor.
package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

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

// Version is the current released version of cmdit, shown in the welcome
// screen and other user-facing surfaces. Bump this together with the
// CHANGELOG / README badge on each release.
const Version = "v0.4.3"

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
	ModeGoToLine
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
	selStart  int
	selEnd    int
	selecting bool // true while mouse-drag selecting

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

	goToLineInput string

	// Error display
	errorMessage string
	errorTime    time.Time

	// Multi-cursor state
	extraCursors []EditorCursor

	// Close request from container (tab/split manager)
	closeRequested bool

	// Config
	config Config

	// Vim mode
	vimState VimState

	// Auto-close
	autoClosed map[int]bool // positions with auto-closed pairs

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

	m.vimState = newVimState()

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
	m.startLSP()

	return m, nil
}

// registerActions populates the command palette actions.
func (m *Model) registerActions() {
	m.paletteActions = []command.Action{
		{ID: "file.save", Label: "Save", Shortcut: "Ctrl+S"},
		{ID: "file.save-as", Label: "Save As", Shortcut: "F3"},
		{ID: "file.open", Label: "Open File", Shortcut: "Ctrl+O"},
		{ID: "file.quit", Label: "Quit", Shortcut: "Ctrl+Q"},
		{ID: "edit.undo", Label: "Undo", Shortcut: "Ctrl+Z"},
		{ID: "edit.redo", Label: "Redo", Shortcut: "Ctrl+Y"},
		{ID: "edit.cut", Label: "Cut", Shortcut: "Ctrl+X"},
		{ID: "edit.copy", Label: "Copy", Shortcut: "Ctrl+C"},
		{ID: "edit.paste", Label: "Paste", Shortcut: "Ctrl+V"},
		{ID: "edit.select-all", Label: "Select All", Shortcut: "Ctrl+A"},
		{ID: "search.find", Label: "Find", Shortcut: "Ctrl+F"},
		{ID: "search.replace", Label: "Replace", Shortcut: "Ctrl+H"},
		{ID: "view.go-line", Label: "Go to Line", Shortcut: "Ctrl+G"},
		{ID: "file.rename", Label: "Rename File", Shortcut: "F2"},
		{ID: "view.toggle-auto-close", Label: "Toggle Auto-Close Brackets", Shortcut: "F4"},
		{ID: "view.toggle-vim-mode", Label: "Toggle Vim Mode", Shortcut: "F5"},
		{ID: "view.next-theme", Label: "Next Theme", Shortcut: "F6"},
		{ID: "view.toggle-word-wrap", Label: "Toggle Word Wrap", Shortcut: "Alt+Z"},
		{ID: "file.toggle-format-on-save", Label: "Toggle Format on Save", Shortcut: ""},
		{ID: "file.toggle-auto-save", Label: "Toggle Auto Save", Shortcut: "F9"},
	}
}

func (m *Model) Init() tea.Cmd {
	return autoSaveTick()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearErrorMsg:
		m.errorMessage = ""
		return m, nil

	case autoSaveMsg:
		if m.modified && m.filename != "" && m.mode == ModeNormal && m.config.AutoSaveEnabled {
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

// Save exports the save operation for external callers (e.g. tab container).
// It routes through the proper save() path so errors are surfaced via the
// status bar instead of being silently dropped and the buffer being wrongly
// marked clean. If there is no filename it falls back to the save-as prompt.
func (m *Model) Save() {
	m.save()
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

	positions := m.sortedCursorPositions()
	var ops []buffer.Operation
	// High → low so earlier positions stay valid after inserts.
	for i := len(positions) - 1; i >= 0; i-- {
		pos := positions[i]
		m.moveGapTo(pos)
		ops = append(ops, buffer.Operation{
			Type: "delete",
			Pos:  pos,
			Text: text,
		})
		m.buf.InsertString(text)
	}
	m.undoStack.PushComposite(ops)

	delta := utf8.RuneCountInString(text)
	m.cursor.Col += delta
	for i := range m.extraCursors {
		m.extraCursors[i].Col += delta
	}
	m.refreshExtraCursorGapPos()
	m.syncGapToCursor()
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
	maxCol := utf8.RuneCountInString(lineText)
	if m.cursor.Col > maxCol {
		m.cursor.Col = maxCol
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
		m.cursor.Col = utf8.RuneCountInString(prevLineText)
	}
	m.syncGapToCursor()
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
}

func (m *Model) moveCursorRight() {
	lineText := m.currentLineText()
	if m.cursor.Col < utf8.RuneCountInString(lineText) {
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
	runes := []rune(lineText)
	if m.cursor.Col == 0 {
		if m.cursor.Line > 0 {
			m.cursor.Line--
			prevLineText := m.lineText(m.cursor.Line)
			m.cursor.Col = utf8.RuneCountInString(prevLineText)
		}
		m.syncGapToCursor()
		return
	}
	for m.cursor.Col > 0 && isSpace(runes[m.cursor.Col-1]) {
		m.cursor.Col--
	}
	for m.cursor.Col > 0 && !isSpace(runes[m.cursor.Col-1]) {
		m.cursor.Col--
	}
	m.syncGapToCursor()
}

func (m *Model) moveCursorWordRight() {
	lineText := m.currentLineText()
	runes := []rune(lineText)
	if m.cursor.Col >= len(runes) {
		if m.cursor.Line < m.buf.LineCount()-1 {
			m.cursor.Line++
			m.cursor.Col = 0
		}
		m.syncGapToCursor()
		return
	}
	for m.cursor.Col < len(runes) && !isSpace(runes[m.cursor.Col]) {
		m.cursor.Col++
	}
	for m.cursor.Col < len(runes) && isSpace(runes[m.cursor.Col]) {
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
			return fmt.Errorf("invalid character: %c", c)
		}
	}
	return nil
}

// --- Multi-cursor helpers ---

// cursorGapPos returns the logical gap index for a (line, col) position.
func (m *Model) cursorGapPos(line, col int) int {
	return m.buf.LineStart(line) + col
}

// refreshExtraCursorGapPos recomputes GapPos for every extra cursor from Line/Col.
func (m *Model) refreshExtraCursorGapPos() {
	for i := range m.extraCursors {
		m.extraCursors[i].GapPos = m.cursorGapPos(m.extraCursors[i].Line, m.extraCursors[i].Col)
	}
}

// sortedCursorPositions returns primary + extra gap positions sorted ascending.
func (m *Model) sortedCursorPositions() []int {
	m.refreshExtraCursorGapPos()
	positions := make([]int, 0, len(m.extraCursors)+1)
	positions = append(positions, m.cursorGapPos(m.cursor.Line, m.cursor.Col))
	for _, c := range m.extraCursors {
		positions = append(positions, c.GapPos)
	}
	sort.Ints(positions)
	return positions
}

// allCursors returns all active cursors (primary + extras) sorted by GapPos ascending.
func (m *Model) allCursors() []EditorCursor {
	m.refreshExtraCursorGapPos()
	primary := EditorCursor{
		Line:   m.cursor.Line,
		Col:    m.cursor.Col,
		GapPos: m.cursorGapPos(m.cursor.Line, m.cursor.Col),
	}
	all := make([]EditorCursor, 0, len(m.extraCursors)+1)
	all = append(all, primary)
	all = append(all, m.extraCursors...)
	sort.Slice(all, func(i, j int) bool {
		return all[i].GapPos < all[j].GapPos
	})
	return all
}

// wordAtCursor returns the word under the primary cursor.
func (m *Model) wordAtCursor() string {
	line := m.currentLineText()
	if len(line) == 0 {
		return ""
	}

	// Operate in rune-space so multibyte sequences are not split.
	runes := []rune(line)
	col := m.cursor.Col
	if col > len(runes) {
		col = len(runes)
	}

	start := col
	for start > 0 && isWordChar(runes[start-1]) {
		start--
	}
	end := col
	for end < len(runes) && isWordChar(runes[end]) {
		end++
	}

	if start == end {
		return ""
	}
	return string(runes[start:end])
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// indexOfRunes finds needle in haystack starting at start (rune indices).
// Returns -1 if not found.
func indexOfRunes(haystack, needle []rune, start int) int {
	if len(needle) == 0 || start < 0 || start > len(haystack) {
		return -1
	}
	n := len(needle)
	for i := start; i <= len(haystack)-n; i++ {
		match := true
		for j := 0; j < n; j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// addNextOccurrence finds the next occurrence of the word at cursor and adds it.
func (m *Model) addNextOccurrence() {
	word := m.wordAtCursor()
	if word == "" {
		return
	}

	m.refreshExtraCursorGapPos()
	existing := make(map[int]bool)
	primaryPos := m.cursorGapPos(m.cursor.Line, m.cursor.Col)
	existing[primaryPos] = true
	for _, c := range m.extraCursors {
		existing[c.GapPos] = true
	}

	lastPos := primaryPos
	for _, c := range m.extraCursors {
		if c.GapPos > lastPos {
			lastPos = c.GapPos
		}
	}

	content := []rune(m.buf.String())
	needle := []rune(word)
	if lastPos+1 >= len(content) {
		return
	}
	searchStart := lastPos + 1

	idx := indexOfRunes(content, needle, searchStart)
	if idx < 0 {
		idx = indexOfRunes(content, needle, 0)
	}
	if idx < 0 || idx >= len(content) {
		return
	}
	if existing[idx] {
		return
	}

	line, col := m.buf.LineCol(idx)
	m.extraCursors = append(m.extraCursors, EditorCursor{
		Line:   line,
		Col:    col,
		GapPos: idx,
	})
}

// clearExtraCursors removes all extra cursors.
func (m *Model) clearExtraCursors() {
	m.extraCursors = nil
}
