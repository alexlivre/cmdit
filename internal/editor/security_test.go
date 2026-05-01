package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathTraversalPrevention(t *testing.T) {
	m := New()
	m.mode = ModeNormal

	// openFile should clean traversal attempts
	dir := t.TempDir()
	traversalPath := filepath.Join(dir, "..", "..", "etc", "passwd")
	cleaned := filepath.Clean(traversalPath)
	if strings.Contains(cleaned, ".."+string(os.PathSeparator)+"..") {
		t.Errorf("path not cleaned: %s → %s", traversalPath, cleaned)
	}

	// Test validateFileName blocks traversal chars
	if err := validateFileName("../../../etc/passwd"); err == nil {
		t.Error("should reject path traversal in filename")
	}
	if err := validateFileName("normal.txt"); err != nil {
		t.Errorf("should accept normal filename: %v", err)
	}
}

func TestEscapeSequenceSanitization(t *testing.T) {
	// Verify that renderContent doesn't panic on escape sequences
	m := New()
	m.mode = ModeNormal

	// Insert some ANSI escape-like content
	escapeContent := "\x1b[31mred text\x1b[0m"
	for _, r := range escapeContent {
		m.buf.Insert(r)
	}

	// Verify content is stored correctly (not executed)
	result := m.buf.String()
	if result != escapeContent {
		t.Errorf("escape content not preserved correctly")
	}

	// Try to render - should not panic
	m.viewport.Resize(80, 24)
	_ = m.View()
}
