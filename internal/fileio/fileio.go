// Package fileio handles loading and saving buffer contents to files.
package fileio

import (
	"os"

	"github.com/alexb/cmdit/internal/buffer"
)

// Load reads a file and returns a buffer with its contents.
func Load(path string) (*buffer.Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return buffer.NewBufferFromString(string(data)), nil
}

// Save writes the buffer contents to a file.
func Save(path string, buf *buffer.Buffer) error {
	return os.WriteFile(path, []byte(buf.String()), 0644)
}

// Rename renames a file from oldPath to newPath.
// Uses os.Rename which fails on Windows if destination exists,
// but silently overwrites on Unix.
func Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
