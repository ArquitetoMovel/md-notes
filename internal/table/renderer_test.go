package table

import (
	"strings"
	"testing"

	"md-notes/internal/theme"
)

func getTestTheme() *theme.CompiledTheme {
	palette, _ := theme.GetPalette("default-dark")
	return theme.CompileTheme(palette, theme.ColorOverrides{})
}

// TC12
func TestRenderer_UnicodeBoxBorders(t *testing.T) {
	th := getTestTheme()

	lines := []string{
		"| Col 1 | Col 2 |",
		"|---|---|",
		"| Val A | Val B |",
	}

	table, ok := DetectTable(lines, 0)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected")
	}

	rendered := table.RenderUnicode(th)

	// Expect:
	// 0: ┌───────┬───────┐
	// 1: │ Col 1 │ Col 2 │
	// 2: ├───────┼───────┤
	// 3: │ Val A │ Val B │
	// 4: └───────┴───────┘
	if len(rendered) != 5 {
		t.Fatalf("expected 5 rendered lines, got %d", len(rendered))
	}

	// Verify top border has ┌, ┬, ┐
	if !strings.Contains(rendered[0], "┌") || !strings.Contains(rendered[0], "┬") || !strings.Contains(rendered[0], "┐") {
		t.Errorf("top border missing box characters: %s", rendered[0])
	}

	// Verify middle divider has ├, ┼, ┤
	if !strings.Contains(rendered[2], "├") || !strings.Contains(rendered[2], "┼") || !strings.Contains(rendered[2], "┤") {
		t.Errorf("middle border missing box characters: %s", rendered[2])
	}

	// Verify bottom border has └, ┴, ┘
	if !strings.Contains(rendered[4], "└") || !strings.Contains(rendered[4], "┴") || !strings.Contains(rendered[4], "┘") {
		t.Errorf("bottom border missing box characters: %s", rendered[4])
	}
}

func TestRenderer_AsciiFallback(t *testing.T) {
	th := getTestTheme()

	lines := []string{
		"| A | B |",
		"|---|---|",
		"| 1 | 2 |",
	}

	table, ok := DetectTable(lines, 0)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected")
	}

	asciiLines := table.RenderASCII(th)

	if len(asciiLines) != 3 {
		t.Fatalf("expected 3 ASCII lines, got %d", len(asciiLines))
	}

	for _, l := range asciiLines {
		if strings.ContainsAny(l, "┌┬┐├┼┤└┴┘") {
			t.Errorf("ASCII fallback contains Unicode box drawing: %s", l)
		}
	}
}

// TC16
func TestRenderer_InlineFormattingInCells(t *testing.T) {
	th := getTestTheme()

	lines := []string{
		"| Item | Descrição |",
		"|---|---|",
		"| Destaque | Texto em **negrito** e `código` |",
	}

	table, ok := DetectTable(lines, 0)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected")
	}

	rendered := table.RenderUnicode(th)

	if len(rendered) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(rendered))
	}

	dataRow := rendered[3]
	if !strings.Contains(dataRow, "Destaque") {
		t.Errorf("data row missing cell text: %s", dataRow)
	}
	if !strings.Contains(dataRow, "negrito") || !strings.Contains(dataRow, "código") {
		t.Errorf("data row missing inline markdown content: %s", dataRow)
	}
}
