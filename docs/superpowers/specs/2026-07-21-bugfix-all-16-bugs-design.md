# Design: Correcao dos 16 Bugs — cmdit v0.4.3

Data: 2026-07-21 | Escopo: Todos os 16 bugs do BUG_REPORT.md | Testes: Regressao por bug

---

## Arquitetura das Mudancas

```
internal/
├── buffer/
│   ├── buffer.go       # PR5: Backspace/MoveGapLeft fixes
│   └── undo.go         # PR1: CompositeOperation, Undo/Redo []Operation
├── editor/
│   ├── actions.go      # PR1: doReplace undo, cut undo; PR6: save/showError
│   ├── editor.go       # PR1: insertTextAtAllCursors composite; PR2: addNextOccurrence fix
│   ├── format.go       # PR6: applyFormat calls undoStack.Clear()
│   ├── keys.go         # PR2: handleBackspace cursor fix; PR4: go-to-line handler
│   ├── view.go         # PR4: applySearchHighlight rune-count; PR6: indent guides order, getSelectedText Builder
│   └── lsp_integration.go  # PR3: context-based lifecycle
├── lsp/
│   └── lsp.go          # PR3: context, timeout, readLoop ctx
└── tests novos em cada pacote
```

---

## PR 1: Integridade do Undo (Bugs #2, #4, #7)

### 1.1 CompositeOperation no undo stack

**Arquivo**: `internal/buffer/undo.go`

Adicionar tipo e metodos:

```go
func (u *UndoStack) PushComposite(ops []Operation) {
    u.stack = u.stack[:u.position]
    u.stack = append(u.stack, Operation{Type: "begin-composite"})
    u.stack = append(u.stack, ops...)
    u.stack = append(u.stack, Operation{Type: "end-composite"})
    u.position = len(u.stack)
}
```

**Mudanca de assinatura**: `Undo() ([]Operation, bool)` e `Redo() ([]Operation, bool)` — retornam slice em vez de `Operation` singular. Internamente, detectam `"begin-composite"`/`"end-composite"` e coletam todas as operacoes do grupo, devolvendo-as em ordem de aplicacao (LIFO revertido para undo, FIFO para redo).

**Undo** — quando encontra `"end-composite"`, coleta todas as ops ate `"begin-composite"` e as reverte (inverte ordem):

```go
func (u *UndoStack) Undo() ([]Operation, bool) {
    if !u.CanUndo() { return nil, false }
    u.position--
    op := u.stack[u.position]
    if op.Type == "end-composite" {
        var ops []Operation
        for u.position > 0 {
            u.position--
            inner := u.stack[u.position]
            if inner.Type == "begin-composite" { break }
            ops = append(ops, inner)
        }
        reverse(ops)
        return ops, true
    }
    if op.Type == "begin-composite" {
        // Em caso de corrupcao, pular
        return u.Undo()
    }
    return []Operation{op}, true
}
```

**Redo** — similar, mas sem inverter a ordem (FIFO):

```go
func (u *UndoStack) Redo() ([]Operation, bool) {
    if !u.CanRedo() { return nil, false }
    op := u.stack[u.position]
    u.position++
    if op.Type == "begin-composite" {
        var ops []Operation
        for u.position < len(u.stack) {
            inner := u.stack[u.position]
            u.position++
            if inner.Type == "end-composite" { break }
            ops = append(ops, inner)
        }
        return ops, true
    }
    return []Operation{op}, true
}
```

**Impacto nos call sites**: `undo()` e `redo()` em `actions.go` precisam iterar `[]Operation` em vez de `Operation` singular.

### 1.2 insertTextAtAllCursors com composite

**Arquivo**: `internal/editor/editor.go:336-362`

Substituir loop de `m.undoStack.Push` individuais por coleta + `PushComposite`:

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
            Type: "delete", Pos: all[i].GapPos, Text: text,
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

### 1.3 doReplace com undo

**Arquivo**: `internal/editor/actions.go:294-308`

Adicionar push antes da delecao:

```go
func (m *Model) doReplace() {
    if len(m.searchMatches) == 0 || m.replaceQuery == "" { return }
    pos := m.searchMatches[m.searchCurrent]
    originalText := m.searchQuery
    m.undoStack.Push(buffer.Operation{
        Type: "insert", Pos: pos, Text: originalText,
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

### 1.4 cut() sem selecao com undo

**Arquivo**: `internal/editor/actions.go:186-194`

```go
// Branch "sem selecao":
m.clipboard.Copy(m.currentLineText())
lineStart := m.buf.LineStart(m.cursor.Line)
lineEnd := m.buf.LineStart(m.cursor.Line + 1)
var sb strings.Builder
for i := lineStart; i < lineEnd && i < m.buf.Len(); i++ {
    sb.WriteRune(m.buf.RuneAt(i))
}
m.undoStack.Push(buffer.Operation{
    Type: "insert", Pos: lineStart, Text: sb.String(),
})
m.moveGapTo(lineStart)
for i := lineStart; i < lineEnd && m.buf.GapPosition() < m.buf.Len(); i++ {
    m.buf.DeleteForward()
}
m.modified = true
```

### 1.5 undo/redo call sites

**Arquivo**: `internal/editor/actions.go:126-166`

```go
func (m *Model) undo() {
    ops, ok := m.undoStack.Undo()
    if !ok { return }
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
    if !ok { return }
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

### Testes PR1

| Teste | Arquivo | Descricao |
|-------|---------|-----------|
| `TestPushCompositeUndoRedo` | undo_test.go | PushComposite com 3 ops, Undo reverte todas, Redo reaplica |
| `TestMultiCursorUndoAtomico` | regression_test.go | 3 cursores, digita "x", Ctrl+Z restaura estado original |
| `TestReplaceUndo` | regression_test.go | Replace, Ctrl+Z, texto original restaurado |
| `TestCutLineUndo` | regression_test.go | Ctrl+X sem selecao, Ctrl+Z, linha restaurada |
| `TestUndoRedoCompositeStack` | undo_test.go | Push + PushComposite + Push, verificar stack integrity |

---

## PR 2: Multi-cursor (Bugs #3, #8)

### 2.1 addNextOccurrence — calculo de posicao

**Arquivo**: `internal/editor/editor.go:634-696`

Substituir linhas 667-675 (calculo convoluto) por logica direta:

```go
func (m *Model) addNextOccurrence() {
    word := m.wordAtCursor()
    if word == "" { return }

    existing := make(map[int]bool)
    existing[m.buf.GapPosition()] = true
    for _, c := range m.extraCursors {
        existing[c.GapPos] = true
    }

    lastPos := m.buf.GapPosition()
    for _, c := range m.extraCursors {
        if c.GapPos > lastPos { lastPos = c.GapPos }
    }

    content := m.buf.String()
    searchStart := lastPos + 1

    // Busca apos a ultima posicao
    idx := strings.Index(content[searchStart:], word)
    if idx >= 0 {
        idx += searchStart // converte para posicao absoluta
    } else {
        // Wrap around: busca do inicio
        idx = strings.Index(content, word)
    }
    if idx < 0 || idx >= len(content) { return }
    if existing[idx] { return }

    line, col := m.buf.LineCol(idx)
    m.extraCursors = append(m.extraCursors, EditorCursor{
        Line: line, Col: col, GapPos: idx,
    })
}
```

### 2.2 handleBackspace — cursor decrement condicional

**Arquivo**: `internal/editor/keys.go:305-356`

Mover `m.cursor.Col--` e `extraCursors[i].Col--` para dentro do loop condicional, apenas para cursores que realmente deletaram:

```go
func (m *Model) handleBackspace() (tea.Model, tea.Cmd) {
    m.clearSelection()
    if m.buf.GapPosition() == 0 { return m, nil }

    if len(m.extraCursors) > 0 {
        all := m.allCursors()
        for i := len(all) - 1; i >= 0; i-- {
            c := all[i]
            if c.GapPos == 0 { continue }
            r := m.buf.RuneAt(c.GapPos - 1)
            m.undoStack.Push(buffer.Operation{
                Type: "insert", Pos: c.GapPos - 1, Text: string(r),
            })
            m.moveGapTo(c.GapPos)
            m.buf.Backspace()
        }
        // Ajustar posicoes apenas dos cursores que estavam no gap (todos no mesmo gap pos)
        m.cursor.Col--
        if m.cursor.Col < 0 { m.cursor.Col = 0 }
        for i := range m.extraCursors {
            m.extraCursors[i].Col--
            if m.extraCursors[i].Col < 0 { m.extraCursors[i].Col = 0 }
        }
        m.modified = true
    } else {
        // ... (inalterado)
    }
    return m, nil
}
```

**Nota**: Como o gap so tem uma posicao e `syncGapToCursor` foi chamado antes, todos os cursores operam no mesmo `GapPos` -- entao se o primeiro cursor deletou, o gap moveu. Isso significa que o codigo atual so funciona corretamente com 1 cursor. O fix real requer reposicionar o gap para cada cursor individualmente. Vou implementar com `moveGapTo` antes de cada backspace no loop, similar ao que `insertTextAtAllCursors` faz.

### Testes PR2

| Teste | Arquivo | Descricao |
|-------|---------|-----------|
| `TestAddNextOccurrenceCorrectPosition` | regression_test.go | Ctrl+D acha ocorrencia correta e wrap-around funciona |
| `TestAddNextOccurrenceNoWrapDuplicate` | regression_test.go | Ctrl+D wrap-around nao duplica cursor existente |
| `TestBackspaceMultiCursorColClamp` | regression_test.go | Backspace com multi-cursor nao produz Col < 0 |
| `TestMultiCursorInsertThenBackspace` | regression_test.go | Insere com 3 cursores, backspace, posicoes corretas |

---

## PR 3: LSP (Bugs #1, #9, #10)

### 3.1 Client com context + done channel

**Arquivo**: `internal/lsp/lsp.go`

Adicionar `ctx context.Context` e `cancel context.CancelFunc` ao `Client`:

```go
type Client struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.ReadCloser
    reader *bufio.Reader
    requestID int
    mu     sync.Mutex
    ctx    context.Context
    cancel context.CancelFunc
    done   chan struct{} // fechado quando readLoop sai

    onDiagnostics func(uri string, diagnostics []Diagnostic)
    onCompletion  func(id int, items []CompletionItem)
    pending map[int]chan Response
}
```

**NewClient**:

```go
func NewClient(serverCmd string, args ...string) (*Client, error) {
    ctx, cancel := context.WithCancel(context.Background())
    c := &Client{
        requestID: 1,
        pending:   make(map[int]chan Response),
        ctx:       ctx,
        cancel:    cancel,
        done:      make(chan struct{}),
    }
    // ... cmd setup ...
    go c.readLoop()
    return c, nil
}
```

**readLoop com ctx**:

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
        // ... parse and dispatch ...
    }
}
```

**Shutdown com cancel**:

```go
func (c *Client) Shutdown() error {
    c.request("shutdown", nil)
    c.notify("exit", nil)
    c.cancel()
    <-c.done // espera readLoop sair
    return c.cmd.Wait()
}
```

### 3.2 request com timeout

```go
func (c *Client) request(method string, params interface{}) (Response, error) {
    c.mu.Lock()
    id := c.requestID
    c.requestID++
    ch := make(chan Response, 1)
    c.pending[id] = ch
    c.mu.Unlock()

    req := Request{JSONRPC: "2.0", ID: id, Method: method, Params: params}
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

### 3.3 Diagnostics: unico lock para contagem

**Arquivo**: `internal/editor/view.go:237-265`

Corrigir TOCTOU — unico bloco de lock:

```go
diagInfo := ""
if m.lspClient != nil {
    m.lspDiagnosticsMu.Lock()
    errCount, warnCount := 0, 0
    for _, diags := range m.diagnostics {
        for _, d := range diags {
            if d.Severity == 1 { errCount++ }
            else if d.Severity == 2 { warnCount++ }
        }
    }
    m.lspDiagnosticsMu.Unlock()
    if errCount > 0 { diagInfo += fmt.Sprintf(" E:%d", errCount) }
    if warnCount > 0 { diagInfo += fmt.Sprintf(" W:%d", warnCount) }
}
```

### Testes PR3

| Teste | Arquivo | Descricao |
|-------|---------|-----------|
| `TestLSPClientShutdownStopsReadLoop` | lsp_test.go (novo) | Shutdown fecha done channel, readLoop termina |
| `TestLSPRequestTimeout` | lsp_test.go | Request com server lento retorna timeout |
| `TestLSPContextCancelExitsReadLoop` | lsp_test.go | ctx.Cancel faz readLoop retornar |
| `TestDiagnosticsConsistentRead` | lsp_integration_test.go (novo) | Race entre onDiagnostics e renderStatus |

---

## PR 4: Search/Replace Highlight + Go-to-Line (Bugs #5, #14)

### 4.1 applySearchHighlight — rune-count

**Arquivo**: `internal/editor/view.go:462`

```go
// Antes:
queryLen := len(m.searchQuery)

// Depois:
queryLen := utf8.RuneCountInString(m.searchQuery)
```

### 4.2 Go-to-Line

**Arquivos**: `internal/editor/keys.go` (novo handler), `internal/editor/actions.go` (executeAction)

**Novo modo** `ModeGoToLine` no editor.go:

```go
const (
    ModeNormal Mode = iota
    ModeConfirm
    ModeSearch
    ModeReplace
    ModePalette
    ModeFilePicker
    ModeSaveAs
    ModeRename
    ModeWelcome
    ModeGoToLine   // NOVO
)
```

**Handler** em keys.go:

```go
case "view.go-line":
    m.enterGoToLine()

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

**View**: adicionar `renderGoToLine()` similar ao `renderSearchBar()`.

**Model**: adicionar campo `goToLineInput string`.

**executeAction**: adicionar `case "view.go-line": m.enterGoToLine()`.

**keys.go dispatch**: adicionar `if m.mode == ModeGoToLine { return m.handleGoToLineKey(msg) }`.

### Testes PR4

| Teste | Arquivo | Descricao |
|-------|---------|-----------|
| `TestSearchHighlightUnicode` | regression_test.go | Busca com "cafe" (acento), highlight correto |
| `TestGoToLine` | regression_test.go | Ctrl+G 5, cursor vai para linha 5 |
| `TestGoToLineOutOfRange` | regression_test.go | Linha 999 em arquivo de 10 linhas mostra erro |
| `TestGoToLineNegative` | regression_test.go | Valor negativo mostra erro |

---

## PR 5: Buffer/Gap (Bugs #11, #15)

### 5.1 MoveGapLeft — valor de retorno correto

**Arquivo**: `internal/buffer/buffer.go:89-97`

```go
func (b *Buffer) MoveGapLeft() rune {
    if b.gapStart == 0 { return 0 }
    b.gapStart--
    b.gapEnd--
    ch := b.data[b.gapStart] // salva antes de sobrescrever
    b.data[b.gapEnd] = ch
    return ch
}
```

### 5.2 Backspace — limpeza defensiva do gap

```go
func (b *Buffer) Backspace() bool {
    if b.gapStart == 0 { return false }
    b.gapStart--
    b.data[b.gapStart] = 0 // limpa valor obsoleto no gap
    b.length--
    return true
}
```

### Testes PR5

| Teste | Arquivo | Descricao |
|-------|---------|-----------|
| `TestMoveGapLeftReturnsCorrectRune` | buffer_test.go | MoveGapLeft retorna o rune correto que foi movido |
| `TestBackspaceThenMoveGapLeftNoGarbage` | buffer_test.go | Backspace + MoveGapLeft nao propaga lixo |

---

## PR 6: Rendering + Save Errors (Bugs #6, #12, #13, #16)

### 6.1 applyFormat — Clear no undo stack

**Arquivo**: `internal/editor/format.go:48-63`

```go
func (m *Model) applyFormat() error {
    formatted, err := m.formatBuffer()
    if err != nil || formatted == m.buf.String() { return nil }

    m.undoStack.Clear() // explicito: format reseta historico

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

### 6.2 getSelectedText — strings.Builder

**Arquivo**: `internal/editor/actions.go:218-232`

```go
func (m *Model) getSelectedText() string {
    if !m.hasSelection() { return "" }
    start, end := m.selStart, m.selEnd
    if start > end { start, end = end, start }
    var sb strings.Builder
    sb.Grow(end - start)
    for i := start; i < end; i++ {
        sb.WriteRune(m.buf.RuneAt(i))
    }
    return sb.String()
}
```

### 6.3 applyIndentGuides — ordem na pipeline

**Arquivo**: `internal/editor/view.go:84-93` e `109-118`

Mover `applyIndentGuides` para ANTES da syntax highlighting. No `renderContent`:

```go
// Antes:
segments := m.highlighter.HighlightLine(lines[i], m.language)
lineText := highlight.RenderSegments(segments)
lineText = m.applyIndentGuides(lineText, lines[i])

// Depois:
lineText := m.applyIndentGuides(lines[i], lines[i]) // raw text
segments := m.highlighter.HighlightLine(lineText, m.language)
lineText = highlight.RenderSegments(segments)
```

A funcao `applyIndentGuides` agora recebe e retorna texto puro (sem ANSI):

```go
func (m *Model) applyIndentGuides(lineText string, rawLine string) string {
    if strings.TrimSpace(rawLine) == "" { return lineText }
    indent := 0
    for _, r := range rawLine {
        if r == ' ' { indent++ } else if r == '\t' { indent += 4 } else { break }
    }
    if indent < 4 { return lineText }

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

### 6.4 save() e SaveConfig — showError nos erros

**Arquivo**: `internal/editor/actions.go:83-98`

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

**executeAction** calls `SaveConfig` sem verificar erros — adicionar:

```go
case "view.toggle-auto-close":
    m.config.AutoCloseEnabled = !m.config.AutoCloseEnabled
    if err := SaveConfig(m.config); err != nil {
        logError(err, "save config")
    }
```

Repetir padrao para todos os `SaveConfig` em actions.go (linhas 53, 56, 68, 71, 74, 77).

### Testes PR6

| Teste | Arquivo | Descricao |
|-------|---------|-----------|
| `TestFormatResetsUndo` | format_test.go | Format on save, undo stack vazio depois |
| `TestGetSelectedTextPerformance` | bench_test.go | Select all em buffer grande usa Builder |
| `TestIndentGuidesBeforeHighlight` | regression_test.go | Indent guide visivel sem quebrar cor da syntax |
| `TestSaveErrorShowsMessage` | regression_test.go | Save em diretorio readonly mostra mensagem |

---

## Ordem de Implementacao

```
PR1 (undo integrity)     — base para PR2, PR4
  ├── PR2 (multi-cursor)  — depende do modelo composite do PR1
  ├── PR3 (LSP)           — independente, mas usa mesmo branch
  ├── PR4 (search + goto) — Bug #4 resolvido no PR1, bug #5 independente
  ├── PR5 (buffer/gap)    — independente, mudancas triviais
  └── PR6 (render + save) — independente, varias mudancas pequenas
```

Cada PR:
- 1 commit por bug (ou 1 commit por arquivo se varios bugs no mesmo topico)
- `fix:` prefixo no commit message
- Referencia ao numero do bug do BUG_REPORT.md

---

## Verificacao final

Apos todos os PRs mergeados:
1. `go test ./... -cover` — cobertura deve subir (novos testes)
2. `go vet ./...` — zero warnings
3. Build cross-platform: `make build-all`
