package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Helper: keyMsg creates a tea.KeyMsg from a rune.
func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// TestVimModeToggle verifies that F5 toggles VimMode on and off.
func TestVimModeToggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// No vim mode initially
	if m.config.VimMode {
		t.Error("expected VimMode to be false initially")
	}

	// F5 toggles on
	m.handleKey(tea.KeyMsg{Type: tea.KeyF5})
	if !m.config.VimMode {
		t.Error("expected VimMode to be true after F5")
	}

	// F5 toggles off
	m.handleKey(tea.KeyMsg{Type: tea.KeyF5})
	if m.config.VimMode {
		t.Error("expected VimMode to be false after second F5")
	}
}

// TestVimModeNavigation tests h/j/k/l movement in normal mode.
func TestVimModeNavigation(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Enable vim mode
	m.config.VimMode = true
	m.vimState.Mode = VimNormal

	// Enter insert mode with "i", type "line1", Enter, "line2"
	m.handleKey(keyRune('i'))
	if m.vimState.Mode != VimInsert {
		t.Fatal("expected VimInsert after 'i'")
	}
	for _, r := range "line1" {
		m.handleKey(keyRune(r))
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	for _, r := range "line2" {
		m.handleKey(keyRune(r))
	}
	// Esc to normal mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if m.vimState.Mode != VimNormal {
		t.Fatal("expected VimNormal after Esc")
	}

	// Cursor ends at (1,0) after typing + Enter sets col to 0.
	// Use $ to go to end of current line.
	m.handleKey(keyRune('$'))
	if m.cursor.Line != 1 || m.cursor.Col != 5 {
		t.Fatalf("expected cursor at (1,5) after $, got (%d,%d)", m.cursor.Line, m.cursor.Col)
	}

	// Move up with "k"
	m.handleKey(keyRune('k'))
	if m.cursor.Line != 0 {
		t.Errorf("expected line 0 after 'k', got %d", m.cursor.Line)
	}
	// Col should be clamped to line 0 length (5)
	if m.cursor.Col != 5 {
		t.Errorf("expected col 5 after 'k', got %d", m.cursor.Col)
	}

	// Move down with "j"
	m.handleKey(keyRune('j'))
	if m.cursor.Line != 1 {
		t.Errorf("expected line 1 after 'j', got %d", m.cursor.Line)
	}
	if m.cursor.Col != 5 {
		t.Errorf("expected col 5 after 'j', got %d", m.cursor.Col)
	}

	// Move left with "h"
	m.handleKey(keyRune('h'))
	if m.cursor.Col != 4 {
		t.Errorf("expected col 4 after 'h', got %d", m.cursor.Col)
	}

	// Move right with "l"
	m.handleKey(keyRune('l'))
	if m.cursor.Col != 5 {
		t.Errorf("expected col 5 after 'l', got %d", m.cursor.Col)
	}
}

// TestVimModeInsert tests i enters insert mode and Esc returns to normal.
func TestVimModeInsert(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.config.VimMode = true
	m.vimState.Mode = VimNormal

	if m.vimState.Mode != VimNormal {
		t.Fatal("expected starting mode VimNormal")
	}

	// 'i' enters insert mode
	m.handleKey(keyRune('i'))
	if m.vimState.Mode != VimInsert {
		t.Errorf("expected VimInsert after 'i', got %v", m.vimState.Mode)
	}

	// Esc returns to normal mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if m.vimState.Mode != VimNormal {
		t.Errorf("expected VimNormal after Esc, got %v", m.vimState.Mode)
	}

	// Typing while in insert mode should produce text
	m.handleKey(keyRune('i')) // enter insert again
	m.handleKey(keyRune('h'))
	m.handleKey(keyRune('i'))

	// Text should be "hi" in buffer
	if m.buf.String() != "hi" {
		t.Errorf("expected 'hi', got %q", m.buf.String())
	}

	// Esc back to normal
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if m.vimState.Mode != VimNormal {
		t.Errorf("expected VimNormal after Esc, got %v", m.vimState.Mode)
	}
}

// TestVimModeDeleteChar tests 'x' in normal mode deletes the character under cursor.
func TestVimModeDeleteChar(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.config.VimMode = true
	m.vimState.Mode = VimNormal

	// Type "abc" in insert mode
	m.handleKey(keyRune('i'))
	for _, r := range "abc" {
		m.handleKey(keyRune(r))
	}
	// After typing "abc" in insert mode, cursor sits at col 3 (after 'c').
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	// Vim Esc retracts cursor one column: now over 'c' at col 2.
	if m.cursor.Col != 2 {
		t.Fatalf("expected cursor col 2 after Esc (Vim retract), got %d", m.cursor.Col)
	}

	// Move right once to be at end (col 3); from col 2, 'l' goes to col 3.
	m.handleKey(keyRune('l'))
	if m.cursor.Col != 3 {
		t.Fatalf("expected cursor col 3 after 'l', got %d", m.cursor.Col)
	}

	// Move back two to be over 'b' (col 1), then 'x' deletes 'b'.
	m.handleKey(keyRune('h'))
	m.handleKey(keyRune('h'))
	if m.cursor.Col != 1 {
		t.Fatalf("expected cursor col 1 over 'b', got %d", m.cursor.Col)
	}

	// Press 'x' to delete character under cursor ('b')
	m.handleKey(keyRune('x'))

	if m.buf.String() != "ac" {
		t.Errorf("expected 'ac' after delete, got %q", m.buf.String())
	}
}

// TestVimModeCountPrefix tests numeric count prefix for repeating commands.
func TestVimModeCountPrefix(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.config.VimMode = true
	m.vimState.Mode = VimNormal

	// Type "abcde" in insert mode
	m.handleKey(keyRune('i'))
	for _, r := range "abcde" {
		m.handleKey(keyRune(r))
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})

	// Go to end of line with $
	m.handleKey(keyRune('$'))
	if m.cursor.Col != 5 {
		t.Fatalf("expected col 5 after $, got %d", m.cursor.Col)
	}

	// Use "3h" to move left 3 times (col 5 -> col 2)
	m.handleKey(keyRune('3'))
	m.handleKey(keyRune('h'))
	if m.cursor.Col != 2 {
		t.Errorf("expected col 2 after '3h', got %d", m.cursor.Col)
	}

	// Use "2x" to delete 2 characters (at col 2, deletes 'c' and 'd')
	m.handleKey(keyRune('2'))
	m.handleKey(keyRune('x'))
	if m.buf.String() != "abe" {
		t.Errorf("expected 'abe' after '2x', got %q", m.buf.String())
	}
}

// TestVimModeCommandSave tests :w command saves the file.
func TestVimModeCommandSave(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Create a temp file
	dir := t.TempDir()
	// Make the model believe it has a file
	m.filename = dir + "/test.txt"

	m.config.VimMode = true
	m.vimState.Mode = VimNormal

	// Type "hello" in insert mode
	m.handleKey(keyRune('i'))
	for _, r := range "hello" {
		m.handleKey(keyRune(r))
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})

	if !m.modified {
		t.Fatal("expected buffer to be modified")
	}

	// Enter command mode with ':'
	m.handleKey(keyRune(':'))
	if m.vimState.Mode != VimCommand {
		t.Fatalf("expected VimCommand, got %v", m.vimState.Mode)
	}
	if m.vimState.CommandBuf != "" {
		t.Fatalf("expected empty command buf, got %q", m.vimState.CommandBuf)
	}

	// Type 'w'
	m.handleKey(keyRune('w'))
	if m.vimState.CommandBuf != "w" {
		t.Fatalf("expected command buf 'w', got %q", m.vimState.CommandBuf)
	}

	// Press Enter to execute :w
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Should be back in normal mode
	if m.vimState.Mode != VimNormal {
		t.Errorf("expected VimNormal after :w, got %v", m.vimState.Mode)
	}

	// File should be saved (modified = false)
	if m.modified {
		t.Error("expected buffer to be saved (modified = false)")
	}
}

// TestVimModeUndoDeleteChar tests that undo after multi-char delete (e.g. 2x)
// restores all deleted characters in a single undo operation.
func TestVimModeUndoDeleteChar(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	m.config.VimMode = true
	m.vimState.Mode = VimNormal

	// Type "abcdef" in insert mode
	m.handleKey(keyRune('i'))
	for _, r := range "abcdef" {
		m.handleKey(keyRune(r))
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})

	// After Esc, cursor sits over 'f' (col 5). Go to col 0 then over 'b' (col 1).
	m.handleKey(keyRune('0')) // col 0
	m.handleKey(keyRune('l')) // col 1
	if m.cursor.Col != 1 {
		t.Fatalf("expected cursor col 1, got %d", m.cursor.Col)
	}

	// Use 3x to delete 3 characters ('b', 'c', 'd')
	m.handleKey(keyRune('3'))
	m.handleKey(keyRune('x'))
	if m.buf.String() != "aef" {
		t.Fatalf("expected 'aef' after 3x, got %q", m.buf.String())
	}

	// Undo: should restore all 3 chars at once
	m.handleKey(keyRune('u'))
	if m.buf.String() != "abcdef" {
		t.Errorf("expected 'abcdef' after undo, got %q", m.buf.String())
	}

	// Verify only one undo step was pushed (undo again should undo the initial typing)
	// But we're in vim mode still, so typing 'u' again would affect the insert mode text
	// Just verify the first undo restored everything.
}
