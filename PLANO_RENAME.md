# PLANO — Renomear Arquivo (file.rename)

> **Feature:** Renomear o arquivo atualmente aberto no editor  
> **Atalho:** `F2` (padrão multiplataforma)  
> **Palette:** `file.rename` — "Renomear arquivo"  
> **Versão alvo:** v0.2.0  
> **Data do plano:** 2026-04-26

---

## 🎯 Objetivo

Permitir que o usuário renomeie o arquivo aberto de forma rápida, modeless e intuitiva — sem sair do editor, sem diálogos modais, seguindo o padrão CUA (F2 = Rename).

---

## 🧠 Decisões de Design

### Por que F2?

| SO | Atalho padrão de rename | Funciona no terminal? |
|----|--------------------------|------------------------|
| **Windows** | F2 (Explorer, VS Code, Notepad++) | ✅ Sim |
| **macOS** | Enter (Finder), mas Enter = newline no editor | ❌ Enter conflita |
| **Linux** | F2 (Nautilus, Dolphin, Thunar) | ✅ Sim |

**Conclusão:** `F2` é o único atalho consistente nas 3 plataformas que não conflita com a edição de texto.

### Por que input inline (e não diálogo modal)?

O cmdit é **modeless** por filosofia. Um diálogo modal (popup no centro da tela) quebraria o fluxo. O input inline na barra inferior (igual ao Search/Replace atual) mantém:

- Consistência visual com os modos `ModeSearch` / `ModeReplace` / `ModeSaveAs`
- O conteúdo do arquivo visível durante a operação
- Escape (`Esc`) para cancelar a qualquer momento

### Comportamento

```
┌─────────────────────────────────────────────────────────────┐
│  1  │ package main                                          │
│  2  │                                                       │
│  3  │ func main() {                                         │
│  ...│   ...                                                 │
│     │                                                       │
├─────────────────────────────────────────────────────────────┤
│  Renomear: main.go → [main_v2.go                    ]       │
└─────────────────────────────────────────────────────────────┘
```

1. Usuário pressiona `F2` (ou seleciona "Renomear arquivo" na palette)
2. Barra inferior mostra: `Renomear: <nome atual> → [__________]`
3. Campo pré-preenchido com o nome base do arquivo atual
4. Usuário edita o nome
5. `Enter` → executa rename, atualiza `filename`, volta ao modo normal
6. `Esc` → cancela, volta ao modo normal

---

## 🔀 Edge Cases (tabela de decisão)

| Cenário | Comportamento |
|---------|---------------|
| **Arquivo nunca salvo** (`filename == ""`) | Mostrar mensagem: "Nenhum arquivo para renomear. Salve primeiro (Ctrl+S)." Não entrar no modo rename. |
| **Buffer com alterações não salvas** | Auto-save antes de renomear (consistente com auto-save). Se save falhar, abortar rename com erro. |
| **Nome vazio** | Mostrar mensagem de erro: "Nome não pode estar vazio." Continuar no modo rename. |
| **Nome igual ao atual** | Sair do modo rename sem fazer nada (no-op). |
| **Destino já existe** | Mostrar mensagem de erro: "Arquivo já existe: <nome>". Continuar no modo rename. |
| **Caracteres inválidos** (`\ / : * ? " < > \|`) | Mostrar mensagem de erro: "Nome contém caracteres inválidos: <chars>". Continuar no modo rename. |
| **Erro de permissão** (os.Rename falha) | Mostrar mensagem de erro: "Erro ao renomear: <erro>". Continuar no modo rename. |
| **Caminho relativo vs absoluto** | Rename é sempre no mesmo diretório. Apenas o nome base muda. `filepath.Dir(currentPath)` + novo nome. |

---

## 📁 Arquivos a Modificar / Criar

| Arquivo | Ação | Descrição |
|---------|------|-----------|
| `internal/fileio/fileio.go` | ✏️ Modificar | Adicionar `func Rename(oldPath, newPath string) error` |
| `internal/fileio/fileio_test.go` | ✏️ Modificar | Testes para `Rename()` |
| `internal/editor/editor.go` | ✏️ Modificar | Novo modo `ModeRename`, campos, handlers, render, registro |
| `internal/editor/editor_test.go` | ✏️ Modificar | Testes de integração para rename |
| `internal/editor/rename.go` | ✨ Novo | Handlers e render específicos do rename (manter editor.go enxuto) |
| `test-automation/TEST_LOG.md` | ✏️ Modificar | Registrar resultados dos testes |
| `CHANGELOG.md` | ✏️ Modificar | Registrar feature na v0.2.0 |
| `README.md` | ✏️ Modificar | Atualizar lista de atalhos |
| `PLANO.md` | ✏️ Modificar | Atualizar tabela de atalhos e fase 6 |

---

## 🔧 Implementação — Passo a Passo

### Step 1: `fileio.Rename()`

```go
// internal/fileio/fileio.go

// Rename renames a file from oldPath to newPath.
// Returns an error if the source doesn't exist or the operation fails.
func Rename(oldPath, newPath string) error {
    return os.Rename(oldPath, newPath)
}
```

### Step 2: Testes do `fileio.Rename()`

Casos de teste em `fileio_test.go`:

- ✅ `TestRenameSuccess` — cria arquivo, renomeia, verifica antigo não existe, novo existe com mesmo conteúdo
- ✅ `TestRenameSourceNotFound` — renomear arquivo inexistente → erro
- ✅ `TestRenameDestinationExists` — renomear para nome já existente → erro (no Windows; no Unix sobrescreve silenciosamente)
- ✅ `TestRenameEmptyName` — nome vazio → erro

### Step 3: Novo modo `ModeRename` + campos no Model

```go
// internal/editor/editor.go

const (
    // ... existing modes ...
    ModeRename  // NEW
)

type Model struct {
    // ... existing fields ...
    
    // Rename state
    renameInput string  // NEW
    renameError string  // NEW (mensagem de erro para exibir na barra)
}
```

### Step 4: Registro da ação `file.rename`

```go
// internal/editor/editor.go — registerActions()

{ID: "file.rename", Label: "Renomear arquivo", Shortcut: "F2"},
```

### Step 5: Handlers de teclado

**`handleKey()` — modo normal:**

```go
case "f2":
    m.enterRename()
    return m, nil
```

**`handleRenameKey()` — novo método:**

```go
func (m *Model) handleRenameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "esc":
        m.mode = ModeNormal
        m.renameError = ""
        return m, nil
    case "enter":
        return m.confirmRename()
    case "backspace":
        if len(m.renameInput) > 0 {
            m.renameInput = m.renameInput[:len(m.renameInput)-1]
        }
        m.renameError = ""
    default:
        if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
            m.renameInput += string(msg.Runes)
            m.renameError = ""
        }
    }
    return m, nil
}
```

### Step 6: Lógica de rename

```go
func (m *Model) enterRename() {
    if m.filename == "" {
        // Sem arquivo — não entra no modo
        return
    }
    m.mode = ModeRename
    m.renameInput = filepath.Base(m.filename)
    m.renameError = ""
}

func (m *Model) confirmRename() (tea.Model, tea.Cmd) {
    newName := strings.TrimSpace(m.renameInput)
    oldPath := m.filename
    
    // Validações
    if newName == "" {
        m.renameError = "Nome não pode estar vazio."
        return m, nil
    }
    if newName == filepath.Base(oldPath) {
        m.mode = ModeNormal
        m.renameError = ""
        return m, nil
    }
    if err := validateFileName(newName); err != nil {
        m.renameError = err.Error()
        return m, nil
    }
    
    // Auto-save se modificado
    if m.modified {
        m.save()
        if m.modified {
            m.renameError = "Erro ao salvar antes de renomear."
            return m, nil
        }
    }
    
    // Executa rename
    dir := filepath.Dir(oldPath)
    newPath := filepath.Join(dir, newName)
    
    if err := fileio.Rename(oldPath, newPath); err != nil {
        if os.IsExist(err) {
            m.renameError = fmt.Sprintf("Arquivo já existe: %s", newName)
        } else {
            m.renameError = fmt.Sprintf("Erro ao renomear: %v", err)
        }
        return m, nil
    }
    
    // Sucesso
    m.filename = newPath
    m.language = highlight.DetectLanguage(newPath)
    m.mode = ModeNormal
    m.renameError = ""
    m.addRecentFile(newPath)
    return m, nil
}
```

### Step 7: Renderização

```go
func (m *Model) renderRenameBar() string {
    oldName := filepath.Base(m.filename)
    s := fmt.Sprintf("Renomear: %s → %s", oldName, m.renameInput)
    if m.renameError != "" {
        s += "  " + lipgloss.NewStyle().
            Foreground(lipgloss.Color("203")).
            Render("(" + m.renameError + ")")
    }
    return m.searchStyle.Render(s) // reusa estilo da search bar
}
```

### Step 8: Integração no `View()` e `Update()`

- `View()`: adicionar `if m.mode == ModeRename` → renderizar com barra de rename
- `Update()`: adicionar `if m.mode == ModeRename` → delegar para `handleRenameKey()`
- `executeAction()`: adicionar case `"file.rename"` → `m.enterRename()`

### Step 9: Validação de nome de arquivo (cross-platform)

```go
func validateFileName(name string) error {
    // Caracteres inválidos em qualquer SO
    invalid := []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}
    for _, c := range invalid {
        if strings.ContainsRune(name, c) {
            return fmt.Errorf("caractere inválido: %c", c)
        }
    }
    return nil
}
```

---

## 🧪 Plano de Testes

### Unitários (`fileio_test.go`)

| Teste | Descrição |
|-------|-----------|
| `TestRenameSuccess` | Cria `a.txt`, renomeia para `b.txt`, verifica conteúdo preservado e `a.txt` não existe |
| `TestRenameSourceNotFound` | `Rename("/nonexistent/a.txt", "b.txt")` → erro |
| `TestRenameEmptyName` | `Rename("a.txt", "")` → erro |
| `TestRenameToSelf` | `Rename("a.txt", "a.txt")` → sucesso (no-op, mas não deveria falhar) |

### Integração (`editor_test.go`)

| Teste | Descrição |
|-------|-----------|
| `TestRenameFile` | Abre arquivo → F2 → digita novo nome → Enter → verifica `filename` atualizado |
| `TestRenameNoFile` | Buffer novo (sem arquivo) → F2 → não entra no modo rename |
| `TestRenameCancelEsc` | Modo rename → Esc → volta ao normal, `filename` inalterado |
| `TestRenameWithUnsavedChanges` | Modifica buffer → F2 → rename → verifica que salvou antes |
| `TestRenameInvalidChars` | Nome com `:` → mensagem de erro |
| `TestRenameEmptyName` | Apagar todo o nome → Enter → mensagem de erro |
| `TestRenameExistingFile` | Renomear para nome de arquivo já existente → mensagem de erro |

---

## 📋 Checklist de Implementação

- [ ] **1.** `fileio.Rename()` + testes → `go test ./internal/fileio/`
- [ ] **2.** `ModeRename` + campos no Model
- [ ] **3.** `registerActions()`: adicionar `file.rename` com `F2`
- [ ] **4.** `handleKey()`: adicionar case `"f2"` → `enterRename()`
- [ ] **5.** `enterRename()`, `handleRenameKey()`, `confirmRename()`
- [ ] **6.** `validateFileName()` cross-platform
- [ ] **7.** `renderRenameBar()` + integração no `View()`
- [ ] **8.** `executeAction()` adicionar case `"file.rename"`
- [ ] **9.** Update(): rotear `ModeRename` para `handleRenameKey()`
- [ ] **10.** Testes de integração no editor
- [ ] **11.** `go test ./...` — todos passando
- [ ] **12.** `go vet ./...` — sem warnings
- [ ] **13.** Atualizar `CHANGELOG.md`, `README.md`, `PLANO.md`
- [ ] **14.** Registrar em `test-automation/TEST_LOG.md`

---

## 🌍 Cross-Platform Considerations

| Aspecto | Windows | macOS | Linux |
|---------|---------|-------|-------|
| Atalho F2 | ✅ Funciona | ✅ Funciona | ✅ Funciona |
| `os.Rename` | ❌ Falha se destino existe | ✅ Sobrescreve silenciosamente | ✅ Sobrescreve silenciosamente |
| Caracteres inválidos | `\ / : * ? " < > \|` | `/` apenas | `/` apenas |
| Case sensitivity | Case-insensitive | Case-sensitive (APFS) | Case-sensitive |
| Tratamento | Mensagem de erro clara | Mensagem de erro clara | Mensagem de erro clara |

**Decisão sobre destino existente:** Validar antes com `os.Stat` em todas as plataformas e rejeitar com mensagem. Consistente e seguro.

---

## 🔗 Dependências

- Nenhuma nova dependência externa
- Usa apenas `os`, `path/filepath`, `strings`, `fmt` da stdlib

---

> **Próximo passo:** Após aprovação deste plano, implementar em `feature/file-rename` branch.
