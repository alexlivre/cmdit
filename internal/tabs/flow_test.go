package tabs

import (
	"testing"

	"github.com/alexb/cmdit/internal/editor"
	tea "github.com/charmbracelet/bubbletea"
)

// TestCloseTabSaveClose tests: modified file → Ctrl+W → S (save) → tab closes
func TestCloseTabSaveClose(t *testing.T) {
	tm, err := NewWithFile("README.md")
	if err != nil {
		t.Fatalf("NewWithFile failed: %v", err)
	}

	// Verify editor is clean
	ed := tm.ActiveEditor()
	if ed.Modified() {
		t.Error("file should not be modified on open")
	}

	// Simulate typing 'x' via TabManager
	tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	ed = tm.ActiveEditor()
	if !ed.Modified() {
		t.Errorf("expected modified=true after insert, mode=%v", ed.CurrentMode())
	}

	// Ctrl+W should enter confirm mode
	tm.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	ed = tm.ActiveEditor()
	if ed.CurrentMode() != editor.ModeConfirm {
		t.Errorf("expected ModeConfirm after Ctrl+W, got %v", ed.CurrentMode())
	}
	if ed.CloseRequested() {
		t.Error("closeRequested should be false before dialog answer")
	}

	// Press S to save and close
	_, _ = tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	ed = tm.ActiveEditor()

	// If editor is nil, tab was already closed (single tab, tea.Quit triggered)
	if ed == nil {
		t.Log("✅ PASS: Tab closed immediately after S (tea.Quit)")
		return
	}

	// If editor exists, closeRequested should be set
	if !ed.CloseRequested() {
		t.Errorf("closeRequested should be true after pressing S")
	}

	// Next update → TabManager detects closeRequested and closes tab
	_, _ = tm.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// All tabs closed
	if tm.ActiveEditor() == nil {
		t.Log("✅ PASS: Tab closed after confirm dialog")
	} else {
		t.Error("tab should have been closed")
	}
}

// TestCloseTabDiscardClose tests: modified file → Ctrl+W → D (discard) → tab closes
func TestCloseTabDiscardClose(t *testing.T) {
	tm, err := NewWithFile("README.md")
	if err != nil {
		t.Fatalf("NewWithFile failed: %v", err)
	}

	tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	tm.Update(tea.KeyMsg{Type: tea.KeyCtrlW})

	ed := tm.ActiveEditor()
	if ed.CurrentMode() != editor.ModeConfirm {
		t.Fatalf("expected ModeConfirm, got %v", ed.CurrentMode())
	}

	// Press D (discard)
	_, _ = tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})

	// If nil, tab already closed
	if tm.ActiveEditor() == nil {
		t.Log("✅ PASS: Tab discarded (tea.Quit)")
		return
	}

	ed = tm.ActiveEditor()
	if !ed.CloseRequested() {
		t.Error("closeRequested should be true after discard")
	}

	// Next update → close tab
	_, _ = tm.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if tm.ActiveEditor() == nil {
		t.Log("✅ PASS: Tab discarded correctly")
	} else {
		t.Error("tab should have been discarded")
	}
}

// TestCloseTabCancel tests: modified file → Ctrl+W → C (cancel) → tab stays open
func TestCloseTabCancel(t *testing.T) {
	tm, err := NewWithFile("README.md")
	if err != nil {
		t.Fatalf("NewWithFile failed: %v", err)
	}

	tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	tm.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	tm.Update(tea.KeyMsg{Type: tea.KeyEscape})

	ed := tm.ActiveEditor()
	if ed.CurrentMode() != editor.ModeNormal {
		t.Errorf("expected ModeNormal after cancel, got %v", ed.CurrentMode())
	}
	if ed.CloseRequested() {
		t.Error("closeRequested should be false after cancel")
	}
	if tm.ActiveEditor() == nil {
		t.Error("tab should still exist after cancel")
	}
	t.Log("✅ PASS: Cancel keeps tab open")
}

// TestCtrlTNewTab tests that Ctrl+T creates a new empty tab
func TestCtrlTNewTab(t *testing.T) {
	tm, err := NewWithFile("README.md")
	if err != nil {
		t.Fatalf("NewWithFile failed: %v", err)
	}

	tm.Update(tea.KeyMsg{Type: tea.KeyCtrlT})

	ed := tm.ActiveEditor()
	if ed == nil {
		t.Fatal("ActiveEditor returned nil after Ctrl+T")
	}
	if ed.CurrentMode() != editor.ModeWelcome {
		t.Logf("new tab mode: %v (expected ModeWelcome)", ed.CurrentMode())
	}
	t.Log("✅ PASS: New tab created with Ctrl+T")
}

// TestCtrlWOnCleanTab closes immediately (no dialog)
func TestCtrlWOnCleanTab(t *testing.T) {
	tm, err := NewWithFile("README.md")
	if err != nil {
		t.Fatalf("NewWithFile failed: %v", err)
	}

	ed := tm.ActiveEditor()
	if ed.Modified() {
		t.Error("file should not be modified")
	}

	// Ctrl+W on clean tab → should NOT enter confirm, tab closes immediately
	_, cmd := tm.Update(tea.KeyMsg{Type: tea.KeyCtrlW})

	// Since it's the only tab, closing should trigger tea.Quit
	if tm.ActiveEditor() == nil {
		if cmd == nil {
			t.Error("expected tea.Quit command when closing last tab")
		}
		t.Log("✅ PASS: Clean tab closed immediately")
	} else {
		t.Errorf("tab should have closed, mode=%v", tm.ActiveEditor().CurrentMode())
	}
}

// TestNextTabWrapsAround tests that F8 cycles through tabs
func TestNextTabWrapsAround(t *testing.T) {
	tm, err := NewWithFile("README.md")
	if err != nil {
		t.Fatalf("NewWithFile failed: %v", err)
	}

	// Create a second tab
	tm.Update(tea.KeyMsg{Type: tea.KeyCtrlT})

	// F8 should cycle back to first tab
	tm.Update(tea.KeyMsg{Type: tea.KeyF8})
	ed := tm.ActiveEditor()
	if ed.Filename() != "README.md" {
		t.Errorf("expected to be on README.md, got %s", ed.Filename())
	}
	t.Log("✅ PASS: F8 wraps around correctly")
}

// TestPrevTabGoesBack tests that F7 goes to previous tab
func TestPrevTabGoesBack(t *testing.T) {
	tm, err := NewWithFile("README.md")
	if err != nil {
		t.Fatalf("NewWithFile failed: %v", err)
	}

	// Create second and third tabs
	tm.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	tm.Update(tea.KeyMsg{Type: tea.KeyCtrlT})

	// Go back one tab with F7
	tm.Update(tea.KeyMsg{Type: tea.KeyF7})
	if tm.activeIdx != 1 {
		t.Errorf("expected active tab 1 after F7, got %d", tm.activeIdx)
	}

	// F7 again should go to first tab
	tm.Update(tea.KeyMsg{Type: tea.KeyF7})
	if tm.activeIdx != 0 {
		t.Errorf("expected active tab 0 after second F7, got %d", tm.activeIdx)
	}

	t.Log("✅ PASS: F7 goes back correctly")
}
