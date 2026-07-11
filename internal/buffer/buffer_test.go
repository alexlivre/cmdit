package buffer

import (
	"testing"
)

func TestNewBufferIsEmpty(t *testing.T) {
	b := NewBuffer()
	if b.Len() != 0 {
		t.Errorf("expected empty buffer, got length %d", b.Len())
	}
	if b.String() != "" {
		t.Errorf("expected empty string, got %q", b.String())
	}
}

func TestInsertSingleRune(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')
	if b.Len() != 1 {
		t.Errorf("expected length 1, got %d", b.Len())
	}
	if b.String() != "a" {
		t.Errorf("expected 'a', got %q", b.String())
	}
}

func TestInsertMultipleRunes(t *testing.T) {
	b := NewBuffer()
	for _, r := range "hello" {
		b.Insert(r)
	}
	if b.Len() != 5 {
		t.Errorf("expected length 5, got %d", b.Len())
	}
	if b.String() != "hello" {
		t.Errorf("expected 'hello', got %q", b.String())
	}
}

func TestDeleteBackspace(t *testing.T) {
	b := NewBufferFromString("hello world")
	// Move gap to end so we can delete
	for i := 0; i < 11; i++ {
		b.MoveGapRight()
	}

	deleted := b.Backspace()
	if !deleted {
		t.Error("expected deletion to succeed")
	}
	if b.String() != "hello worl" {
		t.Errorf("expected 'hello worl', got %q", b.String())
	}
}

func TestDeleteFromEmptyBuffer(t *testing.T) {
	b := NewBuffer()
	deleted := b.Backspace()
	if deleted {
		t.Error("expected deletion to fail on empty buffer")
	}
}

func TestMoveGapRightAndInsert(t *testing.T) {
	b := NewBufferFromString("hello")
	// After NewBufferFromString, the gap is at the end (position 5).
	// Move gap left twice to position 3 (after "hel").
	b.MoveGapLeft()
	b.MoveGapLeft()

	b.InsertString("X")
	if b.String() != "helXlo" {
		t.Errorf("expected 'helXlo', got %q", b.String())
	}
}

func TestNewBufferFromString(t *testing.T) {
	b := NewBufferFromString("hello\nworld")
	if b.String() != "hello\nworld" {
		t.Errorf("expected 'hello\\nworld', got %q", b.String())
	}
	if b.Len() != 11 {
		t.Errorf("expected length 11, got %d", b.Len())
	}
}

func TestLines(t *testing.T) {
	b := NewBufferFromString("line1\nline2\nline3")
	lines := b.Lines()
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestLineCount(t *testing.T) {
	b := NewBufferFromString("a\nb\nc\n")
	if b.LineCount() != 4 {
		t.Errorf("expected 4 lines, got %d", b.LineCount())
	}

	b2 := NewBufferFromString("single line")
	if b2.LineCount() != 1 {
		t.Errorf("expected 1 line, got %d", b2.LineCount())
	}
}

func TestLineStart(t *testing.T) {
	b := NewBufferFromString("abc\ndef\nghi")
	// Line 0 starts at 0
	if b.LineStart(0) != 0 {
		t.Errorf("expected line 0 at 0, got %d", b.LineStart(0))
	}
	// Line 1 starts after "abc\n"
	if b.LineStart(1) != 4 {
		t.Errorf("expected line 1 at 4, got %d", b.LineStart(1))
	}
	// Line 2 starts after "abc\ndef\n"
	if b.LineStart(2) != 8 {
		t.Errorf("expected line 2 at 8, got %d", b.LineStart(2))
	}
}

func TestRuneAt(t *testing.T) {
	b := NewBufferFromString("abc")
	if b.RuneAt(0) != 'a' {
		t.Errorf("expected 'a', got %c", b.RuneAt(0))
	}
	if b.RuneAt(1) != 'b' {
		t.Errorf("expected 'b', got %c", b.RuneAt(1))
	}
	if b.RuneAt(2) != 'c' {
		t.Errorf("expected 'c', got %c", b.RuneAt(2))
	}
	if b.RuneAt(3) != 0 {
		t.Error("expected 0 for out-of-bounds")
	}
	if b.RuneAt(-1) != 0 {
		t.Error("expected 0 for negative index")
	}
}

func TestMoveGapLeft(t *testing.T) {
	b := NewBufferFromString("hello")
	// Gap is at position 0 initially. Move gap right to end.
	for i := 0; i < 5; i++ {
		b.MoveGapRight()
	}
	// Gap at position 5 (end). Move left.
	r := b.MoveGapLeft()
	if r != 'o' {
		t.Errorf("expected 'o', got %c", r)
	}
	// String should still be the same
	if b.String() != "hello" {
		t.Errorf("expected 'hello', got %q", b.String())
	}
}

func TestDeleteForward(t *testing.T) {
	b := NewBufferFromString("hello")
	// Gap is at the end. Move gap to beginning.
	for i := 0; i < 5; i++ {
		b.MoveGapLeft()
	}
	// Now gap at position 0. DeleteForward removes first char.
	deleted := b.DeleteForward()
	if !deleted {
		t.Error("expected deletion to succeed")
	}
	if b.String() != "ello" {
		t.Errorf("expected 'ello', got %q", b.String())
	}
}

func TestByteOffset(t *testing.T) {
	b := NewBufferFromString("abc")
	if b.ByteOffset(0) != 0 {
		t.Errorf("expected 0, got %d", b.ByteOffset(0))
	}
	if b.ByteOffset(3) != 3 {
		t.Errorf("expected 3, got %d", b.ByteOffset(3))
	}

	b2 := NewBufferFromString("olá")
	offset := b2.ByteOffset(2)
	if offset != 2 { // 'o'=1byte + 'l'=1byte = 2
		t.Errorf("expected byte offset 2, got %d", offset)
	}
}

func TestLineCol(t *testing.T) {
	b := NewBufferFromString("ab\ncd\nef")
	// ab\ncd\nef
	// 0:a 1:b 2:\n 3:c 4:d 5:\n 6:e 7:f

	tests := []struct {
		index   int
		expLine int
		expCol  int
	}{
		{0, 0, 0},
		{1, 0, 1},
		{2, 0, 2}, // newline at end of line 0
		{3, 1, 0},
		{4, 1, 1},
		{5, 1, 2},
		{6, 2, 0},
		{7, 2, 1},
	}

	for _, tc := range tests {
		line, col := b.LineCol(tc.index)
		if line != tc.expLine || col != tc.expCol {
			t.Errorf("LineCol(%d) = (%d,%d), expected (%d,%d)",
				tc.index, line, col, tc.expLine, tc.expCol)
		}
	}
}

func TestLineColPastEnd(t *testing.T) {
	b := NewBufferFromString("ab\ncd")
	line, col := b.LineCol(100)
	if line != 1 || col != 2 {
		t.Errorf("LineCol(100) = (%d,%d), expected (1,2)", line, col)
	}
}

func TestLineColNegative(t *testing.T) {
	b := NewBufferFromString("hello")
	line, col := b.LineCol(-1)
	if line != 0 || col != 0 {
		t.Errorf("LineCol(-1) = (%d,%d), expected (0,0)", line, col)
	}
}

func TestMoveGapTo(t *testing.T) {
	b := NewBufferFromString("hello world")

	b.MoveGapTo(5)
	if b.GapPosition() != 5 {
		t.Errorf("expected gap at 5, got %d", b.GapPosition())
	}

	b.MoveGapTo(0)
	if b.GapPosition() != 0 {
		t.Errorf("expected gap at 0, got %d", b.GapPosition())
	}

	b.MoveGapTo(11)
	if b.GapPosition() != 11 {
		t.Errorf("expected gap at 11, got %d", b.GapPosition())
	}

	if b.String() != "hello world" {
		t.Errorf("content changed: %s", b.String())
	}
}

func TestMoveGapTo_OutOfBounds(t *testing.T) {
	b := NewBufferFromString("test")
	b.MoveGapTo(100)
	if b.GapPosition() != 4 {
		t.Errorf("expected clamp to 4, got %d", b.GapPosition())
	}
	b.MoveGapTo(-10)
	if b.GapPosition() != 0 {
		t.Errorf("expected clamp to 0, got %d", b.GapPosition())
	}
}

func TestLines_Cached(t *testing.T) {
	b := NewBufferFromString("line1\nline2\nline3")

	lines := b.Lines()
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Errorf("unexpected lines: %v", lines)
	}

	lines2 := b.Lines()
	if len(lines2) != 3 {
		t.Errorf("cached call failed")
	}
}

func TestLines_InvalidateOnInsert(t *testing.T) {
	b := NewBufferFromString("hello\nworld")
	_ = b.Lines()

	b.MoveGapTo(5)
	b.Insert('\n')

	lines := b.Lines()
	if len(lines) != 3 {
		t.Errorf("expected 3 lines after insert, got %d", len(lines))
	}
}

func TestLineCount_Cached(t *testing.T) {
	b := NewBufferFromString("a\nb\nc\nd")
	if b.LineCount() != 4 {
		t.Errorf("expected 4 lines, got %d", b.LineCount())
	}
}

func TestLineStart_Cached(t *testing.T) {
	b := NewBufferFromString("hello\nworld\nfoo")
	if b.LineStart(0) != 0 {
		t.Errorf("line 0 start: expected 0, got %d", b.LineStart(0))
	}
	if b.LineStart(1) != 6 {
		t.Errorf("line 1 start: expected 6, got %d", b.LineStart(1))
	}
	if b.LineStart(2) != 12 {
		t.Errorf("line 2 start: expected 12, got %d", b.LineStart(2))
	}
}

func TestInsertString_Empty(t *testing.T) {
	b := NewBuffer()
	b.InsertString("")
	if b.Len() != 0 {
		t.Errorf("expected length 0, got %d", b.Len())
	}
}

func TestDeleteForward_AtEnd(t *testing.T) {
	b := NewBufferFromString("test")
	b.MoveGapTo(4)
	if b.DeleteForward() {
		t.Error("expected false when deleting at end")
	}
}

func TestBackspace_AtStart(t *testing.T) {
	b := NewBufferFromString("test")
	b.MoveGapTo(0)
	if b.Backspace() {
		t.Error("expected false when backspacing at start")
	}
}

func TestRuneAt_OutOfBounds(t *testing.T) {
	b := NewBufferFromString("test")
	if b.RuneAt(-1) != 0 {
		t.Error("expected 0 for negative index")
	}
	if b.RuneAt(100) != 0 {
		t.Error("expected 0 for out of bounds index")
	}
}
