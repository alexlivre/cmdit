package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexb/cmdit/internal/buffer"
	"github.com/alexb/cmdit/internal/fileio"
	"github.com/alexb/cmdit/internal/highlight"
)

// --- Welcome key handler ---

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
			m.handleKey(msg)
			return m, nil
		}
	}
	return m, nil
}

// --- Save As key handler ---

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

// --- Rename key handler ---

func (m *Model) enterRename() {
	if m.filename == "" {
		return
	}
	m.mode = ModeRename
	m.rename.Input = filepath.Base(m.filename)
	m.rename.Error = ""
}

func (m *Model) handleRenameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.rename.Error = ""
		return m, nil

	case "enter":
		return m.confirmRename()

	case "backspace":
		if len(m.rename.Input) > 0 {
			m.rename.Input = m.rename.Input[:len(m.rename.Input)-1]
			m.rename.Error = ""
		}

	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.rename.Input += string(msg.Runes)
			m.rename.Error = ""
		}
	}
	return m, nil
}

func (m *Model) confirmRename() (tea.Model, tea.Cmd) {
	newName := strings.TrimSpace(m.rename.Input)
	oldPath := m.filename

	if newName == "" {
		m.rename.Error = "Name cannot be empty."
		return m, nil
	}
	if newName == filepath.Base(oldPath) {
		m.mode = ModeNormal
		m.rename.Error = ""
		return m, nil
	}
	if err := validateFileName(newName); err != nil {
		m.rename.Error = err.Error()
		return m, nil
	}

	if m.modified {
		m.save()
		if m.modified {
			m.rename.Error = "Error saving before rename."
			return m, nil
		}
	}

	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)

	if _, err := os.Stat(newPath); err == nil {
		m.rename.Error = fmt.Sprintf("File already exists: %s", newName)
		return m, nil
	}

	if err := fileio.Rename(oldPath, newPath); err != nil {
		m.rename.Error = fmt.Sprintf("Error renaming: %v", err)
		return m, nil
	}

	m.filename = newPath
	m.language = highlight.DetectLanguage(newPath)
	m.mode = ModeNormal
	m.rename.Error = ""
	m.addRecentFile(newPath)
	return m, nil
}

// --- Confirmation dialog ---

func (m *Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s", "S":
		m.save()
		if m.confirmAction == ConfirmQuit {
			return m, tea.Quit
		}
		m.closeRequested = true
		m.mode = ModeNormal
		return m, nil
	case "d", "D", "n", "N":
		if m.confirmAction == ConfirmQuit {
			return m, tea.Quit
		}
		m.closeRequested = true
		m.mode = ModeNormal
		return m, nil
	case "c", "C", "esc":
		m.mode = ModeNormal
		return m, nil
	}
	return m, nil
}

// --- Mouse handler ---

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

// --- Backspace and Delete handlers ---

func (m *Model) handleBackspace() (tea.Model, tea.Cmd) {
	m.clearSelection()
	if m.buf.GapPosition() == 0 {
		return m, nil
	}

	if len(m.extraCursors) > 0 {
		// Delete at all cursor positions (end to start)
		all := m.allCursors()
		for i := len(all) - 1; i >= 0; i-- {
			c := all[i]
			if c.GapPos == 0 {
				continue
			}
			r := m.buf.RuneAt(c.GapPos - 1)
			m.undoStack.Push(buffer.Operation{
				Type: buffer.OpInsert,
				Pos:  c.GapPos - 1,
				Text: string(r),
			})
			m.moveGapTo(c.GapPos)
			m.buf.Backspace()
		}
		// Restore primary cursor position
		m.cursor.Col--
		if m.cursor.Col < 0 {
			m.cursor.Col = 0
		}
		for i := range m.extraCursors {
			m.extraCursors[i].Col--
			if m.extraCursors[i].Col < 0 {
				m.extraCursors[i].Col = 0
			}
		}
		m.modified = true
	} else {
		r := m.buf.RuneAt(m.buf.GapPosition() - 1)
		m.undoStack.Push(buffer.Operation{
			Type: buffer.OpInsert,
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
	}
	return m, nil
}

func (m *Model) handleDelete() (tea.Model, tea.Cmd) {
	m.clearSelection()
	if m.buf.GapPosition() >= m.buf.Len() {
		return m, nil
	}

	if len(m.extraCursors) > 0 {
		all := m.allCursors()
		for i := len(all) - 1; i >= 0; i-- {
			c := all[i]
			if c.GapPos >= m.buf.Len() {
				continue
			}
			r := m.buf.RuneAt(c.GapPos)
			m.undoStack.Push(buffer.Operation{
				Type: buffer.OpInsert,
				Pos:  c.GapPos,
				Text: string(r),
			})
			m.moveGapTo(c.GapPos)
			m.buf.DeleteForward()
		}
		m.modified = true
	} else {
		r := m.buf.RuneAt(m.buf.GapPosition())
		m.undoStack.Push(buffer.Operation{
			Type: buffer.OpInsert,
			Pos:  m.buf.GapPosition(),
			Text: string(r),
		})
		if m.buf.DeleteForward() {
			m.modified = true
		}
	}
	return m, nil
}
