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

func newTestBuffer() *buffer.Buffer {
	return newTestBufferWithString("hello\nworld\n")
}

func newTestBufferWithString(s string) *buffer.Buffer {
	return buffer.NewBufferFromString(s)
}
