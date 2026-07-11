package editor

import (
	"github.com/alexb/cmdit/internal/command"
)

// SearchState holds search/replace state.
type SearchState struct {
	Query   string
	Matches []int
	Current int
	Replace string
	Last    string
}

// PaletteState holds command palette state.
type PaletteState struct {
	Actions []command.Action
	Query   string
	Results []command.Action
	Sel     int
}

// FilePickerState holds file picker state.
type FilePickerState struct {
	Dir   string
	Files []string
	Sel   int
	Query string
}

// SelectionState holds selection state.
type SelectionState struct {
	Start int
	End   int
}

// RenameState holds rename dialog state.
type RenameState struct {
	Input string
	Error string
}
