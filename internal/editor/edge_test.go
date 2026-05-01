package editor

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBinaryFileDetection(t *testing.T) {
	dir := t.TempDir()

	// Write a binary file (contains null bytes)
	binPath := filepath.Join(dir, "binary.bin")
	os.WriteFile(binPath, []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}, 0644)

	m, err := NewWithFile(binPath)
	// Should handle gracefully - not crash
	if err != nil {
		t.Logf("binary file opened with error (expected): %v", err)
	}
	if m != nil {
		t.Logf("loaded %d bytes from binary file", m.buf.Len())
	}
}

func TestEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	os.WriteFile(path, []byte{}, 0644)

	m, err := NewWithFile(path)
	if err != nil {
		t.Fatalf("empty file should load: %v", err)
	}
	if m.buf.String() != "" {
		t.Errorf("expected empty buffer, got %q", m.buf.String())
	}
}

func TestSingleLineLongFile(t *testing.T) {
	// Generate a single line with 10000 characters
	var line []rune
	for i := 0; i < 10000; i++ {
		line = append(line, 'x')
	}

	m := New()
	m.mode = ModeNormal
	for _, r := range line {
		m.buf.Insert(r)
	}

	if m.buf.LineCount() != 1 {
		t.Errorf("expected 1 line, got %d", m.buf.LineCount())
	}

	// Move cursor around in long line
	m.cursor.SetPos(0, 5000)
	m.syncGapToCursor()
	if m.cursor.Col != 5000 {
		t.Errorf("expected col 5000, got %d", m.cursor.Col)
	}

	// Test undo after many inserts
	m.undoStack.Undo()
	// Buffer should not be empty (only 1 undo)
	if m.buf.Len() < 1000 {
		t.Error("buffer should still be large after single undo")
	}
}

func TestMultiCursorSamePosition(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	for _, r := range "test" {
		m.buf.Insert(r)
	}
	m.cursor.SetPos(0, 0)
	m.syncGapToCursor()

	// Add occurrence when cursor is on 't'
	m.addNextOccurrence()
	// Second add should not duplicate
	initial := len(m.extraCursors)
	m.addNextOccurrence()
	if len(m.extraCursors) != initial {
		t.Errorf("should not add duplicate cursor at same position")
	}
}

func TestUndoRedoToEmpty(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// Use insertText which pushes to undo stack
	m.insertText("a")
	m.insertText("b")
	m.modified = true

	// Undo twice
	m.undo()
	m.undo()

	if m.buf.String() != "" {
		t.Errorf("expected empty after 2 undos, got %q", m.buf.String())
	}

	// Redo twice
	m.redo()
	m.redo()
	if m.buf.String() != "ab" {
		t.Errorf("expected 'ab' after 2 redos, got %q", m.buf.String())
	}
}

func TestHandleKeyWithVimMode(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.config.VimMode = true
	m.vimState = newVimState()

	// Insert 'a' in vim normal mode should do nothing (not i)
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if m.buf.String() != "" {
		t.Errorf("expected empty buffer in vim normal mode, got %q", m.buf.String())
	}

	// 'i' should enter insert mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if m.vimState.Mode != VimInsert {
		t.Errorf("expected VimInsert after 'i', got %v", m.vimState.Mode)
	}
}
