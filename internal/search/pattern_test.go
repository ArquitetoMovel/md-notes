package search

import (
	"testing"
)

// TC09
func TestPattern_SmartCaseDetection(t *testing.T) {
	// 1. Lowercase query -> case-insensitive
	pLower, err := CompilePattern("funcao", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pLower.IsCaseSensitive {
		t.Errorf("expected lowercase query 'funcao' to be case-insensitive")
	}

	testLine := "Funcao funcao FUNCAO outra_coisa"
	matchesLower := pLower.FindInLine(testLine)
	if len(matchesLower) != 3 {
		t.Errorf("expected 3 matches for smart-case lowercase query, got %d", len(matchesLower))
	}

	// 2. Query with uppercase -> case-sensitive
	pUpper, err := CompilePattern("Funcao", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pUpper.IsCaseSensitive {
		t.Errorf("expected query with uppercase 'Funcao' to be case-sensitive")
	}

	matchesUpper := pUpper.FindInLine(testLine)
	if len(matchesUpper) != 1 {
		t.Errorf("expected 1 match for uppercase query, got %d", len(matchesUpper))
	}
}

// TC08, TC15
func TestPattern_RegexAndIncompleteFallback(t *testing.T) {
	// Valid regex: \d{3}-\w+
	pRegex, err := CompilePattern(`\d{3}-\w+`, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pRegex.IsLiteral {
		t.Errorf("expected valid regex to not be literal")
	}

	line := "codigo: 123-abc e 456-def"
	matches := pRegex.FindInLine(line)
	if len(matches) != 2 {
		t.Fatalf("expected 2 regex matches, got %d", len(matches))
	}

	// TC15: Incomplete regex with unclosed parenthesis: "Println("
	pIncomplete, err := CompilePattern("fmt.Println(", true)
	if err != nil {
		t.Fatalf("unexpected error on incomplete regex: %v", err)
	}
	if !pIncomplete.IsLiteral {
		t.Errorf("expected incomplete regex to fallback to literal search")
	}

	codeLine := `fmt.Println("Olá Mundo")`
	matchesLiteral := pIncomplete.FindInLine(codeLine)
	if len(matchesLiteral) != 1 {
		t.Fatalf("expected 1 literal match for 'fmt.Println(', got %d", len(matchesLiteral))
	}
}

// TC06
func TestMatch_StatusString(t *testing.T) {
	sr := &SearchResult{
		Matches: []Match{
			{Line: 0, StartCol: 0, EndCol: 4, Text: "test"},
			{Line: 1, StartCol: 5, EndCol: 9, Text: "test"},
			{Line: 2, StartCol: 2, EndCol: 6, Text: "test"},
		},
		TotalCount:   3,
		CurrentIndex: 2,
		Query:        "test",
	}

	if sr.StatusString() != "[2/3]" {
		t.Errorf("StatusString = %q, want [2/3]", sr.StatusString())
	}

	srEmpty := &SearchResult{TotalCount: 0}
	if srEmpty.StatusString() != "[0/0]" {
		t.Errorf("empty StatusString = %q, want [0/0]", srEmpty.StatusString())
	}
}
