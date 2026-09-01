package markdown

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/buffer"
)

// TokenType represents the semantic type of a markdown token.
type TokenType int

const (
	TokenText TokenType = iota
	TokenMarker
	TokenHeading
	TokenBold
	TokenItalic
	TokenStrike
	TokenCodeInline
	TokenCodeBlock
	TokenQuote
	TokenList
	TokenCheckbox
	TokenLink
)

// String returns the string representation of the token type.
func (t TokenType) String() string {
	switch t {
	case TokenText:
		return "Text"
	case TokenMarker:
		return "Marker"
	case TokenHeading:
		return "Heading"
	case TokenBold:
		return "Bold"
	case TokenItalic:
		return "Italic"
	case TokenStrike:
		return "Strike"
	case TokenCodeInline:
		return "CodeInline"
	case TokenCodeBlock:
		return "CodeBlock"
	case TokenQuote:
		return "Quote"
	case TokenList:
		return "List"
	case TokenCheckbox:
		return "Checkbox"
	case TokenLink:
		return "Link"
	default:
		return "Unknown"
	}
}

// StyledSpan represents a styled segment of text with an associated lipgloss Style and TokenType.
type StyledSpan struct {
	Text  string
	Style lipgloss.Style
	Type  TokenType
}

// TokenizedLine represents a single line of markdown decomposed into styled spans,
// retaining reference to the original buffer.Line.
type TokenizedLine struct {
	Spans       []StyledSpan
	Original    buffer.Line
	IsCodeBlock bool
	IsHeading   bool
	HeadingLvl  int
}

// Render returns the ANSI-formatted string representing the styled line.
func (tl TokenizedLine) Render() string {
	if len(tl.Spans) == 0 {
		return tl.Original.String()
	}
	var sb strings.Builder
	for _, span := range tl.Spans {
		sb.WriteString(span.Style.Render(span.Text))
	}
	return sb.String()
}

// PlainText returns the unstyled plain text of the line.
func (tl TokenizedLine) PlainText() string {
	if len(tl.Spans) == 0 {
		return tl.Original.String()
	}
	var sb strings.Builder
	for _, span := range tl.Spans {
		sb.WriteString(span.Text)
	}
	return sb.String()
}
