package buffer

import (
	"testing"
)

func TestNewUndoStack(t *testing.T) {
	u := NewUndoStack()
	if u.CanUndo() {
		t.Error("new stack should not have undo")
	}
	if u.CanRedo() {
		t.Error("new stack should not have redo")
	}
}

func TestPushAndUndo(t *testing.T) {
	u := NewUndoStack()
	u.Push(Operation{Type: OpInsert, Pos: 0, Text: "hello"})

	if !u.CanUndo() {
		t.Error("should have undo after push")
	}

	op, ok := u.Undo()
	if !ok {
		t.Error("undo should succeed")
	}
	if op.Type != OpInsert || op.Text != "hello" {
		t.Errorf("unexpected operation: %+v", op)
	}
	if u.CanUndo() {
		t.Error("should not have more undo")
	}
}

func TestUndoAndRedo(t *testing.T) {
	u := NewUndoStack()
	u.Push(Operation{Type: OpInsert, Pos: 0, Text: "a"})
	u.Push(Operation{Type: OpInsert, Pos: 1, Text: "b"})

	// Undo last
	op, _ := u.Undo()
	if op.Text != "b" {
		t.Errorf("expected 'b', got %q", op.Text)
	}
	// Undo first
	op, _ = u.Undo()
	if op.Text != "a" {
		t.Errorf("expected 'a', got %q", op.Text)
	}

	// Redo first
	op, _ = u.Redo()
	if op.Text != "a" {
		t.Errorf("expected 'a' on redo, got %q", op.Text)
	}
	// Redo second
	op, _ = u.Redo()
	if op.Text != "b" {
		t.Errorf("expected 'b' on redo, got %q", op.Text)
	}

	if u.CanRedo() {
		t.Error("should not have more redo")
	}
}

func TestPushDiscardsRedoHistory(t *testing.T) {
	u := NewUndoStack()
	u.Push(Operation{Type: OpInsert, Pos: 0, Text: "a"})
	u.Push(Operation{Type: OpInsert, Pos: 1, Text: "b"})

	// Undo one
	u.Undo()
	// Now push a new operation - should discard 'b' redo
	u.Push(Operation{Type: OpInsert, Pos: 1, Text: "c"})

	if u.CanRedo() {
		t.Error("redo history should be discarded after new push")
	}

	op, _ := u.Undo()
	if op.Text != "c" {
		t.Errorf("expected 'c', got %q", op.Text)
	}
}

func TestClear(t *testing.T) {
	u := NewUndoStack()
	u.Push(Operation{Type: OpInsert, Pos: 0, Text: "a"})
	u.Clear()

	if u.CanUndo() {
		t.Error("should be empty after clear")
	}
	if u.Len() != 0 {
		t.Errorf("expected len 0, got %d", u.Len())
	}
}

func TestMultipleOperations(t *testing.T) {
	u := NewUndoStack()
	for i := 0; i < 100; i++ {
		u.Push(Operation{Type: OpInsert, Pos: i, Text: "x"})
	}

	count := 0
	for u.CanUndo() {
		u.Undo()
		count++
	}
	if count != 100 {
		t.Errorf("expected 100 undos, got %d", count)
	}
}
