package editor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/alexb/cmdit/internal/buffer"
)

// formatters maps language names (lowercase) to their formatting command.
var formatters = map[string]struct {
	cmd  string
	args []string
}{
	"go":     {cmd: "gofmt"},
	"python": {cmd: "black", args: []string{"--quiet", "-"}},
	"rust":   {cmd: "rustfmt"},
}

var formatterWarned = make(map[string]bool)

// formatBuffer runs the appropriate formatter on the buffer content.
// Returns the formatted text, or the original text if no formatter
// is available or the formatter is not installed.
func (m *Model) formatBuffer() (string, error) {
	text := m.buf.String()
	lang := strings.ToLower(m.language)

	fmtr, ok := formatters[lang]
	if !ok || fmtr.cmd == "" {
		return text, nil
	}

	if _, err := exec.LookPath(fmtr.cmd); err != nil {
		if !formatterWarned[lang] {
			formatterWarned[lang] = true
			logError(fmt.Errorf("formatter not found: %s", fmtr.cmd), "format-on-save")
		}
		return text, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, fmtr.cmd, fmtr.args...)
	cmd.Stdin = strings.NewReader(text)
	output, err := cmd.Output()
	if err != nil {
		return text, nil
	}

	return string(output), nil
}

// applyFormat formats the buffer and replaces its content.
// The cursor position is approximately restored after formatting.
func (m *Model) applyFormat() error {
	formatted, err := m.formatBuffer()
	if err != nil || formatted == m.buf.String() {
		return nil // nothing to do
	}

	cursorPos := m.buf.GapPosition()
	m.buf = buffer.NewBufferFromString(formatted)

	// Restore cursor approximately
	if cursorPos < m.buf.Len() {
		m.moveGapTo(cursorPos)
	} else {
		m.moveGapTo(m.buf.Len())
	}
	return nil
}
