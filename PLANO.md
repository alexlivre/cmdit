# PLANO DE DESENVOLVIMENTO — cmdit

> **Meta:** Construir um editor de texto TUI modeless, CUA-first, que seja a melhor alternativa ao Vim — usável por leigos em desktop e por devs em servidores SSH.
>
> **Arquitetura:** Go + Bubble Tea + Lip Gloss. Elm-like (Model/Update/View). Gap buffer v1, Rope v2. Command palette como camada de descoberta. Lua para plugins.
>
> **Stack:** Go 1.22+, Bubble Tea, Bubbles, Lip Gloss (charm.sh), Chroma (syntax v1), gopher-lua (plugins v2), Treesitter (v2)

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
| Plugin system | ✅ | ❌ | ❌ | ❌ | ✅ (Lua v2) |
| Vim keymap | ✅ | ❌ | ❌ | ❌ | 🔌 plugin |
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
├── plugins/
│   └── lua/
│       └── runtime.go               # Lua VM (v2)
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
| **8** | Power features | 3-4 sem | Múltiplos cursores, abas, splits, LSP |
| **9** | Plugins Lua | 2-3 sem | Lua runtime, API, marketplace |
| **10** | Produção | 2 sem | Cross-compile, CI/CD, installer, docs |

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

- [ ] **6.2** Implementar Ctrl+Alt+S — Salvar como
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
  - Se sim, avisar: "Ctrl+S bloqueado pelo terminal. Use `stty -ixon` ou Ctrl+Alt+S"

- [ ] **7.7** Executar `go test ./...` e `go vet ./...`
- [ ] **7.8** Build cross-platform: Windows, Linux, macOS

### Marcos
- **M7: v1 release** — Binário único, usável por leigos, pronto para distribuir ✅

---

## 📋 Fase 8 — Power Features (v2)

- [ ] **8.1** Múltiplos cursores
  - Ctrl+D: seleciona próxima ocorrência da palavra atual
  - Ctrl+Click: adiciona cursor na posição
  - Escape: remove todos os cursores extras

- [ ] **8.2** Abas
  - Ctrl+T: nova aba (buffer vazio)
  - Ctrl+W: fecha aba atual (confirma se modificada)
  - Ctrl+Tab / Ctrl+Shift+Tab: alternar abas
  - UI: barra de abas no topo com nome do arquivo

- [ ] **8.3** Splits
  - Horizontal/vertical
  - Redimensionamento com mouse ou Ctrl+W + setas
  - Cada split é um editor independente

- [ ] **8.4** LSP client
  - Cliente LSP em Go (protocolo JSON-RPC sobre stdio)
  - Auto-complete: popup ao digitar (Ctrl+Space)
  - Go-to-definition: F12
  - Diagnostics: sublinhado ondulado em erros
  - Configuração: `~/.cmdit/lsp.json`

- [ ] **8.5** Indent guides
  - Linhas verticais sutis para cada nível de indentação
  - Toggle: `view.toggle-indent-guides`

---

## 📋 Fase 9 — Plugins Lua

- [ ] **9.1** Integrar gopher-lua VM
  - `plugins/lua/runtime.go` — inicializa VM, expõe API
  - Bloquear operações perigosas (file system, network) via sandbox

- [ ] **9.2** Definir API de plugins
  ```lua
  cmdit.map("Ctrl+Shift+M", function() ... end)
  cmdit.buffer.insert("texto")
  cmdit.buffer.get_line(42)
  cmdit.cursor.move(10, 5)
  cmdit.editor.save()
  cmdit.ui.show_message("Pronto!")
  ```

- [ ] **9.3** Event hooks
  - `cmdit.on_open(function(path) ... end)`
  - `cmdit.on_save(function(path) ... end)`
  - `cmdit.on_key(function(key) ... end)`

- [ ] **9.4** Sistema de plugins
  - `cmdit plugin install <nome>` — clona de git, copia para `~/.cmdit/plugins/`
  - `cmdit plugin list` — lista plugins instalados
  - `cmdit plugin remove <nome>` — remove plugin

- [ ] **9.5** Criar plugin exemplo: `vim-keybindings`
  - Mapeia hjkl, `:w`, `:q`, etc.
  - Prova que a API é poderosa o suficiente

---

## 📋 Fase 10 — Produção

- [ ] **10.1** Cross-compile no Makefile
  ```makefile
  build-all:
    GOOS=linux GOARCH=amd64 go build -o bin/cmdit-linux-amd64 ./cmd/cmdit
    GOOS=linux GOARCH=arm64 go build -o bin/cmdit-linux-arm64 ./cmd/cmdit
    GOOS=windows GOARCH=amd64 go build -o bin/cmdit-windows-amd64.exe ./cmd/cmdit
    GOOS=darwin GOARCH=amd64 go build -o bin/cmdit-darwin-amd64 ./cmd/cmdit
    GOOS=darwin GOARCH=arm64 go build -o bin/cmdit-darwin-arm64 ./cmd/cmdit
  ```

- [ ] **10.2** GitHub Actions CI/CD
  - `go test ./...` + `go vet ./...` + `golangci-lint`
  - Build multi-plataforma
  - Release automático ao criar tag

- [ ] **10.3** Script de instalação
  ```bash
  curl -sSL https://cmdit.dev/install.sh | bash
  # Detecta OS/arch, baixa binário, copia para /usr/local/bin
  ```

- [ ] **10.4** Documentação
  - Site: cmdit.dev (Hugo ou Markdown no GitHub Pages)
  - Quickstart: "5 minutos para dominar o cmdit"
  - Guia de atalhos (cheatsheet)
  - Guia de plugins
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
| Plugins | Lua (gopher-lua) | Yaegi, JavaScript | Familiar para usuários Neovim, sandbox fácil |
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
| **M8: v2** | 8+9+10 | LSP, plugins Lua, múltiplos cursores, CI/CD |

---

## 🎮 Atalhos de referência

| Atalho | Ação | Fase |
|--------|------|------|
| Digitar | Inserir texto | 1 |
| `Ctrl+S` | Salvar | 1 |
| `Ctrl+Q` | Sair (confirma se modificado) | 1 |
| `Ctrl+O` | Abrir arquivo | 6 |
| `Ctrl+Alt+S` | Salvar como | 6 |
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
| `F12` | Ir para definição (LSP) | 8 |
| `Ctrl+Space` | Auto-complete (LSP) | 8 |

---

> **Última atualização:** 2026-04-26
> **Versão:** 1.0
