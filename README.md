# cmdit — Editor de texto para humanos no terminal

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Tests](https://img.shields.io/badge/testes-65%20passando-brightgreen)]()
[![Version](https://img.shields.io/badge/versão-v0.3.1-blue)](https://github.com/alexlivre/cmdit/releases/tag/v0.3.1)
[![Licença](https://img.shields.io/badge/licença-MIT-blue)](LICENSE)
[![Plataformas](https://img.shields.io/badge/plataformas-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)]()

**cmdit** é um editor de texto para terminal feito para pessoas — não para máquinas.  
Abre, digita, sai. Funciona como o Bloco de Notas. **Zero curva de aprendizado.**

Projetado para rodar tanto no seu desktop quanto via SSH em servidores.  
Um único binário, zero dependências.

---

## Por que existe?

| Editor | Problema |
|--------|----------|
| **Vim** | Modos, `:wq`, `hjkl`, curva de meses |
| **Nano** | Muito limitado, sem mouse, sem undo |
| **Emacs** | `Ctrl+X Ctrl+C` para sair |
| **Helix** | Paradigma `selection→action` confuso |
| **Micro** | Abandonado, sem LSP nativo |

**cmdit** resolve: poderoso como Vim, simples como Nano, moderno como Helix.

---

## ✨ Funcionalidades

- ✅ **Modeless** — digite como em qualquer editor. Sem modos Normal/Insert.
- ✅ **Mouse** — clique para posicionar cursor, arraste para selecionar, scroll para rolar.
- ✅ **Atalhos familiares** — Ctrl+S, Ctrl+Z, Ctrl+C, Ctrl+V. Padrão CUA.
- ✅ **Undo/Redo ilimitado** — Ctrl+Z / Ctrl+Y.
- ✅ **Command Palette** — Ctrl+P acessa todos os comandos com busca fuzzy.
- ✅ **Syntax Highlighting** — Detecção automática de linguagem. Temas dark/light/monokai/dracula/solarized.
- ✅ **Busca e substituição** — Ctrl+F / Ctrl+H com destaque de todas as ocorrências.
- ✅ **File picker integrado** — Ctrl+O para abrir arquivos navegando por diretórios.
- ✅ **Renomear arquivo** — F2 renomeia o arquivo atual com input inline.
- ✅ **Auto-save** — Salva automaticamente a cada 30 segundos.
- ✅ **Tela de boas-vindas** — Mostra arquivos recentes ao abrir sem arquivo.
- ✅ **Diálogo de confirmação** — Avisa ao sair se houver alterações não salvas.
- ✅ **Abas** — Ctrl+T nova aba, Ctrl+W fecha, Ctrl+Tab alterna, Ctrl+1-9 salta. Indicador ● para arquivos modificados.
- ✅ **Múltiplos cursores** — Ctrl+D adiciona cursor na próxima ocorrência da palavra, Escape limpa. Edição simultânea em todos os cursores.
- ✅ **Splits** — Ctrl+\ divide a tela horizontalmente. Clique para focar painel, Ctrl+\ alterna foco.
- ✅ **Indent guides** — Linhas verticais sutis nos níveis de indentação.
- ✅ **LSP client** — Inicia automaticamente gopls para Go. Diagnósticos de erros e warnings na barra de status. Extensível a Python, Rust, TypeScript.
- ✅ **Single binary** — Um arquivo. Copia e roda. Zero dependências.
- ✅ **Cross-platform** — Windows, Linux, macOS. Desktop e servidor SSH.

### 📋 Planejado (v3)

- 🔲 Treesitter para syntax highlighting preciso
- 🔲 Plugins em Lua
- 🔲 Vim keymap opcional via plugin

---

## ⚡ Instalação

### Pré-compilado (recomendado)

```bash
# Baixe o binário para sua plataforma em:
# https://github.com/alexb/cmdit/releases

# Linux/macOS
chmod +x cmdit
sudo mv cmdit /usr/local/bin/

# Windows
# Mova cmdit.exe para uma pasta no PATH
```

### Go Install

```bash
go install github.com/alexb/cmdit/cmd/cmdit@latest
```

### Build from source

```bash
git clone https://github.com/alexb/cmdit.git
cd cmdit
go build -o cmdit ./cmd/cmdit
```

---

## 🚀 Uso

```bash
# Abrir um arquivo
cmdit README.md

# Criar novo arquivo
cmdit novo-arquivo.txt

# Abrir sem arquivo (tela de boas-vindas)
cmdit
```

---

## ⌨️ Guia de atalhos

### Arquivo

| Atalho | Ação |
|--------|------|
| `Ctrl+S` | Salvar |
| `Ctrl+O` | Abrir arquivo |
| `F3` | Salvar como |
| `F2` | Renomear arquivo |
| `Ctrl+Q` | Sair |

### Abas e Splits

| Atalho | Ação |
|--------|------|
| `Ctrl+T` | Nova aba |
| `Ctrl+W` | Fechar aba |
| `Ctrl+Tab` | Próxima aba |
| `Ctrl+1-9` | Ir para aba N |
| `Ctrl+\` | Criar/alternar split |

### Edição

| Atalho | Ação |
|--------|------|
| `Ctrl+Z` | Desfazer |
| `Ctrl+Y` | Refazer |
| `Ctrl+C` | Copiar (seleção ou linha atual) |
| `Ctrl+X` | Recortar |
| `Ctrl+V` | Colar |
| `Ctrl+A` | Selecionar tudo |
| `Backspace` | Apagar caractere à esquerda |
| `Delete` | Apagar caractere à direita |
| `Tab` | Inserir 4 espaços |
| `Ctrl+D` | Adicionar cursor (próx. ocorrência) |
| `Escape` | Limpar cursores extras |

### Navegação

| Atalho | Ação |
|--------|------|
| `↑ ↓ ← →` | Mover cursor |
| `Ctrl+← / Ctrl+→` | Pular palavras |
| `Home / End` | Início / fim da linha |
| `Ctrl+Home / Ctrl+End` | Início / fim do arquivo |
| `PageUp / PageDown` | Rolar página |
| `Clique do mouse` | Posicionar cursor |
| `Scroll do mouse` | Rolar |

### Busca

| Atalho | Ação |
|--------|------|
| `Ctrl+F` | Buscar |
| `Ctrl+H` | Buscar e substituir |

### Comandos

| Atalho | Ação |
|--------|------|
| `Ctrl+P` | Paleta de comandos |
| `Esc` | Fechar paleta / cancelar |

---

## 🆚 Comparativo com concorrentes

|  | **cmdit** | Vim | Nano | Helix | Micro |
|--|-----------|-----|------|-------|-------|
| Modeless | ✅ | ❌ | ✅ | ⚠️ | ✅ |
| Mouse | ✅ | ❌ | ✅ | ❌ | ✅ |
| Ctrl+C/V/Z | ✅ | ❌ | ❌ | ❌ | ✅ |
| Command Palette | ✅ | ❌ | ❌ | ✅ | ✅ |
| Undo/Redo | ✅ | ✅ | ❌ | ✅ | ✅ |
| Syntax Highlight | ✅ | ✅ | ✅ | ✅ | ✅ |
| File Picker | ✅ | ❌ | ❌ | ❌ | ❌ |
| Auto-save | ✅ | ❌ | ❌ | ❌ | ❌ |
| Tabs | ✅ | ✅ | ❌ | ❌ | ✅ |
| Splits | ✅ | ✅ | ❌ | ✅ | ✅ |
| Multi-cursor | ✅ | ❌ | ❌ | ✅ | ✅ |
| LSP nativo | ✅ | ❌ | ❌ | ✅ | ❌ |
| Single Binary | ✅ | ✅ | ✅ | ✅ | ✅ |
| Curva de aprendizado | 🟢 segundos | 🔴 meses | 🟢 segundos | 🟡 dias | 🟢 minutos |

---

## 🏗️ Arquitetura

```
cmdit/
├── cmd/cmdit/main.go            # Entry point (SplitContainer → TabManager → editor.Model)
├── internal/
│   ├── buffer/                  # Gap buffer + cursor + undo stack
│   ├── clipboard/               # Clipboard interno
│   ├── command/                 # Command palette + action registry
│   ├── editor/                  # Modelo Bubble Tea principal
│   │   ├── editor.go            # Model, Init, Update, helpers
│   │   ├── view.go              # View + renderização
│   │   ├── keys.go              # Key dispatch + modos
│   │   ├── actions.go           # Comandos (save, undo, clipboard, etc.)
│   │   └── lsp_integration.go   # Integração LSP
│   ├── fileio/                  # Load/Save/Rename arquivos
│   ├── highlight/               # Syntax highlighting via Chroma
│   ├── lsp/                     # Cliente LSP (JSON-RPC 2.0)
│   ├── renderer/                # Viewport (scroll, resize)
│   └── tabs/                    # Abas e splits
│       ├── tabmanager.go        # TabManager (container de editor.Model)
│       └── split.go             # SplitContainer (layout de painéis)
├── bin/                         # Binários compilados
├── go.mod / go.sum              # Dependências Go
├── Makefile                     # Build, test, cross-compile
└── PLANO.md                     # Plano de desenvolvimento completo
```

### Stack tecnológica

| Camada | Tecnologia | Motivo |
|--------|-----------|--------|
| Linguagem | **Go 1.23+** | Binário único, cross-compile, performance previsível |
| TUI | **Bubble Tea + Lip Gloss** | Arquitetura Elm-like, componentes composáveis |
| Syntax | **Chroma** | 500+ linguagens, sem CGo, rápido |
| Estrutura de texto | **Gap Buffer** | O(1) para edições localizadas, simples |

### Design decisions

- **Modeless-first**: o editor está sempre em modo de inserção. Comandos são acessados via atalhos CUA ou palette.
- **Progressive disclosure**: funcionalidades avançadas são descobertas gradualmente via command palette.
- **Single binary**: compilação estática nativa em Go. Distribuição trivial (`scp cmdit servidor:/usr/local/bin/`).

---

## 🧪 Testes

```bash
# Executar todos os testes
go test ./...

# Com cobertura
go test ./... -cover

# Resultado atual
# 65 testes passando em 6 pacotes
# buffer (17) | clipboard (4) | editor (14) | fileio (4) | renderer (9) | tabs (6)
# go vet ./... ✅ sem warnings
```

---

## 📦 Build

```bash
# Build local
make build          # → bin/cmdit.exe

# Cross-compile (todas as plataformas)
make build-all      # → bin/cmdit-linux-amd64
                    #   bin/cmdit-windows-amd64.exe
                    #   bin/cmdit-darwin-amd64
                    #   bin/cmdit-darwin-arm64

# Testes + lint
make test
make lint
```

---

## 🗺️ Roadmap

| Fase | Status | Descrição |
|------|--------|-----------|
| 0 | ✅ | Scaffold: Go + Bubble Tea |
| 1 | ✅ | Editor básico: digitar, salvar, sair |
| 2 | ✅ | Navegação: setas, mouse, scroll, viewport |
| 3 | ✅ | Edição power: undo, clipboard, busca |
| 4 | ✅ | Command palette |
| 5 | ✅ | Syntax highlighting via Chroma |
| 6 | ✅ | File picker, welcome screen, recentes |
| 7 | ✅ | Auto-save, cross-compile, polimento |
| 8 | ✅ | Múltiplos cursores, abas, splits, indent guides, LSP |
| 9 | 📋 | Plugins Lua + marketplace |
| 10 | 📋 | Treesitter syntax highlighting |

---

## 🤝 Contribuindo

Contribuições são bem-vindas!

1. Fork o repositório
2. Crie uma branch: `git checkout -b feature/nome`
3. Faça commit: `git commit -m "feat: descrição"`
4. Push: `git push origin feature/nome`
5. Abra um Pull Request

**Convenções de commit:**  
`feat:` nova funcionalidade  
`fix:` correção de bug  
`refactor:` refatoração  
`test:` testes  
`docs:` documentação  
`chore:` manutenção

---

## 📄 Licença

MIT © 2026 Alexb

---

**cmdit** — *editor de texto para humanos.*
