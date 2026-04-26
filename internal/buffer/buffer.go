// Package buffer implements a gap buffer for efficient text editing.
// A gap buffer maintains a "gap" (empty space) at the cursor position,
// so insertions and deletions at the cursor are O(1) without shifting data.
package buffer

import (
	"strings"
	"unicode/utf8"
)

const initialGapSize = 256

// Buffer stores text using a gap buffer data structure.
type Buffer struct {
	data     []rune
	length   int
	gapStart int
	gapEnd   int
}

// NewBuffer creates an empty gap buffer.
func NewBuffer() *Buffer {
	b := &Buffer{
		data:     make([]rune, initialGapSize),
		length:   0,
		gapStart: 0,
		gapEnd:   initialGapSize,
	}
	return b
}

// NewBufferFromString creates a buffer initialized with the given text.
func NewBufferFromString(text string) *Buffer {
	b := NewBuffer()
	for _, r := range text {
		b.Insert(r)
	}
	return b
}

// Insert inserts a rune at the current gap position.
func (b *Buffer) Insert(r rune) {
	if b.gapStart == b.gapEnd {
		b.growGap()
	}
	b.data[b.gapStart] = r
	b.gapStart++
	b.length++
}

// InsertString inserts a string at the current gap position.
func (b *Buffer) InsertString(s string) {
	for _, r := range s {
		b.Insert(r)
	}
}

// Backspace removes the rune immediately left of the cursor (gap start).
// Returns true if a character was deleted.
func (b *Buffer) Backspace() bool {
	if b.gapStart == 0 {
		return false
	}
	b.gapStart--
	b.length--
	return true
}

// Delete removes the rune to the left of the gap (backspace).
// Deprecated: use Backspace instead.
func (b *Buffer) Delete() bool {
	return b.Backspace()
}

// DeleteForward removes the rune to the right of the gap (delete key).
// Returns true if a character was deleted.
func (b *Buffer) DeleteForward() bool {
	if b.gapEnd >= len(b.data) {
		return false
	}
	b.gapEnd++
	b.length--
	return true
}

// MoveGapLeft moves the gap one position to the left.
// This has the effect of moving the cursor left.
// Returns the rune that was moved past (for display purposes), or 0.
func (b *Buffer) MoveGapLeft() rune {
	if b.gapStart == 0 {
		return 0
	}
	b.gapStart--
	b.gapEnd--
	b.data[b.gapEnd] = b.data[b.gapStart]
	return b.data[b.gapStart]
}

// MoveGapRight moves the gap one position to the right.
// This has the effect of moving the cursor right.
func (b *Buffer) MoveGapRight() {
	if b.gapEnd >= len(b.data) {
		return
	}
	b.data[b.gapStart] = b.data[b.gapEnd]
	b.gapStart++
	b.gapEnd++
}

// GapPosition returns the current cursor position (gap start index).
func (b *Buffer) GapPosition() int {
	return b.gapStart
}

// Len returns the logical length of the text (excluding the gap).
func (b *Buffer) Len() int {
	return b.length
}

// RuneAt returns the rune at the given logical index (0-based).
// Returns 0 if index is out of bounds.
func (b *Buffer) RuneAt(index int) rune {
	if index < 0 || index >= b.length {
		return 0
	}
	if index < b.gapStart {
		return b.data[index]
	}
	return b.data[index+(b.gapEnd-b.gapStart)]
}

// String returns the full text content as a string.
func (b *Buffer) String() string {
	var sb strings.Builder
	sb.Grow(b.length)
	for i := 0; i < b.gapStart; i++ {
		sb.WriteRune(b.data[i])
	}
	for i := b.gapEnd; i < len(b.data); i++ {
		sb.WriteRune(b.data[i])
	}
	return sb.String()
}

// Lines returns the text split into lines.
func (b *Buffer) Lines() []string {
	return strings.Split(b.String(), "\n")
}

// LineCount returns the number of lines in the buffer.
func (b *Buffer) LineCount() int {
	count := 1
	for i := 0; i < b.length; i++ {
		if b.RuneAt(i) == '\n' {
			count++
		}
	}
	return count
}

// LineStart returns the logical index of the start of the given line (0-based).
func (b *Buffer) LineStart(line int) int {
	if line <= 0 {
		return 0
	}
	currentLine := 0
	for i := 0; i < b.length; i++ {
		if currentLine == line {
			return i
		}
		if b.RuneAt(i) == '\n' {
			currentLine++
		}
	}
	return b.length
}

// LineCol converts a logical index to (line, col).
func (b *Buffer) LineCol(index int) (int, int) {
	if index < 0 {
		return 0, 0
	}
	if index >= b.length {
		// Return last position
		line := 0
		col := 0
		for i := 0; i < b.length; i++ {
			if b.RuneAt(i) == '\n' {
				line++
				col = 0
			} else {
				col++
			}
		}
		return line, col
	}

	line := 0
	col := 0
	for i := 0; i < index; i++ {
		r := b.RuneAt(i)
		if r == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return line, col
}

// ByteOffset returns the byte offset corresponding to the given logical rune index.
func (b *Buffer) ByteOffset(logicalIndex int) int {
	offset := 0
	for i := 0; i < logicalIndex && i < b.length; i++ {
		r := b.RuneAt(i)
		offset += utf8.RuneLen(r)
	}
	return offset
}

// growGap doubles the buffer size by reallocating.
func (b *Buffer) growGap() {
	newSize := len(b.data) * 2
	newData := make([]rune, newSize)

	// Copy text before gap
	copy(newData, b.data[:b.gapStart])

	// Copy text after gap (placed at the end with new gap in between)
	newGapStart := b.gapStart
	newGapEnd := newSize - (len(b.data) - b.gapEnd)
	copy(newData[newGapEnd:], b.data[b.gapEnd:])

	b.data = newData
	b.gapStart = newGapStart
	b.gapEnd = newGapEnd
}
