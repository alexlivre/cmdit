# PLANO DE DESENVOLVIMENTO — cmdit

> **Meta:** Construir um editor de texto TUI modeless, CUA-first, que seja a melhor alternativa ao Vim — usável por leigos em desktop e por devs em servidores SSH.
>
> **Arquitetura:** Go + Bubble Tea + Lip Gloss. Elm-like (Model/Update/View). Gap buffer v1, Rope v2. Command palette como camada de descoberta. Funcionalidades built-in sem sistema de plugins.
>
> **Stack:** Go 1.22+, Bubble Tea, Bubbles, Lip Gloss (charm.sh), Chroma (syntax v1), Treesitter (v2)

---

## 📊 Comparativo com concorrentes

|  | Vim | Nano | Helix | Micro | **cmdit** |
|--|-----|------|-------|-------|-----------|
| Modeless | ❌ | ✅ | ⚠️ | ✅ | ✅ |
| Mouse | ❌ | ✅ | ❌ | ✅ | ✅ |
| CUA (Ctrl+C/V/Z) | ❌ | ❌ | ❌ | ✅ | ✅ |
| Command Palette | ❌ | ❌ | ✅ | ✅ | ✅ |
| Multiple cursors | ❌ | ❌ | ✅ | ✅ | ✅ |
| LSP nativo | ❌ | ❌ | ✅ | ❌ | ✅ (v2) |
| Treesitter | ❌ | ❌ | ✅ | ❌ | ✅ (v2) |
| Plugin system | ✅ | ❌ | ❌ | ❌ | ❌ (built-in only) |
| Vim keymap | ✅ | ❌ | ❌ | ❌ | ✅ built-in (toggle) |
| Single binary | ✅ | ✅ | ✅ | ✅ | ✅ |
| Curva aprendizado | 🔴 meses | 🟢 segundos | 🟡 dias | 🟢 minutos | 🟢 **segundos** |

---

## 📁 Estrutura de diretórios

```
cmdit/
├── cmd/
│   └── cmdit/
│       └── main.go                  # Entry point
├── internal/
│   ├── editor/
│   │   ├── editor.go                # Bubble Tea Model principal
│   │   └── editor_test.go
│   ├── buffer/
│   │   ├── buffer.go                # Gap buffer (armazenamento de texto)
│   │   ├── buffer_test.go
│   │   ├── cursor.go                # Posição do cursor, seleção
│   │   └── cursor_test.go
│   ├── renderer/
│   │   ├── viewport.go              # Viewport (scroll, resize)
│   │   └── viewport_test.go
│   ├── input/
│   │   ├── keybindings.go           # Mapeamento tecla → ação
│   │   └── keybindings_test.go
│   ├── command/
│   │   ├── palette.go               # Command palette UI
│   │   ├── actions.go               # Registro de ações
│   │   └── palette_test.go
│   ├── highlight/
│   │   ├── syntax.go                # Syntax highlighting (Chroma)
│   │   └── syntax_test.go
│   ├── clipboard/
│   │   ├── clipboard.go             # OSC52 + clipboard do sistema
│   │   └── clipboard_test.go
│   └── fileio/
│       ├── fileio.go                # Load/Save arquivos
│       └── fileio_test.go
├── go.mod
├── go.sum
├── Makefile
└── PLANO.md                         # Este arquivo
```

---

## 🔢 Fases de desenvolvimento

| Fase | Nome | Duração | Entregável |
|------|------|---------|------------|
| **0** | Scaffold | 1 dia | Projeto Go + Bubble Tea hello world |
| **1** | Editor básico | 2-3 sem | Gap buffer, inserção, Ctrl+S, Ctrl+Q |
| **2** | Navegação | 1-2 sem | Setas, mouse, scroll, números de linha |
| **3** | Edição power | 2-3 sem | Undo/Redo, Ctrl+C/V/X, busca, substituir |
| **4** | Command palette | 1-2 sem | Ctrl+P, ações registradas |
| **5** | Syntax highlight | 1-2 sem | Chroma, 5 temas, detecção de linguagem |
| **6** | File operations | 1 sem | Ctrl+O, salvar como, recentes |
| **7** | Polimento v1 | 1-2 sem | Auto-save, welcome, OSC52, status bar |
| **8** | Power features | 3-4 sem | Múltiplos cursores, abas, splits, LSP | ✅ |
| **9** | Features built-in | 1-2 sem | Auto-close brackets, vim mode, format on save, themes, keybindings | 📋 |
| **10** | Qualidade & estabilidade | 1-2 sem | Análise de código, edge cases, cobertura 70%+, robustez | 📋 |
| **11** | Produção | 2 sem | Cross-compile, CI/CD, installer, docs | 📋 |

---

## 📋 Fase 0 — Scaffold

**Objetivo:** Projeto Go compilável com Bubble Tea renderizando na tela.

- [ ] Inicializar módulo Go (`go mod init github.com/user/cmdit`)
- [ ] Adicionar dependências: `bubbletea`, `bubbles`, `lipgloss`
- [ ] Criar `cmd/cmdit/main.go` com Model mínimo:
  ```go
  type model struct { text string }
  func (m model) Init() tea.Cmd { return nil }
  func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
  func (m model) View() string { return m.text }
  ```
- [ ] Renderizar `"cmdit v0.1.0"` na tela
- [ ] Criar `Makefile` com targets: `build`, `run`, `test`, `lint`
- [ ] Executar `go build ./...` — deve compilar sem erros
- [ ] Executar `go run ./cmd/cmdit` — deve exibir "cmdit v0.1.0"

### Marcos
- **M1: Hello World** — `go run .` mostra cmdit na tela ✅

---

## 📋 Fase 1 — Editor básico (MVP real)

**Objetivo:** Digitar texto, salvar em arquivo, sair.

### Tarefas

- [ ] **1.1** Criar `internal/buffer/buffer.go` — Gap buffer
  - Estrutura: `data []rune`, `gapStart int`, `gapEnd int`
  - Métodos: `NewBuffer()`, `Insert(r rune)`, `Delete()`, `MoveLeft()`,
    `MoveRight()`, `MoveUp()`, `MoveDown()`, `String() string`
  - Gap buffer é uma estrutura com buraco no meio do array;
    inserções e deleções ocorrem na posição do gap, evitando shifts

- [ ] **1.2** Criar `internal/buffer/buffer_test.go` — Testes do gap buffer
  - Teste: inserir caracteres e verificar `String()`
  - Teste: inserir no meio, deletar, mover gap
  - Teste: inserir quebra de linha (`\n`)
  - Teste: buffer vazio → `String()` retorna `""`

- [ ] **1.3** Criar `internal/buffer/cursor.go` — Cursor
  - Estrutura: `Line int`, `Col int`
  - Métodos: `SetPos(line, col int)`, `Up()`, `Down()`, `Left()`, `Right()`

- [ ] **1.4** Criar `internal/buffer/cursor_test.go` — Testes do cursor
  - Teste: mover nas 4 direções
  - Teste: não ultrapassar limites (linha 0, coluna 0)

- [ ] **1.5** Criar `internal/editor/editor.go` — Model principal
  - Model Bubble Tea com: `buffer *buffer.Buffer`, `cursor *buffer.Cursor`,
    `filename string`, `modified bool`
  - `Init()` — retorna `nil`
  - `Update(msg)` — processa `tea.KeyMsg`:
    - Caracteres imprimíveis → `buffer.Insert(char)`
    - Backspace → `buffer.Delete()`
    - Enter → `buffer.Insert('\n')`
    - Ctrl+S → salvar
    - Ctrl+Q → sair (se modificado, mostra diálogo)
  - `View()` — renderiza conteúdo do buffer

- [ ] **1.6** Criar `internal/editor/editor_test.go` — Testes do editor
  - Teste: abrir editor, digitar "hello", verificar buffer
  - Teste: Ctrl+S salva arquivo

- [ ] **1.7** Criar `internal/fileio/fileio.go` — Load/Save
  - `Load(path string) (*buffer.Buffer, error)`
  - `Save(path string, buf *buffer.Buffer) error`

- [ ] **1.8** Criar `internal/fileio/fileio_test.go` — Testes de I/O
  - Teste: salvar e carregar arquivo (roundtrip)
  - Teste: carregar arquivo com UTF-8 (acentos)

- [ ] **1.9** Atualizar `cmd/cmdit/main.go` para usar o editor
  - Aceitar argumento de arquivo: `cmdit [arquivo]`
  - Se arquivo existe → carregar; senão → buffer vazio

- [ ] **1.10** Executar `go test ./...` — todos os testes devem passar
- [ ] **1.11** Teste manual: abrir, digitar, Ctrl+S, reabrir → texto persiste

### Marcos
- **M2: Digitar e salvar** — Abre, digita, Ctrl+S, reabre, texto persiste ✅

---

## 📋 Fase 2 — Navegação

**Objetivo:** Mover cursor com teclado e mouse, scroll, números de linha.

- [ ] **2.1** Implementar navegação por setas no editor
  - Left/Right → `cursor.Left()` / `cursor.Right()` + move gap
  - Up/Down → `cursor.Up()` / `cursor.Down()` + move gap
  - Ctrl+Left/Right → pular palavras
  - Home/End → início/fim da linha
  - Ctrl+Home/End → início/fim do arquivo

- [ ] **2.2** Implementar `PageUp/PageDown` — rolar uma página

- [ ] **2.3** Criar `internal/renderer/viewport.go`
  - `scrollY int`, `scrollX int`, `width int`, `height int`
  - Métodos: `ScrollTo(line, col)`, `EnsureVisible(line, col)`, `Resize(w, h)`

- [ ] **2.4** Criar `internal/renderer/viewport_test.go`
  - Teste: ScrollTo mantém linha visível
  - Teste: EnsureVisible ajusta scroll quando cursor sai da tela

- [ ] **2.5** Implementar suporte a mouse
  - `tea.MouseMsg` → clique posiciona cursor
  - Scroll wheel → `viewport.ScrollTo()`

- [ ] **2.6** Implementar números de linha
  - Margem esquerda: `│` + número + ` `
  - Ex: `  1 │ func main() {`
  - Estilo via Lip Gloss (cor cinza)

- [ ] **2.7** Tratar `SIGWINCH` (redimensionamento do terminal)
  - `tea.WindowSizeMsg` → `viewport.Resize(w, h)`

- [ ] **2.8** Executar `go test ./...` — todos os testes passando
- [ ] **2.9** Teste manual: setas, mouse, scroll, redimensionar janela

### Marcos
- **M3: Navegação completa** — Mouse, setas, scroll, números de linha ✅

---

## 📋 Fase 3 — Edição power

**Objetivo:** Undo/Redo, clipboard, busca, substituir.

- [ ] **3.1** Implementar UndoStack em `internal/buffer/`
  - `type Operation struct { Type string; Pos int; Text string }`
  - `Push(op Operation)`, `Undo()`, `Redo()`
  - Cada Insert/Delete gera Operation reversa

- [ ] **3.2** Testar UndoStack
  - Teste: Insert "abc", Undo → buffer vazio
  - Teste: Undo + Redo → texto restaurado
  - Teste: múltiplos Undo consecutivos

- [ ] **3.3** Implementar seleção de texto
  - Shift + setas → expandir seleção
  - Mouse arrasto → selecionar
  - Ctrl+A → selecionar tudo
  - Renderizar seleção com fundo invertido

- [ ] **3.4** Implementar clipboard
  - Ctrl+C: copia seleção (ou linha atual se nada selecionado)
  - Ctrl+X: recorta seleção
  - Ctrl+V: cola do clipboard
  - Clipboard do sistema (Windows/Linux/macOS)

- [ ] **3.5** Criar `internal/clipboard/clipboard.go`
  - `Copy(text string) error`
  - `Paste() (string, error)`
  - Usa biblioteca `golang.design/x/clipboard` ou `atotto/clipboard`

- [ ] **3.6** Implementar busca (Ctrl+F)
  - Abre barra de busca na parte inferior
  - Destaca todas as ocorrências no texto
  - Enter/F3 → próximo match, Shift+F3 → anterior
  - Match atual com cor diferente

- [ ] **3.7** Implementar substituição (Ctrl+H)
  - Dois campos: "buscar" e "substituir"
  - Botões: [Substituir] [Substituir tudo] [Cancelar]

- [ ] **3.8** Conectar teclas no `keybindings.go`
  - Ctrl+Z → Undo, Ctrl+Y → Redo
  - Ctrl+C/X/V → clipboard
  - Ctrl+F → busca, Ctrl+H → substituir
  - Ctrl+A → selecionar tudo

- [ ] **3.9** Executar `go test ./...` — todos os testes passando

### Marcos
- **M4: Edição produtiva** — Undo, clipboard, busca funcionando ✅

---

## 📋 Fase 4 — Command Palette

**Objetivo:** Ctrl+P acessa todos os comandos com busca fuzzy.

- [ ] **4.1** Criar `internal/command/actions.go`
  - `type Action struct { ID, Label, Shortcut string; Handler func(*editor.Model) tea.Cmd }`
  - Registro global: `var AllActions []Action`
  - Função `Register(a Action)` para plugins futuros

- [ ] **4.2** Registrar ações iniciais:
  - `file.save`, `file.open`, `file.save-as`, `file.quit`
  - `edit.undo`, `edit.redo`, `edit.cut`, `edit.copy`, `edit.paste`, `edit.select-all`
  - `search.find`, `search.replace`
  - `view.go-line`, `view.toggle-word-wrap`
  - `help.welcome`, `help.shortcuts`

- [ ] **4.3** Criar `internal/command/palette.go`
  - UI: campo de busca + lista de resultados
  - Filtro fuzzy em tempo real (case-insensitive)
  - Cada item mostra: nome do comando + atalho à direita
  - Enter executa ação, Esc fecha palette
  - Setas navegam na lista

- [ ] **4.4** Criar `internal/command/palette_test.go`
  - Teste: filtro fuzzy retorna comandos relevantes
  - Teste: Enter executa handler correto
  - Teste: Esc fecha sem executar

- [ ] **4.5** Conectar Ctrl+P ao editor

- [ ] **4.6** Executar `go test ./...`

### Marcos
- **M5: Discoverable** — Command palette funcional com todos os comandos ✅

---

## 📋 Fase 5 — Syntax Highlighting

**Objetivo:** Cores no texto baseadas na linguagem do arquivo.

- [ ] **5.1** Criar `internal/highlight/syntax.go`
  - `Detect(filename, content string) string` — retorna linguagem
  - `Tokenize(content, language string) []Token`
  - `ApplyTheme(tokens []Token, theme Theme) []StyledToken`
  - Usar Chroma (`github.com/alecthomas/chroma`)

- [ ] **5.2** Definir 5 temas built-in:
  - `dark` (padrão), `light`, `monokai`, `dracula`, `solarized-dark`

- [ ] **5.3** Integrar highlight na renderização do editor
  - Tokenizar buffer a cada alteração (debounced)
  - Renderizar cada token com cor do tema
  - Se linguagem não reconhecida → sem highlight (texto plano)

- [ ] **5.4** Adicionar indicador de linguagem na barra de status
  - `[Go]` ou `[Plain Text]`

- [ ] **5.5** Criar `internal/highlight/syntax_test.go`
  - Teste: detectar `.go` → `Go`
  - Teste: detectar `.py` → `Python`
  - Teste: detectar `.md` → `Markdown`
  - Teste: tokenizar código Go e verificar tipos de token

- [ ] **5.6** Executar `go test ./...`

### Marcos
- **M6: Bonito** — Syntax highlight com 5 temas, detecção automática ✅

---

## 📋 Fase 6 — File Operations

**Objetivo:** Abrir arquivos, salvar como, arquivos recentes.

- [ ] **6.1** Implementar Ctrl+O — File picker
  - UI: lista de arquivos/diretórios no diretório atual
  - Setas navegam, Enter abre diretório/arquivo
  - Filtro fuzzy enquanto digita
  - Esc cancela

- [ ] **6.2** Implementar F3 — Salvar como
  - Campo de texto para digitar caminho
  - Confirmação se arquivo já existe

- [ ] **6.3** Implementar arquivos recentes
  - Arquivo `~/.cmdit/recent.json` com `[{path, timestamp}]`
  - Mostrar na welcome screen
  - Atualizar ao abrir/salvar arquivo

- [ ] **6.4** Implementar welcome screen
  - Se `cmdit` executado sem argumento → welcome screen
  - Lista de arquivos recentes
  - Atalhos: `Ctrl+O Abrir`, `Ctrl+P Comandos`, `Ctrl+Q Sair`
  - Logo ASCII "cmdit" estilizado

- [ ] **6.5** Executar `go test ./...`

---

## 📋 Fase 7 — Polimento v1

**Objetivo:** Refinar experiência para release v1.

- [ ] **7.1** Implementar auto-save
  - Goroutine: a cada 30s, se `modified == true` → salvar
  - Também ao perder foco (se possível detectar)

- [ ] **7.2** Melhorar barra de status
  - Formato: `[📄 main.go] [modificado ●] [L:42 C:10] [UTF-8] [Go] [Tema:dark]`
  - Cores via Lip Gloss

- [ ] **7.3** Implementar diálogo de confirmação ao sair
  - "Arquivo modificado! Deseja salvar?"
  - Opções: `[S] Salvar  [D] Descartar  [C] Cancelar`

- [ ] **7.4** Implementar OSC52 clipboard
  - Detectar se terminal suporta OSC52
  - Codificar texto em base64, enviar via escape sequence
  - Fallback para clipboard do sistema

- [ ] **7.5** Implementar detecção de cores do terminal
  - `$COLORTERM=truecolor` → true color (16M)
  - `$TERM=*-256color` → 256 cores
  - Senão → 16 cores (fallback)
  - Ajustar tema automaticamente

- [ ] **7.6** Tratar Ctrl+S vs flow control
  - Detectar se flow control está ativo (`stty -a`)
  - Se sim, avisar: "Ctrl+S bloqueado pelo terminal. Use `stty -ixon` ou F3"

- [ ] **7.7** Executar `go test ./...` e `go vet ./...`
- [ ] **7.8** Build cross-platform: Windows, Linux, macOS

### Marcos
- **M7: v1 release** — Binário único, usável por leigos, pronto para distribuir ✅

---

## 📋 Fase 8 — Power Features (v2) ✅

**Objetivo:** Múltiplos cursores, abas, splits, indent guides, LSP client.

### ✅ Implementado (2026-04-29)

- **8.1 Múltiplos cursores** — Ctrl+D adiciona cursor na próxima ocorrência da palavra. Escape limpa. Edição simultânea em todos os cursores (insert, delete, backspace, enter, tab). Indicador na status bar.
- **8.2 Abas** — Ctrl+T nova aba, Ctrl+W fecha (confirma se modificada), Ctrl+Tab alterna, Ctrl+1-9 salta. Barra de abas com indicador ● de modificação. Package `internal/tabs/`.
- **8.3 Splits** — Ctrl+\ divide horizontalmente com novo TabManager. Clique para focar painel. Borda ativa destacada em laranja. SplitContainer suporta modo single-pane.
- **8.4 LSP client** — Package `internal/lsp/` com JSON-RPC 2.0 sobre stdio. Inicia gopls automaticamente para Go. Envia didChange em cada edição. Mostra ✗N e ⚠N na status bar.
- **8.5 Indent guides** — Linhas verticais │ em cada nível de 4 espaços. Renderizado na margem esquerda.

### Arquitetura final

```
SplitContainer (tea.Model)
├── TabManager (left)
│   ├── editor.Model (tab 1)
│   └── editor.Model (tab 2)
└── TabManager (right, opcional)
    └── editor.Model (tab 1)
```

### Arquivos criados/alterados
- `internal/tabs/tabmanager.go` — TabManager (367 linhas)
- `internal/tabs/split.go` — SplitContainer (400 linhas)
- `internal/lsp/lsp.go` — Cliente LSP (340 linhas)
- `internal/editor/` — Refatorado em editor.go + view.go + keys.go + actions.go + lsp_integration.go
- `cmd/cmdit/main.go` — Usa SplitContainer → TabManager → editor.Model

---

## 📋 Fase 9 — Features Built-in

**Objetivo:** Adicionar funcionalidades essenciais de editores modernos como código Go nativo — sem sistema de plugins.

**Filosofia:** Toda funcionalidade relevante deve vir built-in e ser ativada/desativada por toggle. Nada de Lua, nada de marketplace, nada de sandbox. Simples, direto, confiável.

---

### 9.1 Auto-close Brackets & Quotes

**Problema:** Usuário digita `(` e precisa manualmente digitar `)`.

**Solução:** Inserir o par de fechamento automaticamente com smart-skip (se digitar o fechamento manualmente, pula o caractere existente).

- [ ] **9.1.1** Criar `internal/editor/autoclose.go`
  - Função `shouldAutoClose(char rune) (rune, bool)` — retorna o fechamento para `( [ { " ' \``
  - Função `handleAutoClose(m *Model, openChar rune) tea.Cmd`
    - Insere `openChar + closeChar` no buffer
    - Move cursor entre eles
    - Marca posição como "auto-closed" para smart-skip

- [ ] **9.1.2** Implementar smart-skip no `keys.go`
  - Se próximo caractere == char digitado E posição atual é "auto-closed" → pula (não insere duplicado)
  - Funciona para todos os pares: `() [] {} "" '' ``

- [ ] **9.1.3** Adicionar toggle: `F4` ou comando `view.toggle-auto-close`
  - Estado salvo em `config.json`
  - Indicador na status bar: `[AutoClose]`

- [ ] **9.1.4** Testes:
  - `TestAutoCloseParens` — digitar `(` insere `()` com cursor no meio
  - `TestAutoCloseSmartSkip` — digitar `)` após `()` pula ao invés de duplicar
  - `TestAutoCloseToggle` — desabilitar e verificar que não fecha

---

### 9.2 Vim Keybindings (Modo Toggle)

**Problema:** Usuários vindos do Vim não conseguem usar o cmdit sem plugins.

**Solução:** Modo Vim built-in com toggle `F5`. Implementado como uma camada de tradução de teclas no `keys.go`.

- [ ] **9.2.1** Criar `internal/editor/vimmode.go`
  - `type VimMode string` — `Normal`, `Insert`, `Visual`, `Command`
  - `type VimState struct { mode VimMode; count string; lastOp string }`
  - Mapa de keybindings vim:
    ```
    Normal mode:
      h j k l → move cursor
      w b → next/prev word
      0 $ → line start/end
      gg G → file start/end
      i I a A o O → enter insert mode
      x → delete char
      dd → delete line
      yy → copy line
      p P → paste after/before
      u Ctrl+R → undo/redo
      / → search
      :w :q :wq → save/quit/save+quit
      Esc → normal mode (from any)
    ```

- [ ] **9.2.2** Integrar no `keys.go`
  - `if m.vimMode { return dispatchVimKey(m, msg) }`
  - Tradução transparente — resto do editor não sabe que está em modo vim

- [ ] **9.2.3** Comando `view.toggle-vim-mode` no palette + atalho `F5`
  - Estado salvo em `config.json`
  - Indicador na status bar: `[VIM]` quando ativo

- [ ] **9.2.4** Testes:
  - `TestVimModeNavigation` — hjkl movem cursor
  - `TestVimModeInsert` — `i` entra em insert mode
  - `TestVimModeSave` — `:w` salva arquivo
  - `TestVimModeToggle` — `F5` ativa/desativa

---

### 9.3 Format on Save

**Problema:** Desenvolvedores precisam formatar código manualmente.

**Solução:** Executar formatador externo ao salvar, baseado na linguagem detectada.

- [ ] **9.3.1** Criar `internal/editor/format.go`
  - `type FormatterFunc func(text string, lang string) (string, error)`
  - Registro de formatadores:
    ```
    Go      → gofmt (via exec)
    Python  → black / autopep8 (via exec, se disponível)
    Rust    → rustfmt (via exec, se disponível)
    JSON    → json.Indent (built-in Go)
    Markdown→ prettier (via exec, se disponível)
    ```
  - Fallback universal: `editorconfig` (trim trailing whitespace, ensure newline at EOF)

- [ ] **9.3.2** Hook no save (`Ctrl+S` e auto-save)
  - `if config.FormatOnSave { formatAndSave() }`
  - Mostra mensagem na status bar: "Formatado com gofmt" ou "gofmt não encontrado"

- [ ] **9.3.3** Comando `file.toggle-format-on-save` no palette
  - Estado salvo em `config.json`
  - Indicador na status bar: `[Fmt]`

- [ ] **9.3.4** Testes:
  - `TestFormatOnSaveGo` — salvar arquivo .go e verificar que gofmt foi aplicado
  - `TestFormatOnSaveDisabled` — salvar sem formatar
  - `TestFormatFallback` — sem formatador externo, aplica trim trailing whitespace

---

### 9.4 Sistema de Temas (Runtime Switch)

**Problema:** Usuário quer trocar entre tema claro e escuro sem reiniciar.

**Solução:** `F6` alterna entre os 5 temas built-in em tempo real.

- [ ] **9.4.1** Criar `internal/highlight/themes.go`
  - Carregar todos os temas Chroma disponíveis
  - `func SwitchTheme(name string)` — troca tema e força re-render
  - Lista de temas: `dark` (padrão), `light`, `monokai`, `dracula`, `solarized-dark`

- [ ] **9.4.2** Atalho `F6` → próximo tema na rotação
  - Status bar mostra nome do tema: `[dark]`

- [ ] **9.4.3** Comando `view.next-theme` no palette
  - Salva tema preferido em `config.json`

- [ ] **9.4.4** Testes:
  - `TestThemeSwitch` — trocar tema e verificar que tokens têm cores diferentes
  - `TestThemePersist` — reiniciar com tema salvo

---

### 9.5 Custom Keybindings (JSON Config)

**Problema:** Usuário quer remapear teclas sem editar código.

**Solução:** Arquivo `~/.cmdit/config.json` com seção `keybindings`.

- [ ] **9.5.1** Criar `internal/editor/config.go`
  - `type Config struct { AutoClose bool; VimMode bool; FormatOnSave bool; Theme string; Keybindings map[string]string }`
  - `LoadConfig() Config` — carrega de `~/.cmdit/config.json`
  - `SaveConfig(c Config)` — salva alterações
  - Criar diretório `~/.cmdit/` se não existir

- [ ] **9.5.2** Exemplo de `config.json`:
  ```json
  {
    "auto_close_enabled": true,
    "vim_mode": false,
    "format_on_save": true,
    "theme": "dark",
    "keybindings": {
      "ctrl+s": "file.save",
      "ctrl+shift+s": "file.save-as"
    }
  }
  ```

- [ ] **9.5.3** Integrar com `keybindings.go`
  - Resolver ação por ID (`file.save`, `edit.undo`, etc.)
  - Config do usuário sobrescreve defaults

- [ ] **9.5.4** Testes:
  - `TestConfigLoad` — carregar config.json
  - `TestConfigKeybindOverride` — remapear Ctrl+S e verificar
  - `TestConfigDefaults` — sem config.json, usar defaults

---

### 9.6 Word Wrap Toggle

- [ ] **9.6.1** Implementar toggle word wrap
  - Modos: `soft-wrap` (quebra visual), `no-wrap` (scroll horizontal)
  - Atalho: `Alt+Z` (padrão VS Code)
  - Indicador na status bar: `[Wrap]`

---

### Marcos
- **M9: Completo sem plugins** — Auto-close, Vim mode, format on save, 5 temas, keybindings customizáveis

---

## 📋 Fase 10 — Qualidade & Estabilidade de Código

**Objetivo:** Transformar o cmdit de "funcionando" para "confiável". Esta fase não adiciona funcionalidades visíveis — ela garante que tudo que existe funciona corretamente em todos os cenários.

**Filosofia:** Bugs encontrados aqui nunca chegam ao usuário. Cada bug corrigido é um crash evitado. Cada edge case coberto é um usuário que não perde trabalho.

---

### 10.1 Análise Completa de Código

- [ ] **10.1.1** Code review sistemático de todos os arquivos
  - `internal/editor/` — editor.go (654), view.go (438), keys.go (699), actions.go (261), lsp_integration.go (95)
  - `internal/tabs/` — tabmanager.go (367), split.go (400), flow_test.go
  - `internal/lsp/` — lsp.go (340)
  - `internal/buffer/` — buffer.go, cursor.go, undo.go
  - `internal/command/` — palette.go, actions.go
  - `internal/clipboard/` — clipboard.go
  - `internal/fileio/` — fileio.go
  - `internal/highlight/` — syntax.go
  - `cmd/cmdit/main.go`

- [ ] **10.1.2** Identificar e catalogar problemas:
  - Funções muito longas (>50 linhas)
  - Código duplicado (DRY violations)
  - Responsabilidades misturadas (SRP violations)
  - Nomes confusos ou inconsistentes
  - Código morto (nunca executado ou inalcançável)

- [ ] **10.1.3** Documentar débitos técnicos encontrados no CHANGELOG

---

### 10.2 Tratamento de Erros Robusto

- [ ] **10.2.1** Auditar todo `error` retornado
  - Nenhum `_` ignorando erro sem justificativa documentada
  - Adicionar `if err != nil` com log ou tratamento adequado
  - Erros críticos devem ser exibidos ao usuário na status bar

- [ ] **10.2.2** Criar `internal/editor/errors.go`
  - Centralizar mensagens de erro
  - Função `showError(msg string)` → exibe na status bar em vermelho por 3 segundos
  - Função `logError(err error, context string)` → log para `~/.cmdit/cmdit.log`

- [ ] **10.2.3** Tratar panics
  - Adicionar `defer recover()` nos loops principais
  - Log do stack trace antes de crash
  - Mensagem amigável: "cmdit encontrou um erro. Reporte em github.com/alexlivre/cmdit/issues"

- [ ] **10.2.4** Testes de erro:
  - `TestFileLoadNotFound` — abrir arquivo inexistente
  - `TestFileSavePermission` — salvar em diretório sem permissão
  - `TestLSPStartupFailure` — gopls não instalado
  - `TestClipboardFailure` — clipboard indisponível
  - `TestInvalidUTF8` — arquivo com bytes inválidos

---

### 10.3 Edge Cases e Robustez

- [ ] **10.3.1** Arquivos extremos
  - Arquivo vazio (0 bytes)
  - Arquivo muito grande (100MB+) — testar performance de scroll
  - Arquivo com 1 única linha de 100K caracteres
  - Arquivo binário — detectar e avisar
  - Arquivo com caracteres especiais: `\x00`, `\r\n`, BOM

- [ ] **10.3.2** Buffer edge cases
  - Inserir no início do buffer
  - Inserir no final do buffer
  - Deletar o último caractere do buffer
  - Buffer com apenas `\n` (linha vazia)
  - Undo até estado inicial (buffer vazio)
  - Redo após undo (sem inserções no meio)

- [ ] **10.3.3** Tabs edge cases
  - Fechar a última aba (confirmação)
  - Fechar todas as abas → deve mostrar welcome screen
  - Ctrl+Tab com 1 aba (não faz nada)
  - Ctrl+9 com menos de 9 abas (não faz nada)
  - Abrir 50 abas (performance e UI da tab bar)
  - Fechar aba com seleção via teclado (S=salvar, D=descartar, C=cancelar)

- [ ] **10.3.4** Split edge cases
  - Split com 0 abas em um painel
  - Fechar todos os painéis → volta para single pane
  - Redimensionar splits (se implementado)
  - Focar painel vazio

- [ ] **10.3.5** Multi-cursor edge cases
  - Ctrl+D sem palavra sob cursor (não faz nada)
  - Todos cursores na mesma posição (não duplica)
  - Escape limpa cursores extras
  - Editar com 100 cursores (performance)
  - Desfazer operação multi-cursor

- [ ] **10.3.6** LSP edge cases
  - gopls não instalado → fallback silencioso
  - gopls crash → reconectar automaticamente
  - Arquivo não reconhecido → sem LSP
  - Resposta LSP muito grande (timeout handling)
  - Múltiplos arquivos .go abertos em abas diferentes

---

### 10.4 Cobertura de Testes

- [ ] **10.4.1** Medir cobertura atual
  ```bash
  go test -coverprofile=coverage.out ./...
  go tool cover -func=coverage.out
  ```

- [ ] **10.4.2** Meta por pacote:

  | Pacote | Cobertura atual | Meta |
  |--------|-----------------|------|
  | `internal/buffer/` | ? | 85%+ |
  | `internal/editor/` | ? | 70%+ |
  | `internal/tabs/` | 59 testes | 70%+ |
  | `internal/lsp/` | ? | 50%+ |
  | `internal/command/` | ? | 70%+ |
  | `internal/clipboard/` | ? | 60%+ |
  | `internal/fileio/` | ? | 80%+ |
  | `internal/highlight/` | ? | 60%+ |

- [ ] **10.4.3** Escrever testes faltantes até atingir meta
  - Prioridade: buffer → editor → tabs → fileio → command → lsp

- [ ] **10.4.4** Adicionar `go test -race ./...` ao Makefile
  - Detectar data races (especialmente em goroutines de auto-save e LSP)

---

### 10.5 Qualidade de Código

- [ ] **10.5.1** Executar e corrigir `golangci-lint run ./...`
  - errcheck, gosimple, govet, ineffassign, staticcheck, unused
  - 0 warnings é a meta

- [ ] **10.5.2** Padronizar nomes e convenções
  - Funções exportadas com comentários (`// Foo does bar.`)
  - Constantes em UPPER_CASE ou seguindo convenção Go
  - Erros com prefixo do pacote: `buffer: cursor out of bounds`

- [ ] **10.5.3** Refatorar funções longas (>50 linhas)
  - Extrair para funções menores com nomes claros
  - Cada função faz UMA coisa (Single Responsibility)

- [ ] **10.5.4** Eliminar duplicação
  - Buscar padrões repetidos com `grep` ou análise manual
  - Extrair lógica comum para helpers

- [ ] **10.5.5** Melhorar nomes
  - Variáveis de 1 letra só em loops muito curtos
  - Nomes descritivos: `lineCount` não `lc`, `tabManager` não `tm`
  - Funções com verbo: `calculateOffset`, `renderLine`, `saveFile`

---

### 10.6 Performance

- [ ] **10.6.1** Profiling
  ```bash
  go test -bench=. ./... -benchmem
  ```
  - Identificar alocações excessivas
  - Medir tempo de renderização com arquivos grandes

- [ ] **10.6.2** Otimizações-alvo
  - Gap buffer: medir tempo de Insert/Delete
  - Syntax highlight: debounce tokenização (não tokenizar a cada keystroke)
  - LSP: verificar se há goroutines vazando
  - View: medir tempo de renderização da tela completa

- [ ] **10.6.3** Benchmarks
  - `BenchmarkBufferInsert` — inserir 10000 chars
  - `BenchmarkBufferDelete` — deletar 10000 chars
  - `BenchmarkSyntaxHighlight` — tokenizar arquivo de 1000 linhas
  - `BenchmarkRender` — renderizar tela 80x24 com syntax highlight

---

### 10.7 Segurança

- [ ] **10.7.1** Path traversal
  - Validar que `Ctrl+O` e `F3` não permitem `../../../etc/passwd`
  - Resolver caminhos com `filepath.Clean()`

- [ ] **10.7.2** Injeção de escape sequences
  - Sanitizar conteúdo do arquivo antes de renderizar
  - Bloquear sequências ANSI perigosas no buffer

- [ ] **10.7.3** Permissões de arquivo
  - `~/.cmdit/` deve ter permissão 0700
  - `config.json` deve ter permissão 0600
  - Não alterar permissões do arquivo editado ao salvar

---

### Marcos
- **M10: Robusto** — Cobertura 70%+, 0 lint warnings, tratamento de erro completo, edge cases documentados

---

## 📋 Fase 11 — Produção

- [ ] **11.1** Cross-compile no Makefile
  ```makefile
  build-all:
    GOOS=linux GOARCH=amd64 go build -o bin/cmdit-linux-amd64 ./cmd/cmdit
    GOOS=linux GOARCH=arm64 go build -o bin/cmdit-linux-arm64 ./cmd/cmdit
    GOOS=windows GOARCH=amd64 go build -o bin/cmdit-windows-amd64.exe ./cmd/cmdit
    GOOS=darwin GOARCH=amd64 go build -o bin/cmdit-darwin-amd64 ./cmd/cmdit
    GOOS=darwin GOARCH=arm64 go build -o bin/cmdit-darwin-arm64 ./cmd/cmdit
  ```

- [ ] **11.2** GitHub Actions CI/CD
  - `go test ./...` + `go vet ./...` + `golangci-lint`
  - Build multi-plataforma
  - Release automático ao criar tag

- [ ] **11.3** Script de instalação
  ```bash
  curl -sSL https://cmdit.dev/install.sh | bash
  # Detecta OS/arch, baixa binário, copia para /usr/local/bin
  ```

- [ ] **11.4** Documentação
  - Site: cmdit.dev (Hugo ou Markdown no GitHub Pages)
  - Quickstart: "5 minutos para dominar o cmdit"
  - Guia de atalhos (cheatsheet)
  - Guia de configuração (config.json)
  - Guia de migração: "Vim → cmdit: o que muda"

---

## 🧪 Estratégia de testes

| Camada | Tipo | Ferramenta |
|--------|------|-----------|
| Buffer, Cursor, UndoStack | Unitário | `go test` |
| Viewport | Unitário | `go test` |
| Editor (Bubble Tea) | Integração | `go test` + `teatest` |
| Renderização | Snapshot | Comparação de saída ANSI em fixtures |
| Clipboard | Unitário com mock | Interface `Clipboard` |
| File I/O | Unitário com tmpfile | `os.CreateTemp` |

### Princípios de teste
- **TDD**: escrever teste → ver falhar → implementar → ver passar
- **Cobertura mínima**: 70% (foco em buffer, cursor, undo, viewport)
- **Cada `internal/` tem seu `_test.go`**
- **Rodar antes de cada commit**: `go test ./...`

---

## 🔑 Decisões técnicas (registro de arquitetura)

| Decisão | Escolha | Alternativa | Motivo |
|---------|---------|-------------|--------|
| Linguagem | Go | Rust, Python | Binário único, cross-compile, GC previsível |
| TUI framework | Bubble Tea + Lip Gloss | tcell, tview | Elm-like, componentes prontos, time-to-market |
| Estrutura texto v1 | Gap buffer | Rope, Piece table | Simples, boa performance para edição localizada |
| Estrutura texto v2 | Rope | — | Melhor para arquivos grandes e múltiplos cursores |
| Syntax v1 | Chroma (regex) | Treesitter | Simples, sem CGo, 500+ linguagens |
| Syntax v2 | Treesitter | — | Precisão, folding, text objects |
| Plugins | ❌ (Decisão: funcionalidades built-in) | Lua (gopher-lua) | Público-alvo leigo não precisa de plugins; tudo built-in com toggle |
| Clipboard SSH | OSC52 | — | Única solução que funciona via SSH |

---

## 📊 Marcos do projeto

| Marco | Fase | Critério de aceite |
|-------|------|--------------------|
| **M1** | 0 | `go run .` mostra "cmdit v0.1.0" na tela |
| **M2** | 1 | Abre, digita, Ctrl+S, reabre — texto persiste |
| **M3** | 2 | Mouse, setas, scroll, números de linha |
| **M4** | 3 | Undo/Redo, clipboard, Ctrl+F busca |
| **M5** | 4 | Ctrl+P palette com todos os comandos |
| **M6** | 5 | Syntax highlight com 5 temas, detecção automática |
| **M7: v1** | 6+7 | Binário único 3 plataformas, usável por leigos |
| **M8: v2** | 8 | LSP, múltiplos cursores, abas, splits |
| **M9: completo** | 9 | Auto-close, Vim mode, format on save, 5 temas, keybindings |
| **M10: robusto** | 10 | Cobertura 70%+, 0 lint warnings, edge cases tratados |
| **M11: release** | 11 | Binários 6 plataformas, CI/CD, docs, installer |

---

## 🎮 Atalhos de referência

| Atalho | Ação | Fase |
|--------|------|------|
| Digitar | Inserir texto | 1 |
| `Ctrl+S` | Salvar | 1 |
| `Ctrl+Q` | Sair (confirma se modificado) | 1 |
| `Ctrl+O` | Abrir arquivo | 6 |
| `F3` | Salvar como | 6 |
| `F2` | Renomear arquivo | 6 |
| Setas | Mover cursor | 2 |
| `Ctrl+←/→` | Pular palavra | 2 |
| `Home/End` | Início/fim da linha | 2 |
| `Ctrl+Home/End` | Início/fim do arquivo | 2 |
| `PageUp/PageDown` | Rolar página | 2 |
| Clique, scroll | Mouse | 2 |
| `Ctrl+Z` | Undo | 3 |
| `Ctrl+Y` | Redo | 3 |
| `Ctrl+C` | Copiar | 3 |
| `Ctrl+X` | Recortar | 3 |
| `Ctrl+V` | Colar | 3 |
| `Ctrl+A` | Selecionar tudo | 3 |
| `Ctrl+F` | Buscar | 3 |
| `Ctrl+H` | Substituir | 3 |
| `Ctrl+P` | Command palette | 4 |
| `Ctrl+T` | Nova aba | 8 |
| `Ctrl+W` | Fechar aba | 8 |
| `Ctrl+Tab` | Alternar aba | 8 |
| `Ctrl+D` | Selecionar próx. ocorrência | 8 |
| `F4` | Toggle auto-close brackets | 9 |
| `F5` | Toggle modo Vim | 9 |
| `F6` | Próximo tema | 9 |
| `Alt+Z` | Toggle word wrap | 9 |
| `F12` | Ir para definição (LSP) | 8 |
| `Ctrl+Space` | Auto-complete (LSP) | 8 |

---

> **Última atualização:** 2026-05-01
> **Versão:** 2.0 — Reestruturação: Fase 9 (built-in features), Fase 10 (qualidade), Fase 11 (produção)
