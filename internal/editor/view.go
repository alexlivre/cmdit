package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/alexb/cmdit/internal/highlight"
)

// View renders the editor UI.
func (m *Model) View() string {
	if m.mode == ModeWelcome {
		return m.renderWelcome()
	}
	if m.mode == ModePalette {
		return m.renderPalette()
	}
	if m.mode == ModeFilePicker {
		return m.renderFilePicker()
	}
	if m.mode == ModeSaveAs {
		return m.renderSaveAs()
	}
	if m.mode == ModeConfirm {
		return m.renderConfirm()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderContent(),
		m.renderStatus(),
	)
}

// renderContent renders the text content area with line numbers, search bar, and indent guides.
func (m *Model) renderContent() string {
	var parts []string

	// Search/replace bar
	if m.mode == ModeSearch || m.mode == ModeReplace {
		parts = append(parts, m.renderSearchBar())
	}
	// Rename bar
	if m.mode == ModeRename {
		parts = append(parts, m.renderRenameBar())
	}

	// Main content
	lines := m.buf.Lines()
	lineNumWidth := m.lineNumberWidth()
	contentHeight := m.viewport.Height()
	if m.mode == ModeSearch || m.mode == ModeReplace {
		contentHeight -= 1
	}
	if m.mode == ModeRename {
		contentHeight -= 1
	}

	var visibleLines []string
	for i := m.viewport.ScrollY(); i < len(lines) && i < m.viewport.ScrollY()+contentHeight; i++ {
		lineNum := fmt.Sprintf("%*d ", lineNumWidth, i+1)
		styledLineNum := m.lineNumStyle.Render(lineNum)

		// Apply syntax highlighting
		segments := m.highlighter.HighlightLine(lines[i], m.language)
		lineText := highlight.RenderSegments(segments)

		// Apply indent guides
		lineText = m.applyIndentGuides(lineText, lines[i])

		// Apply search highlighting on top
		if len(m.searchMatches) > 0 {
			lineText = m.applySearchHighlight(lineText, i)
		}

		// Horizontal scroll
		if m.viewport.ScrollX() > 0 && m.viewport.ScrollX() < len(lineText) {
			lineText = lineText[m.viewport.ScrollX():]
		}
		if len(lineText) > m.viewport.Width()-lineNumWidth-1 {
			lineText = lineText[:m.viewport.Width()-lineNumWidth-1]
		}

		visibleLines = append(visibleLines, styledLineNum+lineText)
	}

	content := strings.Join(visibleLines, "\n")
	contentStyle := lipgloss.NewStyle().
		Width(m.viewport.Width()).
		Height(contentHeight)

	parts = append(parts, contentStyle.Render(content))
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// applyIndentGuides draws subtle vertical lines at each indent level.
func (m *Model) applyIndentGuides(lineText string, rawLine string) string {
	// Only show guides on lines with content
	if strings.TrimSpace(rawLine) == "" {
		return lineText
	}

	// Count leading whitespace
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

	// Draw guide at each 4-space boundary
	tabWidth := 4
	var result []rune
	runes := []rune(lineText)
	col := 0

	for _, r := range runes {
		if col > 0 && col < indent && col%tabWidth == 0 {
			result = append(result, []rune(m.indentGuideStyle.Render("│"))...)
			// Skip the actual space character, replace with guide
			// But we need to handle this carefully — the guide replaces the space
		} else {
			result = append(result, r)
		}
		col++
	}

	// Simpler approach: overlay guides on the raw line
	// Actually let me use a simpler approach — build from scratch
	if indent < tabWidth {
		return lineText
	}

	var b strings.Builder
	runes2 := []rune(lineText)
	for i, r := range runes2 {
		if i > 0 && i < indent && i%tabWidth == 0 {
			b.WriteString(m.indentGuideStyle.Render("│"))
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// --- Search bar rendering ---

func (m *Model) renderSearchBar() string {
	if m.mode == ModeReplace {
		s := fmt.Sprintf("Buscar: %s  Substituir: %s", m.searchQuery, m.replaceQuery)
		if len(m.searchMatches) > 0 {
			s += fmt.Sprintf("  [%d/%d]", m.searchCurrent+1, len(m.searchMatches))
		}
		return m.searchStyle.Render(s)
	}
	s := fmt.Sprintf("Buscar: %s", m.searchQuery)
	if len(m.searchMatches) > 0 {
		s += fmt.Sprintf("  [%d/%d]", m.searchCurrent+1, len(m.searchMatches))
	}
	return m.searchStyle.Render(s)
}

// --- Status bar ---

func (m *Model) renderStatus() string {
	fname := m.filename
	if fname == "" {
		fname = "[novo]"
	}
	modified := ""
	if m.modified {
		modified = m.statusModified.Render(" ●")
	}
	lang := m.language
	if lang == "" {
		lang = "texto"
	}
	pos := fmt.Sprintf("L:%d C:%d [%s]", m.cursor.Line+1, m.cursor.Col+1, lang)

	// Show multi-cursor count
	mcInfo := ""
	if len(m.extraCursors) > 0 {
		mcInfo = fmt.Sprintf(" %d cursores", len(m.extraCursors)+1)
	}

	// Show diagnostics count
	diagInfo := ""
	if m.lspClient != nil {
		m.lspDiagnosticsMu.Lock()
		totalDiags := 0
		for _, diags := range m.diagnostics {
			totalDiags += len(diags)
		}
		m.lspDiagnosticsMu.Unlock()

		if totalDiags > 0 {
			errCount := 0
			warnCount := 0
			for _, diags := range m.diagnostics {
				for _, d := range diags {
					if d.Severity == 1 {
						errCount++
					} else if d.Severity == 2 {
						warnCount++
					}
				}
			}
			if errCount > 0 {
				diagInfo += fmt.Sprintf(" ✗%d", errCount)
			}
			if warnCount > 0 {
				diagInfo += fmt.Sprintf(" ⚠%d", warnCount)
			}
		}
	}

	// Show vim mode indicator
	vimIndicator := ""
	if m.config.VimMode {
		vimIndicator = " [VIM]"
	}

	return m.statusStyle.Render(fname + modified + "  " + pos + mcInfo + diagInfo + vimIndicator + "  Ctrl+S Salvar  Ctrl+Q Sair")
}

// --- Confirm dialog ---

func (m *Model) renderConfirm() string {
	var msg string
	if m.confirmAction == ConfirmQuit {
		msg = "Arquivo modificado! Deseja salvar antes de sair?"
	} else {
		msg = "Arquivo modificado! Deseja salvar antes de fechar?"
	}
	btns := "\n\n" + m.confirmBtnStyle.Render("[S] Salvar") + "  [D] Descartar  [C] Cancelar"
	dialog := m.confirmStyle.Render(msg + btns)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// --- Palette ---

func (m *Model) renderPalette() string {
	maxItems := 10
	start := m.paletteSel - 5
	if start < 0 {
		start = 0
	}
	end := start + maxItems
	if end > len(m.paletteResults) {
		end = len(m.paletteResults)
	}

	var items []string
	items = append(items, m.paletteInput.Render("> "+m.paletteQuery))

	if len(m.paletteQuery) == 0 {
		items = append(items, m.paletteShortcut.Render("  Digite para buscar comandos..."))
	}
	if len(m.paletteResults) == 0 && len(m.paletteQuery) > 0 {
		items = append(items, m.paletteShortcut.Render("  Nenhum comando encontrado"))
	}

	for i := start; i < end; i++ {
		a := m.paletteResults[i]
		line := fmt.Sprintf("  %-30s", a.Label)
		if a.Shortcut != "" {
			line += " " + m.paletteShortcut.Render(a.Shortcut)
		}
		if i == m.paletteSel {
			items = append(items, m.paletteActive.Render(strings.TrimRight(line, " ")))
		} else {
			items = append(items, line)
		}
	}

	content := strings.Join(items, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}

// --- Welcome Screen ---

func (m *Model) renderWelcome() string {
	logo := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render("  cmdit v0.2.0")
	subtitle := "Editor de texto para humanos"

	var lines []string
	lines = append(lines, logo)
	lines = append(lines, "")
	lines = append(lines, subtitle)
	lines = append(lines, "")

	if len(m.recentFiles) > 0 {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Arquivos recentes:"))
		for i, f := range m.recentFiles {
			if i >= 5 {
				break
			}
			lines = append(lines, fmt.Sprintf("  %d. %s", i+1, f))
		}
		lines = append(lines, "")
	}

	lines = append(lines, "Ctrl+O  Abrir arquivo")
	lines = append(lines, "Ctrl+P  Paleta de comandos")
	lines = append(lines, "Ctrl+Q  Sair")
	lines = append(lines, "")
	lines = append(lines, "Comece a digitar para criar um novo arquivo...")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// --- File Picker ---

func (m *Model) renderFilePicker() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Abrir: %s > %s", m.filePickerDir, m.filePickerQuery))
	lines = append(lines, "")

	start := m.filePickerSel - 10
	if start < 0 {
		start = 0
	}
	end := start + 20
	if end > len(m.filePickerFiles) {
		end = len(m.filePickerFiles)
	}

	for i := start; i < end; i++ {
		line := "  " + m.filePickerFiles[i]
		if i == m.filePickerSel {
			line = m.paletteActive.Render(line)
		}
		lines = append(lines, line)
	}

	if len(m.filePickerFiles) == 0 {
		lines = append(lines, "  (diretorio vazio)")
	}

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}

// --- Save As ---

func (m *Model) renderSaveAs() string {
	var lines []string
	lines = append(lines, "Salvar como:")
	lines = append(lines, "")
	lines = append(lines, "> "+m.saveAsQuery)
	lines = append(lines, "")
	lines = append(lines, "Enter para confirmar, Esc para cancelar")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}

// --- Rename bar ---

func (m *Model) renderRenameBar() string {
	oldName := m.filename
	if oldName == "" {
		oldName = "(novo)"
	}
	if len(oldName) > 30 {
		oldName = "..." + oldName[len(oldName)-27:]
	}
	s := fmt.Sprintf("Renomear: %s → %s", oldName, m.renameInput)
	if m.renameError != "" {
		s += "  " + lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Render("("+m.renameError+")")
	}
	return m.searchStyle.Render(s)
}

// --- Search highlighting ---

func (m *Model) applySearchHighlight(lineText string, lineNum int) string {
	if len(m.searchMatches) == 0 {
		return lineText
	}

	lineStart := m.buf.LineStart(lineNum)
	if lineStart < 0 {
		return lineText
	}

	var result strings.Builder
	pos := 0
	queryLen := len(m.searchQuery)

	for i := 0; i <= len(lineText)-queryLen; i++ {
		logicalPos := lineStart + i
		isMatch := false
		matchIdx := -1

		for j, matchPos := range m.searchMatches {
			if matchPos == logicalPos {
				isMatch = true
				matchIdx = j
				break
			}
		}

		if isMatch {
			result.WriteString(lineText[pos:i])

			matchText := lineText[i : i+queryLen]
			if matchIdx == m.searchCurrent {
				result.WriteString(m.currentMatch.Render(matchText))
			} else {
				result.WriteString(m.matchStyle.Render(matchText))
			}
			pos = i + queryLen
			i += queryLen - 1
		}
	}
	result.WriteString(lineText[pos:])
	return result.String()
}

func (m *Model) navigateToMatch(index int) {
	if index < 0 || index >= len(m.searchMatches) {
		return
	}
	m.searchCurrent = index
	pos := m.searchMatches[index]
	line, col := m.buf.LineCol(pos)
	m.cursor.SetPos(line, col)
	m.syncGapToCursor()
	m.viewport.EnsureVisible(line, col)
}
