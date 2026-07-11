package buffer

type OpType int

const (
	OpInsert OpType = iota
	OpDelete
)

// Operation represents a reversible edit operation.
type Operation struct {
	Type OpType
	Pos  int
	Text string
}

// UndoStack provides unlimited undo/redo for buffer operations.
type UndoStack struct {
	stack    []Operation
	position int // Points to next free slot (number of undos available)
}

// NewUndoStack creates an empty undo stack.
func NewUndoStack() *UndoStack {
	return &UndoStack{
		stack:    make([]Operation, 0, 256),
		position: 0,
	}
}

// Push records a new operation. Discards any redo history.
func (u *UndoStack) Push(op Operation) {
	// Discard redo history
	u.stack = u.stack[:u.position]
	u.stack = append(u.stack, op)
	u.position = len(u.stack)
}

// CanUndo returns true if there are operations to undo.
func (u *UndoStack) CanUndo() bool {
	return u.position > 0
}

// CanRedo returns true if there are operations to redo.
func (u *UndoStack) CanRedo() bool {
	return u.position < len(u.stack)
}

// Undo returns the operation to reverse the most recent edit.
func (u *UndoStack) Undo() (Operation, bool) {
	if !u.CanUndo() {
		return Operation{}, false
	}
	u.position--
	return u.stack[u.position], true
}

// Redo returns the operation to re-apply the most recently undone edit.
func (u *UndoStack) Redo() (Operation, bool) {
	if !u.CanRedo() {
		return Operation{}, false
	}
	op := u.stack[u.position]
	u.position++
	return op, true
}

// Clear removes all undo/redo history.
func (u *UndoStack) Clear() {
	u.stack = u.stack[:0]
	u.position = 0
}

// Len returns the number of undoable operations.
func (u *UndoStack) Len() int {
	return u.position
}
