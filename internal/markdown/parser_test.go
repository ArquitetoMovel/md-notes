package markdown

import (
	"fmt"
	"testing"

	"md-notes/internal/buffer"
)

// TC01, TC02
func TestParser_HeadingsHierarchy(t *testing.T) {
	th := getTestTheme()
	parser := NewParser(th)

	for lvl := 1; lvl <= 6; lvl++ {
		hashes := ""
		for k := 0; k < lvl; k++ {
			hashes += "#"
		}
		lineStr := fmt.Sprintf("%s Título Nível %d", hashes, lvl)
		line := buffer.NewLine(lineStr, buffer.EndingLF)

		var inCodeBlock bool
		var codeLang string
		tl := parser.ParseLine(line, &inCodeBlock, &codeLang)

		if !tl.IsHeading {
			t.Fatalf("expected IsHeading to be true for level %d", lvl)
		}
		if tl.HeadingLvl != lvl {
			t.Fatalf("expected HeadingLvl %d, got %d", lvl, tl.HeadingLvl)
		}

		if len(tl.Spans) < 2 {
			t.Fatalf("expected at least 2 spans for heading level %d, got %d", lvl, len(tl.Spans))
		}

		// Marker span
		expectedMarker := hashes + " "
		if tl.Spans[0].Text != expectedMarker || tl.Spans[0].Type != TokenMarker {
			t.Errorf("heading level %d marker mismatch: got %+v, want text %q", lvl, tl.Spans[0], expectedMarker)
		}
		if !tl.Spans[0].Style.GetBold() {
			t.Errorf("heading level %d marker expected to be bold", lvl)
		}

		// Heading text span
		expectedTitle := fmt.Sprintf("Título Nível %d", lvl)
		if tl.Spans[1].Text != expectedTitle || tl.Spans[1].Type != TokenHeading {
			t.Errorf("heading level %d text mismatch: got %+v, want text %q", lvl, tl.Spans[1], expectedTitle)
		}
	}
}

// TC11
func TestParser_Blockquotes(t *testing.T) {
	th := getTestTheme()
	parser := NewParser(th)

	line := buffer.NewLine("> Esta é uma citação importante", buffer.EndingLF)
	var inCodeBlock bool
	var codeLang string
	tl := parser.ParseLine(line, &inCodeBlock, &codeLang)

	if len(tl.Spans) < 2 {
		t.Fatalf("expected at least 2 spans for blockquote, got %d", len(tl.Spans))
	}

	if tl.Spans[0].Text != "│ " || tl.Spans[0].Type != TokenMarker {
		t.Errorf("blockquote marker mismatch: got %+v, want '│ '", tl.Spans[0])
	}
	if tl.Spans[1].Text != "Esta é uma citação importante" || tl.Spans[1].Type != TokenQuote {
		t.Errorf("blockquote text mismatch: got %+v", tl.Spans[1])
	}
}

// TC09, TC10
func TestParser_ListsAndCheckboxes(t *testing.T) {
	th := getTestTheme()
	parser := NewParser(th)

	var inCodeBlock bool
	var codeLang string

	// Unordered list items: - , * , +
	bullets := []string{"- Item A", "* Item B", "+ Item C"}
	for _, b := range bullets {
		line := buffer.NewLine(b, buffer.EndingLF)
		tl := parser.ParseLine(line, &inCodeBlock, &codeLang)
		if len(tl.Spans) < 2 {
			t.Fatalf("expected at least 2 spans for %q, got %d", b, len(tl.Spans))
		}
		if tl.Spans[0].Type != TokenList {
			t.Errorf("expected TokenList for bullet %q, got %+v", b, tl.Spans[0])
		}
	}

	// Ordered list items: 1. , 2.
	numLines := []string{"1. Primeiro passo", "2. Segundo passo", "10. Décimo passo"}
	for _, n := range numLines {
		line := buffer.NewLine(n, buffer.EndingLF)
		tl := parser.ParseLine(line, &inCodeBlock, &codeLang)
		if len(tl.Spans) < 2 {
			t.Fatalf("expected at least 2 spans for %q, got %d", n, len(tl.Spans))
		}
		if tl.Spans[0].Type != TokenList {
			t.Errorf("expected TokenList for ordered list %q, got %+v", n, tl.Spans[0])
		}
	}

	// Checkboxes: - [ ] and - [x]
	chkPending := buffer.NewLine("- [ ] Tarefa pendente", buffer.EndingLF)
	tlPending := parser.ParseLine(chkPending, &inCodeBlock, &codeLang)
	if len(tlPending.Spans) < 2 {
		t.Fatalf("expected at least 2 spans for checkbox pending, got %d", len(tlPending.Spans))
	}
	if tlPending.Spans[0].Type != TokenCheckbox || tlPending.Spans[0].Text != "- [ ] " {
		t.Errorf("expected TokenCheckbox '- [ ] ', got %+v", tlPending.Spans[0])
	}

	chkDone := buffer.NewLine("- [x] Tarefa concluída", buffer.EndingLF)
	tlDone := parser.ParseLine(chkDone, &inCodeBlock, &codeLang)
	if len(tlDone.Spans) < 2 {
		t.Fatalf("expected at least 2 spans for checkbox done, got %d", len(tlDone.Spans))
	}
	if tlDone.Spans[0].Type != TokenCheckbox || tlDone.Spans[0].Text != "- [x] " {
		t.Errorf("expected TokenCheckbox '- [x] ', got %+v", tlDone.Spans[0])
	}
}

// TC07, TC08
func TestParser_CodeBlockContext(t *testing.T) {
	th := getTestTheme()
	parser := NewParser(th)

	lines := []buffer.Line{
		buffer.NewLine("```go", buffer.EndingLF),
		buffer.NewLine("package main", buffer.EndingLF),
		buffer.NewLine("func main() {", buffer.EndingLF),
		buffer.NewLine("	fmt.Println(123)", buffer.EndingLF),
		buffer.NewLine("}", buffer.EndingLF),
		buffer.NewLine("```", buffer.EndingLF),
	}

	tokenized := parser.ParseBuffer(lines)

	if len(tokenized) != 6 {
		t.Fatalf("expected 6 tokenized lines, got %d", len(tokenized))
	}

	// Line 0: ```go delimiter
	if tokenized[0].IsCodeBlock {
		t.Errorf("delimiter line should have IsCodeBlock false, got true")
	}
	if tokenized[0].Spans[0].Type != TokenMarker {
		t.Errorf("expected TokenMarker for opening delimiter, got %+v", tokenized[0].Spans[0])
	}

	// Lines 1-4: inside code block
	for idx := 1; idx <= 4; idx++ {
		if !tokenized[idx].IsCodeBlock {
			t.Errorf("line %d should have IsCodeBlock true, got false", idx)
		}
		if len(tokenized[idx].Spans) == 0 {
			t.Fatalf("line %d has no spans", idx)
		}
		// Check that code lines have visual indent and code styling
		firstSpan := tokenized[idx].Spans[0]
		if firstSpan.Type != TokenCodeBlock {
			t.Errorf("line %d span 0 expected TokenCodeBlock, got %+v", idx, firstSpan)
		}
	}

	// Line 5: closing ```
	if tokenized[5].IsCodeBlock {
		t.Errorf("closing delimiter line should have IsCodeBlock false, got true")
	}
	if tokenized[5].Spans[0].Type != TokenMarker {
		t.Errorf("expected TokenMarker for closing delimiter, got %+v", tokenized[5].Spans[0])
	}
}
