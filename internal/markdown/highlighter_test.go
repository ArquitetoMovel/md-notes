package markdown

import (
	"testing"
)

// TC14
func TestHighlighter_GoCode(t *testing.T) {
	th := getTestTheme()
	hl := NewHighlighter()

	goLine := "func main() {"
	spans := hl.HighlightLine(goLine, "go", th)

	if len(spans) == 0 {
		t.Fatalf("expected spans for go code line, got 0")
	}

	// First span should have visual indent and keyword 'func'
	if spans[0].Type != TokenCodeBlock {
		t.Errorf("expected TokenCodeBlock, got %v", spans[0].Type)
	}

	foundFunc := false
	foundMain := false
	for _, span := range spans {
		if span.Text == "  func" || span.Text == "func" {
			foundFunc = true
		}
		if span.Text == "main" || span.Text == " main" {
			foundMain = true
		}
	}

	if !foundFunc {
		t.Errorf("did not find 'func' keyword in spans: %+v", spans)
	}
	if !foundMain {
		t.Errorf("did not find 'main' function identifier in spans: %+v", spans)
	}
}

// TC14
func TestHighlighter_PythonAndJSON(t *testing.T) {
	th := getTestTheme()
	hl := NewHighlighter()

	// Python
	pyLine := "def calculate_sum(a, b):"
	pySpans := hl.HighlightLine(pyLine, "python", th)
	if len(pySpans) == 0 {
		t.Fatalf("expected spans for python code line, got 0")
	}
	foundDef := false
	for _, span := range pySpans {
		if span.Text == "  def" || span.Text == "def" {
			foundDef = true
		}
	}
	if !foundDef {
		t.Errorf("expected 'def' in python spans: %+v", pySpans)
	}

	// JSON
	jsonLine := `{"title": "md-notes", "count": 42}`
	jsonSpans := hl.HighlightLine(jsonLine, "json", th)
	if len(jsonSpans) == 0 {
		t.Fatalf("expected spans for json line, got 0")
	}
}

// TC15
func TestHighlighter_FallbackUnrecognizedLang(t *testing.T) {
	th := getTestTheme()
	hl := NewHighlighter()

	unknownLine := "some arbitrary content in unknown language"
	spans := hl.HighlightLine(unknownLine, "linguagem_desconhecida_123", th)

	if len(spans) != 1 {
		t.Fatalf("expected 1 fallback span, got %d: %+v", len(spans), spans)
	}
	if spans[0].Type != TokenCodeBlock {
		t.Errorf("expected TokenCodeBlock, got %v", spans[0].Type)
	}
	if spans[0].Text != "  "+unknownLine {
		t.Errorf("expected indented text '  %s', got %q", unknownLine, spans[0].Text)
	}

	// Empty language fallback
	emptySpans := hl.HighlightLine(unknownLine, "", th)
	if len(emptySpans) != 1 {
		t.Fatalf("expected 1 span for empty lang, got %d", len(emptySpans))
	}
}
