package renderer

import (
	"testing"
)

func TestNewViewport(t *testing.T) {
	v := NewViewport(80, 24)
	if v.Width() != 80 || v.Height() != 24 {
		t.Errorf("expected (80,24), got (%d,%d)", v.Width(), v.Height())
	}
	if v.ScrollY() != 0 || v.ScrollX() != 0 {
		t.Errorf("expected scroll (0,0), got (%d,%d)", v.ScrollY(), v.ScrollX())
	}
}

func TestResize(t *testing.T) {
	v := NewViewport(80, 24)
	v.Resize(100, 40)
	if v.Width() != 100 || v.Height() != 40 {
		t.Errorf("expected (100,40), got (%d,%d)", v.Width(), v.Height())
	}
}

func TestResizeClampsMinimum(t *testing.T) {
	v := NewViewport(80, 24)
	v.Resize(0, 0)
	if v.Width() != 1 || v.Height() != 1 {
		t.Errorf("expected (1,1), got (%d,%d)", v.Width(), v.Height())
	}
}

func TestEnsureVisibleScrollsDown(t *testing.T) {
	v := NewViewport(80, 24) // height=24 means lines 0-23 visible
	// Line 30 is below viewport
	v.EnsureVisible(30, 0)
	if v.ScrollY() != 7 {
		// 30 - 24 + 1 = 7
		t.Errorf("expected scrollY 7, got %d", v.ScrollY())
	}
}

func TestEnsureVisibleScrollsUp(t *testing.T) {
	v := NewViewport(80, 24)
	v.ScrollTo(10, 0)
	v.EnsureVisible(2, 0)
	if v.ScrollY() != 2 {
		t.Errorf("expected scrollY 2, got %d", v.ScrollY())
	}
}

func TestEnsureVisibleAlreadyVisible(t *testing.T) {
	v := NewViewport(80, 24)
	v.ScrollTo(5, 0)
	v.EnsureVisible(10, 0) // line 10 is visible (lines 5-28)
	if v.ScrollY() != 5 {
		t.Errorf("scroll should not change, got %d", v.ScrollY())
	}
}

func TestEnsureVisibleHorizontal(t *testing.T) {
	v := NewViewport(80, 24)
	v.EnsureVisible(0, 90)
	if v.ScrollX() != 11 {
		// 90 - 80 + 1 = 11
		t.Errorf("expected scrollX 11, got %d", v.ScrollX())
	}
}

func TestScrollUpDown(t *testing.T) {
	v := NewViewport(80, 24)
	v.ScrollTo(10, 0)

	v.ScrollUp(3)
	if v.ScrollY() != 7 {
		t.Errorf("expected 7, got %d", v.ScrollY())
	}

	v.ScrollDown(5)
	if v.ScrollY() != 12 {
		t.Errorf("expected 12, got %d", v.ScrollY())
	}
}

func TestScrollUpClampsToZero(t *testing.T) {
	v := NewViewport(80, 24)
	v.ScrollTo(2, 0)
	v.ScrollUp(10)
	if v.ScrollY() != 0 {
		t.Errorf("expected 0, got %d", v.ScrollY())
	}
}
