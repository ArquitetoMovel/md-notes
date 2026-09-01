package search

import (
	"testing"

	"md-notes/internal/buffer"
)

// TC01, TC02
func TestEngine_FindAllMatches(t *testing.T) {
	lines := []buffer.Line{
		buffer.NewLine("Esta é a primeira linha de teste.", buffer.EndingLF),
		buffer.NewLine("Aqui temos outro teste e mais um teste.", buffer.EndingLF),
		buffer.NewLine("Terceira linha sem correspondência.", buffer.EndingLF),
		buffer.NewLine("Última linha com Teste maiúsculo.", buffer.EndingLF),
	}

	engine := NewSearchEngine()

	// Incremental typing simulation: "t", "te", "test", "teste"
	resT := engine.Search(lines, "t", false)
	if resT.TotalCount == 0 {
		t.Errorf("expected matches for 't', got 0")
	}

	resTeste := engine.Search(lines, "teste", false)
	// Smart-case: "teste" should match "teste" (1 on line 0, 2 on line 1, 1 on line 3 ("Teste")) = 4 matches!
	if resTeste.TotalCount != 4 {
		t.Fatalf("expected 4 matches for 'teste', got %d", resTeste.TotalCount)
	}

	// Match 0: line 0, col 26 ("teste")
	if resTeste.Matches[0].Line != 0 {
		t.Errorf("match 0 expected on line 0, got line %d", resTeste.Matches[0].Line)
	}
	// Match 1: line 1, col 17 ("teste")
	if resTeste.Matches[1].Line != 1 {
		t.Errorf("match 1 expected on line 1, got line %d", resTeste.Matches[1].Line)
	}
	// Match 3: line 3, col 17 ("Teste")
	if resTeste.Matches[3].Line != 3 || resTeste.Matches[3].Text != "Teste" {
		t.Errorf("match 3 expected 'Teste' on line 3, got %q on line %d", resTeste.Matches[3].Text, resTeste.Matches[3].Line)
	}
}

// TC03, TC04, TC05
func TestEngine_CyclicNavigation_NextAndPrev(t *testing.T) {
	lines := []buffer.Line{
		buffer.NewLine("alpha beta", buffer.EndingLF),
		buffer.NewLine("gamma beta delta", buffer.EndingLF),
		buffer.NewLine("beta omega", buffer.EndingLF),
	}

	engine := NewSearchEngine()
	res := engine.Search(lines, "beta", false)

	if res.TotalCount != 3 {
		t.Fatalf("expected 3 matches for 'beta', got %d", res.TotalCount)
	}

	// Initial position at match 0 (line 0, col 6)
	// 1. Next from (0, 6) -> should jump to match 1 (line 1, col 6)
	m1, wrap := engine.Next(0, 6)
	if m1 == nil || m1.Line != 1 || m1.StartCol != 6 || wrap {
		t.Errorf("expected match 1 at (1, 6) without wrap, got %+v, wrap=%v", m1, wrap)
	}
	if engine.Result.CurrentIndex != 2 {
		t.Errorf("expected CurrentIndex 2, got %d", engine.Result.CurrentIndex)
	}

	// 2. Next from (1, 6) -> match 2 (line 2, col 0)
	m2, wrap := engine.Next(1, 6)
	if m2 == nil || m2.Line != 2 || m2.StartCol != 0 || wrap {
		t.Errorf("expected match 2 at (2, 0) without wrap, got %+v, wrap=%v", m2, wrap)
	}
	if engine.Result.CurrentIndex != 3 {
		t.Errorf("expected CurrentIndex 3, got %d", engine.Result.CurrentIndex)
	}

	// 3. TC05: Next from (2, 0) -> should wrap around to match 0 (line 0, col 6)
	m0, wrap := engine.Next(2, 0)
	if m0 == nil || m0.Line != 0 || m0.StartCol != 6 || !wrap {
		t.Errorf("expected wrap around to match 0 at (0, 6), got %+v, wrap=%v", m0, wrap)
	}
	if engine.Result.CurrentIndex != 1 {
		t.Errorf("expected CurrentIndex 1 after wrap, got %d", engine.Result.CurrentIndex)
	}

	// 4. TC04: Prev from (0, 6) -> should wrap to match 2 (line 2, col 0)
	mPrev, wrap := engine.Prev(0, 6)
	if mPrev == nil || mPrev.Line != 2 || mPrev.StartCol != 0 || !wrap {
		t.Errorf("expected wrap around on Prev to (2, 0), got %+v, wrap=%v", mPrev, wrap)
	}
	if engine.Result.CurrentIndex != 3 {
		t.Errorf("expected CurrentIndex 3 on Prev wrap, got %d", engine.Result.CurrentIndex)
	}
}

// TC14
func TestEngine_ReverseSearch(t *testing.T) {
	lines := []buffer.Line{
		buffer.NewLine("item 1", buffer.EndingLF),
		buffer.NewLine("item 2", buffer.EndingLF),
		buffer.NewLine("item 3", buffer.EndingLF),
	}

	engine := NewSearchEngine()
	// Reverse search initialized with isReverse = true (simulating '?')
	res := engine.Search(lines, "item", true)
	if res.TotalCount != 3 {
		t.Fatalf("expected 3 matches, got %d", res.TotalCount)
	}

	// Next() in reverse mode moves backward
	m, wrap := engine.Next(2, 0)
	if m == nil || m.Line != 1 || wrap {
		t.Errorf("expected Next in reverse mode from (2, 0) to jump to line 1, got %+v, wrap=%v", m, wrap)
	}
}
