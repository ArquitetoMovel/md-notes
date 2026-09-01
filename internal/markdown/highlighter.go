package markdown

import (
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/theme"
)

// Highlighter manages syntax highlighting for code blocks using Chroma v2 with lexer caching.
type Highlighter struct {
	mu         sync.RWMutex
	lexerCache map[string]chroma.Lexer
}

// NewHighlighter creates a new syntax highlighter.
func NewHighlighter() *Highlighter {
	return &Highlighter{
		lexerCache: make(map[string]chroma.Lexer),
	}
}

// getLexer returns a cached chroma.Lexer for the given language alias.
func (h *Highlighter) getLexer(lang string) chroma.Lexer {
	norm := strings.ToLower(strings.TrimSpace(lang))
	if norm == "" {
		return nil
	}

	h.mu.RLock()
	l, ok := h.lexerCache[norm]
	h.mu.RUnlock()
	if ok {
		return l
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check after acquiring write lock
	if l, ok := h.lexerCache[norm]; ok {
		return l
	}

	l = lexers.Get(norm)
	if l == nil {
		l = lexers.Match(norm)
	}
	if l != nil {
		l = chroma.Coalesce(l)
		h.lexerCache[norm] = l
	}
	return l
}

// HighlightLine tokenizes a single line of code with syntax highlighting using Chroma.
func (h *Highlighter) HighlightLine(code string, lang string, th *theme.CompiledTheme) []StyledSpan {
	bg := ""
	fg := ""
	if th != nil {
		bg = th.Palette.CodeBg
		fg = th.Palette.CodeFg
	}

	baseCodeStyle := lipgloss.NewStyle()
	if bg != "" {
		baseCodeStyle = baseCodeStyle.Background(lipgloss.Color(bg))
	}
	if fg != "" {
		baseCodeStyle = baseCodeStyle.Foreground(lipgloss.Color(fg))
	}

	lexer := h.getLexer(lang)
	if lexer == nil {
		// Fallback for unrecognized language: single span with 2-space visual indent
		return []StyledSpan{
			{
				Text:  "  " + code,
				Style: baseCodeStyle,
				Type:  TokenCodeBlock,
			},
		}
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return []StyledSpan{
			{
				Text:  "  " + code,
				Style: baseCodeStyle,
				Type:  TokenCodeBlock,
			},
		}
	}

	var spans []StyledSpan
	first := true

	for token := iterator(); token != chroma.EOF; token = iterator() {
		text := token.Value
		if text == "" {
			continue
		}

		if first {
			text = "  " + text
			first = false
		}

		style := h.tokenToStyle(token.Type, th, baseCodeStyle)
		spans = append(spans, StyledSpan{
			Text:  text,
			Style: style,
			Type:  TokenCodeBlock,
		})
	}

	if len(spans) == 0 {
		return []StyledSpan{
			{
				Text:  "  " + code,
				Style: baseCodeStyle,
				Type:  TokenCodeBlock,
			},
		}
	}

	return spans
}

func (h *Highlighter) tokenToStyle(tt chroma.TokenType, th *theme.CompiledTheme, baseStyle lipgloss.Style) lipgloss.Style {
	if th == nil {
		return baseStyle
	}

	style := baseStyle

	switch {
	case tt == chroma.KeywordType:
		// Keyword types (int, string, bool, etc.)
		style = style.Foreground(lipgloss.Color(th.Palette.H2)).Bold(true)
	case tt.InCategory(chroma.Keyword):
		// Keywords (func, package, return, if, else, class, def)
		style = style.Foreground(lipgloss.Color(th.Palette.H5)).Bold(true)
	case tt.InCategory(chroma.NameFunction), tt.InCategory(chroma.NameClass):
		// Function names and class names
		style = style.Foreground(lipgloss.Color(th.Palette.H1)).Bold(true)
	case tt.InCategory(chroma.Name):
		// Identifiers, variables, types
		if tt == chroma.NameBuiltin {
			style = style.Foreground(lipgloss.Color(th.Palette.H2))
		} else {
			style = style.Foreground(lipgloss.Color(th.Palette.CodeFg))
		}
	case tt.InCategory(chroma.LiteralString):
		// String literals
		style = style.Foreground(lipgloss.Color(th.Palette.H3))
	case tt.InCategory(chroma.LiteralNumber):
		// Number literals
		style = style.Foreground(lipgloss.Color(th.Palette.H4))
	case tt.InCategory(chroma.Comment):
		// Comments
		style = style.Foreground(lipgloss.Color(th.Palette.Muted)).Italic(true)
	case tt.InCategory(chroma.Operator), tt.InCategory(chroma.Punctuation):
		// Operators and punctuation
		style = style.Foreground(lipgloss.Color(th.Palette.Muted))
	default:
		style = style.Foreground(lipgloss.Color(th.Palette.CodeFg))
	}

	return style
}
