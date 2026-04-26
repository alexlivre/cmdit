package buffer

import (
	"testing"
)

func TestNewCursor(t *testing.T) {
	c := NewCursor()
	if c.Line != 0 || c.Col != 0 {
		t.Errorf("expected (0,0), got (%d,%d)", c.Line, c.Col)
	}
}

func TestCursorUp(t *testing.T) {
	c := &Cursor{Line: 2, Col: 5}
	c.Up()
	if c.Line != 1 || c.Col != 5 {
		t.Errorf("expected (1,5), got (%d,%d)", c.Line, c.Col)
	}
}

func TestCursorUpAtTop(t *testing.T) {
	c := &Cursor{Line: 0, Col: 5}
	c.Up()
	if c.Line != 0 {
		t.Errorf("expected line 0, got %d", c.Line)
	}
}

func TestCursorDown(t *testing.T) {
	c := &Cursor{Line: 0, Col: 3}
	c.Down()
	if c.Line != 1 || c.Col != 3 {
		t.Errorf("expected (1,3), got (%d,%d)", c.Line, c.Col)
	}
}

func TestCursorLeft(t *testing.T) {
	c := &Cursor{Line: 0, Col: 5}
	c.Left()
	if c.Col != 4 {
		t.Errorf("expected col 4, got %d", c.Col)
	}
}

func TestCursorLeftAtStart(t *testing.T) {
	c := &Cursor{Line: 0, Col: 0}
	c.Left()
	if c.Col != 0 {
		t.Errorf("expected col 0, got %d", c.Col)
	}
}

func TestCursorRight(t *testing.T) {
	c := &Cursor{Line: 0, Col: 3}
	c.Right()
	if c.Col != 4 {
		t.Errorf("expected col 4, got %d", c.Col)
	}
}

func TestSetPos(t *testing.T) {
	c := NewCursor()
	c.SetPos(3, 10)
	if c.Line != 3 || c.Col != 10 {
		t.Errorf("expected (3,10), got (%d,%d)", c.Line, c.Col)
	}
}

func TestSetPosNegative(t *testing.T) {
	c := NewCursor()
	c.SetPos(-1, -5)
	if c.Line != 0 || c.Col != 0 {
		t.Errorf("expected (0,0) for negative, got (%d,%d)", c.Line, c.Col)
	}
}
