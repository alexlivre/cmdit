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
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}

	// Wrap in SplitContainer for split support
	m := tabs.NewSingle(tm)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
