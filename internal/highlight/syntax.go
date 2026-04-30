// Package highlight provides syntax highlighting using Chroma.
package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

// Theme names.
const (
	ThemeDark          = "dark"
	ThemeLight         = "light"
	ThemeMonokai       = "monokai"
	ThemeDracula       = "dracula"
	ThemeSolarizedDark = "solarized-dark"
)

// Highlighter tokenizes text and applies a color theme.
type Highlighter struct {
	themeName string
	style     *chroma.Style
}

// NewHighlighter creates a highlighter with the given theme name.
func NewHighlighter(themeName string) *Highlighter {
	h := &Highlighter{themeName: themeName}
	h.loadStyle()
	return h
}

func (h *Highlighter) loadStyle() {
	switch h.themeName {
	case ThemeDark:
		h.style = styles.Get("monokai")
		if h.style == nil {
			h.style = styles.Fallback
		}
	case ThemeLight:
		h.style = styles.Get("github")
		if h.style == nil {
			h.style = styles.Fallback
		}
	case ThemeMonokai:
		h.style = styles.Get("monokai")
		if h.style == nil {
			h.style = styles.Fallback
		}
	case ThemeDracula:
		h.style = styles.Get("dracula")
		if h.style == nil {
			h.style = styles.Fallback
		}
	case ThemeSolarizedDark:
		h.style = styles.Get("solarized-dark")
		if h.style == nil {
			h.style = styles.Fallback
		}
	default:
		h.style = styles.Fallback
	}
}

// SetTheme changes the highlighting theme.
func (h *Highlighter) SetTheme(themeName string) {
	h.themeName = themeName
	h.loadStyle()
}

// Theme returns the current theme name.
func (h *Highlighter) Theme() string {
	return h.themeName
}

// DetectLanguage guesses the language based on filename.
func DetectLanguage(filename string) string {
	if filename == "" {
		return ""
	}
	lexer := lexers.Match(filename)
	if lexer != nil {
		return lexer.Config().Name
	}

	// Fallback: check extension manually
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".go"):
		return "Go"
	case strings.HasSuffix(lower, ".py"):
		return "Python"
	case strings.HasSuffix(lower, ".js"):
		return "JavaScript"
	case strings.HasSuffix(lower, ".ts"):
		return "TypeScript"
	case strings.HasSuffix(lower, ".html"):
		return "HTML"
	case strings.HasSuffix(lower, ".css"):
		return "CSS"
	case strings.HasSuffix(lower, ".json"):
		return "JSON"
	case strings.HasSuffix(lower, ".md"):
		return "Markdown"
	case strings.HasSuffix(lower, ".yaml"), strings.HasSuffix(lower, ".yml"):
		return "YAML"
	case strings.HasSuffix(lower, ".rs"):
		return "Rust"
	case strings.HasSuffix(lower, ".java"):
		return "Java"
	case strings.HasSuffix(lower, ".c"), strings.HasSuffix(lower, ".h"):
		return "C"
	case strings.HasSuffix(lower, ".cpp"), strings.HasSuffix(lower, ".hpp"):
		return "C++"
	case strings.HasSuffix(lower, ".sh"), strings.HasSuffix(lower, ".bash"):
		return "Bash"
	case strings.HasSuffix(lower, ".txt"):
		return "plaintext"
	}
	return ""
}

// HighlightToken converts a Chroma token type to a Lip Gloss style.
func (h *Highlighter) HighlightToken(tt chroma.TokenType) lipgloss.Style {
	entry := h.style.Get(tt)
	style := lipgloss.NewStyle()

	if entry.IsZero() {
		return style
	}

	if entry.Colour.IsSet() {
		style = style.Foreground(lipgloss.Color(entry.Colour.String()))
	}
	if entry.Background.IsSet() {
		style = style.Background(lipgloss.Color(entry.Background.String()))
	}
	if entry.Bold == chroma.Yes {
		style = style.Bold(true)
	}
	if entry.Italic == chroma.Yes {
		style = style.Italic(true)
	}
	if entry.Underline == chroma.Yes {
		style = style.Underline(true)
	}

	return style
}

// Tokenize splits text into Chroma tokens.
func Tokenize(text, language string) ([]chroma.Token, error) {
	if language == "" || language == "plaintext" {
		return []chroma.Token{
			{Type: chroma.Text, Value: text},
		}, nil
	}

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	iterator, err := lexer.Tokenise(nil, text)
	if err != nil {
		return nil, err
	}

	return iterator.Tokens(), nil
}

// HighlightLine tokenizes a single line and returns styled segments.
func (h *Highlighter) HighlightLine(line string, language string) []StyledSegment {
	if language == "" || language == "plaintext" {
		return []StyledSegment{{Text: line, Style: lipgloss.NewStyle()}}
	}

	tokens, err := Tokenize(line, language)
	if err != nil {
		return []StyledSegment{{Text: line, Style: lipgloss.NewStyle()}}
	}

	var segments []StyledSegment
	for _, t := range tokens {
		if t.Value == "" {
			continue
		}
		style := h.HighlightToken(t.Type)
		segments = append(segments, StyledSegment{Text: t.Value, Style: style})
	}
	return segments
}

// StyledSegment is a piece of text with a Lip Gloss style.
type StyledSegment struct {
	Text  string
	Style lipgloss.Style
}

// RenderSegments renders styled segments into a single styled string.
func RenderSegments(segments []StyledSegment) string {
	var result strings.Builder
	for _, seg := range segments {
		result.WriteString(seg.Style.Render(seg.Text))
	}
	return result.String()
}
