package config

import (
	"os"
	"path/filepath"
	"runtime"

	"md-notes/internal/theme"
)

// EditorConfig holds editor-specific behavioral preferences.
type EditorConfig struct {
	TabSize     int  `toml:"tab_size"`
	LineNumbers bool `toml:"line_numbers"`
}

// ColorOverrides contains optional hex or ANSI 256 color code overrides for theme tokens.
type ColorOverrides = theme.ColorOverrides

// Config represents the root configuration file structure.
type Config struct {
	Theme  string         `toml:"theme"`
	Editor EditorConfig   `toml:"editor"`
	Colors ColorOverrides `toml:"colors"`
}

// DefaultConfig returns the default configuration with standard values.
func DefaultConfig() *Config {
	return &Config{
		Theme: "default-dark",
		Editor: EditorConfig{
			TabSize:     4,
			LineNumbers: true,
		},
		Colors: ColorOverrides{},
	}
}

// GetConfigDir returns the standard configuration directory for md-notes following XDG / AppData conventions.
// On Linux and macOS, it resolves to ~/.config/md-notes (or $XDG_CONFIG_HOME/md-notes if set).
// On Windows, it resolves to %APPDATA%\md-notes.
func GetConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		baseDir, err := os.UserConfigDir()
		if err != nil {
			home, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return "", err
			}
			baseDir = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(baseDir, "md-notes"), nil
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "md-notes"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "md-notes"), nil
}

// GetConfigFilePath returns the absolute path to config.toml in the standard configuration directory.
func GetConfigFilePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}
