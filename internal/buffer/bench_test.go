package buffer

import "testing"

func BenchmarkLines_Small(b *testing.B) {
	buf := NewBufferFromString("line1\nline2\nline3\nline4\nline5")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Lines()
	}
}

func BenchmarkLines_Large(b *testing.B) {
	text := ""
	for i := 0; i < 10000; i++ {
		text += "line content here\n"
	}
	buf := NewBufferFromString(text)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Lines()
	}
}

func BenchmarkLineCount(b *testing.B) {
	text := ""
	for i := 0; i < 10000; i++ {
		text += "line\n"
	}
	buf := NewBufferFromString(text)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.LineCount()
	}
}

func BenchmarkMoveGapTo(b *testing.B) {
	buf := NewBufferFromString("hello world this is a test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.MoveGapTo(0)
		buf.MoveGapTo(26)
	}
}

func BenchmarkInsert(b *testing.B) {
	buf := NewBuffer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Insert('x')
	}
}
