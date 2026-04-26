package buffer

// Cursor represents the position within a buffer.
type Cursor struct {
	Line int
	Col  int
}

// NewCursor creates a cursor at position 0,0 (line 0, column 0).
func NewCursor() *Cursor {
	return &Cursor{Line: 0, Col: 0}
}

// Up moves the cursor up one line. Does not go above line 0.
func (c *Cursor) Up() {
	if c.Line > 0 {
		c.Line--
	}
}

// Down moves the cursor down one line (caller must clamp to max line).
func (c *Cursor) Down() {
	c.Line++
}

// Left moves the cursor left one column. Does not go below 0.
func (c *Cursor) Left() {
	if c.Col > 0 {
		c.Col--
	}
}

// Right moves the cursor right one column.
func (c *Cursor) Right() {
	c.Col++
}

// SetPos sets the cursor to a specific line and column.
func (c *Cursor) SetPos(line, col int) {
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}
	c.Line = line
	c.Col = col
}
