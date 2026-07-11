package editor

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func BenchmarkBufferInsert(b *testing.B) {
	for n := 0; n < b.N; n++ {
		m := New()
		for i := 0; i < 1000; i++ {
			m.buf.Insert('x')
		}
	}
}

func BenchmarkBufferInsertWithCursor(b *testing.B) {
	for n := 0; n < b.N; n++ {
		m := New()
		m.mode = ModeNormal
		for i := 0; i < 1000; i++ {
			m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		}
	}
}

func BenchmarkRenderSmallFile(b *testing.B) {
	m := New()
	m.mode = ModeNormal
	m.viewport.Resize(80, 24)

	// 50 lines of text
	for l := 0; l < 50; l++ {
		line := fmt.Sprintf("Line %d: some text for rendering test", l)
		for _, r := range line {
			m.buf.Insert(r)
		}
		m.buf.Insert('\n')
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = m.View()
	}
}

func BenchmarkRenderLargeFile(b *testing.B) {
	m := New()
	m.mode = ModeNormal
	m.viewport.Resize(80, 24)

	// 500 lines of text
	for l := 0; l < 500; l++ {
		line := fmt.Sprintf("Line %d: some text for rendering benchmark test", l)
		for _, r := range line {
			m.buf.Insert(r)
		}
		m.buf.Insert('\n')
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = m.View()
	}
}

func BenchmarkSearchInLargeFile(b *testing.B) {
	m := New()
	m.mode = ModeSearch

	// 200 lines with searchable content
	for l := 0; l < 200; l++ {
		line := fmt.Sprintf("Line %d: target search word", l)
		for _, r := range line {
			m.buf.Insert(r)
		}
		m.buf.Insert('\n')
	}

	m.search.Query = "target"

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m.doSearch()
	}
}
