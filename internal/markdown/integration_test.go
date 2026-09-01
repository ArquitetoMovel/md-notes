package markdown

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"md-notes/internal/buffer"
	"md-notes/internal/theme"
)

// TC12, TC13
func TestMarkdown_ThemeSwitching(t *testing.T) {
	palDracula, _ := theme.GetPalette("dracula")
	thDracula := theme.CompileTheme(palDracula, theme.ColorOverrides{})

	palNord, _ := theme.GetPalette("nord")
	thNord := theme.CompileTheme(palNord, theme.ColorOverrides{})

	parser := NewParser(thDracula)

	headingLine := buffer.NewLine("# Arquitetura do Sistema", buffer.EndingLF)
	var inCodeBlock bool
	var codeLang string

	tlDracula := parser.ParseLine(headingLine, &inCodeBlock, &codeLang)
	renderedDracula := tlDracula.Render()

	// Switch theme to Nord
	parser.SetTheme(thNord)
	inCodeBlock = false
	codeLang = ""
	tlNord := parser.ParseLine(headingLine, &inCodeBlock, &codeLang)
	renderedNord := tlNord.Render()

	if renderedDracula == "" || renderedNord == "" {
		t.Fatalf("expected non-empty rendered strings")
	}

	// Verify both plain texts are identical
	if tlDracula.PlainText() != tlNord.PlainText() {
		t.Errorf("plain text mismatch: %q vs %q", tlDracula.PlainText(), tlNord.PlainText())
	}
}

// TC12, TC18
func TestMarkdown_FullDocumentParsing(t *testing.T) {
	th := getTestTheme()
	parser := NewParser(th)

	rawDoc := `# Introdução ao Projeto

Este é um documento de teste para o **md-notes**, um editor de terminal com suporte a *Markdown*.

## Funcionalidades
- [x] Syntax highlighting em tempo real com marcadores atenuados
- [ ] Alinhamento dinâmico de tabelas
- Citações em bloco com barra decorativa:

> O software bem arquitetado é aquele que evolui sem fricção.

Acesse a [Documentação Oficial](https://golang.org) ou https://github.com para mais detalhes.

### Código de Exemplo
` + "```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello md-notes!\")\n}\n```\n" + `
Texto final com ~~tachado~~ e ` + "`código inline`" + ` aqui.
`

	linesStr := strings.Split(rawDoc, "\n")
	bufLines := make([]buffer.Line, len(linesStr))
	for i, s := range linesStr {
		bufLines[i] = buffer.NewLine(s, buffer.EndingLF)
	}

	tokenized := parser.ParseBuffer(bufLines)
	if len(tokenized) != len(bufLines) {
		t.Fatalf("expected %d tokenized lines, got %d", len(bufLines), len(tokenized))
	}

	// Validate heading
	if !tokenized[0].IsHeading || tokenized[0].HeadingLvl != 1 {
		t.Errorf("line 0 should be H1 heading, got %+v", tokenized[0])
	}

	// Validate code block lines
	inCodeCount := 0
	for _, tl := range tokenized {
		if tl.IsCodeBlock {
			inCodeCount++
		}
	}
	if inCodeCount < 5 {
		t.Errorf("expected at least 5 lines inside code block, got %d", inCodeCount)
	}
}

// TC18
func TestMarkdown_ParsingPerformance(t *testing.T) {
	th := getTestTheme()
	parser := NewParser(th)

	// Create 1000 lines document with diverse markdown elements
	numLines := 1000
	bufLines := make([]buffer.Line, numLines)

	for i := 0; i < numLines; i++ {
		var content string
		switch i % 10 {
		case 0:
			content = fmt.Sprintf("# Título Principal Nível 1 - Seção %d", i)
		case 1:
			content = fmt.Sprintf("Parágrafo com texto em **negrito** e *itálico* e `código inline` na linha %d.", i)
		case 2:
			content = fmt.Sprintf("- [x] Tarefa concluída número %d com link [Google](https://google.com)", i)
		case 3:
			content = fmt.Sprintf("> Citação elegante e reflexiva sobre o item %d", i)
		case 4:
			content = "```go"
		case 5:
			content = fmt.Sprintf("func ProcessItem%d(val int) string {", i)
		case 6:
			content = fmt.Sprintf("\treturn fmt.Sprintf(\"Item: %%d\", val)")
		case 7:
			content = "}"
		case 8:
			content = "```"
		case 9:
			content = fmt.Sprintf("Linha regular de texto com ~~tachado~~ e números 1. 2. 3. na posição %d.", i)
		}
		bufLines[i] = buffer.NewLine(content, buffer.EndingLF)
	}

	// Warm-up
	parser.ParseBuffer(bufLines)

	// Measure 50 iterations
	iterations := 50
	start := time.Now()
	for k := 0; k < iterations; k++ {
		_ = parser.ParseBuffer(bufLines)
	}
	elapsed := time.Since(start)

	totalParsedLines := numLines * iterations
	avgPerLine := elapsed / time.Duration(totalParsedLines)

	t.Logf("Parsed %d lines in %v (average %v per line)", totalParsedLines, elapsed, avgPerLine)

	// PRD Requirement: < 1ms (1000 microseconds) per line
	if avgPerLine > time.Millisecond {
		t.Fatalf("parsing too slow: %v per line (must be < 1ms)", avgPerLine)
	}
}

func BenchmarkMarkdown_ParseBuffer(b *testing.B) {
	th := getTestTheme()
	parser := NewParser(th)

	numLines := 500
	bufLines := make([]buffer.Line, numLines)
	for i := 0; i < numLines; i++ {
		bufLines[i] = buffer.NewLine(fmt.Sprintf("Linha **%d** com *estilo* e `código`", i), buffer.EndingLF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.ParseBuffer(bufLines)
	}
}
