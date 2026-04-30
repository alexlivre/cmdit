package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SplitDirection indicates the layout orientation.
type SplitDirection int

const (
	SplitHorizontal SplitDirection = iota
	SplitVertical
)

// SplitContainer holds two TabManagers in a split layout.
// When right is nil, it acts as a single pane.
type SplitContainer struct {
	left       *TabManager
	right      *TabManager // nil means single pane
	activeSide int         // 0 = left/top, 1 = right/bottom
	direction  SplitDirection
	ratio      float64 // 0.0 to 1.0, portion for left/top
	width      int
	height     int

	borderStyle    lipgloss.Style
	activeBorder   lipgloss.Style
	inactiveBorder lipgloss.Style
}

// NewSplit creates a SplitContainer with two TabManagers.
func NewSplit(left, right *TabManager, direction SplitDirection) *SplitContainer {
	return &SplitContainer{
		left:           left,
		right:          right,
		activeSide:     0,
		direction:      direction,
		ratio:          0.5,
		borderStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		activeBorder:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		inactiveBorder: lipgloss.NewStyle().Foreground(lipgloss.Color("237")),
	}
}

// NewSingle creates a SplitContainer with a single TabManager (no split).
func NewSingle(tm *TabManager) *SplitContainer {
	return &SplitContainer{
		left:           tm,
		right:          nil,
		activeSide:     0,
		direction:      SplitHorizontal,
		ratio:          0.5,
		borderStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		activeBorder:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		inactiveBorder: lipgloss.NewStyle().Foreground(lipgloss.Color("237")),
	}
}

// HasSplit returns true if there is a right pane.
func (s *SplitContainer) HasSplit() bool {
	return s.right != nil
}

// Split creates a right pane from the current left pane and a new TabManager.
func (s *SplitContainer) Split(direction SplitDirection) {
	if s.right != nil {
		return // already split
	}
	s.direction = direction
	s.right = New()
	s.ratio = 0.5
}

// Init implements tea.Model.
func (s *SplitContainer) Init() tea.Cmd {
	if s.right == nil {
		return s.left.Init()
	}
	return tea.Batch(s.left.Init(), s.right.Init())
}

// Update implements tea.Model.
func (s *SplitContainer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return s.handleResize(msg)

	case tea.MouseMsg:
		return s.handleMouse(msg)

	case tea.KeyMsg:
		return s.handleKey(msg)
	}

	// Single pane mode
	if s.right == nil {
		newLeft, cmd := s.left.Update(msg)
		if lm, ok := newLeft.(*TabManager); ok {
			s.left = lm
		}
		return s, cmd
	}

	// Delegate to active side
	var cmd tea.Cmd
	if s.activeSide == 0 {
		newLeft, c := s.left.Update(msg)
		if lm, ok := newLeft.(*TabManager); ok {
			s.left = lm
		}
		cmd = c
	} else {
		newRight, c := s.right.Update(msg)
		if rm, ok := newRight.(*TabManager); ok {
			s.right = rm
		}
		cmd = c
	}

	return s, cmd
}

func (s *SplitContainer) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	s.width = msg.Width
	s.height = msg.Height

	if s.right == nil {
		newLeft, cmd := s.left.Update(tea.WindowSizeMsg{Width: msg.Width, Height: msg.Height})
		if lm, ok := newLeft.(*TabManager); ok {
			s.left = lm
		}
		return s, cmd
	}

	leftW, leftH, rightW, rightH := s.calculateSizes()

	newLeft, cmdL := s.left.Update(tea.WindowSizeMsg{Width: leftW, Height: leftH})
	if lm, ok := newLeft.(*TabManager); ok {
		s.left = lm
	}

	newRight, cmdR := s.right.Update(tea.WindowSizeMsg{Width: rightW, Height: rightH})
	if rm, ok := newRight.(*TabManager); ok {
		s.right = rm
	}

	return s, tea.Batch(cmdL, cmdR)
}

// View implements tea.Model.
func (s *SplitContainer) View() string {
	if s.right == nil {
		return s.left.View()
	}

	if s.width == 0 || s.height == 0 {
		return ""
	}

	leftW, leftH, rightW, rightH := s.calculateSizes()

	leftView := s.left.View()
	rightView := s.right.View()

	// Constrain views to their sizes
	leftView = constrainView(leftView, leftW, leftH)
	rightView = constrainView(rightView, rightW, rightH)

	// Build split with border
	if s.direction == SplitHorizontal {
		return s.renderHorizontal(leftView, rightView, leftW, leftH, rightW, rightH)
	}
	return s.renderVertical(leftView, rightView, leftW, leftH, rightW, rightH)
}

// renderHorizontal renders left-right split with a vertical border.
func (s *SplitContainer) renderHorizontal(left, right string, leftW, leftH, rightW, rightH int) string {
	leftLines := splitLines(left, leftH)
	rightLines := splitLines(right, rightH)

	borderChar := "│"
	if s.activeSide == 0 {
		borderChar = s.activeBorder.Render("│")
	} else {
		borderChar = s.inactiveBorder.Render("│")
	}

	var result []string
	maxLines := leftH
	if rightH > maxLines {
		maxLines = rightH
	}

	for i := 0; i < maxLines; i++ {
		leftLine := ""
		rightLine := ""

		if i < len(leftLines) {
			leftLine = padRight(leftLines[i], leftW)
		} else {
			leftLine = blankLine(leftW)
		}

		if i < len(rightLines) {
			rightLine = padRight(rightLines[i], rightW)
		} else {
			rightLine = blankLine(rightW)
		}

		result = append(result, leftLine+borderChar+rightLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, result...)
}

// renderVertical renders top-bottom split with a horizontal border.
func (s *SplitContainer) renderVertical(top, bottom string, topW, topH, bottomW, bottomH int) string {
	topLines := splitLines(top, topH)
	bottomLines := splitLines(bottom, bottomH)

	borderLine := ""
	if s.activeSide == 0 {
		borderLine = s.activeBorder.Render(string(repeatChar('─', topW)))
	} else {
		borderLine = s.inactiveBorder.Render(string(repeatChar('─', topW)))
	}

	var result []string
	result = append(result, topLines...)
	result = append(result, borderLine)
	result = append(result, bottomLines...)

	return lipgloss.JoinVertical(lipgloss.Left, result...)
}

// calculateSizes determines the dimensions for left and right sides.
func (s *SplitContainer) calculateSizes() (int, int, int, int) {
	if s.direction == SplitHorizontal {
		leftW := int(float64(s.width-1) * s.ratio)
		if leftW < 5 {
			leftW = 5
		}
		rightW := s.width - 1 - leftW
		if rightW < 5 {
			rightW = 5
			leftW = s.width - 1 - rightW
		}
		return leftW, s.height, rightW, s.height
	}

	// Vertical
	topH := int(float64(s.height-1) * s.ratio)
	if topH < 3 {
		topH = 3
	}
	bottomH := s.height - 1 - topH
	if bottomH < 3 {
		bottomH = 3
		topH = s.height - 1 - bottomH
	}
	return s.width, topH, s.width, bottomH
}

// handleKey processes split management keys.
func (s *SplitContainer) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Split management
	switch key {
	case "ctrl+\\":
		if s.right == nil {
			// Create split
			s.Split(SplitHorizontal)
			return s, tea.Batch(s.left.Init(), s.right.Init())
		}
		// Toggle focus between splits
		s.activeSide = 1 - s.activeSide
		return s, nil
	}

	// Single pane mode: delegate to left
	if s.right == nil {
		newLeft, cmd := s.left.Update(msg)
		if lm, ok := newLeft.(*TabManager); ok {
			s.left = lm
		}
		return s, cmd
	}

	// Delegate to active side
	var cmd tea.Cmd
	if s.activeSide == 0 {
		newLeft, c := s.left.Update(msg)
		if lm, ok := newLeft.(*TabManager); ok {
			s.left = lm
		}
		cmd = c
	} else {
		newRight, c := s.right.Update(msg)
		if rm, ok := newRight.(*TabManager); ok {
			s.right = rm
		}
		cmd = c
	}

	return s, cmd
}

// handleMouse processes mouse events for clicking between splits.
func (s *SplitContainer) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if s.right != nil {
		leftW, _, _, _ := s.calculateSizes()

		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if msg.X <= leftW {
				s.activeSide = 0
			} else {
				s.activeSide = 1
			}
		}
	}

	// Single pane or delegate to active side
	if s.right == nil || s.activeSide == 0 {
		newLeft, cmd := s.left.Update(msg)
		if lm, ok := newLeft.(*TabManager); ok {
			s.left = lm
		}
		return s, cmd
	}
	newRight, cmd := s.right.Update(msg)
	if rm, ok := newRight.(*TabManager); ok {
		s.right = rm
	}
	return s, cmd
}

// --- Helpers ---

func splitLines(s string, maxLines int) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for len(lines) < maxLines {
		lines = append(lines, "")
	}
	return lines
}

func padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return s + string(repeatChar(' ', width-len(runes)))
}

func blankLine(width int) string {
	return string(repeatChar(' ', width))
}

func repeatChar(c rune, n int) []rune {
	result := make([]rune, n)
	for i := range result {
		result[i] = c
	}
	return result
}

func constrainView(view string, width, height int) string {
	lines := strings.Split(view, "\n")

	// Constrain width
	var constrained []string
	for _, line := range lines {
		runes := []rune(line)
		if len(runes) > width {
			constrained = append(constrained, string(runes[:width]))
		} else if len(runes) < width {
			constrained = append(constrained, line+string(repeatChar(' ', width-len(runes))))
		} else {
			constrained = append(constrained, line)
		}
	}

	// Constrain height
	if len(constrained) > height {
		constrained = constrained[:height]
	}
	for len(constrained) < height {
		constrained = append(constrained, string(repeatChar(' ', width)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, constrained...)
}
