package editor

import (
	"fmt"
	"strings"
	"unicode/utf8"

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
	if m.mode == ModeGoToLine {
		return m.renderGoToLine()
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

				lineWithGuides := m.applyIndentGuides(wl, wl)
				segments := m.highlighter.HighlightLine(lineWithGuides, m.language)
				lineText := highlight.RenderSegments(segments)

				if len(m.searchMatches) > 0 {
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

			rawLine := lines[i]
			displayLine := rawLine
			if m.hasSelection() {
				displayLine = m.applySelectionToLine(rawLine, i)
			}
			lineWithGuides := m.applyIndentGuides(displayLine, rawLine)
			segments := m.highlighter.HighlightLine(lineWithGuides, m.language)
			lineText := highlight.RenderSegments(segments)

			if len(m.searchMatches) > 0 {
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

	if indent < 4 {
		return lineText
	}

	runes := []rune(lineText)
	var b strings.Builder
	b.Grow(len(runes) + indent/4*3)
	for i, r := range runes {
		if i > 0 && i < indent && i%4 == 0 {
			b.WriteString(m.indentGuideStyle.Render("│"))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// applySelectionToLine highlights the selected range on a raw (unstyled) line.
func (m *Model) applySelectionToLine(rawLine string, lineIdx int) string {
	lineStart := m.buf.LineStart(lineIdx)
	runes := []rune(rawLine)
	lineEnd := lineStart + len(runes)

	start, end := m.selStart, m.selEnd
	if start > end {
		start, end = end, start
	}
	if end <= lineStart || start >= lineEnd {
		return rawLine
	}

	selStart := start - lineStart
	if selStart < 0 {
		selStart = 0
	}
	selEnd := end - lineStart
	if selEnd > len(runes) {
		selEnd = len(runes)
	}
	if selStart >= selEnd {
		return rawLine
	}

	var b strings.Builder
	if selStart > 0 {
		b.WriteString(string(runes[:selStart]))
	}
	b.WriteString(m.selectionStyle.Render(string(runes[selStart:selEnd])))
	if selEnd < len(runes) {
		b.WriteString(string(runes[selEnd:]))
	}
	return b.String()
}

// --- Go-to-Line ---

func (m *Model) renderGoToLine() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Go to line: %s", m.goToLineInput))
	lines = append(lines, "")
	lines = append(lines, "Enter to confirm, Esc to cancel")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
}

// --- Search bar rendering ---

func (m *Model) renderSearchBar() string {
	if m.mode == ModeReplace {
		s := fmt.Sprintf("Find: %s  Replace: %s", m.searchQuery, m.replaceQuery)
		if len(m.searchMatches) > 0 {
			s += fmt.Sprintf("  [%d/%d]", m.searchCurrent+1, len(m.searchMatches))
		}
		return m.searchStyle.Render(s)
	}
	s := fmt.Sprintf("Find: %s", m.searchQuery)
	if len(m.searchMatches) > 0 {
		s += fmt.Sprintf("  [%d/%d]", m.searchCurrent+1, len(m.searchMatches))
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
		mcInfo = fmt.Sprintf(" %d cursors", len(m.extraCursors)+1)
	}

	// Show diagnostics count
	diagInfo := ""
	if m.lspClient != nil {
		m.lspDiagnosticsMu.Lock()
		errCount, warnCount := 0, 0
		for _, diags := range m.diagnostics {
			for _, d := range diags {
				if d.Severity == 1 {
					errCount++
				} else if d.Severity == 2 {
					warnCount++
				}
			}
		}
		m.lspDiagnosticsMu.Unlock()

		if errCount > 0 {
			diagInfo += fmt.Sprintf(" E:%d", errCount)
		}
		if warnCount > 0 {
			diagInfo += fmt.Sprintf(" W:%d", warnCount)
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

// --- Confirm dialog ---

func (m *Model) renderConfirm() string {
	var msg string
	if m.confirmAction == ConfirmQuit {
		msg = "File modified! Save before quitting?"
	} else {
		msg = "File modified! Save before closing?"
	}
	btns := "\n\n" + m.confirmBtnStyle.Render("[S] Save") + "  [D] Discard  [C] Cancel"
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
		items = append(items, m.paletteShortcut.Render("  Type to search commands..."))
	}
	if len(m.paletteResults) == 0 && len(m.paletteQuery) > 0 {
		items = append(items, m.paletteShortcut.Render("  No commands found"))
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
	logo := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render("  cmdit " + Version)
	subtitle := "Text editor for humans"

	var lines []string
	lines = append(lines, logo)
	lines = append(lines, "")
	lines = append(lines, subtitle)
	lines = append(lines, "")

	if len(m.recentFiles) > 0 {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Recent files:"))
		for i, f := range m.recentFiles {
			if i >= 5 {
				break
			}
			lines = append(lines, fmt.Sprintf("  %d. %s", i+1, f))
		}
		lines = append(lines, "")
	}

	lines = append(lines, "Ctrl+O  Open file")
	lines = append(lines, "Ctrl+P  Command palette")
	lines = append(lines, "Ctrl+Q  Quit")
	lines = append(lines, "")
	lines = append(lines, "Start typing to create a new file...")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// --- File Picker ---

func (m *Model) renderFilePicker() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Open: %s > %s", m.filePickerDir, m.filePickerQuery))
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
	lines = append(lines, "Save as:")
	lines = append(lines, "")
	lines = append(lines, "> "+m.saveAsQuery)
	lines = append(lines, "")
	lines = append(lines, "Enter to confirm, Esc to cancel")

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.paletteStyle.Render(content))
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
	s := fmt.Sprintf("Rename: %s → %s", oldName, m.renameInput)
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
	queryLen := utf8.RuneCountInString(m.searchQuery)

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
