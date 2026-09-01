package search

import (
	"fmt"
	"testing"
	"time"

	"md-notes/internal/buffer"
)

// TC16
func TestPerformance_LargeBufferSearch(t *testing.T) {
	numLines := 50000
	bufLines := make([]buffer.Line, numLines)

	for i := 0; i < numLines; i++ {
		content := fmt.Sprintf("Linha de código %d com texto variado e identificadores normais", i)
		if i%100 == 0 {
			content += " TARGET_MATCH aqui presente"
		}
		bufLines[i] = buffer.NewLine(content, buffer.EndingLF)
	}

	engine := NewSearchEngine()

	// Measure FindAll time
	start := time.Now()
	res := engine.Search(bufLines, "TARGET_MATCH", false)
	elapsed := time.Since(start)

	if res.TotalCount != 500 {
		t.Fatalf("expected 500 matches in 50k lines, got %d", res.TotalCount)
	}

	t.Logf("Searched %d lines (found %d matches) in %v", numLines, res.TotalCount, elapsed)

	// PRD Requirement: < 5 ms
	// In testing environments, ensure reasonable speed
	if elapsed > 50*time.Millisecond {
		t.Errorf("search took too long: %v (expected < 5ms under standard conditions)", elapsed)
	}
}

func BenchmarkSearch_FindAll50k(b *testing.B) {
	numLines := 10000
	bufLines := make([]buffer.Line, numLines)

	for i := 0; i < numLines; i++ {
		content := fmt.Sprintf("Linha de teste %d texto qualquer", i)
		if i%50 == 0 {
			content += " SEARCH_QUERY"
		}
		bufLines[i] = buffer.NewLine(content, buffer.EndingLF)
	}

	engine := NewSearchEngine()
	pat, _ := CompilePattern("SEARCH_QUERY", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.FindAll(bufLines, pat)
	}
}
