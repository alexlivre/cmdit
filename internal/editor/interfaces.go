package editor

import (
	"github.com/alexb/cmdit/internal/buffer"
	"github.com/alexb/cmdit/internal/command"
	"github.com/alexb/cmdit/internal/highlight"
)

// FileLoader handles file I/O operations.
type FileLoader interface {
	Load(path string) (*buffer.Buffer, error)
	Save(path string, buf *buffer.Buffer) error
	Rename(oldPath, newPath string) error
}

// SyntaxHighlighter provides syntax highlighting.
type SyntaxHighlighter interface {
	HighlightLine(line string, language string) []highlight.StyledSegment
	SetTheme(theme string)
	DetectLanguage(filename string) string
	Theme() string
}

// CommandRegistry provides command palette actions.
type CommandRegistry interface {
	All() []command.Action
	Filter(query string) []command.Action
	Register(action command.Action)
}
