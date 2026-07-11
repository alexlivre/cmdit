package editor

import (
	tea "github.com/charmbracelet/bubbletea"
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

	case "f9":
		if m.mode == ModeNormal {
			m.executeAction("file.toggle-auto-save")
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
		m.selection.Start = 0
		m.selection.End = m.buf.Len()
		return m, nil

	// Search/Replace
	case "ctrl+f":
		m.mode = ModeSearch
		m.search.Query = ""
		m.search.Matches = nil
		m.search.Current = 0
		return m, nil

	case "ctrl+h":
		m.mode = ModeReplace
		m.search.Query = ""
		m.search.Replace = ""
		m.search.Matches = nil
		m.search.Current = 0
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
