# Test Plan — cmdit v0.4.0 (Fase 9)

## Setup
- Create temp file, open with cmdit
- Use interactive terminal to send key commands

## Test Sequence

### 1. Core Editor Features
| # | Teste | Ação | Resultado esperado |
|---|-------|------|-------------------|
| 1.1 | Abrir arquivo | `cmdit test.txt` | Abre editor com "test.txt" vazio |
| 1.2 | Digitar texto | Digitar "Hello World" | Texto aparece na tela |
| 1.3 | Salvar | Ctrl+S | Arquivo salvo, indicador "●" some |
| 1.4 | Navegação | Setas, Home, End | Cursor move corretamente |
| 1.5 | Selecionar tudo | Ctrl+A | Todo texto selecionado |
| 1.6 | Copiar | Ctrl+C | Texto copiado |
| 1.7 | Colar | Ctrl+V | Texto colado |
| 1.8 | Undo | Ctrl+Z | Última ação desfeita |
| 1.9 | Redo | Ctrl+Y | Ação refeita |
| 1.10 | Sair | Ctrl+Q | Fecha o editor |

### 2. Fase 8 — Power Features
| # | Teste | Ação | Resultado esperado |
|---|-------|------|-------------------|
| 2.1 | Nova aba | Ctrl+T | Nova aba vazia |
| 2.2 | Alternar aba | Ctrl+Tab | Troca entre abas |
| 2.3 | Fechar aba | Ctrl+W | Fecha aba atual |
| 2.4 | Pular aba | Ctrl+1 | Vai para aba 1 |
| 2.5 | Multi-cursor | Ctrl+D | Adiciona cursor na próx. ocorrência |
| 2.6 | Limpar cursores | Escape | Remove cursores extras |
| 2.7 | Split | Ctrl+\ | Divide a tela |

### 3. Fase 9 — Features Built-in
| # | Teste | Ação | Resultado esperado |
|---|-------|------|-------------------|
| 3.1 | Auto-close ( | Digitar `(` | Insere `()` com cursor no meio |
| 3.2 | Smart-skip `)` | Digitar `)` | Pula sobre `)` existente |
| 3.3 | Auto-close `"` | Digitar `"` | Insere `""` |
| 3.4 | Auto-close `{` | Digitar `{` | Insere `{}` |
| 3.5 | Vim toggle | F5 | Ativa modo vim |
| 3.6 | Vim hjkl | h, j, k, l | Cursor move |
| 3.7 | Vim :w | `:w` + Enter | Salva arquivo |
| 3.8 | Vim :q | `:q` + Enter | Fecha editor |
| 3.9 | Theme switch | F6 | Theme muda |
| 3.10 | Word wrap | Alt+Z | Wrap ativa/desativa |
| 3.11 | Format on save | Ctrl+S (com código Go) | Código formatado |

### 4. Funcionalidades Diversas
| # | Teste | Ação | Resultado esperado |
|---|-------|------|-------------------|
| 4.1 | Command palette | Ctrl+P | Palette abre |
| 4.2 | Busca | Ctrl+F | Input de busca aparece |
| 4.3 | File picker | Ctrl+O | Navegador de arquivos |
| 4.4 | Renomear | F2 | Input de renomeação |
| 4.5 | Save as | F3 | Input de salvar como |

## Após Testes
- Relatar bugs encontrados
- Verificar binário final
