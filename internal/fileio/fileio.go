// Package fileio handles loading and saving buffer contents to files.
package fileio

import (
	"fmt"
	"os"

	"github.com/alexb/cmdit/internal/buffer"
)

const maxFileSize = 100 * 1024 * 1024 // 100MB

// Load reads a file and returns a buffer with its contents.
func Load(path string) (*buffer.Buffer, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d bytes)", info.Size(), maxFileSize)
	}

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
