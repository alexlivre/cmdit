package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all user-configurable settings persisted to ~/.cmdit/config.json.
type Config struct {
	AutoCloseEnabled bool              `json:"auto_close_enabled"`
	VimMode          bool              `json:"vim_mode"`
	FormatOnSave     bool              `json:"format_on_save"`
	WordWrap         bool              `json:"word_wrap"`
	AutoSaveEnabled  bool              `json:"auto_save_enabled"`
	Theme            string            `json:"theme"`
	Keybindings      map[string]string `json:"keybindings"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		AutoCloseEnabled: true,
		VimMode:          false,
		FormatOnSave:     false,
		WordWrap:         false,
		AutoSaveEnabled:  true,
		Theme:            "dark",
		Keybindings:      map[string]string{},
	}
}

// configPath returns the path to ~/.cmdit/config.json.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cmdit", "config.json"), nil
}

// LoadConfig reads config from ~/.cmdit/config.json.
// Returns defaults if the file does not exist.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	path, err := configPath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	if cfg.Keybindings == nil {
		cfg.Keybindings = map[string]string{}
	}

	return cfg, nil
}

// SaveConfig writes config to ~/.cmdit/config.json.
func SaveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
