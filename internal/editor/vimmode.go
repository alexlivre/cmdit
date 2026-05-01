package editor

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexb/cmdit/internal/buffer"
)

// VimMode represents the current vim submode.
type VimMode int

const (
	VimNormal  VimMode = iota
	VimInsert
	VimVisual
	VimCommand
)

// VimState holds the vim mode state machine.
type VimState struct {
	Mode       VimMode
	Count      string // accumulated digits for count prefix
	CommandBuf string // ":" command buffer
	LastOp     string // for two-key sequences like gg, dd, yy
}

func newVimState() VimState {
	return VimState{Mode: VimNormal}
}

// --- Main dispatcher ---

// dispatchVimKey routes a key event through the vim state machine.
func (m *Model) dispatchVimKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	switch m.vimState.Mode {
	case VimNormal:
		return m.vimNormal(key)
	case VimInsert:
		return m.vimInsert(key)
	case VimVisual:
		return m.vimVisual(key)
	case VimCommand:
		return m.vimCommand(key)
	}
	return nil
}

// --- Normal mode ---

func (m *Model) vimNormal(key string) tea.Cmd {
	// Count prefix: accumulate digits
	if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
		m.vimState.Count += key
		return nil
	}

	// Handle F5 toggle directly (so vim mode can be turned off from within vim)
	if key == "f5" {
		m.executeAction("view.toggle-vim-mode")
		return nil
	}

	count := m.parseVimCount()

	switch key {
	// Navigation
	case "h":
		m.vimMoveLeft(count)
	case "j":
		m.vimMoveDown(count)
	case "k":
		m.vimMoveUp(count)
	case "l":
		m.vimMoveRight(count)
	case "w":
		m.vimNextWord(count)
	case "b":
		m.vimPrevWord(count)
	case "0":
		m.cursor.Col = 0
		m.syncGapToCursor()
	case "$":
		lineText := m.currentLineText()
		m.cursor.Col = len(lineText)
		m.syncGapToCursor()

	// Two-key sequences
	case "g":
		if m.vimState.LastOp == "g" {
			m.cursor.SetPos(0, 0)
			m.syncGapToCursor()
			m.vimState.LastOp = ""
		} else {
			m.vimState.LastOp = "g"
			return nil
		}
	case "G":
		lastLine := m.buf.LineCount() - 1
		lastLineText := m.lineText(lastLine)
		m.cursor.SetPos(lastLine, len(lastLineText))
		m.syncGapToCursor()
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)

	// Insert modes: transition to VimInsert
	case "i":
		m.vimState.Mode = VimInsert
	case "I":
		m.cursor.Col = 0
		m.syncGapToCursor()
		m.vimState.Mode = VimInsert
	case "a":
		m.moveCursorRight()
		m.vimState.Mode = VimInsert
	case "A":
		lineText := m.currentLineText()
		m.cursor.Col = len(lineText)
		m.syncGapToCursor()
		m.vimState.Mode = VimInsert
	case "o":
		// Open line below: go to end, insert newline, enter insert
		m.cursor.Col = len(m.currentLineText())
		m.syncGapToCursor()
		m.buf.Insert('\n')
		m.modified = true
		m.sendDidChange()
		m.cursor.Line++
		m.cursor.Col = 0
		m.syncGapToCursor()
		m.vimState.Mode = VimInsert
		return nil
	case "O":
		// Open line above: go to start, insert newline, move up, enter insert
		m.cursor.Col = 0
		m.syncGapToCursor()
		m.buf.Insert('\n')
		m.modified = true
		m.sendDidChange()
		m.cursor.Col = 0
		m.syncGapToCursor()
		// Cursor stays at start of new line (which is the previous line)
		m.vimState.Mode = VimInsert
		return nil

	// Delete
	case "x":
		m.vimDeleteChar(count)
	case "d":
		if m.vimState.LastOp == "d" {
			m.vimDeleteLine()
			m.vimState.LastOp = ""
		} else {
			m.vimState.LastOp = "d"
			return nil
		}

	// Yank
	case "y":
		if m.vimState.LastOp == "y" {
			m.vimYankLine()
			m.vimState.LastOp = ""
		} else {
			m.vimState.LastOp = "y"
			return nil
		}

	// Paste
	case "p":
		m.vimPasteAfter()
	case "P":
		m.vimPasteBefore()

	// Undo
	case "u":
		m.undo()

	// Search (reuse existing search mode)
	case "/":
		m.mode = ModeSearch

	// Command mode
	case ":":
		m.vimState.Mode = VimCommand
		m.vimState.CommandBuf = ""

	// Visual mode
	case "v":
		m.vimState.Mode = VimVisual
	}

	m.vimState.Count = ""
	return nil
}

// --- Insert mode ---

func (m *Model) vimInsert(key string) tea.Cmd {
	switch key {
	case "esc":
		m.vimState.Mode = VimNormal
		return nil
	}
	// All other keys pass through to normal key handling
	return nil
}

// --- Command mode ---

func (m *Model) vimCommand(key string) tea.Cmd {
	switch key {
	case "esc":
		m.vimState.Mode = VimNormal
		m.vimState.CommandBuf = ""
	case "enter":
		cmd := strings.TrimSpace(m.vimState.CommandBuf)
		m.vimState.Mode = VimNormal
		m.vimState.CommandBuf = ""
		switch cmd {
		case "w":
			m.save()
		case "q":
			if m.modified {
				m.mode = ModeConfirm
				m.confirmAction = ConfirmQuit
			} else {
				m.closeRequested = true
			}
		case "wq":
			m.save()
			m.closeRequested = true
		case "q!":
			m.closeRequested = true
		}
	case "backspace":
		if len(m.vimState.CommandBuf) > 0 {
			m.vimState.CommandBuf = m.vimState.CommandBuf[:len(m.vimState.CommandBuf)-1]
		}
	default:
		if len(key) == 1 && key[0] >= 32 {
			m.vimState.CommandBuf += key
		}
	}
	return nil
}

// --- Visual mode ---

func (m *Model) vimVisual(key string) tea.Cmd {
	switch key {
	case "esc":
		m.vimState.Mode = VimNormal
	case "y":
		m.copy()
		m.vimState.Mode = VimNormal
	case "d":
		m.cut()
		m.vimState.Mode = VimNormal
	}
	return nil
}

// --- Helpers ---

func (m *Model) parseVimCount() int {
	if m.vimState.Count == "" {
		return 1
	}
	n := 0
	for _, ch := range m.vimState.Count {
		n = n*10 + int(ch-'0')
	}
	if n == 0 {
		return 1
	}
	return n
}

func (m *Model) vimMoveLeft(n int) {
	for i := 0; i < n; i++ {
		m.moveCursorLeft()
	}
}

func (m *Model) vimMoveRight(n int) {
	for i := 0; i < n; i++ {
		m.moveCursorRight()
	}
}

func (m *Model) vimMoveDown(n int) {
	for i := 0; i < n; i++ {
		m.moveCursorDown()
	}
}

func (m *Model) vimMoveUp(n int) {
	for i := 0; i < n; i++ {
		m.moveCursorUp()
	}
}

func (m *Model) vimNextWord(n int) {
	for i := 0; i < n; i++ {
		m.moveCursorWordRight()
	}
}

func (m *Model) vimPrevWord(n int) {
	for i := 0; i < n; i++ {
		m.moveCursorWordLeft()
	}
}

func (m *Model) vimDeleteChar(n int) {
	for i := 0; i < n; i++ {
		r := m.buf.RuneAt(m.buf.GapPosition())
		m.undoStack.Push(buffer.Operation{
			Type: "insert",
			Pos:  m.buf.GapPosition(),
			Text: string(r),
		})
		m.buf.DeleteForward()
	}
	m.modified = true
	m.sendDidChange()
	m.clampCursor()
}

func (m *Model) vimDeleteLine() {
	lineText := m.currentLineText()

	// Push undo operation for the line content
	m.undoStack.Push(buffer.Operation{
		Type: "insert",
		Pos:  m.buf.LineStart(m.cursor.Line),
		Text: lineText + "\n",
	})

	// Move gap to line start
	m.cursor.Col = 0
	m.syncGapToCursor()

	// Delete characters on this line
	lineLen := len(lineText)
	for i := 0; i < lineLen; i++ {
		m.buf.DeleteForward()
	}
	// Delete the newline (if present)
	if m.buf.GapPosition() < m.buf.Len() {
		r := m.buf.RuneAt(m.buf.GapPosition())
		if r == '\n' {
			m.buf.DeleteForward()
		}
	}

	m.modified = true
	m.sendDidChange()
	m.clampCursor()
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
}

func (m *Model) vimYankLine() {
	lineText := m.currentLineText()
	m.clipboard.Copy(lineText + "\n")
}

func (m *Model) vimPasteAfter() {
	if !m.clipboard.HasText() {
		return
	}
	text := m.clipboard.Paste()
	// Move cursor right one position (paste after current position)
	if m.cursor.Col < len(m.currentLineText()) {
		m.cursor.Col++
		m.syncGapToCursor()
	} else {
		m.cursor.Col = len(m.currentLineText())
		m.syncGapToCursor()
	}

	m.undoStack.Push(buffer.Operation{
		Type: "delete",
		Pos:  m.buf.GapPosition(),
		Text: text,
	})
	m.buf.InsertString(text)
	m.modified = true
	m.sendDidChange()
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
}

func (m *Model) vimPasteBefore() {
	if !m.clipboard.HasText() {
		return
	}
	text := m.clipboard.Paste()

	m.undoStack.Push(buffer.Operation{
		Type: "delete",
		Pos:  m.buf.GapPosition(),
		Text: text,
	})
	m.buf.InsertString(text)
	m.modified = true
	m.sendDidChange()
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
}
