package editor

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexb/cmdit/internal/lsp"
)

// startLSP starts an LSP server for the current file's language.
func (m *Model) startLSP() {
	if m.filename == "" || m.language == "" {
		return
	}

	var serverCmd string
	var args []string

	switch strings.ToLower(m.language) {
	case "go":
		serverCmd = "gopls"
	case "python":
		serverCmd = "pyright-langserver"
		args = []string{"--stdio"}
	case "rust":
		serverCmd = "rust-analyzer"
	case "typescript", "javascript":
		serverCmd = "typescript-language-server"
		args = []string{"--stdio"}
	default:
		return
	}

	// Check if server is installed
	if _, err := exec.LookPath(serverCmd); err != nil {
		return
	}

	client, err := lsp.NewClient(serverCmd, args...)
	if err != nil {
		return
	}

	m.lspClient = client
	m.lspVersion = 0
	m.diagnostics = make(map[int][]lsp.Diagnostic)

	// Register diagnostics handler
	client.OnDiagnostics(func(uri string, diagnostics []lsp.Diagnostic) {
		m.lspDiagnosticsMu.Lock()
		defer m.lspDiagnosticsMu.Unlock()

		m.diagnostics = make(map[int][]lsp.Diagnostic)
		for _, d := range diagnostics {
			line := d.Range.Start.Line
			m.diagnostics[line] = append(m.diagnostics[line], d)
		}
	})

	// Get the project root
	rootURI := filepath.Dir(m.filename)
	rootURI = "file:///" + filepath.ToSlash(rootURI)

	_, err = client.Initialize(rootURI)
	if err != nil {
		client.Shutdown()
		m.lspClient = nil
		return
	}

	// Send didOpen
	uri := "file:///" + filepath.ToSlash(m.filename)
	client.DidOpen(uri, m.language, m.buf.String())
}

// sendDidChange notifies the LSP server that the document changed.
func (m *Model) sendDidChange() {
	if m.lspClient == nil {
		return
	}

	m.lspVersion++
	uri := "file:///" + filepath.ToSlash(m.filename)
	m.lspClient.DidChange(uri, m.lspVersion, m.buf.String())
}

// stopLSP shuts down the LSP server.
func (m *Model) stopLSP() {
	if m.lspClient != nil {
		m.lspClient.Shutdown()
		m.lspClient = nil
	}
}
