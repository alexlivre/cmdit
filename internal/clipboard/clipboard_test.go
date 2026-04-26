package clipboard

import (
	"testing"
)

func TestNewClipboard(t *testing.T) {
	c := New()
	if c.HasText() {
		t.Error("new clipboard should be empty")
	}
}

func TestCopyPaste(t *testing.T) {
	c := New()
	c.Copy("hello")
	if !c.HasText() {
		t.Error("should have text after copy")
	}
	if c.Paste() != "hello" {
		t.Errorf("expected 'hello', got %q", c.Paste())
	}
}

func TestClear(t *testing.T) {
	c := New()
	c.Copy("test")
	c.Clear()
	if c.HasText() {
		t.Error("should be empty after clear")
	}
}

func TestOverwrite(t *testing.T) {
	c := New()
	c.Copy("first")
	c.Copy("second")
	if c.Paste() != "second" {
		t.Errorf("expected 'second', got %q", c.Paste())
	}
}
