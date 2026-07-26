package editor

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

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
	case "file.open":
		m.enterFilePicker()
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
	case "file.rename":
		m.enterRename()
	case "view.go-line":
		m.enterGoToLine()
	case "view.toggle-auto-close":
		m.config.AutoCloseEnabled = !m.config.AutoCloseEnabled
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-auto-close")
		}
	case "view.toggle-vim-mode":
		m.config.VimMode = !m.config.VimMode
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-vim-mode")
		}
	case "view.next-theme":
		themes := []string{"dark", "light", "monokai", "dracula", "solarized-dark"}
		current := m.config.Theme
		for i, t := range themes {
			if t == current {
				m.config.Theme = themes[(i+1)%len(themes)]
				break
			}
		}
		m.highlighter.SetTheme(m.config.Theme)
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config next-theme")
		}
	case "view.toggle-word-wrap":
		m.config.WordWrap = !m.config.WordWrap
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-word-wrap")
		}
	case "file.toggle-format-on-save":
		m.config.FormatOnSave = !m.config.FormatOnSave
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-format-on-save")
		}
	case "file.toggle-auto-save":
		m.config.AutoSaveEnabled = !m.config.AutoSaveEnabled
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-auto-save")
		}
	}
}

// --- File operations ---

func (m *Model) save() {
	if m.filename == "" {
		m.enterSaveAs()
		return
	}
	if err := fileio.Save(m.filename, m.buf); err != nil {
		m.showError(fmt.Sprintf("Save error: %v", err))
		return
	}
	m.modified = false

	if m.config.FormatOnSave {
		before := m.buf.String()
		m.applyFormat()
		if m.buf.String() != before {
			if err := fileio.Save(m.filename, m.buf); err != nil {
				m.showError(fmt.Sprintf("Format save error: %v", err))
				m.modified = true
			}
		}
	}
}

func (m *Model) openFile(path string) {
	// Sanitize path to prevent traversal
	path = filepath.Clean(path)

	// Stop any existing LSP
	m.stopLSP()

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

	// Start LSP for supported languages
	m.startLSP()
}

// --- Undo/Redo ---

func (m *Model) undo() {
	ops, ok := m.undoStack.Undo()
	if !ok {
		return
	}
	for _, op := range ops {
		m.moveGapTo(op.Pos)
		switch op.Type {
		case "insert":
			m.buf.InsertString(op.Text)
		case "delete":
			for i := 0; i < utf8.RuneCountInString(op.Text); i++ {
				m.buf.DeleteForward()
			}
		}
	}
	m.modified = true
	m.cursor.Line, m.cursor.Col = m.buf.LineCol(m.buf.GapPosition())
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
	m.sendDidChange()
}

func (m *Model) redo() {
	ops, ok := m.undoStack.Redo()
	if !ok {
		return
	}
	for _, op := range ops {
		m.moveGapTo(op.Pos)
		switch op.Type {
		case "insert":
			for i := 0; i < utf8.RuneCountInString(op.Text); i++ {
				m.buf.DeleteForward()
			}
		case "delete":
			m.buf.InsertString(op.Text)
		}
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
	m.selStart = -1
	m.selEnd = -1
}

func (m *Model) cut() {
	if m.hasSelection() {
		m.copy()
		m.deleteSelection()
	} else {
		m.clipboard.Copy(m.currentLineText())
		lineStart := m.buf.LineStart(m.cursor.Line)
		lineEnd := m.buf.LineStart(m.cursor.Line + 1)
		var sb strings.Builder
		for i := lineStart; i < lineEnd && i < m.buf.Len(); i++ {
			sb.WriteRune(m.buf.RuneAt(i))
		}
		m.undoStack.Push(buffer.Operation{
			Type: "insert",
			Pos:  lineStart,
			Text: sb.String(),
		})
		m.moveGapTo(lineStart)
		for i := lineStart; i < lineEnd && m.buf.GapPosition() < m.buf.Len(); i++ {
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
		m.cursor.Col += utf8.RuneCountInString(text)
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
	sb.Grow(end - start)
	for i := start; i < end; i++ {
		sb.WriteRune(m.buf.RuneAt(i))
	}
	return sb.String()
}

func (m *Model) clearSelection() {
	m.selStart = -1
	m.selEnd = -1
	m.selecting = false
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

// --- Search ---

func (m *Model) doSearch() {
	if m.searchQuery == "" {
		return
	}
	m.lastSearch = m.searchQuery

	// Search in rune-space so match indices are logical rune indices
	// (RuneAt/navigation work in runes, not bytes).
	contentRunes := []rune(m.buf.String())
	queryRunes := []rune(strings.ToLower(m.searchQuery))

	m.searchMatches = nil
	for i := 0; i+len(queryRunes) <= len(contentRunes); i++ {
		match := true
		for j := 0; j < len(queryRunes); j++ {
			if unicode.ToLower(contentRunes[i+j]) != queryRunes[j] {
				match = false
				break
			}
		}
		if match {
			m.searchMatches = append(m.searchMatches, i)
		}
	}
	m.searchCurrent = 0
}

func (m *Model) doReplace() {
	if len(m.searchMatches) == 0 || m.replaceQuery == "" {
		return
	}

	pos := m.searchMatches[m.searchCurrent]
	m.undoStack.PushComposite([]buffer.Operation{
		{Type: "insert", Pos: pos, Text: m.searchQuery},
		{Type: "delete", Pos: pos, Text: m.replaceQuery},
	})
	m.moveGapTo(pos)
	for i := 0; i < utf8.RuneCountInString(m.searchQuery); i++ {
		m.buf.DeleteForward()
	}
	m.buf.InsertString(m.replaceQuery)
	m.modified = true

	m.doSearch()
}
