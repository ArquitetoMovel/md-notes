package table

import (
	"md-notes/internal/buffer"
)

// NextCell calculates the target row and column index when navigating forward with Tab.
// If currently on the last cell of the last data row, isNewRow is set to true.
func NextCell(t *Table, curRow, curCol int) (nextRow, nextCol int, isNewRow bool) {
	cols := t.ColumnCount()
	if cols == 0 {
		return 0, 0, false
	}

	totalDataRows := len(t.Rows)
	if totalDataRows == 0 {
		// Only header exists
		if curCol+1 < cols {
			return 0, curCol + 1, false
		}
		return 0, 0, true
	}

	if curCol+1 < cols {
		return curRow, curCol + 1, false
	}

	// Move to next row
	if curRow+1 < totalDataRows {
		return curRow + 1, 0, false
	}

	// At end of table, signal creation of new row
	return totalDataRows, 0, true
}

// PrevCell calculates the target row and column index when navigating backward with Shift+Tab.
func PrevCell(t *Table, curRow, curCol int) (prevRow, prevCol int) {
	cols := t.ColumnCount()
	if cols == 0 {
		return 0, 0
	}

	if curCol > 0 {
		return curRow, curCol - 1
	}

	if curRow > 0 {
		return curRow - 1, cols - 1
	}

	// At the very first cell, stay at (0, 0)
	return 0, 0
}

// InsertRow adds a new empty data row after afterRowIdx.
func (t *Table) InsertRow(afterRowIdx int) {
	cols := t.ColumnCount()
	newCells := make([]Cell, cols)
	for i := 0; i < cols; i++ {
		newCells[i] = NewCell("")
	}
	newRow := Row{Cells: newCells}

	if afterRowIdx < 0 {
		// Insert at beginning of data rows
		t.Rows = append([]Row{newRow}, t.Rows...)
	} else if afterRowIdx >= len(t.Rows) {
		// Append at end of data rows
		t.Rows = append(t.Rows, newRow)
	} else {
		// Insert in between
		t.Rows = append(t.Rows[:afterRowIdx+1], append([]Row{newRow}, t.Rows[afterRowIdx+1:]...)...)
	}

	t.ColWidths = CalculateColumnWidths(t)
}

// InsertColumn inserts a new empty column in header, alignments, and all data rows.
func (t *Table) InsertColumn(colIdx int, atRight bool) {
	cols := t.ColumnCount()
	targetIdx := colIdx
	if atRight {
		targetIdx = colIdx + 1
	}
	if targetIdx < 0 {
		targetIdx = 0
	}
	if targetIdx > cols {
		targetIdx = cols
	}

	// 1. Header
	emptyCell := NewCell("")
	if targetIdx >= len(t.Header.Cells) {
		t.Header.Cells = append(t.Header.Cells, emptyCell)
	} else {
		t.Header.Cells = append(t.Header.Cells[:targetIdx], append([]Cell{emptyCell}, t.Header.Cells[targetIdx:]...)...)
	}

	// 2. Alignments
	if targetIdx >= len(t.Alignments) {
		t.Alignments = append(t.Alignments, AlignLeft)
	} else {
		t.Alignments = append(t.Alignments, AlignLeft)
	}

	// 3. Data rows
	for rIdx := range t.Rows {
		if targetIdx >= len(t.Rows[rIdx].Cells) {
			t.Rows[rIdx].Cells = append(t.Rows[rIdx].Cells, emptyCell)
		} else {
			t.Rows[rIdx].Cells = append(t.Rows[rIdx].Cells[:targetIdx], append([]Cell{emptyCell}, t.Rows[rIdx].Cells[targetIdx:]...)...)
		}
	}

	t.ColWidths = CalculateColumnWidths(t)
}

// DeleteColumn removes the column at colIdx from header, alignments, and all data rows.
func (t *Table) DeleteColumn(colIdx int) {
	cols := t.ColumnCount()
	if cols <= 1 || colIdx < 0 || colIdx >= cols {
		return
	}

	// 1. Header
	if colIdx < len(t.Header.Cells) {
		t.Header.Cells = append(t.Header.Cells[:colIdx], t.Header.Cells[colIdx+1:]...)
	}

	// 2. Alignments
	if colIdx < len(t.Alignments) {
		t.Alignments = append(t.Alignments[:colIdx], t.Alignments[colIdx+1:]...)
	}

	// 3. Data rows
	for rIdx := range t.Rows {
		if colIdx < len(t.Rows[rIdx].Cells) {
			t.Rows[rIdx].Cells = append(t.Rows[rIdx].Cells[:colIdx], t.Rows[rIdx].Cells[colIdx+1:]...)
		}
	}

	t.ColWidths = CalculateColumnWidths(t)
}

// GetCellCursorCol returns the horizontal character offset for the given cell column index.
func GetCellCursorCol(t *Table, colIdx int) int {
	if t == nil || colIdx < 0 {
		return 2
	}
	widths := t.ColWidths
	if len(widths) == 0 {
		widths = CalculateColumnWidths(t)
	}

	offset := 2 // starts after "| "
	for i := 0; i < colIdx && i < len(widths); i++ {
		offset += widths[i] + 3 // width + " | "
	}
	return offset
}

// ApplyTableToBuffer writes the formatted Markdown table lines back to the buffer.
func ApplyTableToBuffer(buf *buffer.Buffer, t *Table) {
	if buf == nil || t == nil {
		return
	}

	formatted := FormatMarkdownTable(t)
	oldLen := t.EndLine - t.StartLine + 1
	newLen := len(formatted)

	minLen := oldLen
	if newLen < minLen {
		minLen = newLen
	}

	// Replace existing overlapping lines
	for i := 0; i < minLen; i++ {
		lineIdx := t.StartLine + i
		newLine := buffer.NewLine(formatted[i], buffer.EndingLF)
		_ = buf.SetLine(lineIdx, newLine)
	}

	// If new table has more lines, insert them
	if newLen > oldLen {
		for i := oldLen; i < newLen; i++ {
			lineIdx := t.StartLine + i
			newLine := buffer.NewLine(formatted[i], buffer.EndingLF)
			buf.InsertLine(lineIdx, newLine)
		}
	}

	// If old table was longer, delete excess lines
	if oldLen > newLen {
		for i := 0; i < (oldLen - newLen); i++ {
			_, _ = buf.DeleteLine(t.StartLine + newLen)
		}
	}

	t.EndLine = t.StartLine + newLen - 1
}
