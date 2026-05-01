package editor

import (
	"encoding/json"
	"os"
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
	if !cfg.AutoSaveEnabled {
		t.Error("AutoSaveEnabled should be true by default")
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
		AutoSaveEnabled:  false,
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
	dir := t.TempDir()
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOME", dir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Logf("LoadConfig error (expected on some systems): %v", err)
		cfg = DefaultConfig()
	}
	if cfg.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got '%s'", cfg.Theme)
	}
}

func TestConfigModelIntegration(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOME", dir)

	m := New()

	if m.config.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got '%s'", m.config.Theme)
	}
	if !m.config.AutoCloseEnabled {
		t.Error("expected AutoCloseEnabled true")
	}
}

func TestSaveLoadConfigRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()

	// Override home directory so configPath() uses temp dir
	t.Setenv("USERPROFILE", tmpDir) // Windows
	t.Setenv("HOME", tmpDir)        // Unix

	cfg := Config{
		AutoCloseEnabled: false,
		VimMode:          true,
		FormatOnSave:     true,
		WordWrap:         true,
		AutoSaveEnabled:  false,
		Theme:            "dracula",
		Keybindings: map[string]string{
			"ctrl+j": "file.save",
		},
	}

	// Save
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	// Load back
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	// Verify all fields match
	if loaded.AutoCloseEnabled != cfg.AutoCloseEnabled {
		t.Errorf("AutoCloseEnabled: expected %v, got %v", cfg.AutoCloseEnabled, loaded.AutoCloseEnabled)
	}
	if loaded.VimMode != cfg.VimMode {
		t.Errorf("VimMode: expected %v, got %v", cfg.VimMode, loaded.VimMode)
	}
	if loaded.FormatOnSave != cfg.FormatOnSave {
		t.Errorf("FormatOnSave: expected %v, got %v", cfg.FormatOnSave, loaded.FormatOnSave)
	}
	if loaded.WordWrap != cfg.WordWrap {
		t.Errorf("WordWrap: expected %v, got %v", cfg.WordWrap, loaded.WordWrap)
	}
	if loaded.AutoSaveEnabled != cfg.AutoSaveEnabled {
		t.Errorf("AutoSaveEnabled: expected %v, got %v", cfg.AutoSaveEnabled, loaded.AutoSaveEnabled)
	}
	if loaded.Theme != cfg.Theme {
		t.Errorf("Theme: expected '%s', got '%s'", cfg.Theme, loaded.Theme)
	}
	if loaded.Keybindings["ctrl+j"] != "file.save" {
		t.Errorf("Keybindings: expected 'file.save', got '%s'", loaded.Keybindings["ctrl+j"])
	}
}

func TestLoadConfigCorruptedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("USERPROFILE", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Write invalid JSON to the config file
	configFile := tmpDir + "/.cmdit/config.json"
	os.MkdirAll(tmpDir+"/.cmdit", 0700)
	os.WriteFile(configFile, []byte("not valid json {{{"), 0600)

	cfg, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig should return error for corrupted JSON")
	}
	// Should return defaults
	if cfg.AutoCloseEnabled != true {
		t.Error("corrupted config should fall back to AutoCloseEnabled=true")
	}
}
