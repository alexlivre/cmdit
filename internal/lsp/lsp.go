// Package lsp implements a Language Server Protocol client for cmdit.
package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// --- JSON-RPC 2.0 Types ---

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents a JSON-RPC error.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Notification is a JSON-RPC 2.0 notification (no id).
type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// --- LSP Protocol Types ---

// Position represents a position in a text document.
type Position struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based UTF-16 offset
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a file.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic represents a diagnostic message (error, warning, etc.).
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1=Error, 2=Warning, 3=Information, 4=Hint
	Message  string `json:"message"`
	Source   string `json:"source,omitempty"`
}

// PublishDiagnosticsParams is sent from server to client.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// --- Initialize Params ---

// InitializeParams is sent for the initialize request.
type InitializeParams struct {
	ProcessID    int                `json:"processId"`
	RootURI      string             `json:"rootUri,omitempty"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

// ClientCapabilities describes what the client supports.
type ClientCapabilities struct {
	TextDocument TextDocumentCapabilities `json:"textDocument,omitempty"`
}

// TextDocumentCapabilities describes text document features.
type TextDocumentCapabilities struct {
	Completion CompletionCapabilities `json:"completion,omitempty"`
}

// CompletionCapabilities describes completion support.
type CompletionCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration"`
}

// --- Initialize Result ---

// InitializeResult is the response to initialize.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ServerCapabilities describes what the server supports.
type ServerCapabilities struct {
	TextDocumentSync   int                `json:"textDocumentSync,omitempty"` // 0=None, 1=Full, 2=Incremental
	CompletionProvider *CompletionOptions `json:"completionProvider,omitempty"`
	DefinitionProvider bool               `json:"definitionProvider,omitempty"`
}

// CompletionOptions describes completion capabilities.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// --- DidOpen Text Document ---

// TextDocumentItem is sent for textDocument/didOpen.
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpenTextDocumentParams is the params for textDocument/didOpen.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// --- DidChange Text Document ---

// TextDocumentContentChangeEvent describes a change.
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// DidChangeTextDocumentParams is sent for textDocument/didChange.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// VersionedTextDocumentIdentifier identifies a document with a version.
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// --- DidClose Text Document ---

// DidCloseTextDocumentParams is sent for textDocument/didClose.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// TextDocumentIdentifier identifies a text document.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// --- Completion ---

// CompletionParams is sent for textDocument/completion.
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label      string `json:"label"`
	Kind       int    `json:"kind,omitempty"`
	Detail     string `json:"detail,omitempty"`
	InsertText string `json:"insertText,omitempty"`
}

// CompletionList is a list of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// --- Definition ---

// DefinitionParams is sent for textDocument/definition.
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// --- Client ---

// Client manages communication with an LSP server.
type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader

	requestID int
	mu        sync.Mutex

	// Message handlers
	onDiagnostics func(uri string, diagnostics []Diagnostic)
	onCompletion  func(id int, items []CompletionItem)

	// Pending responses
	pending map[int]chan Response
}

// NewClient creates a new LSP client and starts the server.
// serverCmd is the command to run (e.g., "gopls").
func NewClient(serverCmd string, args ...string) (*Client, error) {
	c := &Client{
		requestID: 1,
		pending:   make(map[int]chan Response),
	}

	cmd := exec.Command(serverCmd, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("lsp stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("lsp stdout: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin
	c.stdout = stdout
	c.reader = bufio.NewReader(stdout)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("lsp start: %w", err)
	}

	// Start reading responses in background
	go c.readLoop()

	return c, nil
}

// Initialize sends the initialize request and waits for the response.
func (c *Client) Initialize(rootURI string) (*ServerCapabilities, error) {
	params := InitializeParams{
		ProcessID: 0,
		RootURI:   rootURI,
		Capabilities: ClientCapabilities{
			TextDocument: TextDocumentCapabilities{
				Completion: CompletionCapabilities{
					DynamicRegistration: false,
				},
			},
		},
	}

	resp, err := c.request("initialize", params)
	if err != nil {
		return nil, err
	}

	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse initialize result: %w", err)
	}

	// Send initialized notification
	c.notify("initialized", struct{}{})

	return &result.Capabilities, nil
}

// DidOpen notifies the server that a file was opened.
func (c *Client) DidOpen(uri, languageID, text string) error {
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: languageID,
			Version:    0,
			Text:       text,
		},
	}
	return c.notify("textDocument/didOpen", params)
}

// DidChange notifies the server that the document changed.
func (c *Client) DidChange(uri string, version int, text string) error {
	params := DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			URI:     uri,
			Version: version,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: text},
		},
	}
	return c.notify("textDocument/didChange", params)
}

// DidClose notifies the server that the file was closed.
func (c *Client) DidClose(uri string) error {
	params := DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	}
	return c.notify("textDocument/didClose", params)
}

// OnDiagnostics registers a handler for diagnostics notifications.
func (c *Client) OnDiagnostics(handler func(uri string, diagnostics []Diagnostic)) {
	c.onDiagnostics = handler
}

// OnCompletion registers a handler for completion responses.
func (c *Client) OnCompletion(handler func(id int, items []CompletionItem)) {
	c.onCompletion = handler
}

// Shutdown sends the shutdown request and terminates the server.
func (c *Client) Shutdown() error {
	c.request("shutdown", nil)
	c.notify("exit", nil)
	return c.cmd.Wait()
}

// --- Internal ---

func (c *Client) request(method string, params interface{}) (Response, error) {
	c.mu.Lock()
	id := c.requestID
	c.requestID++
	ch := make(chan Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.writeMessage(req); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return Response{}, err
	}

	// Wait for response
	resp := <-ch
	if resp.Error != nil {
		return resp, fmt.Errorf("lsp error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	return resp, nil
}

func (c *Client) notify(method string, params interface{}) error {
	notif := Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	return c.writeMessage(notif)
}

func (c *Client) writeMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := c.stdin.Write(data); err != nil {
		return err
	}
	return nil
}

// readLoop continuously reads messages from the server.
func (c *Client) readLoop() {
	for {
		msg, err := c.readMessage()
		if err != nil {
			// Server likely terminated
			return
		}

		// Try parsing as a notification first
		var notif Notification
		if err := json.Unmarshal(msg, &notif); err == nil && notif.Method != "" {
			c.handleNotification(notif)
			continue
		}

		// Try parsing as a response
		var resp Response
		if err := json.Unmarshal(msg, &resp); err == nil {
			c.handleResponse(resp)
			continue
		}
	}
}

func (c *Client) readMessage() ([]byte, error) {
	// Read headers
	var contentLength int
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // end of headers
		}
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(val)
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("no Content-Length header")
	}

	// Read body
	body := make([]byte, contentLength)
	_, err := io.ReadFull(c.reader, body)
	return body, err
}

func (c *Client) handleResponse(resp Response) {
	c.mu.Lock()
	ch, ok := c.pending[resp.ID]
	delete(c.pending, resp.ID)
	c.mu.Unlock()

	if ok {
		ch <- resp
	}
}

func (c *Client) handleNotification(notif Notification) {
	switch notif.Method {
	case "textDocument/publishDiagnostics":
		data, _ := json.Marshal(notif.Params)
		var params PublishDiagnosticsParams
		if err := json.Unmarshal(data, &params); err == nil {
			if c.onDiagnostics != nil {
				c.onDiagnostics(params.URI, params.Diagnostics)
			}
		}
	}
}
