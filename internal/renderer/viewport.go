// Package renderer handles viewport scrolling and text rendering.
package renderer

// Viewport tracks the visible region of the text buffer.
type Viewport struct {
	scrollY int // First visible line
	scrollX int // First visible column
	width   int // Terminal width
	height  int // Terminal height (content area)
}

// NewViewport creates a viewport with the given dimensions.
func NewViewport(width, height int) *Viewport {
	return &Viewport{
		scrollY: 0,
		scrollX: 0,
		width:   width,
		height:  height,
	}
}

// ScrollY returns the current vertical scroll position.
func (v *Viewport) ScrollY() int {
	return v.scrollY
}

// ScrollX returns the current horizontal scroll position.
func (v *Viewport) ScrollX() int {
	return v.scrollX
}

// Width returns the viewport width.
func (v *Viewport) Width() int {
	return v.width
}

// Height returns the viewport height.
func (v *Viewport) Height() int {
	return v.height
}

// Resize updates the viewport dimensions.
func (v *Viewport) Resize(width, height int) {
	v.width = width
	if v.width < 1 {
		v.width = 1
	}
	v.height = height
	if v.height < 1 {
		v.height = 1
	}
}

// EnsureVisible adjusts scroll so that (line, col) is visible in the viewport.
func (v *Viewport) EnsureVisible(line, col int) {
	if line < v.scrollY {
		v.scrollY = line
	}
	if line >= v.scrollY+v.height {
		v.scrollY = line - v.height + 1
	}
	if col < v.scrollX {
		v.scrollX = col
	}
	if col >= v.scrollX+v.width {
		v.scrollX = col - v.width + 1
	}
	if v.scrollY < 0 {
		v.scrollY = 0
	}
	if v.scrollX < 0 {
		v.scrollX = 0
	}
}

// ScrollUp scrolls up by n lines.
func (v *Viewport) ScrollUp(n int) {
	v.scrollY -= n
	if v.scrollY < 0 {
		v.scrollY = 0
	}
}

// ScrollDown scrolls down by n lines.
func (v *Viewport) ScrollDown(n int) {
	v.scrollY += n
}

// ScrollTo sets the scroll position.
func (v *Viewport) ScrollTo(line, col int) {
	v.scrollY = line
	v.scrollX = col
	if v.scrollY < 0 {
		v.scrollY = 0
	}
	if v.scrollX < 0 {
		v.scrollX = 0
	}
}
