package theme

import (
	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/config"
)

// CompiledTheme contains pre-compiled immutable lipgloss.Style instances for 60 FPS zero-allocation rendering.
type CompiledTheme struct {
	Palette       Palette
	H1            lipgloss.Style
	H2            lipgloss.Style
	H3            lipgloss.Style
	H4            lipgloss.Style
	H5            lipgloss.Style
	H6            lipgloss.Style
	Muted         lipgloss.Style
	Bold          lipgloss.Style
	Italic        lipgloss.Style
	CodeBlock     lipgloss.Style
	CodeInline    lipgloss.Style
	TableBorder   lipgloss.Style
	TableHeader   lipgloss.Style
	TableCell     lipgloss.Style
	StatusBar     lipgloss.Style
	StatusBarMode lipgloss.Style
	SearchMatch   lipgloss.Style
	LineNumber    lipgloss.Style
	ActiveLineNo  lipgloss.Style
}

func applyOverrides(p Palette, overrides config.ColorOverrides) Palette {
	if overrides.H1 != "" {
		p.H1 = overrides.H1
	}
	if overrides.H2 != "" {
		p.H2 = overrides.H2
	}
	if overrides.H3 != "" {
		p.H3 = overrides.H3
	}
	if overrides.H4 != "" {
		p.H4 = overrides.H4
	}
	if overrides.H5 != "" {
		p.H5 = overrides.H5
	}
	if overrides.H6 != "" {
		p.H6 = overrides.H6
	}
	if overrides.Muted != "" {
		p.Muted = overrides.Muted
	}
	if overrides.Bold != "" {
		p.Bold = overrides.Bold
	}
	if overrides.Italic != "" {
		p.Italic = overrides.Italic
	}
	if overrides.CodeBg != "" {
		p.CodeBg = overrides.CodeBg
	}
	if overrides.CodeFg != "" {
		p.CodeFg = overrides.CodeFg
	}
	if overrides.TableBorder != "" {
		p.TableBorder = overrides.TableBorder
	}
	if overrides.TableHeader != "" {
		p.TableHeader = overrides.TableHeader
	}
	if overrides.StatusBarBg != "" {
		p.StatusBarBg = overrides.StatusBarBg
	}
	if overrides.StatusBarFg != "" {
		p.StatusBarFg = overrides.StatusBarFg
	}
	if overrides.SearchMatchBg != "" {
		p.SearchMatchBg = overrides.SearchMatchBg
	}
	if overrides.SearchMatchFg != "" {
		p.SearchMatchFg = overrides.SearchMatchFg
	}
	return p
}

// CompileTheme compiles the given palette and user overrides into a CompiledTheme struct.
func CompileTheme(palette Palette, overrides config.ColorOverrides) *CompiledTheme {
	p := applyOverrides(palette, overrides)

	return &CompiledTheme{
		Palette:       p,
		H1:            lipgloss.NewStyle().Foreground(lipgloss.Color(p.H1)).Bold(true),
		H2:            lipgloss.NewStyle().Foreground(lipgloss.Color(p.H2)).Bold(true),
		H3:            lipgloss.NewStyle().Foreground(lipgloss.Color(p.H3)).Bold(true),
		H4:            lipgloss.NewStyle().Foreground(lipgloss.Color(p.H4)).Bold(true),
		H5:            lipgloss.NewStyle().Foreground(lipgloss.Color(p.H5)).Bold(true),
		H6:            lipgloss.NewStyle().Foreground(lipgloss.Color(p.H6)).Bold(true),
		Muted:         lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)),
		Bold:          lipgloss.NewStyle().Foreground(lipgloss.Color(p.Bold)).Bold(true),
		Italic:        lipgloss.NewStyle().Foreground(lipgloss.Color(p.Italic)).Italic(true),
		CodeBlock:     lipgloss.NewStyle().Background(lipgloss.Color(p.CodeBg)).Foreground(lipgloss.Color(p.CodeFg)),
		CodeInline:    lipgloss.NewStyle().Background(lipgloss.Color(p.CodeBg)).Foreground(lipgloss.Color(p.CodeFg)),
		TableBorder:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.TableBorder)),
		TableHeader:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.TableHeader)).Bold(true),
		TableCell:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.CodeFg)),
		StatusBar:     lipgloss.NewStyle().Background(lipgloss.Color(p.StatusBarBg)).Foreground(lipgloss.Color(p.StatusBarFg)),
		StatusBarMode: lipgloss.NewStyle().Background(lipgloss.Color(p.H1)).Foreground(lipgloss.Color(p.StatusBarBg)).Bold(true),
		SearchMatch:   lipgloss.NewStyle().Background(lipgloss.Color(p.SearchMatchBg)).Foreground(lipgloss.Color(p.SearchMatchFg)).Bold(true),
		LineNumber:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)),
		ActiveLineNo:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.H1)).Bold(true),
	}
}
