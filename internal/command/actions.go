// Package command provides the command palette and action registry.
package command

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Action represents a registered command.
type Action struct {
	ID      string
	Label   string
	Shortcut string
}

// Handler is a function that handles an action.
type Handler func() tea.Cmd

// Registry stores all available actions.
type Registry struct {
	actions []Action
}

// NewRegistry creates an empty action registry.
func NewRegistry() *Registry {
	return &Registry{
		actions: make([]Action, 0, 64),
	}
}

// Register adds an action to the registry.
func (r *Registry) Register(a Action) {
	r.actions = append(r.actions, a)
}

// All returns all registered actions.
func (r *Registry) All() []Action {
	return r.actions
}

// Filter returns actions whose Label or ID contain the query (case-insensitive).
func (r *Registry) Filter(query string) []Action {
	if query == "" {
		return r.actions
	}

	var result []Action
	qlower := toLower(query)
	for _, a := range r.actions {
		if contains(toLower(a.Label), qlower) || contains(toLower(a.ID), qlower) {
			result = append(result, a)
		}
	}
	return result
}

func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}
