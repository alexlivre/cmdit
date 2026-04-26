# Test Log — 2026-04-26

## Resumo (2026-04-26 - Ctrl+Shift+S → Ctrl+Alt+S)
- Total de testes: 59
- Passaram: 59 ✅
- Falharam: 0 ❌
- Alteração: atalho Salvar como alterado de Ctrl+Shift+S para Ctrl+Alt+S
- Motivo: mesma limitação — Alt é detectável (ESC prefix), Shift não

## Resumo (2026-04-26 - Ctrl+Shift+P → Ctrl+P)
- Total de testes: 59
- Passaram: 59 ✅
- Falharam: 0 ❌
- Alteração: atalho da paleta de comandos alterado de Ctrl+Shift+P para Ctrl+P
- Motivo: Ctrl+Shift+P é indetectável — protocolo ASCII não distingue Ctrl+Shift+letra de Ctrl+letra

## Resumo (v0.2.0 - rename feature)
- Cobertura: ~70%

## Resultados Detalhados

| Pacote | Testes | Status |
|--------|--------|--------|
| internal/buffer | 14 | ✅ |
| internal/clipboard | 3 | ✅ |
| internal/editor | 24 | ✅ |
| internal/fileio | 8 | ✅ |
| internal/renderer | 10 | ✅ |
| **TOTAL** | **59** | **✅** |

## Feature: Rename File (2026-04-26)

### Novos testes

| Teste | Pacote | Status |
|-------|--------|--------|
| TestRenameSuccess | fileio | ✅ |
| TestRenameSourceNotFound | fileio | ✅ |
| TestRenameEmptyName | fileio | ✅ |
| TestRenameDestinationExists | fileio | ✅ |
| TestRenameNoFile | editor | ✅ |
| TestRenameEnterAndCancel | editor | ✅ |
| TestRenameSuccess | editor | ✅ |
| TestRenameEmptyName | editor | ✅ |
| TestRenameUnchanged | editor | ✅ |
| TestRenameInvalidChars | editor | ✅ |
| TestRenameExistingFile | editor | ✅ |
| TestRenameUnsavedChanges | editor | ✅ |
| TestPaletteActionRename | editor | ✅ |

### Ciclo de Verificação

| Horário | Ação | Resultado |
|---------|------|-----------|
| 2026-04-26 | go test ./... | Todos passando ✅ |
| 2026-04-26 | go vet ./... | Sem warnings ✅ |
| 2026-04-26 | go build ./... | Compilação limpa ✅ |

## Status Final: APROVADO ✅
