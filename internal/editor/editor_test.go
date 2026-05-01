package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewEditor(t *testing.T) {
	m := New()
	if m.buf == nil {
		t.Error("expected non-nil buffer")
	}
	if m.cursor == nil {
		t.Error("expected non-nil cursor")
	}
	if m.undoStack == nil {
		t.Error("expected non-nil undo stack")
	}
	if m.clipboard == nil {
		t.Error("expected non-nil clipboard")
	}
	if m.mode != ModeWelcome {
		t.Errorf("expected ModeWelcome, got %v", m.mode)
	}
}

func TestInsertText(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	if m.buf.String() != "hi" {
		t.Errorf("expected 'hi', got %q", m.buf.String())
	}
	if !m.modified {
		t.Error("expected modified to be true after insert")
	}
}

func TestBackspace(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})

	if m.buf.String() != "a" {
		t.Errorf("expected 'a', got %q", m.buf.String())
	}
}

func TestBackspaceAtStart(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.buf.String() != "" {
		t.Errorf("expected empty, got %q", m.buf.String())
	}
}

func TestEnterKey(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if m.buf.String() != "a\nb" {
		t.Errorf("expected 'a\\nb', got %q", m.buf.String())
	}
	if m.cursor.Line != 1 {
		t.Errorf("expected cursor line 1, got %d", m.cursor.Line)
	}
}

func TestUndo(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlZ})
	if m.buf.String() != "h" {
		t.Errorf("expected 'h' after undo, got %q", m.buf.String())
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlZ})
	if m.buf.String() != "" {
		t.Errorf("expected empty after 2 undos, got %q", m.buf.String())
	}
}

func TestRedo(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlZ})

	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlY})
	if m.buf.String() != "a" {
		t.Errorf("expected 'a' after redo, got %q", m.buf.String())
	}
}

func TestCopyPaste(t *testing.T) {
	m := New()
	// Type "hello"
	for _, r := range "hello" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Select all
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	// Copy
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Paste
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlV})

	if strings.Count(m.buf.String(), "hello") != 2 {
		t.Errorf("expected 'hellohello', got %q", m.buf.String())
	}
}

func TestCopyLine(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	// Copy current line (no selection copies line)
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.clipboard.HasText() {
		t.Error("expected clipboard to have text")
	}
}

func TestQuitUnsavedShowsConfirm(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlQ})

	if m.mode != ModeConfirm {
		t.Errorf("expected ModeConfirm, got %v", m.mode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestQuitSavedNoConfirm(t *testing.T) {
	m := New()
	// Switch to normal mode first
	m.mode = ModeNormal
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlQ})

	if m.mode != ModeNormal {
		t.Error("expected no confirm for unmodified buffer")
	}
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

func TestConfirmDiscard(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlQ})

	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Error("expected quit cmd after discard")
	}
}

func TestConfirmCancel(t *testing.T) {
	m := New()
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlQ})

	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if m.mode != ModeNormal {
		t.Errorf("expected ModeNormal after cancel, got %v", m.mode)
	}
}

func TestSearchMode(t *testing.T) {
	m := New()
	for _, r := range "hello world" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Enter search mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlF})
	if m.mode != ModeSearch {
		t.Errorf("expected ModeSearch, got %v", m.mode)
	}
}

func TestSearchExecute(t *testing.T) {
	m := New()
	for _, r := range "hello world hello" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Enter search mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlF})
	// Type search query (in search mode)
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	// Press Enter to search
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if len(m.searchMatches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(m.searchMatches))
	}
}

// --- Rename Tests ---

func TestRenameNoFile(t *testing.T) {
	m := New()
	// F2 on a new buffer (no file) should not enter rename mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyF2})

	if m.mode == ModeRename {
		t.Error("expected not to enter ModeRename when no file is open")
	}
	if m.mode != ModeWelcome {
		t.Errorf("expected ModeWelcome, got %v", m.mode)
	}
}

func TestRenameEnterAndCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	m := New()
	m.filename = path
	m.mode = ModeNormal

	// F2 enters rename mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyF2})
	if m.mode != ModeRename {
		t.Errorf("expected ModeRename, got %v", m.mode)
	}
	if m.renameInput != "test.txt" {
		t.Errorf("expected renameInput 'test.txt', got %q", m.renameInput)
	}

	// Esc cancels and returns to normal
	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if m.mode != ModeNormal {
		t.Errorf("expected ModeNormal after cancel, got %v", m.mode)
	}
	if m.filename != path {
		t.Errorf("filename should not change after cancel")
	}
}

func TestRenameSuccess(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")
	newPath := filepath.Join(dir, "new.txt")

	if err := os.WriteFile(oldPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	m := New()
	m.filename = oldPath
	m.mode = ModeNormal

	// Enter rename mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyF2})
	if m.renameInput != "old.txt" {
		t.Errorf("expected renameInput 'old.txt', got %q", m.renameInput)
	}

	// Clear and type new name
	for i := 0; i < len("old.txt"); i++ {
		m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	for _, r := range "new.txt" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press Enter to confirm
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if m.mode != ModeNormal {
		t.Errorf("expected ModeNormal after rename, got %v", m.mode)
	}
	if m.filename != newPath {
		t.Errorf("expected filename %q, got %q", newPath, m.filename)
	}

	// Old file should not exist
	if _, err := os.Stat(oldPath); err == nil {
		t.Error("old file still exists after rename")
	}
	// New file should exist
	if _, err := os.Stat(newPath); err != nil {
		t.Error("new file does not exist after rename")
	}
}

func TestRenameEmptyName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hi"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	m := New()
	m.filename = path
	m.mode = ModeNormal

	// Enter rename mode
	m.handleKey(tea.KeyMsg{Type: tea.KeyF2})
	// Clear input with backspace
	for i := 0; i < len("test.txt"); i++ {
		m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	// Press Enter with empty name
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if m.renameError == "" {
		t.Error("expected error for empty name")
	}
	if m.mode != ModeRename {
		t.Errorf("expected to stay in ModeRename, got %v", m.mode)
	}
	if m.filename != path {
		t.Errorf("filename should not change on error")
	}
}

func TestRenameUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hi"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	m := New()
	m.filename = path
	m.mode = ModeNormal

	// Enter rename mode and press Enter immediately (same name)
	m.handleKey(tea.KeyMsg{Type: tea.KeyF2})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if m.mode != ModeNormal {
		t.Errorf("expected ModeNormal, got %v", m.mode)
	}
	if m.filename != path {
		t.Errorf("filename should not change")
	}
}

func TestRenameInvalidChars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hi"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	for _, invalid := range []string{"a:b", "a<b", "a>b", "a|b", "a?b", "a*b", "a\"b"} {
		m := New()
		m.filename = path
		m.mode = ModeNormal

		m.handleKey(tea.KeyMsg{Type: tea.KeyF2})
		// Clear and type invalid name
		for i := 0; i < len("test.txt"); i++ {
			m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
		}
		for _, r := range invalid {
			m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
		m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

		if m.renameError == "" {
			t.Errorf("expected error for invalid name %q", invalid)
		}
		if m.mode != ModeRename {
			t.Errorf("expected to stay in ModeRename for %q, got %v", invalid, m.mode)
		}
	}
}

func TestRenameExistingFile(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")
	existingPath := filepath.Join(dir, "existing.txt")

	if err := os.WriteFile(oldPath, []byte("old"), 0644); err != nil {
		t.Fatalf("failed to create old file: %v", err)
	}
	if err := os.WriteFile(existingPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	m := New()
	m.filename = oldPath
	m.mode = ModeNormal

	m.handleKey(tea.KeyMsg{Type: tea.KeyF2})
	// Clear and type existing name
	for i := 0; i < len("old.txt"); i++ {
		m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	for _, r := range "existing.txt" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if m.renameError == "" {
		t.Error("expected error when destination exists")
	}
	if m.mode != ModeRename {
		t.Errorf("expected to stay in ModeRename, got %v", m.mode)
	}
	if m.filename != oldPath {
		t.Errorf("filename should not change when dest exists")
	}
}

func TestRenameUnsavedChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	m, err := NewWithFile(path)
	if err != nil {
		t.Fatalf("NewWithFile failed: %v", err)
	}
	if m.filename != path {
		t.Fatalf("expected filename %q, got %q", path, m.filename)
	}

	// Type additional content
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if !m.modified {
		t.Fatal("expected modified to be true")
	}
	if m.buf.String() != "originalx" {
		t.Fatalf("expected buffer 'originalx', got %q", m.buf.String())
	}

	newPath := filepath.Join(dir, "renamed.txt")
	m.handleKey(tea.KeyMsg{Type: tea.KeyF2})
	// Clear and type new name
	for i := 0; i < len("test.txt"); i++ {
		m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	for _, r := range "renamed.txt" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if m.mode != ModeNormal {
		t.Errorf("expected ModeNormal after rename, got %v", m.mode)
	}
	if m.filename != newPath {
		t.Errorf("expected filename %q, got %q", newPath, m.filename)
	}

	// Verify new file contains both original + added content
	data, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("new file not found: %v", err)
	}
	if string(data) != "originalx" {
		t.Errorf("expected 'originalx', got %q", string(data))
	}
}

func TestPaletteActionRename(t *testing.T) {
	m := New()
	m.filename = "test.txt"
	m.mode = ModeNormal

	m.executeAction("file.rename")

	if m.mode != ModeRename {
		t.Errorf("expected ModeRename via palette, got %v", m.mode)
	}
	if m.renameInput != "test.txt" {
		t.Errorf("expected renameInput 'test.txt', got %q", m.renameInput)
	}
}

// --- Phase 10 coverage tests ---

func TestHandleDelete(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf.Insert('a')
	m.buf.Insert('b')
	m.buf.Insert('c')
	m.cursor.SetPos(0, 3)
	m.syncGapToCursor()

	// Delete last character
	m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.buf.String() != "ab" {
		t.Errorf("expected 'ab', got %q", m.buf.String())
	}

	// Delete at beginning (no op)
	m.cursor.SetPos(0, 0)
	m.syncGapToCursor()
	m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.buf.String() != "ab" {
		t.Errorf("expected 'ab' unchanged, got %q", m.buf.String())
	}

	// Delete key
	m.handleKey(tea.KeyMsg{Type: tea.KeyDelete})
	if m.buf.String() != "b" {
		t.Errorf("expected 'b', got %q", m.buf.String())
	}
}

func TestMoveCursorWordLeftRight(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	for _, r := range "hello world" {
		m.buf.Insert(r)
	}
	m.cursor.SetPos(0, 11)
	m.syncGapToCursor()

	// Word left from end
	m.moveCursorWordLeft()
	if m.cursor.Col != 6 {
		t.Errorf("expected col 6 (start of 'world'), got %d", m.cursor.Col)
	}

	// Word left again
	m.moveCursorWordLeft()
	if m.cursor.Col != 0 {
		t.Errorf("expected col 0, got %d", m.cursor.Col)
	}

	// Word right
	m.moveCursorWordRight()
	if m.cursor.Col != 6 {
		t.Errorf("expected col 6, got %d", m.cursor.Col)
	}
}

func TestWordAtCursorAndMultiCursor(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	for _, r := range "test word test" {
		m.buf.Insert(r)
	}
	m.cursor.SetPos(0, 0)
	m.syncGapToCursor()

	// wordAtCursor at start
	word := m.wordAtCursor()
	if word != "test" {
		t.Errorf("expected 'test', got '%s'", word)
	}

	// addNextOccurrence
	m.addNextOccurrence()
	if len(m.extraCursors) != 1 {
		t.Errorf("expected 1 extra cursor, got %d", len(m.extraCursors))
	}

	// clear extra cursors
	m.clearExtraCursors()
	if len(m.extraCursors) != 0 {
		t.Errorf("expected 0 extra cursors, got %d", len(m.extraCursors))
	}
}

func TestCutAndDeleteSelection(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// deleteSelection with gap positioned at start of selection
	for _, r := range "delete this" {
		m.buf.Insert(r)
	}
	m.selStart = 0
	m.selEnd = 6
	m.cursor.SetPos(0, 6)
	m.moveGapTo(0)
	m.deleteSelection()
	if m.buf.String() != " this" {
		t.Errorf("expected ' this' after deleteSelection, got %q", m.buf.String())
	}

	// cut (without selection falls back to cut line)
	m = New()
	for _, r := range "cut me" {
		m.buf.Insert(r)
	}
	m.cursor.SetPos(0, 0)
	m.cut()
	if m.buf.String() != "" {
		t.Errorf("expected empty after cut line, got %q", m.buf.String())
	}
}

func TestDoReplace(t *testing.T) {
	m := New()
	m.mode = ModeSearch
	for _, r := range "foo bar foo" {
		m.buf.Insert(r)
	}

	m.searchQuery = "foo"
	m.replaceQuery = "baz"
	m.doSearch()

	// Verify search found matches
	if len(m.searchMatches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(m.searchMatches))
	}

	// Replace first
	m.searchCurrent = 0
	m.doReplace()
	if m.buf.String() != "baz bar foo" {
		t.Errorf("expected 'baz bar foo', got %q", m.buf.String())
	}
}

func TestHandleMouse(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf.Insert('a')
	m.buf.Insert('\n')
	m.buf.Insert('b')
	m.cursor.SetPos(0, 0)

	// Mouse click on col 0 of line 1 (1-indexed since line 0 = status bar offset)
	m.handleMouse(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 0, Y: 0})
	// Y=0 maps to line 0 of content
	if m.cursor.Line != 0 || m.cursor.Col != 0 {
		t.Errorf("expected (0,0), got (%d,%d)", m.cursor.Line, m.cursor.Col)
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save_test.txt")

	m := New()
	m.filename = path
	m.mode = ModeNormal
	for _, r := range "save content" {
		m.buf.Insert(r)
	}
	m.modified = true

	m.save()
	if m.modified {
		t.Error("expected modified=false after save")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not saved: %v", err)
	}
	if string(data) != "save content" {
		t.Errorf("expected 'save content', got %q", string(data))
	}
}
