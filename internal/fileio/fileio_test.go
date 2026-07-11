package fileio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexb/cmdit/internal/buffer"
)

func TestSaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	buf := newTestBuffer()
	if err := Save(path, buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.String() != buf.String() {
		t.Errorf("roundtrip mismatch:\n  expected: %q\n  got:      %q", buf.String(), loaded.String())
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestSaveAndLoadUTF8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "utf8.txt")

	content := "olá, mundo! こんにちは"
	buf := newTestBufferWithString(content)
	if err := Save(path, buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.String() != content {
		t.Errorf("UTF-8 roundtrip mismatch:\n  expected: %q\n  got:      %q", content, loaded.String())
	}
}

func TestSaveFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perms.txt")

	buf := newTestBuffer()
	if err := Save(path, buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// On Windows, file permissions work differently; check file exists instead.
	perm := info.Mode().Perm()
	if perm != 0644 && perm != 0666 {
		t.Errorf("unexpected permissions: got %o", perm)
	}
}

func TestRenameSuccess(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")
	newPath := filepath.Join(dir, "new.txt")

	content := "hello\nworld\n"
	if err := os.WriteFile(oldPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := Rename(oldPath, newPath); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Old file should not exist
	if _, err := os.Stat(oldPath); err == nil {
		t.Error("old file still exists after rename")
	}

	// New file should exist with same content
	data, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("new file not found: %v", err)
	}
	if string(data) != content {
		t.Errorf("content mismatch:\n  expected: %q\n  got:      %q", content, string(data))
	}
}

func TestRenameSourceNotFound(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "nonexistent.txt")
	newPath := filepath.Join(dir, "new.txt")

	err := Rename(oldPath, newPath)
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestRenameEmptyName(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")

	// Create file first
	if err := os.WriteFile(oldPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := Rename(oldPath, "")
	if err == nil {
		t.Error("expected error for empty destination name")
	}
}

func TestRenameDestinationExists(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")
	existingPath := filepath.Join(dir, "existing.txt")

	if err := os.WriteFile(oldPath, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create old file: %v", err)
	}
	if err := os.WriteFile(existingPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	err := Rename(oldPath, existingPath)
	// On Windows: error (file exists)
	// On Unix: success (silent overwrite)
	if err != nil {
		// Windows behavior — acceptable
		t.Logf("Rename to existing file returned error (expected on Windows): %v", err)
	} else {
		// Unix behavior — verify overwrite
		data, err := os.ReadFile(existingPath)
		if err != nil {
			t.Fatalf("existing file not found after overwrite: %v", err)
		}
		if string(data) != "old content" {
			t.Errorf("content mismatch after overwrite:\n  expected: %q\n  got:      %q", "old content", string(data))
		}
	}
}

func newTestBuffer() *buffer.Buffer {
	return newTestBufferWithString("hello\nworld\n")
}

func newTestBufferWithString(s string) *buffer.Buffer {
	return buffer.NewBufferFromString(s)
}

func TestLoad_TooLarge(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "large.txt")
	content := make([]byte, maxFileSize+1)
	for i := range content {
		content[i] = 'x'
	}
	os.WriteFile(tmpFile, content, 0644)

	_, err := Load(tmpFile)
	if err == nil {
		t.Error("expected error for large file")
	}
}

func TestLoad_WithinLimit(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "normal.txt")
	os.WriteFile(tmpFile, []byte("hello world"), 0644)

	buf, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "hello world" {
		t.Errorf("unexpected content: %s", buf.String())
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.txt")
	os.WriteFile(tmpFile, []byte{}, 0644)

	buf, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty buffer, got length %d", buf.Len())
	}
}

func TestSave_EmptyBuffer(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.txt")
	b := buffer.NewBuffer()
	err := Save(tmpFile, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tmpFile)
	if len(data) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(data))
	}
}
