package editor

import (
	"encoding/json"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.AutoCloseEnabled {
		t.Error("AutoCloseEnabled should be true by default")
	}
	if cfg.VimMode {
		t.Error("VimMode should be false by default")
	}
	if cfg.FormatOnSave {
		t.Error("FormatOnSave should be false by default")
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme should be 'dark' by default, got '%s'", cfg.Theme)
	}
	if cfg.WordWrap {
		t.Error("WordWrap should be false by default")
	}
	if cfg.Keybindings == nil {
		t.Error("Keybindings should not be nil by default")
	}
	if len(cfg.Keybindings) != 0 {
		t.Error("Keybindings should be empty by default")
	}
}

func TestConfigRoundtrip(t *testing.T) {
	cfg := Config{
		AutoCloseEnabled: false,
		VimMode:          true,
		FormatOnSave:     true,
		WordWrap:         true,
		Theme:            "monokai",
		Keybindings: map[string]string{
			"ctrl+s": "file.save",
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.AutoCloseEnabled {
		t.Error("AutoCloseEnabled should be false after roundtrip")
	}
	if !decoded.VimMode {
		t.Error("VimMode should be true after roundtrip")
	}
	if decoded.Keybindings["ctrl+s"] != "file.save" {
		t.Error("Keybindings should survive roundtrip")
	}
}

func TestLoadConfigDefaultsWhenNoFile(t *testing.T) {
	// We can't easily override UserHomeDir, but we can test that
	// LoadConfig() doesn't panic and returns a valid config.
	// The function handles errors gracefully.
	cfg, err := LoadConfig()
	if err != nil {
		// This is OK on systems without home dir
		t.Logf("LoadConfig error (expected on some systems): %v", err)
		cfg = DefaultConfig()
	}
	if cfg.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got '%s'", cfg.Theme)
	}
}

func TestConfigModelIntegration(t *testing.T) {
	m := New()

	if m.config.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got '%s'", m.config.Theme)
	}
	if !m.config.AutoCloseEnabled {
		t.Error("expected AutoCloseEnabled true")
	}
}
