package highlight

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewHighlighter(t *testing.T) {
	h := NewHighlighter(ThemeDark)
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.themeName != ThemeDark {
		t.Errorf("expected theme %s, got %s", ThemeDark, h.themeName)
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"main.go", "Go"},
		{"script.py", "Python"},
		{"app.js", "JavaScript"},
		{"style.css", "CSS"},
		{"README.md", "markdown"},
		{"unknown.xyz", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := DetectLanguage(tt.filename)
		if got != tt.expected {
			t.Errorf("DetectLanguage(%s) = %s, want %s", tt.filename, got, tt.expected)
		}
	}
}

func TestHighlightLine(t *testing.T) {
	h := NewHighlighter(ThemeDark)
	segments := h.HighlightLine("func main() {}", "Go")
	if len(segments) == 0 {
		t.Error("expected non-empty segments for Go code")
	}
}

func TestHighlightLine_Empty(t *testing.T) {
	h := NewHighlighter(ThemeDark)
	segments := h.HighlightLine("", "Go")
	if len(segments) != 0 {
		t.Errorf("expected empty segments for empty line, got %d", len(segments))
	}
}

func TestHighlightLine_Plaintext(t *testing.T) {
	h := NewHighlighter(ThemeDark)
	segments := h.HighlightLine("hello world", "plaintext")
	if len(segments) != 1 {
		t.Errorf("expected 1 segment for plaintext, got %d", len(segments))
	}
	if segments[0].Text != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", segments[0].Text)
	}
}

func TestHighlightLine_UnknownLanguage(t *testing.T) {
	h := NewHighlighter(ThemeDark)
	segments := h.HighlightLine("hello world", "unknown_lang_xyz")
	if len(segments) == 0 {
		t.Error("expected at least one segment for unknown language")
	}
}

func TestSetTheme(t *testing.T) {
	h := NewHighlighter(ThemeDark)
	h.SetTheme(ThemeLight)
	if h.Theme() != ThemeLight {
		t.Errorf("expected theme %s, got %s", ThemeLight, h.Theme())
	}
}

func TestRenderSegments(t *testing.T) {
	segments := []StyledSegment{
		{Text: "hello", Style: lipgloss.NewStyle().Foreground(lipgloss.Color("1"))},
		{Text: " world", Style: lipgloss.NewStyle()},
	}
	result := RenderSegments(segments)
	if result == "" {
		t.Error("expected non-empty rendered string")
	}
}
