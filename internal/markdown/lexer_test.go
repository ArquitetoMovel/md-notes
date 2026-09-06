package markdown

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/theme"
)

func getTestTheme() *theme.CompiledTheme {
	palette, _ := theme.GetPalette("default-dark")
	return theme.CompileTheme(palette, theme.ColorOverrides{})
}

// TC03, TC05, TC06
func TestLexer_BoldAndItalic(t *testing.T) {
	th := getTestTheme()
	lexer := NewLexer(th)

	// Bold test
	boldInput := "Este é um texto **importante e em negrito** aqui"
	spans := lexer.TokenizeInline(boldInput, lipgloss.NewStyle())

	// Expect: "Este é um texto ", "**", "importante e em negrito", "**", " aqui"
	if len(spans) != 5 {
		t.Fatalf("expected 5 spans, got %d: %+v", len(spans), spans)
	}

	if spans[0].Text != "Este é um texto " || spans[0].Type != TokenText {
		t.Errorf("span 0 mismatch: %+v", spans[0])
	}
	if spans[1].Text != "**" || spans[1].Type != TokenMarker {
		t.Errorf("span 1 (marker) mismatch: %+v", spans[1])
	}
	if spans[2].Text != "importante e em negrito" || spans[2].Type != TokenBold {
		t.Errorf("span 2 (bold) mismatch: %+v", spans[2])
	}
	if spans[3].Text != "**" || spans[3].Type != TokenMarker {
		t.Errorf("span 3 (marker) mismatch: %+v", spans[3])
	}
	if spans[4].Text != " aqui" || spans[4].Type != TokenText {
		t.Errorf("span 4 mismatch: %+v", spans[4])
	}

	// Italic with * and _
	italicInput := "*ênfase um* e _ênfase dois_"
	spansItalic := lexer.TokenizeInline(italicInput, lipgloss.NewStyle())

	// Expect: "*", "ênfase um", "*", " e ", "_", "ênfase dois", "_"
	if len(spansItalic) != 7 {
		t.Fatalf("expected 7 spans, got %d: %+v", len(spansItalic), spansItalic)
	}
	if spansItalic[0].Text != "*" || spansItalic[0].Type != TokenMarker {
		t.Errorf("span 0 marker mismatch: %+v", spansItalic[0])
	}
	if spansItalic[1].Text != "ênfase um" || spansItalic[1].Type != TokenItalic {
		t.Errorf("span 1 italic mismatch: %+v", spansItalic[1])
	}
	if spansItalic[2].Text != "*" || spansItalic[2].Type != TokenMarker {
		t.Errorf("span 2 marker mismatch: %+v", spansItalic[2])
	}
	if spansItalic[3].Text != " e " || spansItalic[3].Type != TokenText {
		t.Errorf("span 3 text mismatch: %+v", spansItalic[3])
	}
	if spansItalic[4].Text != "_" || spansItalic[4].Type != TokenMarker {
		t.Errorf("span 4 marker mismatch: %+v", spansItalic[4])
	}
	if spansItalic[5].Text != "ênfase dois" || spansItalic[5].Type != TokenItalic {
		t.Errorf("span 5 italic mismatch: %+v", spansItalic[5])
	}
	if spansItalic[6].Text != "_" || spansItalic[6].Type != TokenMarker {
		t.Errorf("span 6 marker mismatch: %+v", spansItalic[6])
	}
}

// TC04
func TestLexer_InlineCode(t *testing.T) {
	th := getTestTheme()
	lexer := NewLexer(th)

	input := "execute o comando `go test` para validar"
	spans := lexer.TokenizeInline(input, lipgloss.NewStyle())

	if len(spans) != 5 {
		t.Fatalf("expected 5 spans, got %d: %+v", len(spans), spans)
	}

	if spans[1].Text != "`" || spans[1].Type != TokenMarker {
		t.Errorf("expected opening ` marker, got %+v", spans[1])
	}
	if spans[2].Text != "go test" || spans[2].Type != TokenCodeInline {
		t.Errorf("expected CodeInline 'go test', got %+v", spans[2])
	}
	if spans[3].Text != "`" || spans[3].Type != TokenMarker {
		t.Errorf("expected closing ` marker, got %+v", spans[3])
	}
}

// TC03
func TestLexer_Strikethrough(t *testing.T) {
	th := getTestTheme()
	lexer := NewLexer(th)

	input := "texto ~~tachado~~ normal"
	spans := lexer.TokenizeInline(input, lipgloss.NewStyle())

	if len(spans) != 5 {
		t.Fatalf("expected 5 spans, got %d: %+v", len(spans), spans)
	}

	if spans[1].Text != "~~" || spans[1].Type != TokenMarker {
		t.Errorf("expected opening ~~ marker, got %+v", spans[1])
	}
	if spans[2].Text != "tachado" || spans[2].Type != TokenStrike {
		t.Errorf("expected TokenStrike 'tachado', got %+v", spans[2])
	}
	if spans[3].Text != "~~" || spans[3].Type != TokenMarker {
		t.Errorf("expected closing ~~ marker, got %+v", spans[3])
	}
}

// TC16
func TestLexer_LinksAndURLs(t *testing.T) {
	th := getTestTheme()
	lexer := NewLexer(th)

	input := "Acesse o site [Documentação](https://golang.org) para mais detalhes."
	spans := lexer.TokenizeInline(input, lipgloss.NewStyle())

	// Expect:
	// "Acesse o site " (Text)
	// "[" (Marker)
	// "Documentação" (Link)
	// "]" (Marker)
	// "(" (Marker)
	// "https://golang.org" (Marker)
	// ")" (Marker)
	// " para mais detalhes." (Text)
	if len(spans) != 8 {
		t.Fatalf("expected 8 spans, got %d: %+v", len(spans), spans)
	}

	if spans[1].Text != "[" || spans[1].Type != TokenMarker {
		t.Errorf("expected [ marker, got %+v", spans[1])
	}
	if spans[2].Text != "Documentação" || spans[2].Type != TokenLink {
		t.Errorf("expected TokenLink 'Documentação', got %+v", spans[2])
	}
	if spans[3].Text != "]" || spans[3].Type != TokenMarker {
		t.Errorf("expected ] marker, got %+v", spans[3])
	}
	if spans[4].Text != "(" || spans[4].Type != TokenMarker {
		t.Errorf("expected ( marker, got %+v", spans[4])
	}
	if spans[5].Text != "https://golang.org" || spans[5].Type != TokenMarker {
		t.Errorf("expected url marker, got %+v", spans[5])
	}
	if spans[6].Text != ")" || spans[6].Type != TokenMarker {
		t.Errorf("expected ) marker, got %+v", spans[6])
	}

	// Direct URL
	directInput := "Visite https://charm.sh hoje!"
	directSpans := lexer.TokenizeInline(directInput, lipgloss.NewStyle())
	if len(directSpans) != 3 {
		t.Fatalf("expected 3 spans for direct URL, got %d: %+v", len(directSpans), directSpans)
	}
	if directSpans[1].Text != "https://charm.sh" || directSpans[1].Type != TokenLink {
		t.Errorf("expected TokenLink for raw URL, got %+v", directSpans[1])
	}
}

// TC17
func TestLexer_NestedStyles(t *testing.T) {
	th := getTestTheme()
	lexer := NewLexer(th)

	input := "**negrito com `código inline` e *itálico* aninhado**"
	spans := lexer.TokenizeInline(input, lipgloss.NewStyle())

	// Expect:
	// "**" (Marker)
	// "negrito com " (Bold)
	// "`" (Marker)
	// "código inline" (CodeInline)
	// "`" (Marker)
	// " e " (Bold)
	// "*" (Marker)
	// "itálico" (Italic)
	// "*" (Marker)
	// " aninhado" (Bold)
	// "**" (Marker)

	if len(spans) != 11 {
		t.Fatalf("expected 11 spans for nested styles, got %d: %+v", len(spans), spans)
	}

	if spans[0].Text != "**" || spans[0].Type != TokenMarker {
		t.Errorf("span 0 mismatch: %+v", spans[0])
	}
	if spans[1].Text != "negrito com " || spans[1].Type != TokenBold {
		t.Errorf("span 1 mismatch: %+v", spans[1])
	}
	if spans[2].Text != "`" || spans[2].Type != TokenMarker {
		t.Errorf("span 2 mismatch: %+v", spans[2])
	}
	if spans[3].Text != "código inline" || spans[3].Type != TokenCodeInline {
		t.Errorf("span 3 mismatch: %+v", spans[3])
	}
	if spans[4].Text != "`" || spans[4].Type != TokenMarker {
		t.Errorf("span 4 mismatch: %+v", spans[4])
	}
	if spans[5].Text != " e " || spans[5].Type != TokenBold {
		t.Errorf("span 5 mismatch: %+v", spans[5])
	}
	if spans[6].Text != "*" || spans[6].Type != TokenMarker {
		t.Errorf("span 6 mismatch: %+v", spans[6])
	}
	if spans[7].Text != "itálico" || spans[7].Type != TokenItalic {
		t.Errorf("span 7 mismatch: %+v", spans[7])
	}
	if spans[8].Text != "*" || spans[8].Type != TokenMarker {
		t.Errorf("span 8 mismatch: %+v", spans[8])
	}
	if spans[9].Text != " aninhado" || spans[9].Type != TokenBold {
		t.Errorf("span 9 mismatch: %+v", spans[9])
	}
	if spans[10].Text != "**" || spans[10].Type != TokenMarker {
		t.Errorf("span 10 mismatch: %+v", spans[10])
	}
}

func TestLexer_MarkerStyles_BoldAndItalic(t *testing.T) {
	th := getTestTheme()
	lexer := NewLexer(th)

	// Test bold **
	spansBoldAsterisk := lexer.TokenizeInline("**negrito**", lipgloss.NewStyle())
	if len(spansBoldAsterisk) != 3 {
		t.Fatalf("expected 3 spans for **negrito**, got %d", len(spansBoldAsterisk))
	}
	if !spansBoldAsterisk[0].Style.GetBold() || !spansBoldAsterisk[2].Style.GetBold() {
		t.Errorf("expected ** markers to be bold")
	}

	// Test bold __
	spansBoldUnderscore := lexer.TokenizeInline("__negrito__", lipgloss.NewStyle())
	if len(spansBoldUnderscore) != 3 {
		t.Fatalf("expected 3 spans for __negrito__, got %d", len(spansBoldUnderscore))
	}
	if !spansBoldUnderscore[0].Style.GetBold() || !spansBoldUnderscore[2].Style.GetBold() {
		t.Errorf("expected __ markers to be bold")
	}

	// Test italic *
	spansItalicAsterisk := lexer.TokenizeInline("*italico*", lipgloss.NewStyle())
	if len(spansItalicAsterisk) != 3 {
		t.Fatalf("expected 3 spans for *italico*, got %d", len(spansItalicAsterisk))
	}
	if !spansItalicAsterisk[0].Style.GetItalic() || !spansItalicAsterisk[2].Style.GetItalic() {
		t.Errorf("expected * markers to be italic")
	}

	// Test italic _
	spansItalicUnderscore := lexer.TokenizeInline("_italico_", lipgloss.NewStyle())
	if len(spansItalicUnderscore) != 3 {
		t.Fatalf("expected 3 spans for _italico_, got %d", len(spansItalicUnderscore))
	}
	if !spansItalicUnderscore[0].Style.GetItalic() || !spansItalicUnderscore[2].Style.GetItalic() {
		t.Errorf("expected _ markers to be italic")
	}
}

