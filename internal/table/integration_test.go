package table

import (
	"strings"
	"testing"

	"md-notes/internal/buffer"
)

// TC11, TC17
func TestTable_BufferIntegrationAndCursorStability(t *testing.T) {
	buf := buffer.NewEmptyBuffer("table_test.md")

	initialDoc := []string{
		"# Tabela de Exemplo",
		"",
		"| Nome | Cargo | Salário |",
		"|:---|:---|---:|",
		"| Alexandre | Arquiteto | 15000 |",
		"| Beatriz | Engenheira | 12000 |",
		"",
		"Fim do documento.",
	}

	for i, line := range initialDoc {
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(line, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(line, buffer.EndingLF))
		}
	}

	// 1. Detect table at line 4 ("Alexandre | Arquiteto")
	tbl, ok := DetectTable(strings.Split(buf.RawText(), "\n"), 4)
	if !ok || tbl == nil {
		t.Fatalf("expected table to be detected at line 4")
	}

	if tbl.StartLine != 2 || tbl.EndLine != 5 {
		t.Fatalf("expected table bounds [2, 5], got [%d, %d]", tbl.StartLine, tbl.EndLine)
	}

	// 2. Modify cell content and reformat
	tbl.Rows[1].Cells[1] = NewCell("Engenheira Sênior")
	ApplyTableToBuffer(buf, tbl)

	// Verify buffer lines after apply
	raw := buf.RawText()
	if !strings.Contains(raw, "Engenheira Sênior") {
		t.Fatalf("expected updated cell in buffer")
	}
	if !strings.Contains(raw, "# Tabela de Exemplo") || !strings.Contains(raw, "Fim do documento.") {
		t.Fatalf("buffer lost surrounding lines after table apply")
	}

	// TC11: Confirm pure Markdown in buffer
	if strings.ContainsAny(raw, "┌┬┐├┼┤└┴┘") {
		t.Fatalf("buffer contains Unicode box characters instead of pure Markdown")
	}

	// TC17: Cursor position calculation
	cursorCol := GetCellCursorCol(tbl, 1)
	if cursorCol <= 0 {
		t.Errorf("invalid cursor column calculation: %d", cursorCol)
	}

	// Render Unicode with theme
	th := getTestTheme()
	rendered := tbl.RenderUnicode(th)
	if len(rendered) != 6 {
		t.Fatalf("expected 6 rendered lines (top + head + mid + 2 rows + bot), got %d", len(rendered))
	}
}
