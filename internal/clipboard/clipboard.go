// Package clipboard provides clipboard operations for the editor.
// Currently uses an internal clipboard; system clipboard via OSC52 is planned for Phase 7.
package clipboard

// Clipboard stores cut/copied text.
type Clipboard struct {
	text string
}

// New creates a new clipboard.
func New() *Clipboard {
	return &Clipboard{}
}

// Copy stores text in the clipboard.
func (c *Clipboard) Copy(text string) {
	c.text = text
}

// Paste returns the clipboard contents.
func (c *Clipboard) Paste() string {
	return c.text
}

// HasText returns true if the clipboard has content.
func (c *Clipboard) HasText() bool {
	return len(c.text) > 0
}

// Clear empties the clipboard.
func (c *Clipboard) Clear() {
	c.text = ""
}
