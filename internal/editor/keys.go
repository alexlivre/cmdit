package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexb/cmdit/internal/buffer"
	"github.com/alexb/cmdit/internal/command"
	"github.com/alexb/cmdit/internal/fileio"
	"github.com/alexb/cmdit/internal/highlight"
)

// --- Main key handler ---

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Vim mode: intercept ALL keys when vim mode is on
	if m.config.VimMode {
		// In insert mode, handle Esc specially; pass rest to normal flow
		if m.vimState.Mode == VimInsert {
			switch msg.String() {
			case "esc":
				m.vimState.Mode = VimNormal
				return m, nil
			}
			// Fall through to normal key handling for insert mode typing.
			// Vim dispatch is skipped to avoid double-handling.
		} else {
			// Normal / Visual / Command: dispatch to vim handler
			return m, m.dispatchVimKey(msg)
		}
	}

	// Custom keybindings: check config overrides BEFORE normal dispatch
	if len(m.config.Keybindings) > 0 {
		keyStr := msg.String()
		if actionID, ok := m.config.Keybindings[keyStr]; ok {
			m.executeAction(actionID)
			return m, nil
		}
	}

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
	if m.mode == ModeRename {
		return m.handleRenameKey(msg)
	}
	if m.mode == ModeSearch || m.mode == ModeReplace {
		return m.handleSearchKey(msg)
	}
	if m.mode == ModeConfirm {
		return m.handleConfirmKey(msg)
	}

	// Normal mode key dispatch
	switch msg.String() {
	case "ctrl+p":
		m.enterPalette()
		return m, nil

	case "ctrl+o":
		m.enterFilePicker()
		return m, nil

	case "f2":
		m.enterRename()
		return m, nil
	case "f3":
		m.enterSaveAs()
		return m, nil

	case "f4":
		if m.mode == ModeNormal {
			m.executeAction("view.toggle-auto-close")
		}
		return m, nil

	case "f5":
		if m.mode == ModeNormal {
			m.executeAction("view.toggle-vim-mode")
		}
		return m, nil

	case "f6":
		if m.mode == ModeNormal {
			m.executeAction("view.next-theme")
		}
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

	// Multi-cursor
	case "ctrl+d":
		m.addNextOccurrence()
		return m, nil

	// Escape - clear multi-cursors, then selection
	case "esc":
		if len(m.extraCursors) > 0 {
			m.clearExtraCursors()
			return m, nil
		}
		m.clearSelection()
		return m, nil

	case "backspace":
		return m.handleBackspace()

	case "delete":
		return m.handleDelete()

	case "enter":
		m.clearSelection()
		m.insertTextAtAllCursors("\n")
		// Reset cursor position
		m.cursor.Line++
		m.cursor.Col = 0
		// Update extra cursors
		for i := range m.extraCursors {
			m.extraCursors[i].Line++
			m.extraCursors[i].Col = 0
		}
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		return m, nil

	case "tab":
		m.clearSelection()
		m.insertTextAtAllCursors("    ")
		m.cursor.Col += 4
		for i := range m.extraCursors {
			m.extraCursors[i].Col += 4
		}
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

	case "alt+z":
		if m.mode == ModeNormal {
			m.executeAction("view.toggle-word-wrap")
		}
		return m, nil

	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.clearSelection()
			ch := msg.Runes[0]

			// Auto-close brackets/quotes (before normal insert)
			if m.config.AutoCloseEnabled {
				if m.handleSmartSkip(ch) {
					return m, nil
				}
				if _, ok := shouldAutoClose(ch); ok {
					return m.handleAutoClose(ch)
				}
			}

			// Normal insert
			text := string(msg.Runes)
			m.insertTextAtAllCursors(text)
			m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		}
		return m, nil
	}
}

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
				Type: "insert",
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
				Type: "insert",
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
			Type: "insert",
			Pos:  m.buf.GapPosition(),
			Text: string(r),
		})
		if m.buf.DeleteForward() {
			m.modified = true
		}
	}
	return m, nil
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
			// For now, tab stays in replace mode
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
			if m.mode == ModeReplace && msg.Runes[0] >= 32 {
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

// --- Palette key handler ---

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

// --- File Picker key handler ---

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
	m.renameInput = filepath.Base(m.filename)
	m.renameError = ""
}

func (m *Model) handleRenameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.renameError = ""
		return m, nil

	case "enter":
		return m.confirmRename()

	case "backspace":
		if len(m.renameInput) > 0 {
			m.renameInput = m.renameInput[:len(m.renameInput)-1]
			m.renameError = ""
		}

	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.renameInput += string(msg.Runes)
			m.renameError = ""
		}
	}
	return m, nil
}

func (m *Model) confirmRename() (tea.Model, tea.Cmd) {
	newName := strings.TrimSpace(m.renameInput)
	oldPath := m.filename

	if newName == "" {
		m.renameError = "Nome nao pode estar vazio."
		return m, nil
	}
	if newName == filepath.Base(oldPath) {
		m.mode = ModeNormal
		m.renameError = ""
		return m, nil
	}
	if err := validateFileName(newName); err != nil {
		m.renameError = err.Error()
		return m, nil
	}

	if m.modified {
		m.save()
		if m.modified {
			m.renameError = "Erro ao salvar antes de renomear."
			return m, nil
		}
	}

	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)

	if _, err := os.Stat(newPath); err == nil {
		m.renameError = fmt.Sprintf("Arquivo ja existe: %s", newName)
		return m, nil
	}

	if err := fileio.Rename(oldPath, newPath); err != nil {
		m.renameError = fmt.Sprintf("Erro ao renomear: %v", err)
		return m, nil
	}

	m.filename = newPath
	m.language = highlight.DetectLanguage(newPath)
	m.mode = ModeNormal
	m.renameError = ""
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
