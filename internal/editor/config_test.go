package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestSaveLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".cmdit", "config.json")

	cfg := Config{
		Theme:            "dracula",
		AutoCloseEnabled: false,
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.MkdirAll(filepath.Dir(configFile), 0700)
	os.WriteFile(configFile, data, 0600)

	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	readData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	var decoded Config
	json.Unmarshal(readData, &decoded)

	if decoded.Theme != "dracula" {
		t.Errorf("expected theme 'dracula', got '%s'", decoded.Theme)
	}
	if decoded.AutoCloseEnabled {
		t.Error("expected AutoCloseEnabled false")
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
