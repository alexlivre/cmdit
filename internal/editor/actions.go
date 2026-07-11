package editor

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/alexb/cmdit/internal/buffer"
	"github.com/alexb/cmdit/internal/fileio"
	"github.com/alexb/cmdit/internal/highlight"
)

// --- Action execution ---

func (m *Model) executeAction(id string) {
	switch id {
	case "file.save":
		m.save()
	case "file.save-as":
		m.enterSaveAs()
	case "file.quit":
		if m.modified {
			m.mode = ModeConfirm
			m.confirmAction = ConfirmQuit
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
		m.selection.Start = 0
		m.selection.End = m.buf.Len()
	case "search.find":
		m.mode = ModeSearch
		m.search.Query = ""
		m.search.Matches = nil
	case "search.replace":
		m.mode = ModeReplace
		m.search.Query = ""
		m.search.Replace = ""
		m.search.Matches = nil
	case "file.rename":
		m.enterRename()
	case "view.toggle-auto-close":
		m.config.AutoCloseEnabled = !m.config.AutoCloseEnabled
		SaveConfig(m.config)
	case "view.toggle-vim-mode":
		m.config.VimMode = !m.config.VimMode
		SaveConfig(m.config)
	case "view.next-theme":
		// Cycle through themes
		themes := []string{"dark", "light", "monokai", "dracula", "solarized-dark"}
		current := m.config.Theme
		for i, t := range themes {
			if t == current {
				m.config.Theme = themes[(i+1)%len(themes)]
				break
			}
		}
		m.highlighter.SetTheme(m.config.Theme)
		SaveConfig(m.config)
	case "view.toggle-word-wrap":
		m.config.WordWrap = !m.config.WordWrap
		SaveConfig(m.config)
	case "file.toggle-format-on-save":
		m.config.FormatOnSave = !m.config.FormatOnSave
		SaveConfig(m.config)
	case "file.toggle-auto-save":
		m.config.AutoSaveEnabled = !m.config.AutoSaveEnabled
		SaveConfig(m.config)
	}
}

// --- File operations ---

func (m *Model) save() {
	if m.filename == "" {
		m.enterSaveAs()
		return
	}
	if err := fileio.Save(m.filename, m.buf); err != nil {
		m.showError(fmt.Sprintf("Save failed: %v", err))
		return
	}
	m.modified = false

	if m.config.FormatOnSave {
		m.applyFormat()
		if err := fileio.Save(m.filename, m.buf); err != nil {
			m.showError(fmt.Sprintf("Save after format failed: %v", err))
		}
	}
}

func (m *Model) openFile(path string) {
	path = filepath.Clean(path)

	m.stopLSP()

	b, err := fileio.Load(path)
	if err != nil {
		m.showError(fmt.Sprintf("Failed to open file: %v", err))
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

	// Start LSP for supported languages
	m.startLSP()
}

// --- Undo/Redo ---

func (m *Model) undo() {
	op, ok := m.undoStack.Undo()
	if !ok {
		return
	}

	m.moveGapTo(op.Pos)
	switch op.Type {
	case buffer.OpInsert:
		m.buf.InsertString(op.Text)
	case buffer.OpDelete:
		for i := 0; i < len(op.Text); i++ {
			m.buf.DeleteForward()
		}
	}
	m.modified = true
	m.cursor.Line, m.cursor.Col = m.buf.LineCol(m.buf.GapPosition())
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
	m.sendDidChange()
}

func (m *Model) redo() {
	op, ok := m.undoStack.Redo()
	if !ok {
		return
	}

	m.moveGapTo(op.Pos)
	switch op.Type {
	case buffer.OpInsert:
		for i := 0; i < len(op.Text); i++ {
			m.buf.DeleteForward()
		}
	case buffer.OpDelete:
		m.buf.InsertString(op.Text)
	}
	m.modified = true
	m.cursor.Line, m.cursor.Col = m.buf.LineCol(m.buf.GapPosition())
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
	m.sendDidChange()
}

// --- Clipboard ---

func (m *Model) copy() {
	if m.hasSelection() {
		text := m.getSelectedText()
		m.clipboard.Copy(text)
	} else {
		text := m.currentLineText()
		m.clipboard.Copy(text)
	}
	m.selection.Start = -1
	m.selection.End = -1
}

func (m *Model) cut() {
	if m.hasSelection() {
		m.copy()
		m.deleteSelection()
	} else {
		m.clipboard.Copy(m.currentLineText())
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
			Type: buffer.OpDelete,
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
	return m.selection.Start >= 0 && m.selection.End >= 0 && m.selection.Start != m.selection.End
}

func (m *Model) getSelectedText() string {
	if !m.hasSelection() {
		return ""
	}
	start := m.selection.Start
	end := m.selection.End
	if start > end {
		start, end = end, start
	}
	var sb strings.Builder
	sb.Grow(end - start)
	for i := start; i < end; i++ {
		sb.WriteRune(m.buf.RuneAt(i))
	}
	return sb.String()
}

func (m *Model) clearSelection() {
	m.selection.Start = -1
	m.selection.End = -1
}

func (m *Model) deleteSelection() {
	if !m.hasSelection() {
		return
	}
	text := m.getSelectedText()
	m.undoStack.Push(buffer.Operation{
		Type: buffer.OpInsert,
		Pos:  m.selection.Start,
		Text: text,
	})

	start := m.selection.Start
	end := m.selection.End
	if start > end {
		start, end = end, start
	}

	m.moveGapTo(start)
	for i := 0; i < end-start; i++ {
		m.buf.DeleteForward()
	}
	m.selection.Start = -1
	m.selection.End = -1
	m.modified = true
}

// --- Search ---

func (m *Model) doSearch() {
	if m.search.Query == "" {
		m.search.Matches = nil
		return
	}
	m.search.Last = m.search.Query

	content := m.buf.String()
	contentRunes := []rune(content)
	queryRunes := []rune(strings.ToLower(m.search.Query))
	queryLen := len(queryRunes)

	m.search.Matches = nil
	m.search.Current = 0

	for i := 0; i <= len(contentRunes)-queryLen; i++ {
		match := true
		for j := 0; j < queryLen; j++ {
			if unicode.ToLower(contentRunes[i+j]) != queryRunes[j] {
				match = false
				break
			}
		}
		if match {
			byteOffset := len(string(contentRunes[:i]))
			m.search.Matches = append(m.search.Matches, byteOffset)
		}
	}
}

func (m *Model) doReplace() {
	if len(m.search.Matches) == 0 || m.search.Replace == "" {
		return
	}

	pos := m.search.Matches[m.search.Current]
	m.moveGapTo(pos)
	for i := 0; i < len(m.search.Query); i++ {
		m.buf.DeleteForward()
	}
	m.buf.InsertString(m.search.Replace)
	m.modified = true

	m.doSearch()
}
