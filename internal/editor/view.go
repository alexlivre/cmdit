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

	textWidth := m.viewport.Width() - lineNumWidth - 1
	if textWidth < 5 {
		textWidth = 5
	}

	var visibleLines []string

	if m.config.WordWrap {
		displayLineCount := 0
		for i := m.viewport.ScrollY(); i < len(lines) && displayLineCount < contentHeight; i++ {
			wrapped := wrapText(lines[i], textWidth)
			for wi, wl := range wrapped {
				if displayLineCount >= contentHeight {
					break
				}
				var prefix string
				if wi == 0 {
					lineNum := fmt.Sprintf("%*d ", lineNumWidth, i+1)
					prefix = m.lineNumStyle.Render(lineNum)
				} else {
					prefix = strings.Repeat(" ", lineNumWidth+1)
				}

				// Apply syntax highlighting
				segments := m.highlighter.HighlightLine(wl, m.language)
				lineText := highlight.RenderSegments(segments)

				// Apply indent guides
				lineText = m.applyIndentGuides(lineText, wl)

				// Apply search highlighting on top
				if len(m.search.Matches) > 0 {
					lineText = m.applySearchHighlight(lineText, i)
				}

				if len(lineText) > textWidth {
					lineText = lineText[:textWidth]
				}

				visibleLines = append(visibleLines, prefix+lineText)
				displayLineCount++
			}
		}
	} else {
		for i := m.viewport.ScrollY(); i < len(lines) && i < m.viewport.ScrollY()+contentHeight; i++ {
			lineNum := fmt.Sprintf("%*d ", lineNumWidth, i+1)
			styledLineNum := m.lineNumStyle.Render(lineNum)

			// Apply syntax highlighting
			segments := m.highlighter.HighlightLine(lines[i], m.language)
			lineText := highlight.RenderSegments(segments)

			// Apply indent guides
			lineText = m.applyIndentGuides(lineText, lines[i])

			// Apply search highlighting on top
			if len(m.search.Matches) > 0 {
				lineText = m.applySearchHighlight(lineText, i)
			}

			// Horizontal scroll
			if m.viewport.ScrollX() > 0 && m.viewport.ScrollX() < len(lineText) {
				lineText = lineText[m.viewport.ScrollX():]
			}
			if len(lineText) > textWidth {
				lineText = lineText[:textWidth]
			}

			visibleLines = append(visibleLines, styledLineNum+lineText)
		}
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

	tabWidth := 4
	if indent < tabWidth {
		return lineText
	}

	var b strings.Builder
	runes := []rune(lineText)
	for i, r := range runes {
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
		s := fmt.Sprintf("Find: %s  Replace: %s", m.search.Query, m.search.Replace)
		if len(m.search.Matches) > 0 {
			s += fmt.Sprintf("  [%d/%d]", m.search.Current+1, len(m.search.Matches))
		}
		return m.searchStyle.Render(s)
	}
	s := fmt.Sprintf("Find: %s", m.search.Query)
	if len(m.search.Matches) > 0 {
		s += fmt.Sprintf("  [%d/%d]", m.search.Current+1, len(m.search.Matches))
	}
	return m.searchStyle.Render(s)
}

// --- Status bar ---

func (m *Model) renderStatus() string {
	fname := m.filename
	if fname == "" {
		fname = "[new]"
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

	// Show auto-save indicator
	autoSaveIndicator := ""
	if m.config.AutoSaveEnabled {
		autoSaveIndicator = " [AutoSave]"
	}

	// Show theme
	themeIndicator := fmt.Sprintf(" [Theme:%s]", m.config.Theme)

	status := fname + modified + "  " + pos + mcInfo + diagInfo + autoSaveIndicator + vimIndicator + themeIndicator + "  Ctrl+S Save  Ctrl+Q Quit"

	// Show error overlay if active
	if m.errorMessage != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("160")).Padding(0, 2)
		return m.statusStyle.Render(status) + "\n" + errStyle.Render(" "+m.errorMessage+" ")
	}

	return m.statusStyle.Render(status)
}

// --- Rename bar ---

func (m *Model) renderRenameBar() string {
	oldName := m.filename
	if oldName == "" {
		oldName = "(new)"
	}
	if len(oldName) > 30 {
		oldName = "..." + oldName[len(oldName)-27:]
	}
	s := fmt.Sprintf("Rename: %s → %s", oldName, m.rename.Input)
	if m.rename.Error != "" {
		s += "  " + lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Render("("+m.rename.Error+")")
	}
	return m.searchStyle.Render(s)
}

// --- Search highlighting ---

func (m *Model) applySearchHighlight(lineText string, lineNum int) string {
	if len(m.search.Matches) == 0 {
		return lineText
	}

	lineStart := m.buf.LineStart(lineNum)
	if lineStart < 0 {
		return lineText
	}

	matchMap := make(map[int]int)
	for i, pos := range m.search.Matches {
		matchMap[pos] = i
	}

	var result strings.Builder
	pos := 0
	queryLen := len(m.search.Query)

	for i := 0; i <= len(lineText)-queryLen; i++ {
		logicalPos := lineStart + i
		if matchIdx, ok := matchMap[logicalPos]; ok {
			result.WriteString(lineText[pos:i])

			matchText := lineText[i : i+queryLen]
			if matchIdx == m.search.Current {
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
	if index < 0 || index >= len(m.search.Matches) {
		return
	}
	m.search.Current = index
	pos := m.search.Matches[index]
	line, col := m.buf.LineCol(pos)
	m.cursor.SetPos(line, col)
	m.syncGapToCursor()
	m.viewport.EnsureVisible(line, col)
}

// wrapText wraps text at word boundaries to fit within the given width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	var lines []string
	for len(text) > width {
		breakAt := width
		// Try to find a word boundary (space/tab) in the second half of the line
		for i := width - 1; i >= width/2; i-- {
			if i < len(text) && (text[i] == ' ' || text[i] == '\t') {
				breakAt = i + 1
				break
			}
		}
		lines = append(lines, text[:breakAt])
		text = text[breakAt:]
	}
	if len(text) > 0 {
		lines = append(lines, text)
	}
	return lines
}
