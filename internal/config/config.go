package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// EditorConfig holds editor-specific behavioral preferences.
type EditorConfig struct {
	TabSize     int  `toml:"tab_size"`
	LineNumbers bool `toml:"line_numbers"`
}

// ColorOverrides contains optional hex or ANSI 256 color code overrides for theme tokens.
type ColorOverrides struct {
	H1            string `toml:"h1,omitempty"`
	H2            string `toml:"h2,omitempty"`
	H3            string `toml:"h3,omitempty"`
	H4            string `toml:"h4,omitempty"`
	H5            string `toml:"h5,omitempty"`
	H6            string `toml:"h6,omitempty"`
	Muted         string `toml:"muted,omitempty"`
	Bold          string `toml:"bold,omitempty"`
	Italic        string `toml:"italic,omitempty"`
	CodeBg        string `toml:"code_bg,omitempty"`
	CodeFg        string `toml:"code_fg,omitempty"`
	TableBorder   string `toml:"table_border,omitempty"`
	TableHeader   string `toml:"table_header,omitempty"`
	StatusBarBg   string `toml:"status_bar_bg,omitempty"`
	StatusBarFg   string `toml:"status_bar_fg,omitempty"`
	SearchMatchBg string `toml:"search_match_bg,omitempty"`
	SearchMatchFg string `toml:"search_match_fg,omitempty"`
}

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
