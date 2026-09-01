package table

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// CalculateColumnWidths computes the maximum visual width of each column in the table.
func CalculateColumnWidths(t *Table) []int {
	cols := t.ColumnCount()
	if cols == 0 {
		return nil
	}

	widths := make([]int, cols)

	// Check header
	for i, c := range t.Header.Cells {
		if i < cols {
			w := runewidth.StringWidth(c.Text)
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	// Check data rows
	for _, row := range t.Rows {
		for i, c := range row.Cells {
			if i < cols {
				w := runewidth.StringWidth(c.Text)
				if w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	// Ensure minimum width of 3 for valid delimiter hyphens
	for i := range widths {
		if widths[i] < 3 {
			widths[i] = 3
		}
	}

	return widths
}

// padCell aligns and pads cell content to the target width based on column Alignment.
func padCell(text string, width int, align Alignment) string {
	cellW := runewidth.StringWidth(text)
	if cellW >= width {
		return text
	}
	diff := width - cellW
	switch align {
	case AlignRight:
		return strings.Repeat(" ", diff) + text
	case AlignCenter:
		leftPad := diff / 2
		rightPad := diff - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	case AlignLeft:
		fallthrough
	default:
		return text + strings.Repeat(" ", diff)
	}
}

// formatDelimiterCell generates the delimiter string (e.g. :---:, ---:, ---) for a column.
func formatDelimiterCell(width int, align Alignment) string {
	if width < 3 {
		width = 3
	}
	switch align {
	case AlignCenter:
		hyphens := width - 2
		if hyphens < 1 {
			hyphens = 1
		}
		return ":" + strings.Repeat("-", hyphens) + ":"
	case AlignRight:
		hyphens := width - 1
		if hyphens < 1 {
			hyphens = 1
		}
		return strings.Repeat("-", hyphens) + ":"
	case AlignLeft:
		fallthrough
	default:
		return strings.Repeat("-", width)
	}
}

// FormatMarkdownTable formats the entire Table into clean, padded GFM Markdown lines.
func FormatMarkdownTable(t *Table) []string {
	cols := t.ColumnCount()
	if cols == 0 {
		return nil
	}

	widths := CalculateColumnWidths(t)
	t.ColWidths = widths

	var result []string

	// 1. Format Header
	var headerCells []string
	for i := 0; i < cols; i++ {
		text := ""
		if i < len(t.Header.Cells) {
			text = t.Header.Cells[i].Text
		}
		align := AlignLeft
		if i < len(t.Alignments) {
			align = t.Alignments[i]
		}
		headerCells = append(headerCells, padCell(text, widths[i], align))
	}
	result = append(result, "| "+strings.Join(headerCells, " | ")+" |")

	// 2. Format Delimiter Row
	var delimCells []string
	for i := 0; i < cols; i++ {
		align := AlignLeft
		if i < len(t.Alignments) {
			align = t.Alignments[i]
		}
		delimCells = append(delimCells, formatDelimiterCell(widths[i], align))
	}
	result = append(result, "| "+strings.Join(delimCells, " | ")+" |")

	// 3. Format Data Rows
	for _, row := range t.Rows {
		var rowCells []string
		for i := 0; i < cols; i++ {
			text := ""
			if i < len(row.Cells) {
				text = row.Cells[i].Text
			}
			align := AlignLeft
			if i < len(t.Alignments) {
				align = t.Alignments[i]
			}
			rowCells = append(rowCells, padCell(text, widths[i], align))
		}
		result = append(result, "| "+strings.Join(rowCells, " | ")+" |")
	}

	return result
}

// FormatGFM returns the clean GFM Markdown formatted lines for the table.
func (t *Table) FormatGFM() []string {
	return FormatMarkdownTable(t)
}
