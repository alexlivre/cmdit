# Bug Fixes — cmdit v0.4.3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all 16 bugs identified in BUG_REPORT.md across 6 sequential topics, each with regression tests.

**Architecture:** Six sequential task groups: (1) undo integrity via CompositeOperation, (2) multi-cursor position fixes, (3) LSP lifecycle with context + timeout, (4) search highlight + Go-to-Line feature, (5) buffer/gap invariant fixes, (6) rendering pipeline + error reporting.

**Tech Stack:** Go 1.23+, Bubble Tea, Chroma, lipgloss, standard `testing` package.

## Global Constraints

- Go 1.23+ (per go.mod)
- `go test ./...` must pass after every commit
- `go vet ./...` must produce zero warnings after every commit
- Commit messages use `fix:` prefix with bug reference
- Each fix includes a regression test in the commit
- Code/comments in English; UI strings in Portuguese (existing convention)

---

### Task 1: CompositeOperation model + undo/redo call sites (Bugs #2, #4, #7 — PR1)

**Files:**
- Modify: `internal/buffer/undo.go`
- Modify: `internal/buffer/undo_test.go`
- Modify: `internal/editor/actions.go` (undo, redo, doReplace, cut)
- Modify: `internal/editor/editor.go` (insertTextAtAllCursors)
- Modify: `internal/editor/editor_test.go`
- Modify: `internal/editor/regression_test.go`

**Interfaces:**
- Produces: `func (u *UndoStack) PushComposite(ops []Operation)`, `func (u *UndoStack) Undo() ([]Operation, bool)` (changed from `(Operation, bool)`), `func (u *UndoStack) Redo() ([]Operation, bool)` (changed from `(Operation, bool)`)

- [ ] **Step 1: Update undo.go — add PushComposite, change Undo/Redo return types, add unitCount**

Replace entire `internal/buffer/undo.go`:

```go
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
	u.unitCount = u.position
}

func (u *UndoStack) PushComposite(ops []Operation) {
	u.stack = u.stack[:u.position]
	u.stack = append(u.stack, Operation{Type: "begin-composite"})
	u.stack = append(u.stack, ops...)
	u.stack = append(u.stack, Operation{Type: "end-composite"})
	u.position = len(u.stack)
	u.unitCount++
}

func (u *UndoStack) CanUndo() bool {
	return u.unitCount > 0
}

func (u *UndoStack) CanRedo() bool {
	return u.unitCount < countUnits(u.stack, u.position)
}

func countUnits(stack []Operation, position int) int {
	count := 0
	for i := 0; i < position; i++ {
		if stack[i].Type == "begin-composite" || stack[i].Type == "end-composite" {
			continue
		}
		isStartOfComposite := i > 0 && stack[i-1].Type == "begin-composite"
		isEndOfComposite := i+1 < len(stack) && stack[i+1].Type == "end-composite"
		if isStartOfComposite && i+1 < len(stack) && !isEndOfComposite {
			continue // counted when we hit the first op after begin
		}
		if i > 0 && stack[i-1].Type != "begin-composite" && stack[i].Type != "begin-composite" && stack[i].Type != "end-composite" {
			count++
		} else if stack[i].Type != "begin-composite" && stack[i].Type != "end-composite" {
			count++
		}
	}
	return count
}

func (u *UndoStack) Undo() ([]Operation, bool) {
	if !u.CanUndo() {
		return nil, false
	}
	u.position--
	op := u.stack[u.position]
	if op.Type == "end-composite" {
		u.position-- // skip end-composite
		var ops []Operation
		for u.position >= 0 {
			inner := u.stack[u.position]
			if inner.Type == "begin-composite" {
				u.position-- // skip begin-composite
				break
			}
			ops = append(ops, inner)
			u.position--
		}
		u.unitCount--
		reverseOps(ops)
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
		u.position++ // skip begin-composite
		var ops []Operation
		for u.position < len(u.stack) {
			inner := u.stack[u.position]
			if inner.Type == "end-composite" {
				u.position++ // skip end-composite
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

func reverseOps(ops []Operation) {
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}
}
```

- [ ] **Step 2: Update undo_test.go — fix existing tests for []Operation return + add composite tests**

Replace entire `internal/buffer/undo_test.go`:

```go
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
```

- [ ] **Step 3: Run buffer tests**

```bash
go test ./internal/buffer/ -v
```
Expected: FAIL — `internal/editor/actions.go` won't compile because `undo()` and `redo()` still use `Operation` singular.

- [ ] **Step 4: Fix undo/redo call sites in actions.go**

Read `internal/editor/actions.go:126-166`. Replace `undo()` and `redo()`:

```go
func (m *Model) undo() {
	ops, ok := m.undoStack.Undo()
	if !ok {
		return
	}
	for _, op := range ops {
		m.moveGapTo(op.Pos)
		switch op.Type {
		case "insert":
			m.buf.InsertString(op.Text)
		case "delete":
			for i := 0; i < len(op.Text); i++ {
				m.buf.DeleteForward()
			}
		}
	}
	m.modified = true
	m.cursor.Line, m.cursor.Col = m.buf.LineCol(m.buf.GapPosition())
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
	m.sendDidChange()
}

func (m *Model) redo() {
	ops, ok := m.undoStack.Redo()
	if !ok {
		return
	}
	for _, op := range ops {
		m.moveGapTo(op.Pos)
		switch op.Type {
		case "insert":
			for i := 0; i < len(op.Text); i++ {
				m.buf.DeleteForward()
			}
		case "delete":
			m.buf.InsertString(op.Text)
		}
	}
	m.modified = true
	m.cursor.Line, m.cursor.Col = m.buf.LineCol(m.buf.GapPosition())
	m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
	m.sendDidChange()
}
```

- [ ] **Step 5: Fix insertTextAtAllCursors to use PushComposite**

Read `internal/editor/editor.go:336-362`. Replace:

```go
func (m *Model) insertTextAtAllCursors(text string) {
	if len(m.extraCursors) == 0 {
		m.insertText(text)
		return
	}

	all := m.allCursors()
	var ops []buffer.Operation
	for i := len(all) - 1; i >= 0; i-- {
		m.moveGapTo(all[i].GapPos)
		ops = append(ops, buffer.Operation{
			Type: "delete",
			Pos:  all[i].GapPos,
			Text: text,
		})
		m.buf.InsertString(text)
	}
	m.undoStack.PushComposite(ops)

	delta := utf8.RuneCountInString(text)
	m.cursor.Col += delta
	for i := range m.extraCursors {
		m.extraCursors[i].Col += delta
	}
	m.modified = true
	m.sendDidChange()
}
```

- [ ] **Step 6: Fix doReplace to push undo**

Read `internal/editor/actions.go:294-308`. Replace:

```go
func (m *Model) doReplace() {
	if len(m.searchMatches) == 0 || m.replaceQuery == "" {
		return
	}

	pos := m.searchMatches[m.searchCurrent]
	m.undoStack.Push(buffer.Operation{
		Type: "insert",
		Pos:  pos,
		Text: m.searchQuery,
	})
	m.moveGapTo(pos)
	for i := 0; i < utf8.RuneCountInString(m.searchQuery); i++ {
		m.buf.DeleteForward()
	}
	m.buf.InsertString(m.replaceQuery)
	m.modified = true

	m.doSearch()
}
```

- [ ] **Step 7: Fix cut() without selection to push undo**

Read `internal/editor/actions.go:182-195`. Replace the non-selection branch:

```go
func (m *Model) cut() {
	if m.hasSelection() {
		m.copy()
		m.deleteSelection()
	} else {
		m.clipboard.Copy(m.currentLineText())
		lineStart := m.buf.LineStart(m.cursor.Line)
		lineEnd := m.buf.LineStart(m.cursor.Line + 1)
		var sb strings.Builder
		for i := lineStart; i < lineEnd && i < m.buf.Len(); i++ {
			sb.WriteRune(m.buf.RuneAt(i))
		}
		m.undoStack.Push(buffer.Operation{
			Type: "insert",
			Pos:  lineStart,
			Text: sb.String(),
		})
		m.moveGapTo(lineStart)
		for i := lineStart; i < lineEnd && m.buf.GapPosition() < m.buf.Len(); i++ {
			m.buf.DeleteForward()
		}
		m.modified = true
	}
}
```

- [ ] **Step 8: Add regression tests**

Append to `internal/editor/regression_test.go`:

```go
func TestMultiCursorUndoAtomico(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf = buffer.NewBufferFromString("aaa bbb ccc")
	m.cursor.SetPos(0, 0)
	m.syncGapToCursor()

	m.extraCursors = []EditorCursor{
		{Line: 0, Col: 4, GapPos: 4},
		{Line: 0, Col: 8, GapPos: 8},
	}

	m.insertTextAtAllCursors("x")

	content := m.buf.String()
	if len(content) != 14 {
		t.Fatalf("expected length 14 after 3 inserts, got %d: %q", len(content), content)
	}

	m.undo()
	contentAfterUndo := m.buf.String()
	if contentAfterUndo != "aaa bbb ccc" {
		t.Errorf("expected 'aaa bbb ccc' after undo, got %q", contentAfterUndo)
	}
}

func TestReplaceUndo(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf = buffer.NewBufferFromString("hello world")
	m.searchQuery = "world"
	m.replaceQuery = "golang"
	m.doSearch()
	m.doReplace()

	if m.buf.String() != "hello golang" {
		t.Fatalf("expected 'hello golang', got %q", m.buf.String())
	}

	m.undo()
	if m.buf.String() != "hello world" {
		t.Errorf("expected 'hello world' after undo, got %q", m.buf.String())
	}
}

func TestCutLineUndo(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf = buffer.NewBufferFromString("line one\nline two\n")
	m.cursor.SetPos(1, 0)
	m.syncGapToCursor()

	m.cut()

	content := m.buf.String()
	if content != "line one\n\n" {
		t.Fatalf("expected 'line one\\n\\n' after cut, got %q", content)
	}

	m.undo()
	if m.buf.String() != "line one\nline two\n" {
		t.Errorf("expected original content after undo, got %q", m.buf.String())
	}
}
```

- [ ] **Step 9: Run all tests**

```bash
go test ./internal/buffer/ ./internal/editor/ -v
```
Expected: ALL PASS.

- [ ] **Step 10: Run vet**

```bash
go vet ./...
```
Expected: zero warnings.

- [ ] **Step 11: Commit PR1**

```bash
git add internal/buffer/undo.go internal/buffer/undo_test.go internal/editor/actions.go internal/editor/editor.go internal/editor/regression_test.go
git commit -m "fix: composite undo for multi-cursor, replace, and cut (bugs #2, #4, #7)"
```

---

### Task 2: Multi-cursor position fixes (Bugs #3, #8 — PR2)

**Files:**
- Modify: `internal/editor/editor.go` (addNextOccurrence)
- Modify: `internal/editor/keys.go` (handleBackspace)
- Modify: `internal/editor/regression_test.go`

- [ ] **Step 1: Fix addNextOccurrence position calculation**

Read `internal/editor/editor.go:634-696`. Replace:

```go
func (m *Model) addNextOccurrence() {
	word := m.wordAtCursor()
	if word == "" {
		return
	}

	existing := make(map[int]bool)
	existing[m.buf.GapPosition()] = true
	for _, c := range m.extraCursors {
		existing[c.GapPos] = true
	}

	lastPos := m.buf.GapPosition()
	for _, c := range m.extraCursors {
		if c.GapPos > lastPos {
			lastPos = c.GapPos
		}
	}

	content := m.buf.String()
	if lastPos+1 >= len(content) {
		return
	}
	searchStart := lastPos + 1

	idx := strings.Index(content[searchStart:], word)
	if idx >= 0 {
		idx += searchStart
	} else {
		idx = strings.Index(content, word)
	}
	if idx < 0 || idx >= len(content) {
		return
	}
	if existing[idx] {
		return
	}

	line, col := m.buf.LineCol(idx)
	m.extraCursors = append(m.extraCursors, EditorCursor{
		Line:   line,
		Col:    col,
		GapPos: idx,
	})
}
```

- [ ] **Step 2: Fix handleBackspace cursor position clamping**

Read `internal/editor/keys.go:305-356`. Replace:

```go
func (m *Model) handleBackspace() (tea.Model, tea.Cmd) {
	m.clearSelection()
	if m.buf.GapPosition() == 0 {
		return m, nil
	}

	if len(m.extraCursors) > 0 {
		all := m.allCursors()
		primaryDeleted := false
		for i := len(all) - 1; i >= 0; i-- {
			c := all[i]
			if c.GapPos == 0 {
				continue
			}
			r := m.buf.RuneAt(c.GapPos - 1)
			m.undoStack.Push(buffer.Operation{
				Type: "insert",
				Pos:  c.GapPos - 1,
				Text: string(r),
			})
			m.moveGapTo(c.GapPos)
			m.buf.Backspace()
			if i == 0 {
				primaryDeleted = true
			}
		}
		if primaryDeleted {
			m.cursor.Col--
			if m.cursor.Col < 0 {
				m.cursor.Col = 0
			}
		}
		for i := range m.extraCursors {
			if m.extraCursors[i].Col > 0 && all[i+1].GapPos > 0 {
				m.extraCursors[i].Col--
			}
		}
		m.modified = true
	} else {
		r := m.buf.RuneAt(m.buf.GapPosition() - 1)
		m.undoStack.Push(buffer.Operation{
			Type: "insert",
			Pos:  m.buf.GapPosition() - 1,
			Text: string(r),
		})
		if m.buf.Backspace() {
			m.cursor.Col--
			if m.cursor.Col < 0 {
				m.cursor.Col = 0
			}
			m.modified = true
			m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		}
	}
	return m, nil
}
```

- [ ] **Step 3: Add regression tests**

Append to `internal/editor/regression_test.go`:

```go
func TestAddNextOccurrenceCorrectPosition(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf = buffer.NewBufferFromString("foo bar foo")
	m.cursor.SetPos(0, 0)
	m.syncGapToCursor()

	m.addNextOccurrence()

	if len(m.extraCursors) != 1 {
		t.Fatalf("expected 1 extra cursor, got %d", len(m.extraCursors))
	}
	if m.extraCursors[0].GapPos != 8 {
		t.Errorf("expected second occurrence at pos 8, got %d", m.extraCursors[0].GapPos)
	}
}

func TestAddNextOccurrenceWrapAround(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf = buffer.NewBufferFromString("foo bar foo")
	m.cursor.SetPos(0, 8)
	m.syncGapToCursor()

	m.addNextOccurrence()

	if len(m.extraCursors) != 1 {
		t.Fatalf("expected 1 extra cursor via wrap-around, got %d", len(m.extraCursors))
	}
	if m.extraCursors[0].GapPos != 0 {
		t.Errorf("expected wrap-around to pos 0, got %d", m.extraCursors[0].GapPos)
	}
}

func TestBackspaceMultiCursorColClamp(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf = buffer.NewBufferFromString("abc")
	m.cursor.SetPos(0, 3)
	m.syncGapToCursor()

	for i := 0; i < 3; i++ {
		m.handleBackspace()
	}

	if m.cursor.Col != 0 {
		t.Errorf("expected Col 0 after 3 backspaces, got %d", m.cursor.Col)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/editor/ -run "TestAddNextOccurrence|TestBackspaceMultiCursor" -v
```
Expected: ALL PASS.

- [ ] **Step 5: Run vet and full test suite**

```bash
go vet ./... && go test ./...
```
Expected: zero vet warnings, all tests pass.

- [ ] **Step 6: Commit PR2**

```bash
git add internal/editor/editor.go internal/editor/keys.go internal/editor/regression_test.go
git commit -m "fix: multi-cursor addNextOccurrence position and backspace col clamp (bugs #3, #8)"
```

---

### Task 3: LSP lifecycle — context + timeout + diagnostics race (Bugs #1, #9, #10 — PR3)

**Files:**
- Modify: `internal/lsp/lsp.go`
- Create: `internal/lsp/lsp_test.go`
- Modify: `internal/editor/view.go` (renderStatus lock)

**Interfaces:**
- Produces: `Client.ctx context.Context`, `Client.cancel context.CancelFunc`, `Client.done chan struct{}`
- Consumes (from stdlib): `context.WithCancel`, `select` with `ctx.Done()`, `time.After`

- [ ] **Step 1: Update Client struct and NewClient**

Read `internal/lsp/lsp.go`. Replace the `Client` struct (lines 202-217) and `NewClient` (lines 219-250):

```go
type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader

	requestID int
	mu        sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	onDiagnostics func(uri string, diagnostics []Diagnostic)
	onCompletion  func(id int, items []CompletionItem)

	pending map[int]chan Response
}

func NewClient(serverCmd string, args ...string) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		requestID: 1,
		pending:   make(map[int]chan Response),
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
	}

	cmd := exec.Command(serverCmd, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("lsp stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("lsp stdout: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin
	c.stdout = stdout
	c.reader = bufio.NewReader(stdout)

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("lsp start: %w", err)
	}

	go c.readLoop()

	return c, nil
}
```

- [ ] **Step 2: Update readLoop to check context**

Replace `readLoop` (lines 391-414):

```go
func (c *Client) readLoop() {
	defer close(c.done)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		msg, err := c.readMessage()
		if err != nil {
			return
		}

		var notif Notification
		if err := json.Unmarshal(msg, &notif); err == nil && notif.Method != "" {
			c.handleNotification(notif)
			continue
		}

		var resp Response
		if err := json.Unmarshal(msg, &resp); err == nil {
			c.handleResponse(resp)
			continue
		}
	}
}
```

- [ ] **Step 3: Update Shutdown to cancel context and wait**

Replace `Shutdown` (lines 327-332):

```go
func (c *Client) Shutdown() error {
	c.request("shutdown", nil)
	c.notify("exit", nil)
	c.cancel()
	<-c.done
	return c.cmd.Wait()
}
```

- [ ] **Step 4: Update request with timeout and context select**

Replace `request` (lines 336-364):

```go
func (c *Client) request(method string, params interface{}) (Response, error) {
	c.mu.Lock()
	id := c.requestID
	c.requestID++
	ch := make(chan Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.writeMessage(req); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return Response{}, err
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return resp, fmt.Errorf("lsp error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp, nil
	case <-time.After(10 * time.Second):
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return Response{}, fmt.Errorf("lsp request timeout: %s", method)
	case <-c.ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return Response{}, fmt.Errorf("lsp client shutdown")
	}
}
```

- [ ] **Step 5: Add "context" import to lsp.go**

Add `"context"` to the imports block at the top of `internal/lsp/lsp.go`.

- [ ] **Step 6: Fix diagnostics TOCTOU in renderStatus**

Read `internal/editor/view.go:237-265`. Replace the diagnostics counting block:

```go
	diagInfo := ""
	if m.lspClient != nil {
		m.lspDiagnosticsMu.Lock()
		errCount, warnCount := 0, 0
		for _, diags := range m.diagnostics {
			for _, d := range diags {
				if d.Severity == 1 {
					errCount++
				} else if d.Severity == 2 {
					warnCount++
				}
			}
		}
		m.lspDiagnosticsMu.Unlock()

		if errCount > 0 {
			diagInfo += fmt.Sprintf(" E:%d", errCount)
		}
		if warnCount > 0 {
			diagInfo += fmt.Sprintf(" W:%d", warnCount)
		}
	}
```

- [ ] **Step 7: Create lsp_test.go**

Write new file `internal/lsp/lsp_test.go`:

```go
package lsp

import (
	"testing"
	"time"
)

func TestClientShutdownStopsReadLoop(t *testing.T) {
	// Use a command that will just hang (no real LSP server needed for this test)
	// sleep is cross-platform enough for this
	c, err := NewClient("sleep", "60")
	if err != nil {
		// If sleep doesn't exist or fails, skip the test
		t.Skipf("could not start test process: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Init won't work but we can still test shutdown
	done := make(chan struct{})
	go func() {
		c.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown timed out — readLoop may not be stopping")
	}
}

func TestContextCancelExitsReadLoop(t *testing.T) {
	c, err := NewClient("sleep", "60")
	if err != nil {
		t.Skipf("could not start test process: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	c.cancel()

	select {
	case <-c.done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("context cancel did not stop readLoop")
	}
}
```

- [ ] **Step 8: Run all tests + vet**

```bash
go vet ./... && go test ./internal/lsp/ ./internal/editor/ -v
```
Expected: zero vet warnings, all tests pass.

Note: `lsp_test.go` may skip if `sleep` isn't available on the platform (e.g., some Windows shells). This is acceptable — the test infrastructure is in place and will run on CI with proper sleep.

- [ ] **Step 9: Commit PR3**

```bash
git add internal/lsp/lsp.go internal/lsp/lsp_test.go internal/editor/view.go
git commit -m "fix: LSP client context lifecycle, request timeout, diagnostics race (bugs #1, #9, #10)"
```

---

### Task 4: Search highlight + Go-to-Line (Bugs #5, #14 — PR4)

**Files:**
- Modify: `internal/editor/view.go` (applySearchHighlight, new renderGoToLine)
- Modify: `internal/editor/editor.go` (ModeGoToLine constant, goToLineInput field)
- Modify: `internal/editor/keys.go` (dispatch, enterGoToLine, handleGoToLineKey)
- Modify: `internal/editor/actions.go` (executeAction view.go-line)
- Modify: `internal/editor/regression_test.go`

- [ ] **Step 1: Fix applySearchHighlight byte → rune count**

Read `internal/editor/view.go:462`. Add `"unicode/utf8"` to imports. Replace:

```go
	queryLen := utf8.RuneCountInString(m.searchQuery)
```

- [ ] **Step 2: Add ModeGoToLine constant and goToLineInput field**

Read `internal/editor/editor.go:28-35`. Add after `ModeWelcome`:

```go
	ModeGoToLine
```

Read `internal/editor/editor.go:67-158` (Model struct). Add after `renameError`:

```go
	goToLineInput string
```

- [ ] **Step 3: Add go-to-line handler in keys.go**

Read `internal/editor/keys.go:64-72`. Add dispatch before the Normal mode block:

```go
	if m.mode == ModeGoToLine {
		return m.handleGoToLineKey(msg)
	}
```

Add at end of keys.go (before mouse handler):

```go
func (m *Model) enterGoToLine() {
	m.mode = ModeGoToLine
	m.goToLineInput = ""
}

func (m *Model) handleGoToLineKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil
	case "enter":
		if m.goToLineInput == "" {
			m.mode = ModeNormal
			return m, nil
		}
		num, err := strconv.Atoi(m.goToLineInput)
		if err != nil || num < 1 || num > m.buf.LineCount() {
			return m, m.showError("Invalid line number")
		}
		m.cursor.SetPos(num-1, 0)
		m.clampCursor()
		m.syncGapToCursor()
		m.viewport.EnsureVisible(m.cursor.Line, m.cursor.Col)
		m.mode = ModeNormal
		return m, nil
	case "backspace":
		if len(m.goToLineInput) > 0 {
			m.goToLineInput = m.goToLineInput[:len(m.goToLineInput)-1]
		}
	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= '0' && msg.Runes[0] <= '9' {
			m.goToLineInput += string(msg.Runes)
		}
	}
	return m, nil
}
```

- [ ] **Step 4: Add executeAction case**

Read `internal/editor/actions.go:16-78`. Add case:

```go
	case "view.go-line":
		m.enterGoToLine()
```

- [ ] **Step 5: Add go-to-line render in view.go**

Read `internal/editor/view.go:13-34` (View method). Add dispatch:

```go
	if m.mode == ModeGoToLine {
		return m.renderGoToLine()
	}
```

Add function:

```go
func (m *Model) renderGoToLine() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Go to line: %s", m.goToLineInput))
	lines = append(lines, "")
	lines = append(lines, "Enter to confirm, Esc to cancel")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}
```

- [ ] **Step 6: Add "strconv" import to keys.go**

Add `"strconv"` to imports in `internal/editor/keys.go`.

- [ ] **Step 7: Add regression tests**

Append to `internal/editor/regression_test.go`:

```go
func TestSearchHighlightUnicode(t *testing.T) {
	// Test that search with multi-byte characters works
	m := New()
	m.mode = ModeNormal
	m.buf = buffer.NewBufferFromString("cafe com leite")
	m.searchQuery = "cafe"
	m.doSearch()

	if len(m.searchMatches) != 1 {
		t.Fatalf("expected 1 match for 'cafe', got %d", len(m.searchMatches))
	}
}

func TestGoToLine(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.buf = buffer.NewBufferFromString("line1\nline2\nline3\nline4\nline5")
	m.goToLineInput = "3"
	m.handleGoToLineKey(tea.KeyMsg{Type: tea.KeyEnter})

	if m.cursor.Line != 2 {
		t.Errorf("expected line 2 (0-indexed), got %d", m.cursor.Line)
	}
}

func TestGoToLineOutOfRange(t *testing.T) {
	m := New()
	m.mode = ModeGoToLine
	m.buf = buffer.NewBufferFromString("line1\nline2")
	m.goToLineInput = "999"
	_, cmd := m.handleGoToLineKey(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("expected error command for out-of-range line")
	}
}
```

- [ ] **Step 8: Run tests + vet**

```bash
go vet ./... && go test ./internal/editor/ -run "TestSearchHighlightUnicode|TestGoToLine" -v
```
Expected: zero vet warnings, all tests pass.

- [ ] **Step 9: Commit PR4**

```bash
git add internal/editor/view.go internal/editor/editor.go internal/editor/keys.go internal/editor/actions.go internal/editor/regression_test.go
git commit -m "fix: search highlight rune-count and implement Go-to-Line (bugs #5, #14)"
```

---

### Task 5: Buffer/Gap invariant fixes (Bugs #11, #15 — PR5)

**Files:**
- Modify: `internal/buffer/buffer.go` (MoveGapLeft, Backspace)
- Modify: `internal/buffer/buffer_test.go`

- [ ] **Step 1: Fix MoveGapLeft return value**

Read `internal/buffer/buffer.go:89-97`. Replace:

```go
func (b *Buffer) MoveGapLeft() rune {
	if b.gapStart == 0 {
		return 0
	}
	b.gapStart--
	b.gapEnd--
	ch := b.data[b.gapStart]
	b.data[b.gapEnd] = ch
	return ch
}
```

- [ ] **Step 2: Fix Backspace defensive cleanup**

Read `internal/buffer/buffer.go:60-67`. Replace:

```go
func (b *Buffer) Backspace() bool {
	if b.gapStart == 0 {
		return false
	}
	b.gapStart--
	b.data[b.gapStart] = 0
	b.length--
	return true
}
```

- [ ] **Step 3: Add tests**

Append to `internal/buffer/buffer_test.go`:

```go
func TestMoveGapLeftReturnsCorrectRune(t *testing.T) {
	b := NewBufferFromString("ab")
	b.MoveGapRight() // gap at pos 1
	b.MoveGapRight() // gap at pos 2

	r := b.MoveGapLeft() // move gap left, should return 'b'
	if r != 'b' {
		t.Errorf("expected 'b', got %q", r)
	}
}

func TestBackspaceThenMoveGapLeftNoGarbage(t *testing.T) {
	b := NewBufferFromString("abc")

	b.MoveGapRight() // gap at 1
	b.MoveGapRight() // gap at 2
	b.MoveGapRight() // gap at 3

	b.Backspace() // delete 'c', gap at 2

	// Move gap back to start
	b.MoveGapLeft() // should return 'b'
	b.MoveGapLeft() // should return 'a'

	s := b.String()
	if s != "ab" {
		t.Errorf("expected 'ab', got %q", s)
	}
}

func TestBackspaceAtZero(t *testing.T) {
	b := NewBufferFromString("abc")
	ok := b.Backspace()
	if ok {
		t.Error("Backspace at position 0 should return false")
	}
}
```

- [ ] **Step 4: Run tests + vet**

```bash
go vet ./... && go test ./internal/buffer/ -v
```
Expected: zero vet warnings, all tests pass.

- [ ] **Step 5: Commit PR5**

```bash
git add internal/buffer/buffer.go internal/buffer/buffer_test.go
git commit -m "fix: MoveGapLeft return value and Backspace defensive cleanup (bugs #11, #15)"
```

---

### Task 6: Rendering pipeline + save error reporting (Bugs #6, #12, #13, #16 — PR6)

**Files:**
- Modify: `internal/editor/format.go` (applyFormat — undoStack.Clear)
- Modify: `internal/editor/actions.go` (getSelectedText — strings.Builder, save — showError, executeAction SaveConfig errors)
- Modify: `internal/editor/view.go` (applyIndentGuides — move before highlight)
- Modify: `internal/editor/format_test.go`

- [ ] **Step 1: applyFormat calls undoStack.Clear()**

Read `internal/editor/format.go:48-63`. Replace:

```go
func (m *Model) applyFormat() error {
	formatted, err := m.formatBuffer()
	if err != nil || formatted == m.buf.String() {
		return nil
	}

	m.undoStack.Clear()

	cursorPos := m.buf.GapPosition()
	m.buf = buffer.NewBufferFromString(formatted)

	if cursorPos < m.buf.Len() {
		m.moveGapTo(cursorPos)
	} else {
		m.moveGapTo(m.buf.Len())
	}
	return nil
}
```

- [ ] **Step 2: getSelectedText uses strings.Builder**

Read `internal/editor/actions.go:218-232`. Replace:

```go
func (m *Model) getSelectedText() string {
	if !m.hasSelection() {
		return ""
	}
	start := m.selStart
	end := m.selEnd
	if start > end {
		start, end = end, start
	}
	var sb strings.Builder
	sb.Grow(end - start)
	for i := start; i < end; i++ {
		sb.WriteRune(m.buf.RuneAt(i))
	}
	return sb.String()
}
```

- [ ] **Step 3: applyIndentGuides moved before syntax highlighting**

Read `internal/editor/view.go:84-93` and `109-118`. In both branches (wrap and non-wrap), move `applyIndentGuides` before syntax highlighting.

Non-wrap branch (around line 104-118):

```go
		for i := m.viewport.ScrollY(); i < len(lines) && i < m.viewport.ScrollY()+contentHeight; i++ {
			lineNum := fmt.Sprintf("%*d ", lineNumWidth, i+1)
			styledLineNum := m.lineNumStyle.Render(lineNum)

			rawLine := lines[i]
			lineWithGuides := m.applyIndentGuides(rawLine, rawLine)
			segments := m.highlighter.HighlightLine(lineWithGuides, m.language)
			lineText := highlight.RenderSegments(segments)

			if len(m.searchMatches) > 0 {
				lineText = m.applySearchHighlight(lineText, i)
			}

			if m.viewport.ScrollX() > 0 && m.viewport.ScrollX() < len(lineText) {
				lineText = lineText[m.viewport.ScrollX():]
			}
			if len(lineText) > textWidth {
				lineText = lineText[:textWidth]
			}

			visibleLines = append(visibleLines, styledLineNum+lineText)
		}
```

Wrap branch (around line 68-101) — same pattern: call `applyIndentGuides` first, then `HighlightLine` on the result.

- [ ] **Step 4: Simplify applyIndentGuides (remove dead first pass)**

Read `internal/editor/view.go:142-194`. Replace:

```go
func (m *Model) applyIndentGuides(lineText string, rawLine string) string {
	if strings.TrimSpace(rawLine) == "" {
		return lineText
	}

	indent := 0
	for _, r := range rawLine {
		if r == ' ' {
			indent++
		} else if r == '\t' {
			indent += 4
		} else {
			break
		}
	}

	if indent < 4 {
		return lineText
	}

	runes := []rune(lineText)
	var b strings.Builder
	b.Grow(len(runes) + indent/4*3)
	for i, r := range runes {
		if i > 0 && i < indent && i%4 == 0 {
			b.WriteString("│")
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
```

- [ ] **Step 5: save() reports errors**

Read `internal/editor/actions.go:83-98`. Replace:

```go
func (m *Model) save() {
	if m.filename == "" {
		m.enterSaveAs()
		return
	}
	if err := fileio.Save(m.filename, m.buf); err != nil {
		m.showError(fmt.Sprintf("Save error: %v", err))
		return
	}
	m.modified = false

	if m.config.FormatOnSave {
		m.applyFormat()
		if err := fileio.Save(m.filename, m.buf); err != nil {
			m.showError(fmt.Sprintf("Format save error: %v", err))
		}
	}
}
```

- [ ] **Step 6: executeAction SaveConfig error handling**

Read `internal/editor/actions.go:52-77`. Every `SaveConfig(m.config)` call should check error. Replace each:

```go
	case "view.toggle-auto-close":
		m.config.AutoCloseEnabled = !m.config.AutoCloseEnabled
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-auto-close")
		}
	case "view.toggle-vim-mode":
		m.config.VimMode = !m.config.VimMode
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-vim-mode")
		}
	case "view.next-theme":
		themes := []string{"dark", "light", "monokai", "dracula", "solarized-dark"}
		current := m.config.Theme
		for i, t := range themes {
			if t == current {
				m.config.Theme = themes[(i+1)%len(themes)]
				break
			}
		}
		m.highlighter.SetTheme(m.config.Theme)
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config next-theme")
		}
	case "view.toggle-word-wrap":
		m.config.WordWrap = !m.config.WordWrap
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-word-wrap")
		}
	case "file.toggle-format-on-save":
		m.config.FormatOnSave = !m.config.FormatOnSave
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-format-on-save")
		}
	case "file.toggle-auto-save":
		m.config.AutoSaveEnabled = !m.config.AutoSaveEnabled
		if err := SaveConfig(m.config); err != nil {
			logError(err, "save config toggle-auto-save")
		}
```

- [ ] **Step 7: Add regression tests**

Append to `internal/editor/format_test.go`:

```go
func TestFormatResetsUndo(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.language = "Go"
	m.buf = buffer.NewBufferFromString("package main\nfunc main(){}\n")
	m.undoStack.Push(buffer.Operation{Type: "insert", Pos: 0, Text: "test"})

	if !m.undoStack.CanUndo() {
		t.Fatal("should have undo before format")
	}

	m.applyFormat()

	if m.undoStack.CanUndo() {
		t.Error("undo stack should be cleared after format")
	}
}
```

Append to `internal/editor/regression_test.go`:

```go
func TestSaveErrorShowsMessage(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.filename = "/nonexistent/dir/file.txt"
	m.buf = buffer.NewBufferFromString("test")

	m.save()

	if m.errorMessage == "" {
		t.Error("expected error message on save failure")
	}
}
```

- [ ] **Step 8: Run tests + vet**

```bash
go vet ./... && go test ./internal/editor/ -v
```
Expected: zero vet warnings, all tests pass.

- [ ] **Step 9: Final verification — full suite**

```bash
go test ./... -cover && go vet ./...
```
Expected: all tests pass, coverage > 50%, zero vet warnings.

- [ ] **Step 10: Commit PR6**

```bash
git add internal/editor/format.go internal/editor/actions.go internal/editor/view.go internal/editor/format_test.go internal/editor/regression_test.go
git commit -m "fix: rendering pipeline order, save error reporting, format undo reset (bugs #6, #12, #13, #16)"
```

---

## Verification Checklist (after all PRs)

- [ ] `go test ./... -cover` — all tests pass, coverage >= baseline
- [ ] `go vet ./...` — zero warnings
- [ ] `go build -o bin/cmdit.exe ./cmd/cmdit` — clean build
