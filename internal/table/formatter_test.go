package table

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

// TC06, TC07
func TestFormatter_UnicodeAndEmojiWidth(t *testing.T) {
	// TC06: Accented Portuguese text
	accText := "Atenção e Configuração"
	accWidth := runewidth.StringWidth(accText)
	if accWidth != 22 {
		t.Errorf("expected width 22 for %q, got %d", accText, accWidth)
	}

	// TC07: Double-width emoji 🚀 (width = 2)
	emojiText := "Status: 🚀 Concluído"
	emojiWidth := runewidth.StringWidth(emojiText)
	// "Status: " (8) + "🚀" (2) + " Concluído" (10) = 20
	if emojiWidth != 20 {
		t.Errorf("expected width 20 for %q, got %d", emojiText, emojiWidth)
	}

	cell := NewCell(emojiText)
	if cell.Width != 20 {
		t.Errorf("expected cell.Width 20, got %d", cell.Width)
	}
}

// TC03, TC04, TC05, TC11
func TestFormatter_FormatGFMOutput(t *testing.T) {
	lines := []string{
		"| Nome | Cargo | Idade |",
		"|:---|:---:|---:|",
		"| Alexandre | Arquiteto de Software | 32 |",
		"| Ana | Dev | 28 |",
		"| 🚀 Robô | IA | 1 |",
	}

	table, ok := DetectTable(lines, 0)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected")
	}

	formatted := FormatMarkdownTable(table)

	if len(formatted) != 5 {
		t.Fatalf("expected 5 formatted lines, got %d", len(formatted))
	}

	// Widths:
	// Col 0: "Alexandre" (9), "🚀 Robô" (2+1+4=7) -> max 9
	// Col 1: "Arquiteto de Software" (21) -> max 21
	// Col 2: "Idade" (5) -> max 5
	expectedWidths := []int{9, 21, 5}
	for i, w := range expectedWidths {
		if table.ColWidths[i] != w {
			t.Errorf("col %d width = %d, want %d", i, table.ColWidths[i], w)
		}
	}

	// Delimiter row should have :--- (or ---), :---:, ---:
	delimLine := formatted[1]
	if !strings.Contains(delimLine, ":") {
		t.Errorf("delimiter line missing alignment markers: %s", delimLine)
	}

	// Verify all formatted lines start with '| ' and end with ' |'
	for _, fl := range formatted {
		if !strings.HasPrefix(fl, "| ") || !strings.HasSuffix(fl, " |") {
			t.Errorf("formatted line has invalid border format: %q", fl)
		}
		// Confirm pure Markdown (no Unicode box drawing characters in GFM output)
		if strings.ContainsAny(fl, "┌┬┐├┼┤└┴┘") {
			t.Errorf("pure GFM output contains box drawing characters: %s", fl)
		}
	}

	// Check alignment in data row 1 (Alexandre, Arquiteto de Software, 32)
	// Col 0 (left): "Alexandre" has 0 right spaces (width 9)
	// Col 1 (center): "Arquiteto de Software" has 0 pad (width 21)
	// Col 2 (right): "32" (width 2) has 3 leading spaces (total width 5 -> "   32")
	row1 := formatted[2]
	if !strings.Contains(row1, "   32 |") {
		t.Errorf("right aligned cell 2 mismatch in row 1: %s", row1)
	}

	// Check row 2 (Ana, Dev, 28)
	// Col 0 (left): "Ana" (3) -> "Ana      " (6 spaces)
	// Col 1 (center): "Dev" (3) -> total pad 18 -> 9 left, 9 right -> "|          Dev          |"
	// Col 2 (right): "28" (2) -> 3 spaces -> "|    28 |"
	row2 := formatted[3]
	if !strings.Contains(row2, "| Ana       |") {
		t.Errorf("left aligned cell 0 mismatch in row 2: %s", row2)
	}
	if !strings.Contains(row2, "|          Dev          |") {
		t.Errorf("center aligned cell 1 mismatch in row 2: %s", row2)
	}
	if !strings.Contains(row2, "|    28 |") {
		t.Errorf("right aligned cell 2 mismatch in row 2: %s", row2)
	}
}
