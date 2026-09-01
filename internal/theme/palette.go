package theme

import (
	"strings"
)

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

// Palette defines the raw color definitions and metadata for a theme.
type Palette struct {
	Name          string
	IsDark        bool
	H1            string
	H2            string
	H3            string
	H4            string
	H5            string
	H6            string
	Muted         string
	Bold          string
	Italic        string
	CodeBg        string
	CodeFg        string
	TableBorder   string
	TableHeader   string
	StatusBarBg   string
	StatusBarFg   string
	SearchMatchBg string
	SearchMatchFg string
	CursorBg      string
}

var builtinPalettes = map[string]Palette{
	"default-dark": {
		Name:          "default-dark",
		IsDark:        true,
		H1:            "#7AA2F7",
		H2:            "#7DCFFF",
		H3:            "#9ECE6A",
		H4:            "#E0AF68",
		H5:            "#BB9AF7",
		H6:            "#73DACA",
		Muted:         "#565F89",
		Bold:          "#F7768E",
		Italic:        "#E0AF68",
		CodeBg:        "#24283B",
		CodeFg:        "#C0CAF5",
		TableBorder:   "#565F89",
		TableHeader:   "#7AA2F7",
		StatusBarBg:   "#1F2335",
		StatusBarFg:   "#C0CAF5",
		SearchMatchBg: "#E0AF68",
		SearchMatchFg: "#1A1B26",
		CursorBg:      "#C0CAF5",
	},
	"default-light": {
		Name:          "default-light",
		IsDark:        false,
		H1:            "#2E7DE9",
		H2:            "#007197",
		H3:            "#587539",
		H4:            "#8C6C3E",
		H5:            "#7847BD",
		H6:            "#387068",
		Muted:         "#8A8987",
		Bold:          "#B15C00",
		Italic:        "#8C6C3E",
		CodeBg:        "#E1E2E7",
		CodeFg:        "#373B41",
		TableBorder:   "#8A8987",
		TableHeader:   "#2E7DE9",
		StatusBarBg:   "#D5D6DB",
		StatusBarFg:   "#373B41",
		SearchMatchBg: "#F5D547",
		SearchMatchFg: "#1F1D2E",
		CursorBg:      "#373B41",
	},
	"dracula": {
		Name:          "dracula",
		IsDark:        true,
		H1:            "#BD93F9",
		H2:            "#8BE9FD",
		H3:            "#50FA7B",
		H4:            "#FFB86C",
		H5:            "#FF79C6",
		H6:            "#F1FA8C",
		Muted:         "#6272A4",
		Bold:          "#FF79C6",
		Italic:        "#F1FA8C",
		CodeBg:        "#282A36",
		CodeFg:        "#F8F8F2",
		TableBorder:   "#6272A4",
		TableHeader:   "#BD93F9",
		StatusBarBg:   "#44475A",
		StatusBarFg:   "#F8F8F2",
		SearchMatchBg: "#F1FA8C",
		SearchMatchFg: "#282A36",
		CursorBg:      "#F8F8F2",
	},
	"nord": {
		Name:          "nord",
		IsDark:        true,
		H1:            "#88C0D0",
		H2:            "#81A1C1",
		H3:            "#A3BE8C",
		H4:            "#EBCB8B",
		H5:            "#B48EAD",
		H6:            "#D08770",
		Muted:         "#4C566A",
		Bold:          "#ECEFF4",
		Italic:        "#EBCB8B",
		CodeBg:        "#2E3440",
		CodeFg:        "#ECEFF4",
		TableBorder:   "#4C566A",
		TableHeader:   "#88C0D0",
		StatusBarBg:   "#3B4252",
		StatusBarFg:   "#ECEFF4",
		SearchMatchBg: "#EBCB8B",
		SearchMatchFg: "#2E3440",
		CursorBg:      "#D8DEE9",
	},
	"catppuccin-mocha": {
		Name:          "catppuccin-mocha",
		IsDark:        true,
		H1:            "#CBA6F7",
		H2:            "#89B4FA",
		H3:            "#A6E3A1",
		H4:            "#F9E2AF",
		H5:            "#F38BA8",
		H6:            "#94E2D5",
		Muted:         "#6C7086",
		Bold:          "#F5C2E7",
		Italic:        "#F9E2AF",
		CodeBg:        "#1E1E2E",
		CodeFg:        "#CDD6F4",
		TableBorder:   "#585B70",
		TableHeader:   "#CBA6F7",
		StatusBarBg:   "#313244",
		StatusBarFg:   "#CDD6F4",
		SearchMatchBg: "#F9E2AF",
		SearchMatchFg: "#11111B",
		CursorBg:      "#F5E0DC",
	},
	"catppuccin-macchiato": {
		Name:          "catppuccin-macchiato",
		IsDark:        true,
		H1:            "#C6A0F6",
		H2:            "#8AADF4",
		H3:            "#A6DA95",
		H4:            "#EED49F",
		H5:            "#ED8796",
		H6:            "#8BD5CA",
		Muted:         "#6E738D",
		Bold:          "#F5BDE6",
		Italic:        "#EED49F",
		CodeBg:        "#24273A",
		CodeFg:        "#CAD3F5",
		TableBorder:   "#5B6078",
		TableHeader:   "#C6A0F6",
		StatusBarBg:   "#363A4F",
		StatusBarFg:   "#CAD3F5",
		SearchMatchBg: "#EED49F",
		SearchMatchFg: "#181926",
		CursorBg:      "#F4DBD6",
	},
	"monokai": {
		Name:          "monokai",
		IsDark:        true,
		H1:            "#FD971F",
		H2:            "#66D9EF",
		H3:            "#A6E22E",
		H4:            "#F92672",
		H5:            "#AE81FF",
		H6:            "#E6DB74",
		Muted:         "#75715E",
		Bold:          "#F8F8F2",
		Italic:        "#E6DB74",
		CodeBg:        "#272822",
		CodeFg:        "#F8F8F2",
		TableBorder:   "#75715E",
		TableHeader:   "#66D9EF",
		StatusBarBg:   "#3E3D32",
		StatusBarFg:   "#F8F8F2",
		SearchMatchBg: "#FFE792",
		SearchMatchFg: "#272822",
		CursorBg:      "#F8F8F0",
	},
}

// BuiltinThemes returns the list of all supported built-in theme names.
func BuiltinThemes() []string {
	return []string{
		"default-dark",
		"default-light",
		"dracula",
		"nord",
		"catppuccin-mocha",
		"catppuccin-macchiato",
		"monokai",
	}
}

// GetPalette returns the palette associated with the given theme name.
// If the theme is not recognized, it returns the default-dark palette and false.
func GetPalette(themeName string) (Palette, bool) {
	norm := strings.ToLower(strings.TrimSpace(themeName))
	p, ok := builtinPalettes[norm]
	if !ok {
		return builtinPalettes["default-dark"], false
	}
	return p, true
}
