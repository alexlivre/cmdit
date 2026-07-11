// Package buffer implements a gap buffer for efficient text editing.
// A gap buffer maintains a "gap" (empty space) at the cursor position,
// so insertions and deletions at the cursor are O(1) without shifting data.
package buffer

import (
	"sort"
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

	lineOffsets []int
	cacheValid  bool
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
	b.cacheValid = false
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
	b.cacheValid = false
	return true
}



// DeleteForward removes the rune to the right of the gap (delete key).
// Returns true if a character was deleted.
func (b *Buffer) DeleteForward() bool {
	if b.gapEnd >= len(b.data) {
		return false
	}
	b.gapEnd++
	b.length--
	b.cacheValid = false
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

// MoveGapTo moves the gap to the specified position using copy() for O(n) performance.
func (b *Buffer) MoveGapTo(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > b.length {
		pos = b.length
	}

	current := b.gapStart
	if pos == current {
		return
	}

	gapSize := b.gapEnd - b.gapStart

	if pos > current {
		n := pos - current
		copy(b.data[b.gapStart:], b.data[b.gapEnd:b.gapEnd+n])
		b.gapStart += n
		b.gapEnd += n
	} else {
		n := current - pos
		copy(b.data[pos+gapSize:], b.data[pos:pos+n])
		b.gapStart = pos
		b.gapEnd = pos + gapSize
	}
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

// Lines returns the text split into lines (cached).
func (b *Buffer) Lines() []string {
	b.buildLineCache()
	runes := []rune(b.String())
	lines := make([]string, len(b.lineOffsets))
	for i := range b.lineOffsets {
		start := b.lineOffsets[i]
		var end int
		if i+1 < len(b.lineOffsets) {
			end = b.lineOffsets[i+1] - 1
		} else {
			end = b.length
		}
		if start > end {
			start = end
		}
		if end > len(runes) {
			end = len(runes)
		}
		if start > len(runes) {
			start = len(runes)
		}
		lines[i] = string(runes[start:end])
	}
	return lines
}

// LineCount returns the number of lines in the buffer (cached).
func (b *Buffer) LineCount() int {
	b.buildLineCache()
	return len(b.lineOffsets)
}

// LineStart returns the logical index of the start of the given line (cached).
func (b *Buffer) LineStart(line int) int {
	b.buildLineCache()
	if line <= 0 {
		return 0
	}
	if line >= len(b.lineOffsets) {
		return b.length
	}
	return b.lineOffsets[line]
}

// LineCol converts a logical index to (line, col) using cache.
func (b *Buffer) LineCol(index int) (int, int) {
	b.buildLineCache()
	if index < 0 {
		return 0, 0
	}
	if index >= b.length {
		index = b.length
	}

	line := sort.SearchInts(b.lineOffsets, index+1) - 1
	if line < 0 {
		line = 0
	}
	col := index - b.lineOffsets[line]
	return line, col
}

func (b *Buffer) buildLineCache() {
	if b.cacheValid {
		return
	}
	b.lineOffsets = []int{0}
	for i := 0; i < b.length; i++ {
		if b.RuneAt(i) == '\n' {
			b.lineOffsets = append(b.lineOffsets, i+1)
		}
	}
	b.cacheValid = true
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
