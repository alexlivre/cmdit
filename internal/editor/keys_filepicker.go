package editor

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// --- File Picker key handler ---

func (m *Model) enterFilePicker() {
	m.mode = ModeFilePicker
	m.filePicker.Dir = "."
	m.filePicker.Query = ""
	m.loadDirectory()
	m.filePicker.Sel = 0
}

func (m *Model) loadDirectory() {
	entries, err := os.ReadDir(m.filePicker.Dir)
	if err != nil {
		m.filePicker.Files = nil
		return
	}
	m.filePicker.Files = nil
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		if m.filePicker.Query == "" || strings.Contains(strings.ToLower(name), strings.ToLower(m.filePicker.Query)) {
			m.filePicker.Files = append(m.filePicker.Files, name)
		}
	}
}

func (m *Model) handleFilePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		if m.buf.Len() == 0 && m.filename == "" {
			m.mode = ModeWelcome
		}
		return m, nil

	case "enter":
		if m.filePicker.Sel >= 0 && m.filePicker.Sel < len(m.filePicker.Files) {
			name := m.filePicker.Files[m.filePicker.Sel]
			fullPath := filepath.Join(m.filePicker.Dir, strings.TrimSuffix(name, "/"))
			info, err := os.Stat(fullPath)
			if err == nil && info.IsDir() {
				m.filePicker.Dir = fullPath
				m.filePicker.Sel = 0
				m.loadDirectory()
				return m, nil
			}
			m.openFile(fullPath)
		}
		return m, nil

	case "up":
		m.filePicker.Sel--
		if m.filePicker.Sel < 0 {
			m.filePicker.Sel = len(m.filePicker.Files) - 1
		}

	case "down":
		m.filePicker.Sel++
		if m.filePicker.Sel >= len(m.filePicker.Files) {
			m.filePicker.Sel = 0
		}

	case "backspace":
		if len(m.filePicker.Query) > 0 {
			m.filePicker.Query = m.filePicker.Query[:len(m.filePicker.Query)-1]
			m.loadDirectory()
			m.filePicker.Sel = 0
		}

	default:
		if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
			m.filePicker.Query += string(msg.Runes)
			m.loadDirectory()
			m.filePicker.Sel = 0
		}
	}
	return m, nil
}
