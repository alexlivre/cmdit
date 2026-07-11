package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// showError displays an error message on the status bar for a few seconds.
func (m *Model) showError(msg string) tea.Cmd {
	m.errorMessage = msg
	m.errorTime = time.Now()

	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

type clearErrorMsg struct{}

// logError writes an error to ~/.cmdit/cmdit.log with context.
func logError(err error, context string) {
	home, e := os.UserHomeDir()
	if e != nil {
		return
	}
	logDir := filepath.Join(home, ".cmdit")
	os.MkdirAll(logDir, 0700)

	f, e := os.OpenFile(filepath.Join(logDir, "cmdit.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if e != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] %s: %v\n%s\n", timestamp, context, err, debug.Stack())
}


