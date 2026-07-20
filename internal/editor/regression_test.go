package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Regression tests for bugs found during systematic debugging.
// Each test reproduces a bug BEFORE the fix to confirm root cause.

// Bug A/C: paste() uses len(text) (bytes) to advance cursor Col, but Col is a
// rune count. Pasting multibyte text misaligns the cursor.
func TestPasteMultibyteCursorCol(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.cursor.SetPos(0, 0)
	// Load clipboard with multi-byte text. "áéí" = 3 runes, 6 bytes in UTF-8.
	m.clipboard.Copy("áéí")
	m.paste()
	wantRunes := 3
	if m.cursor.Col != wantRunes {
		t.Errorf("paste multibyte: Col = %d (bytes), want %d (runes)", m.cursor.Col, wantRunes)
	}
	if got := m.buf.String(); got != "áéí" {
		t.Errorf("paste multibyte: buf = %q, want %q", got, "áéí")
	}
}

// Bug G: insertText via default key branch never advances cursor.Col in the
// single-cursor path. TestInsertText does not check Col, so this is silent.
// Even worse for multibyte runes because the byte/rune distinction would
// also matter — here we use ASCII to isolate the regression.
func TestTypeSingleCursorColAdvancesOnInsert(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.cursor.SetPos(0, 0)
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if m.cursor.Col != 1 {
		t.Errorf("type 'a': Col = %d, want 1", m.cursor.Col)
	}
}

// Bug A2/C: paste() advances cursor.Col by len(text) (bytes) instead of rune count.
func TestTypeMultibyteRuneCursorCol(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.cursor.SetPos(0, 0)
	// Type a single rune "á" via the rune path. The default branch converts
	// msg.Runes to a string and inserts. Cursor Col is updated by the buffer
	// path? Let's assert: typing 1 rune should yield Col=1.
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'á'}})
	if m.cursor.Col != 1 {
		t.Errorf("type multibyte: Col = %d, want 1", m.cursor.Col)
	}
	if got := m.buf.String(); got != "á" {
		t.Errorf("type multibyte: buf = %q, want %q", got, "á")
	}
}

// Bug D: addNextOccurrence wrap-around — the confusional actualPos math
// double-counts the searchStart offset when the match wraps, resulting in
// either no cursor added or wrong cursor position. We test the wrap case.
func TestAddNextOccurrenceWrap(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	// text: foo bar foo ; cursor at position 0 (first "foo")
	m.buf.InsertString("foo bar foo")
	m.syncGapToCursor()
	// Position cursor at the second "foo" so wrap search hits the first.
	m.cursor.SetPos(0, 8) // start of second "foo"
	m.syncGapToCursor()
	m.addNextOccurrence()
	if len(m.extraCursors) == 0 {
		t.Fatal("addNextOccurrence wrap: expected an extra cursor, got none")
	}
	got := m.extraCursors[0].GapPos
	if got != 0 {
		t.Errorf("addNextOccurrence wrap: GapPos = %d, want 0 (first foo)", got)
	}
}

// Bug: handleDelete with extraCursors. Captured allCursors use a primary
// GapPos captured BEFORE the loop. After the first iteration deletes a rune
// BEFORE the primary's gap, that gap's logical position is no longer valid:
// the code does not decrement it, so the wrong rune is deleted.
func TestHandleDeleteExtraCursorsPositions(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf.InsertString("abcdef")
	m.syncGapToCursor()
	// primary at col 2 (logical gap=2), extra at col 0 (GapPos=0)
	m.cursor.SetPos(0, 2)
	m.syncGapToCursor()
	m.extraCursors = append(m.extraCursors, EditorCursor{Line: 0, Col: 0, GapPos: 0})
	m.handleKey(tea.KeyMsg{Type: tea.KeyDelete})
	if got := m.buf.String(); got != "bcef" {
		t.Errorf("handleDelete multi: buf = %q, want \"bcef\" (stale pos, demonstrates bug)", got)
	}
}

// Bug: MoveGapLeft return value is documented as returning the rune that was
// "moved past", but it actually returns the same rune it copied. Verify it
// returns a non-zero rune when cursor moves left over one.
func TestMoveGapLeftReturnValue(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf.InsertString("abc")
	// gap is at 3. Move left: should return 'c'.
	got := m.buf.MoveGapLeft()
	if got != 'c' {
		t.Errorf("MoveGapLeft: got %q, want 'c'", got)
	}
}

// Bug B: doSearch substring iteration uses byte indices over a lowercased
// copy of the content. If the query is ASCII, this is fine, but if the
// content has multibyte runes *before* the match, the match index reported
// is in BYTES, not RUNES — and searchMatches is later used as logical rune
// indices to navigate. We test the rune-vs-byte mismatch.
func TestDoSearchMultibyteContentMatchesAreRuneIndices(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	// "ábc": á = 2 bytes, b = 1 byte, c = 1 byte.
	// Searching for "b" should yield rune-index 1, not byte-index 2.
	m.buf.InsertString("ábc")
	m.syncGapToCursor()
	m.searchQuery = "b"
	m.doSearch()
	if len(m.searchMatches) != 1 {
		t.Fatalf("doSearch: expected 1 match, got %d", len(m.searchMatches))
	}
	if m.searchMatches[0] != 1 {
		t.Errorf("doSearch multibyte: match index = %d (bytes?), want 1 (rune)", m.searchMatches[0])
	}
}

// Bug E: wordAtCursor (and word-left/right navigation) used `rune(line[i])` on
// a UTF-8 string — byte indexing a multibyte sequence corrupts the rune. The
// fix switches to []rune(line). We verify with text where `é` is NOT a wordchar
// so the word boundary lands on a multibyte rune.
func TestWordAtCursorMultibyte(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	// runes: c a f é w o r l d  (9 runes, 10 bytes)
	m.buf.InsertString("caféworld")
	m.syncGapToCursor()
	m.cursor.SetPos(0, 6) // rune idx 6 = 'r'
	m.syncGapToCursor()
	w := m.wordAtCursor()
	if w != "world" {
		t.Errorf("wordAtCursor multibyte: got %q, want \"world\"", w)
	}
}
