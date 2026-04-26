package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexb/cmdit/internal/editor"
)

func main() {
	var m *editor.Model
	var err error

	if len(os.Args) > 1 {
		m, err = editor.NewWithFile(os.Args[1])
	} else {
		m = editor.New()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir arquivo: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
