package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

// autoClosePairs maps opening characters to their closing counterparts.
var autoClosePairs = map[rune]rune{
	'(': ')',
	'[': ']',
	'{': '}',
	'"': '"',
	'\'': '\'',
	'`': '`',
}

// shouldAutoClose returns (closingRune, true) if the character should trigger auto-close.
func shouldAutoClose(r rune) (rune, bool) {
	closer, ok := autoClosePairs[r]
	return closer, ok
}

// handleAutoClose inserts a pair of characters and positions the cursor between them.
func (m *Model) handleAutoClose(openChar rune) (tea.Model, tea.Cmd) {
	closer, ok := shouldAutoClose(openChar)
	if !ok {
		return m, nil
	}

	// Record position before insert for smart-skip
	cursorPos := m.buf.GapPosition()

	// Insert both characters
	m.buf.Insert(openChar)
	m.buf.Insert(closer)

	// Move cursor back to sit between the pair
	m.buf.MoveGapLeft()
	m.cursor.Col++

	// Mark position as auto-closed
	m.markAutoClosed(cursorPos)

	m.modified = true
	m.sendDidChange()
	return m, nil
}

// markAutoClosed records that position 'pos' has an auto-closed pair.
func (m *Model) markAutoClosed(pos int) {
	if m.autoClosed == nil {
		m.autoClosed = make(map[int]bool)
	}
	m.autoClosed[pos] = true
}

// isAutoClosed checks if the character at position 'pos' was auto-closed.
func (m *Model) isAutoClosed(pos int) bool {
	if m.autoClosed == nil {
		return false
	}
	return m.autoClosed[pos]
}

// clearAutoClosedAt removes marker when pair is broken.
func (m *Model) clearAutoClosedAt(pos int) {
	if m.autoClosed != nil {
		delete(m.autoClosed, pos)
	}
}

// handleSmartSkip: if typed char matches next char and current position was auto-closed, skip.
func (m *Model) handleSmartSkip(char rune) bool {
	cursorPos := m.buf.GapPosition()
	nextChar := m.buf.RuneAt(cursorPos)
	if nextChar == char {
		// The opening char is at (cursorPos-1)
		if m.isAutoClosed(cursorPos - 1) {
			m.buf.MoveGapRight()
			m.cursor.Col++
			return true
		}
	}
	return false
}
