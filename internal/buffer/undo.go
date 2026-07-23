package buffer

type Operation struct {
	Type string
	Pos  int
	Text string
}

type UndoStack struct {
	stack     []Operation
	position  int
	unitCount int
}

func NewUndoStack() *UndoStack {
	return &UndoStack{
		stack:    make([]Operation, 0, 256),
		position: 0,
	}
}

func (u *UndoStack) Push(op Operation) {
	u.stack = u.stack[:u.position]
	u.stack = append(u.stack, op)
	u.position = len(u.stack)
	u.unitCount = u.countTotalUnits() // recompute after truncation
}

func (u *UndoStack) PushComposite(ops []Operation) {
	u.stack = u.stack[:u.position]
	u.stack = append(u.stack, Operation{Type: "begin-composite"})
	u.stack = append(u.stack, ops...)
	u.stack = append(u.stack, Operation{Type: "end-composite"})
	u.position = len(u.stack)
	u.unitCount = u.countTotalUnits()
}

func (u *UndoStack) countTotalUnits() int {
	total := 0
	inComposite := false
	for _, op := range u.stack {
		switch op.Type {
		case "begin-composite":
			inComposite = true
			total++
		case "end-composite":
			inComposite = false
		default:
			if !inComposite {
				total++
			}
		}
	}
	return total
}

func (u *UndoStack) CanUndo() bool {
	return u.unitCount > 0
}

func (u *UndoStack) CanRedo() bool {
	return u.unitCount < u.countTotalUnits()
}

func (u *UndoStack) Undo() ([]Operation, bool) {
	if !u.CanUndo() {
		return nil, false
	}
	u.position--
	op := u.stack[u.position]
	if op.Type == "end-composite" {
		u.position--
		var ops []Operation
		for u.position >= 0 {
			inner := u.stack[u.position]
			if inner.Type == "begin-composite" {
				break // leave position at begin-composite for redo
			}
			ops = append(ops, inner)
			u.position--
		}
		u.unitCount--
		return ops, true
	}
	if op.Type == "begin-composite" {
		return u.Undo()
	}
	u.unitCount--
	return []Operation{op}, true
}

func (u *UndoStack) Redo() ([]Operation, bool) {
	if !u.CanRedo() {
		return nil, false
	}
	op := u.stack[u.position]
	if op.Type == "begin-composite" {
		u.position++
		var ops []Operation
		for u.position < len(u.stack) {
			inner := u.stack[u.position]
			if inner.Type == "end-composite" {
				u.position++
				break
			}
			ops = append(ops, inner)
			u.position++
		}
		u.unitCount++
		return ops, true
	}
	u.position++
	u.unitCount++
	return []Operation{op}, true
}

func (u *UndoStack) Clear() {
	u.stack = u.stack[:0]
	u.position = 0
	u.unitCount = 0
}

func (u *UndoStack) Len() int {
	return u.unitCount
}


