package markdown

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/buffer"
)

func TestTokenizedLine_RenderAndPlainText(t *testing.T) {
	styleBold := lipgloss.NewStyle().Bold(true)
	styleMuted := lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89"))

	origLine := buffer.NewLine("# Title", buffer.EndingLF)
	tl := TokenizedLine{
		Spans: []StyledSpan{
			{Text: "# ", Style: styleMuted, Type: TokenMarker},
			{Text: "Title", Style: styleBold, Type: TokenHeading},
		},
		Original:   origLine,
		IsHeading:  true,
		HeadingLvl: 1,
	}

	// Plain text should be exactly "# Title"
	if tl.PlainText() != "# Title" {
		t.Fatalf("expected PlainText to be '# Title', got %q", tl.PlainText())
	}

	// Rendered text should contain "# " and "Title" and ANSI escapes
	rendered := tl.Render()
	if rendered == "" {
		t.Fatalf("expected non-empty rendered string")
	}

	if tl.Original.String() != "# Title" {
		t.Fatalf("expected original line string to be '# Title', got %q", tl.Original.String())
	}
}

func TestTokenizedLine_EmptySpans(t *testing.T) {
	origLine := buffer.NewLine("Plain line content", buffer.EndingLF)
	tl := TokenizedLine{
		Spans:    nil,
		Original: origLine,
	}

	if tl.PlainText() != "Plain line content" {
		t.Fatalf("expected PlainText '%s', got '%s'", "Plain line content", tl.PlainText())
	}
	if tl.Render() != "Plain line content" {
		t.Fatalf("expected Render '%s', got '%s'", "Plain line content", tl.Render())
	}
}

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		token TokenType
		want  string
	}{
		{TokenText, "Text"},
		{TokenMarker, "Marker"},
		{TokenHeading, "Heading"},
		{TokenBold, "Bold"},
		{TokenItalic, "Italic"},
		{TokenStrike, "Strike"},
		{TokenCodeInline, "CodeInline"},
		{TokenCodeBlock, "CodeBlock"},
		{TokenQuote, "Quote"},
		{TokenList, "List"},
		{TokenCheckbox, "Checkbox"},
		{TokenLink, "Link"},
		{TokenType(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.token.String(); got != tt.want {
			t.Errorf("TokenType(%d).String() = %q, want %q", tt.token, got, tt.want)
		}
	}
}
