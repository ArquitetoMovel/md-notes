package table

import (
	"testing"

	"md-notes/internal/buffer"
)

// TC08
func TestNavigation_TabNextCell(t *testing.T) {
	lines := []string{
		"| Col 0 | Col 1 | Col 2 |",
		"|---|---|---|",
		"| A0 | B0 | C0 |",
		"| A1 | B1 | C1 |",
	}

	table, ok := DetectTable(lines, 0)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected")
	}

	// Case 1: Col 0 -> Col 1 on row 0
	nextRow, nextCol, isNewRow := NextCell(table, 0, 0)
	if nextRow != 0 || nextCol != 1 || isNewRow {
		t.Errorf("expected (0, 1, false), got (%d, %d, %v)", nextRow, nextCol, isNewRow)
	}

	// Case 2: Last col (2) on row 0 -> Col 0 on row 1
	nextRow, nextCol, isNewRow = NextCell(table, 0, 2)
	if nextRow != 1 || nextCol != 0 || isNewRow {
		t.Errorf("expected (1, 0, false), got (%d, %d, %v)", nextRow, nextCol, isNewRow)
	}

	// Case 3: Last col (2) on last row (1) -> triggers new row
	nextRow, nextCol, isNewRow = NextCell(table, 1, 2)
	if nextRow != 2 || nextCol != 0 || !isNewRow {
		t.Errorf("expected (2, 0, true), got (%d, %d, %v)", nextRow, nextCol, isNewRow)
	}
}

// TC09
func TestNavigation_ShiftTabPrevCell(t *testing.T) {
	lines := []string{
		"| Col 0 | Col 1 | Col 2 |",
		"|---|---|---|",
		"| A0 | B0 | C0 |",
		"| A1 | B1 | C1 |",
	}

	table, ok := DetectTable(lines, 0)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected")
	}

	// Case 1: Col 1 -> Col 0 on row 1
	prevRow, prevCol := PrevCell(table, 1, 1)
	if prevRow != 1 || prevCol != 0 {
		t.Errorf("expected (1, 0), got (%d, %d)", prevRow, prevCol)
	}

	// Case 2: Col 0 on row 1 -> Col 2 on row 0
	prevRow, prevCol = PrevCell(table, 1, 0)
	if prevRow != 0 || prevCol != 2 {
		t.Errorf("expected (0, 2), got (%d, %d)", prevRow, prevCol)
	}

	// Case 3: Col 0 on row 0 -> stays at (0, 0)
	prevRow, prevCol = PrevCell(table, 0, 0)
	if prevRow != 0 || prevCol != 0 {
		t.Errorf("expected (0, 0), got (%d, %d)", prevRow, prevCol)
	}
}

// TC10
func TestNavigation_TabAutoCreateRow(t *testing.T) {
	lines := []string{
		"| Item | Preço |",
		"|---|---|",
		"| Maçã | 3.50 |",
	}

	buf := buffer.NewEmptyBuffer("doc.md")
	for i, l := range lines {
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(l, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(l, buffer.EndingLF))
		}
	}

	table, ok := DetectTable(lines, 0)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected")
	}

	nextRow, _, isNewRow := NextCell(table, 0, 1)
	if !isNewRow {
		t.Fatalf("expected isNewRow to be true")
	}

	// Insert row and apply to buffer
	table.InsertRow(nextRow - 1)
	ApplyTableToBuffer(buf, table)

	if len(table.Rows) != 2 {
		t.Errorf("expected 2 data rows, got %d", len(table.Rows))
	}
	if buf.LineCount() != 4 {
		t.Errorf("expected 4 lines in buffer, got %d", buf.LineCount())
	}

	lastLine, _ := buf.GetLine(3)
	// Last line should be empty cells with pipes: "|     |       |"
	if !tableIsRowPipes(lastLine.String()) {
		t.Errorf("last line in buffer is not a valid empty table row: %q", lastLine.String())
	}
}

func tableIsRowPipes(line string) bool {
	return len(splitCells(line)) >= 2
}

// TC13, TC14, TC15
func TestNavigation_InsertAndDeleteColumn(t *testing.T) {
	lines := []string{
		"| Col 0 | Col 1 |",
		"|---|---|",
		"| A | B |",
	}

	table, ok := DetectTable(lines, 0)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected")
	}

	// TC13: Insert column at right of col 0 -> new column at index 1
	table.InsertColumn(0, true)
	if table.ColumnCount() != 3 {
		t.Fatalf("expected 3 columns after insert right, got %d", table.ColumnCount())
	}
	if len(table.Header.Cells) != 3 || len(table.Rows[0].Cells) != 3 {
		t.Errorf("row cells count mismatch after insert column right")
	}

	// TC14: Insert column at left of col 0 -> new column at index 0
	table.InsertColumn(0, false)
	if table.ColumnCount() != 4 {
		t.Fatalf("expected 4 columns after insert left, got %d", table.ColumnCount())
	}

	// TC15: Delete column at index 1
	table.DeleteColumn(1)
	if table.ColumnCount() != 3 {
		t.Fatalf("expected 3 columns after delete, got %d", table.ColumnCount())
	}

	// Cursor offset calculation
	offset0 := GetCellCursorCol(table, 0)
	offset1 := GetCellCursorCol(table, 1)
	if offset0 != 2 {
		t.Errorf("offset0 expected 2, got %d", offset0)
	}
	if offset1 <= offset0 {
		t.Errorf("offset1 should be greater than offset0, got %d vs %d", offset1, offset0)
	}
}
