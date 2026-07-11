package editor

import (
	"strings"
	"testing"

	"github.com/alexb/cmdit/internal/buffer"
)

func TestGetSelectedText_Performance(t *testing.T) {
	m := New()
	largeText := strings.Repeat("a", 10000)
	m.buf = buffer.NewBufferFromString(largeText)
	m.selection.Start = 0
	m.selection.End = 10000

	result := m.getSelectedText()
	if len(result) != 10000 {
		t.Errorf("expected 10000 chars, got %d", len(result))
	}
}

func TestGetSelectedText_Empty(t *testing.T) {
	m := New()
	m.selection.Start = -1
	m.selection.End = -1
	if m.getSelectedText() != "" {
		t.Error("expected empty string for no selection")
	}
}

func TestGetSelectedText_Partial(t *testing.T) {
	m := New()
	m.buf = buffer.NewBufferFromString("hello world")
	m.selection.Start = 6
	m.selection.End = 11
	if got := m.getSelectedText(); got != "world" {
		t.Errorf("expected 'world', got '%s'", got)
	}
}
