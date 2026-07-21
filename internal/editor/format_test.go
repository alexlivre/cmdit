package editor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexb/cmdit/internal/buffer"
)

func TestFormatOnSaveToggle(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	initial := m.config.FormatOnSave

	m.executeAction("file.toggle-format-on-save")

	if m.config.FormatOnSave == initial {
		t.Errorf("FormatOnSave should have toggled, but remained %v", initial)
	}
}

func TestFormatNoFormatter(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.language = "plaintext"
	m.buf.InsertString("hello world")

	result, err := m.formatBuffer()
	if err != nil {
		t.Fatalf("formatBuffer: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected unchanged text, got '%s'", result)
	}
}

func TestFormatGoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")

	code := "package main\nfunc main(){x:=1\nfmt.Println(x)}\n"
	os.WriteFile(path, []byte(code), 0644)

	m, err := NewWithFile(path)
	if err != nil {
		t.Fatalf("NewWithFile: %v", err)
	}
	m.mode = ModeNormal
	m.config.FormatOnSave = true

	m.save()

	data, _ := os.ReadFile(path)
	if string(data) == code {
		t.Error("file should be formatted")
	}
}

func TestFormatResetsUndo(t *testing.T) {
	m := New()
	m.mode = ModeNormal
	m.language = "Go"
	m.buf = buffer.NewBufferFromString("package main\nfunc main(){}\n")
	m.undoStack.Push(buffer.Operation{Type: "insert", Pos: 0, Text: "test"})

	if !m.undoStack.CanUndo() {
		t.Fatal("should have undo before format")
	}

	m.applyFormat()

	if m.undoStack.CanUndo() {
		t.Error("undo stack should be cleared after format")
	}
}
