// Package command provides the command palette and action registry.
package command

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Action represents a registered command.
type Action struct {
	ID       string
	Label    string
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

	qlower := strings.ToLower(query)
	var result []Action
	for _, a := range r.actions {
		if strings.Contains(strings.ToLower(a.Label), qlower) || strings.Contains(strings.ToLower(a.ID), qlower) {
			result = append(result, a)
		}
	}
	return result
}
