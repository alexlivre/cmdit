package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexb/cmdit/internal/tabs"
)

func main() {
	var tm *tabs.TabManager
	var err error

	if len(os.Args) > 1 {
		tm, err = tabs.NewWithFile(os.Args[1])
	} else {
		tm = tabs.New()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir arquivo: %v\n", err)
		os.Exit(1)
	}

	// Test mode: NO WithAltScreen — renders directly to main screen buffer
	m := tabs.NewSingle(tm)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
