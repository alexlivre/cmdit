# Relatorio de Bugs e Features Incompletas -- cmdit

## Resumo

Analise completa do codigo-fonte do cmdit (v0.4.2) revelou **16 bugs** (4 de severidade alta, 8 media, 4 baixa), **3 features parcialmente implementadas** e diversas **vulnerabilidades de seguranca e usabilidade**. O editor esta funcional e bem estruturado, mas ha lacunas importantes em multi-cursor, LSP, search highlight e tratamento de erros.

---

## Bugs Encontrados

### 1. Goroutine leak no LSP `readLoop`
- **Severidade**: Alta
- **Localizacao**: `internal/lsp/lsp.go:247`, `internal/editor/lsp_integration.go:67`
- **Descricao**: `NewClient` inicia uma goroutine `readLoop` (linha 247) que nunca e explicitamente parada. Quando `startLSP` falha no `client.Initialize` (lsp_integration.go:67) e chama `client.Shutdown()`, a goroutine continua rodando. Se o pipe do servidor nunca fechar, a goroutine vaza para sempre.
- **Impacto**: Memory leak de goroutines por toda a sessao da aplicacao. Multiplas aberturas de arquivos acumulam goroutines zumbis.
- **Sugestao**: Adicionar um `context.Context` ou `chan struct{}` de cancelamento ao `Client`, usar `select` no `readLoop` para saida limpa.

### 2. Undo quebrado com multi-cursor
- **Severidade**: Alta
- **Localizacao**: `internal/editor/editor.go:336-362` (`insertTextAtAllCursors`)
- **Descricao**: Com N cursores, `insertTextAtAllCursors` faz N chamadas a `m.undoStack.Push()` (uma por cursor). Ao desfazer (Ctrl+Z), so UMA operacao e revertida por vez. O usuario precisa de N Ctrl+Z para desfazer uma unica digitacao multi-cursor. Alem disso, entre os undos parciais, o buffer fica em estado inconsistente (texto inserido em alguns cursores mas nao em outros).
- **Impacto**: Experiencia de undo quebrada -- comportamento imprevisivel e estado de buffer inconsistente.
- **Sugestao**: Agrupar as N insercoes em uma unica operacao composta (`Operation` com slice de posicoes), ou criar um tipo `CompositeOperation`.

### 3. `addNextOccurrence` com calculo de posicao incorreto
- **Severidade**: Alta
- **Localizacao**: `internal/editor/editor.go:634-696` (linhas 667-675)
- **Descricao**: O calculo de `actualPos` e excessivamente complexo e contem um bug. A expressao `searchStart-lastPos+searchStart` (linha 671) reduz a `2*searchStart - lastPos`, que nao faz sentido dimensionalmente. Alem disso, ha condicoes de corrida entre as branches `idx` que podem produzir a posicao errada para matches encontrados apos wrap-around.
- **Impacto**: Cursor extra adicionado na posicao errada ao usar Ctrl+D.
- **Sugestao**: Simplificar:
  ```go
  if idx < searchStart {
      actualPos = idx  // came from wrap-around, idx is absolute
  } else {
      actualPos = searchStart + idx  // idx is relative to searchStart
  }
  ```

### 4. `doReplace` sem push para undo stack
- **Severidade**: Alta
- **Localizacao**: `internal/editor/actions.go:294-308`
- **Descricao**: `doReplace` deleta o termo de busca e insere o de substituicao (linhas 300-305) mas nunca registra operacao no `undoStack`. Ctrl+Z apos replace nao reverte a alteracao.
- **Impacto**: Perda de integridade do undo -- replace e irreversivel.
- **Sugestao**: Adicionar `m.undoStack.Push` antes das operacoes de delecao/insercao, similar ao que `insertText` faz.

### 5. `applySearchHighlight` usa byte-length como queryLen
- **Severidade**: Media
- **Localizacao**: `internal/editor/view.go:462`
- **Descricao**: `queryLen := len(m.searchQuery)` retorna o comprimento em bytes. Porem `m.searchMatches` armazena posicoes logicas de **runas**. Com caracteres multibyte (ex: "cafe" com acento, emojis), os indices nao coincidem e o highlight e aplicado na posicao errada.
- **Impacto**: Highlight de busca quebrado para queries com caracteres nao-ASCII.
- **Sugestao**: Usar `utf8.RuneCountInString(m.searchQuery)` (como `doReplace` ja faz em actions.go:301).

### 6. `applyFormat` descarta historico de undo
- **Severidade**: Media
- **Localizacao**: `internal/editor/format.go:55`
- **Descricao**: `applyFormat` cria um buffer novo (`buffer.NewBufferFromString(formatted)`) descartando todo o historico de undo. Apos format-on-save, o usuario nao pode desfazer edicoes anteriores.
- **Impacto**: Undo resetado silenciosamente apos formatacao automatica.
- **Sugestao**: Apos criar novo buffer, chamar `m.undoStack.Clear()` explicitamente para deixar claro, ou melhor: aplicar diff no buffer existente preservando o historico.

### 7. `cut()` sem push para undo stack
- **Severidade**: Media
- **Localizacao**: `internal/editor/actions.go:186-194`
- **Descricao**: Quando nao ha selecao, `cut()` deleta a linha inteira (linhas 188-192) sem registrar no undoStack. Apenas a delecao de selecao (linha 184) registra undo via `deleteSelection`.
- **Impacto**: Ctrl+X sem selecao e irreversivel.
- **Sugestao**: Adicionar `m.undoStack.Push` antes da delecao da linha, similar ao `deleteSelection`.

### 8. `handleBackspace` com extra cursors decrementa cursor incorretamente
- **Severidade**: Media
- **Localizacao**: `internal/editor/keys.go:329-338`
- **Descricao**: O `cursor.Col` (linha 329) e `extraCursors[i].Col` (linha 334) sao sempre decrementados, mesmo quando o backspace foi pulado para aquele cursor (ex: GapPos == 0). O cursor primario fica com Col=-1 se estiver na posicao 0.
- **Impacto**: Posicao de cursor incorreta apos backspace com multi-cursor na borda esquerda.
- **Sugestao**: Mover o decremento para dentro do loop `for i := len(all) - 1; i >= 0; i--`, apenas para os cursores onde a delecao realmente ocorreu.

### 9. Race condition ao ler LSP diagnostics
- **Severidade**: Media
- **Localizacao**: `internal/editor/view.go:239-244`
- **Descricao**: `renderStatus` faz lock, conta `totalDiags`, unlock, e depois (linhas 249-256) itera sobre `m.diagnostics` **sem lock** para contar erros/warnings. TOCTOU: os diagnostics podem mudar entre as duas leituras.
- **Impacto**: Potencial leitura inconsistente de dados (panico improvavel, mas counts incorretos).
- **Sugestao**: Fazer toda a iteracao de contagem dentro de um unico bloco `Lock()/defer Unlock()`.

### 10. LSP `request` bloqueia indefinidamente sem timeout
- **Severidade**: Media
- **Localizacao**: `internal/lsp/lsp.go:359`
- **Descricao**: `resp := <-ch` bloqueia para sempre se o servidor LSP nunca responder ou o `readLoop` ja tiver saido. Nao ha mecanismo de timeout.
- **Impacto**: A aplicacao trava se o servidor LSP falhar silenciosamente.
- **Sugestao**: Adicionar `select` com `time.After` (ex: 10s timeout) e cleanup do pending channel.

### 11. `Backspace()` deixa lixo no gap
- **Severidade**: Media
- **Localizacao**: `internal/buffer/buffer.go:60-67`
- **Descricao**: A implementacao de `Backspace()` apenas decrementa `gapStart` sem limpar o rune deletado. Embora tecnicamente correto para gap buffer (area do gap e indefinida), `MoveGapLeft` (linha 95) copia `b.data[b.gapStart]` para `b.data[b.gapEnd]`. Se `Backspace` foi chamado e depois `MoveGapLeft` novamente, o valor obsoleto pode ser copiado para uma posicao ativa do buffer.
- **Impacto**: Corrupcao de dados em sequencias especificas: Backspace seguido de movimento do gap sobre a mesma regiao.
- **Sugestao**: Nao e necessario limpar, mas `MoveGapLeft` deveria verificar se esta lendo de dentro do gap (gapStart vs gapEnd).

### 12. `getSelectedText` e O(n^2) -- nao usa strings.Builder
- **Severidade**: Baixa
- **Localizacao**: `internal/editor/actions.go:227-230`
- **Descricao**: Constroi string com `sb += string(...)` em loop, realocando a cada iteracao. Para selecoes grandes (ex: select all em arquivo de 1MB), performance degrada significativamente.
- **Impacto**: Lenta para selecoes grandes (Select All + Copy).
- **Sugestao**: Usar `strings.Builder`.

### 13. `applyIndentGuides` quebra styling ANSI
- **Severidade**: Baixa
- **Localizacao**: `internal/editor/view.go:183-193`
- **Descricao**: O codigo substitui caracteres de espaco nas posicoes de indentacao por `│` estilizado. Se o caractere original ja tiver um estilo (ex: highlight de sintaxe), a substituicao corrompe a renderizacao pois mistura bytes ANSI.
- **Impacto**: Indent guides podem quebrar visualmente em linhas com syntax highlighting.
- **Sugestao**: Aplicar indent guides ANTES da syntax highlighting na pipeline de renderizacao.

### 14. `paletteActions` contem `view.go-line` mas sem handler
- **Severidade**: Media
- **Localizacao**: `internal/editor/editor.go:240`, `internal/editor/actions.go:16-78`
- **Descricao**: A acao `view.go-line` (Ctrl+G) e registrada no palette (linha 240 do editor.go) mas nao tem case no `executeAction` (actions.go:16-78). Selecionar "Go to Line" no command palette nao faz nada.
- **Impacto**: Funcionalidade "Go to Line" listada como disponivel mas nao funcional.
- **Sugestao**: Implementar um `m.enterGoToLine()` similar ao `enterRename()`, ou remover a acao do registro.

### 15. `MoveGapLeft` retorna valor errado
- **Severidade**: Baixa
- **Localizacao**: `internal/buffer/buffer.go:89-97`
- **Descricao**: Apos mover o gap, `b.data[b.gapStart]` na linha 96 contem o caractere que ESTAVA em `gapStart-1`, que acabou de ser movido -- mas esse valor foi sobrescrito na linha 95 quando o caractere do gap foi copiado para `b.data[b.gapEnd]`. O retorno correto seria o caractere que foi movido, mas ele agora esta em `b.data[b.gapEnd]` (apos o decremento na linha 94).
- **Impacto**: Valor de retorno incorreto -- felizmente nenhum chamador usa o valor retornado atualmente. Baixo impacto.
- **Sugestao**: Salvar `ch := b.data[b.gapStart]` antes das operacoes de movimento.

### 16. `save()` e `SaveConfig()` ignoram erros silenciosamente
- **Severidade**: Baixa
- **Localizacao**: `internal/editor/actions.go:83-98`, `internal/editor/actions.go:52-77`
- **Descricao**: Erros de `fileio.Save` (linha 88) e `SaveConfig` (diversas chamadas em actions.go) sao ignorados. O usuario nao recebe feedback de falha ao salvar.
- **Impacto**: Usuario pode pensar que salvou, mas os dados foram perdidos.
- **Sugestao**: Adicionar `m.showError(...)` nos casos de erro.

### 17. `undo()`/`redo()` usam `len()` (bytes) para deletar runas, corrompendo buffer com UTF-8 multibyte
- **Severidade**: Alta
- **Localizacao**: `internal/editor/actions.go:153, 173`
- **Descricao**: `undo()` e `redo()` usam `len(op.Text)` para iterar chamadas a `DeleteForward()`. `len()` em Go retorna bytes, mas `DeleteForward()` remove uma runa por chamada. Para caracteres multibyte (acentos, emojis, CJK), o byte-count excede o rune-count, fazendo undo/redo deletar mais runas do que deveria, corrompendo o buffer.
- **Impacto**: Corrupcao silenciosa do buffer ao desfazer/refazer texto com caracteres nao-ASCII.
- **Correcao**: `utf8.RuneCountInString(op.Text)` em `actions.go:153, 173` + teste de regressao `TestUndoRedoMultibyteRunes`.

### 18. `Save()` exportado engole erros e marca arquivo como salvo mesmo apos falha
- **Severidade**: Media
- **Localizacao**: `internal/editor/editor.go:310-324`
- **Descricao**: O metodo publico `Save()` do Model ignorava erros de `fileio.Save` e marcava `m.modified = false` incondicionalmente. Se o disco estivesse cheio ou arquivo readonly, o usuario perdia alteracoes achando que salvou.
- **Impacto**: Perda de dados com falhas de I/O.
- **Correcao**: `Save()` agora delega para `save()` (actions.go:97) que trata erros e mostra mensagem na barra de status.

### 19. Welcome screen mostra versao errada (v0.4.1 em vez de v0.4.2)
- **Severidade**: Baixa
- **Localizacao**: `internal/editor/view.go:333`
- **Descricao**: Welcome screen exibia "cmdit v0.4.1" enquanto o release e v0.4.2 e o CHANGELOG afirmava que foi atualizado.
- **Correcao**: Constante `Version = "v0.4.2"` em `editor.go`, usada no lugar do literal.

---

## Funcionalidades Incompletas

### 1. "Go to Line" (Ctrl+G) -- Parcialmente implementada
- **Funcionalidade**: Go to Line via Ctrl+G ou Command Palette
- **Status**: Registrada no palette, nao implementada
- **Evidencia**: `view.go-line` esta em `registerActions` (editor.go:240) mas ausente em `executeAction` (actions.go:16-78). Nao ha `ModeGoToLine` nem handler.
- **Prioridade**: Media

### 2. System clipboard -- Nao implementada
- **Funcionalidade**: Clipboard do sistema (OSC52 ou similar)
- **Status**: Nao implementada
- **Evidencia**: `internal/clipboard/clipboard.go:2` documenta "system clipboard via OSC52 is planned for Phase 7". O clipboard atual e apenas interno.
- **Prioridade**: Baixa (nao listado como feature ✅)

### 3. LSP auto-complete popup -- Nao implementada (conforme planejado)
- **Funcionalidade**: Auto-complete popup (LSP)
- **Status**: Nao implementada (conforme README: 🔲)
- **Evidencia**: Nao ha codigo de UI para popup de autocomplete. O `lsp.Client` tem `OnCompletion` handler registravel mas nunca e usado. Nao ha `textDocument/completion` request enviado.
- **Prioridade**: Media (planejado para v5)

### 4. Treesitter -- Nao implementada (conforme planejado)
- **Funcionalidade**: Treesitter para syntax highlighting preciso
- **Status**: Nao implementada (conforme README: 🔲)
- **Evidencia**: Syntax highlighting atual usa exclusivamente Chroma. Nao ha codigo relacionado a treesitter.
- **Prioridade**: Baixa (planejado para v5)

### 5. Rope data structure -- Nao implementada (conforme planejado)
- **Funcionalidade**: Rope data structure para arquivos grandes
- **Status**: Nao implementada (conforme README: 🔲)
- **Evidencia**: A estrutura de dados atual e exclusivamente Gap Buffer (`internal/buffer/buffer.go`). Nao ha codigo de rope.
- **Prioridade**: Baixa (planejado para v5)

---

## Melhorias Recomendadas

### Seguranca
1. **Path traversal em `openFile`**: Embora use `filepath.Clean`, nao valida se o resultado esta dentro de um diretorio esperado. Adicionar verificacao de que o path resolvido nao sai do diretorio de trabalho ou home.
2. **`recent.json` sem validacao**: O arquivo e lido como newline-separated mas o nome sugere JSON. Se um atacante modificar `recent.json` com paths maliciosos, o editor os exibe na welcome screen.

### Performance
1. **Chroma tokenization por linha**: `HighlightLine` chama `Tokenize` a cada linha em cada frame de renderizacao. Para arquivos grandes, isso e extremamente caro. Cache de tokens por linha, invalidado apenas quando a linha e modificada.
2. **`m.buf.Lines()` aloca slice inteiro**: Chamado a cada `renderContent`, cria slice de todas as linhas mesmo para viewport pequeno. Usar acesso por indice seria mais eficiente.

### Tratamento de Erros
1. **`logError` ignora erros silenciosamente**: Se `os.OpenFile` falhar, o erro e perdido (linhas 36-38 de errors.go). Deveria tentar `os.Stderr` como fallback.
2. **`safeRun` descarta tipo de retorno**: Em caso de panic, retorna `clearErrorMsg` mesmo quando o chamador espera outro tipo. O chamador pode nao tratar `clearErrorMsg`.

### Usabilidade
1. **Status bar overflow**: `renderStatus` concatena muitos campos (linha 282) -- em terminais estreitos, a barra transborda e corta informacao util.
2. **Welcome screen mostra "v0.4.1" hardcoded**: Linha 351 usa string literal, nao reflete a versao real do binario.
3. **Mensagens de erro em portugues**: Consistente, mas `cursores` e `diretorio vazio` na barra de status nao seguem o ingles do codigo-fonte.

### Testes
1. **Faltam testes para**: Vim mode (`vimmode.go` sem `_test.go` especifico), LSP integration, command palette rendering, auto-close brackets, multi-cursor undo consistency, indent guides, file picker navigation, rename flow completo.
2. **Testes existentes sao bons**: 20 arquivos de teste cobrindo buffer, clipboard, editor, fileio, renderer, tabs.

---

## Anexo: Checklist de Conformidade com README

| Funcionalidade | Status README | Status Real | Observacoes |
|---|---|---|---|
| Modeless | ✅ | ✅ Implementado | Correto |
| Mouse | ✅ | ✅ Implementado | Click, scroll funcionam |
| Familiar shortcuts (Ctrl+S/Z/C/V) | ✅ | ✅ Implementado | Ctrl+S, Ctrl+Z, Ctrl+Y, Ctrl+C, Ctrl+X, Ctrl+V |
| Unlimited Undo/Redo | ✅ | ⚠️ Parcial | Quebrado com multi-cursor (Bug #2) |
| Command Palette (Ctrl+P) | ✅ | ✅ Implementado | Fuzzy search funcionando |
| Syntax Highlighting | ✅ | ✅ Implementado | 5 temas via Chroma |
| Find and Replace (Ctrl+F/H) | ✅ | ⚠️ Parcial | Replace sem undo (Bug #4), highlight bug com nao-ASCII (Bug #5) |
| File Picker (Ctrl+O) | ✅ | ✅ Implementado | Navegacao de diretorios ok |
| Rename File (F2) | ✅ | ✅ Implementado | Confirmacao e validacao ok |
| Auto-save (30s, F9) | ✅ | ✅ Implementado | Tick a cada 30s, indicador [AutoSave] |
| Welcome Screen | ✅ | ✅ Implementado | Recent files mostrados |
| Confirmation Dialog | ✅ | ✅ Implementado | S/D/C para save/discard/cancel |
| Tabs (Ctrl+T/W/F7/F8/Ctrl+1-9) | ✅ | ✅ Implementado | Indicador ● para modificados |
| Multiple Cursors (Ctrl+D, Esc) | ✅ | ⚠️ Parcial | Bugs #2, #3, #8 afetam funcionalidade |
| Splits (Ctrl+\) | ✅ | ✅ Implementado | Split horizontal + toggle focus |
| Indent Guides | ✅ | ⚠️ Parcial | Implementacao quebra styling ANSI (Bug #13) |
| LSP Client (gopls) | ✅ | ⚠️ Parcial | Funciona, mas goroutine leak (Bug #1) e sem timeout (Bug #10) |
| Auto-close Brackets | ✅ | ✅ Implementado | Smart-skip funcionando |
| Vim Mode (F5) | ✅ | ✅ Implementado | Normal/Insert/Visual/Command |
| Theme Switching (F6) | ✅ | ✅ Implementado | 5 temas, ciclico |
| Word Wrap (Alt+Z) | ✅ | ✅ Implementado | Alterna wrapping no viewport |
| Format on Save | ✅ | ✅ Implementado | gofmt/black/rustfmt, mas reseta undo (Bug #6) |
| JSON Configuration | ✅ | ✅ Implementado | Load/Save com defaults |
| Customizable Keybindings | ✅ | ✅ Implementado | Via config.json keybindings map |
| Single Binary | ✅ | ✅ (com CGo?) | Chroma pode precisar de CGo dependendo da build |
| Cross-platform | ✅ | ✅ Implementado | Windows/Linux/macOS |
| Treesitter | 🔲 | Nao implementado | Conforme planejado |
| Rope Data Structure | 🔲 | Nao implementado | Conforme planejado |
| Auto-complete Popup (LSP) | 🔲 | Nao implementado | Conforme planejado |
