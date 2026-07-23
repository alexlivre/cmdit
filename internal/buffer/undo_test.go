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
	u.Push(Operation{Type: "insert", Pos: 0, Text: "hello"})

	if !u.CanUndo() {
		t.Error("should have undo after push")
	}

	ops, ok := u.Undo()
	if !ok {
		t.Error("undo should succeed")
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	if ops[0].Type != "insert" || ops[0].Text != "hello" {
		t.Errorf("unexpected operation: %+v", ops[0])
	}
	if u.CanUndo() {
		t.Error("should not have more undo")
	}
}

func TestUndoAndRedo(t *testing.T) {
	u := NewUndoStack()
	u.Push(Operation{Type: "insert", Pos: 0, Text: "a"})
	u.Push(Operation{Type: "insert", Pos: 1, Text: "b"})

	ops, _ := u.Undo()
	if ops[0].Text != "b" {
		t.Errorf("expected 'b', got %q", ops[0].Text)
	}
	ops, _ = u.Undo()
	if ops[0].Text != "a" {
		t.Errorf("expected 'a', got %q", ops[0].Text)
	}

	ops, _ = u.Redo()
	if ops[0].Text != "a" {
		t.Errorf("expected 'a' on redo, got %q", ops[0].Text)
	}
	ops, _ = u.Redo()
	if ops[0].Text != "b" {
		t.Errorf("expected 'b' on redo, got %q", ops[0].Text)
	}

	if u.CanRedo() {
		t.Error("should not have more redo")
	}
}

func TestPushDiscardsRedoHistory(t *testing.T) {
	u := NewUndoStack()
	u.Push(Operation{Type: "insert", Pos: 0, Text: "a"})
	u.Push(Operation{Type: "insert", Pos: 1, Text: "b"})

	u.Undo()
	u.Push(Operation{Type: "insert", Pos: 1, Text: "c"})

	if u.CanRedo() {
		t.Error("redo history should be discarded after new push")
	}

	ops, _ := u.Undo()
	if ops[0].Text != "c" {
		t.Errorf("expected 'c', got %q", ops[0].Text)
	}
}

func TestClear(t *testing.T) {
	u := NewUndoStack()
	u.Push(Operation{Type: "insert", Pos: 0, Text: "a"})
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
		u.Push(Operation{Type: "insert", Pos: i, Text: "x"})
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

func TestPushCompositeUndoRedo(t *testing.T) {
	u := NewUndoStack()

	ops := []Operation{
		{Type: "insert", Pos: 0, Text: "a"},
		{Type: "insert", Pos: 1, Text: "b"},
		{Type: "insert", Pos: 2, Text: "c"},
	}
	u.PushComposite(ops)
	if u.Len() != 1 {
		t.Fatalf("PushComposite should count as 1 undo unit, got %d", u.Len())
	}

	undoOps, ok := u.Undo()
	if !ok {
		t.Fatal("expected undo to succeed")
	}
	if len(undoOps) != 3 {
		t.Fatalf("expected 3 operations in composite undo, got %d", len(undoOps))
	}
	if undoOps[0].Text != "c" {
		t.Errorf("expected first undo op to be 'c' (reversed), got %q", undoOps[0].Text)
	}
	if undoOps[2].Text != "a" {
		t.Errorf("expected last undo op to be 'a' (reversed), got %q", undoOps[2].Text)
	}

	redoOps, ok := u.Redo()
	if !ok {
		t.Fatal("expected redo to succeed")
	}
	if len(redoOps) != 3 {
		t.Fatalf("expected 3 operations in composite redo, got %d", len(redoOps))
	}
}

func TestUndoRedoCompositeStack(t *testing.T) {
	u := NewUndoStack()

	u.Push(Operation{Type: "insert", Pos: 0, Text: "x"})
	u.PushComposite([]Operation{
		{Type: "insert", Pos: 0, Text: "y"},
		{Type: "insert", Pos: 0, Text: "z"},
	})
	u.Push(Operation{Type: "insert", Pos: 0, Text: "w"})

	if u.Len() != 3 {
		t.Fatalf("expected 3 undo units, got %d", u.Len())
	}

	ops, _ := u.Undo()
	if len(ops) != 1 || ops[0].Text != "w" {
		t.Fatal("first undo should be single 'w'")
	}

	ops, _ = u.Undo()
	if len(ops) != 2 {
		t.Fatalf("second undo should be composite of 2, got %d", len(ops))
	}

	ops, _ = u.Undo()
	if len(ops) != 1 || ops[0].Text != "x" {
		t.Fatal("third undo should be single 'x'")
	}

	if u.CanUndo() {
		t.Fatal("stack should be empty")
	}
}
