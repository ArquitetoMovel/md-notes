package table

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// Alignment defines the text alignment for a table column.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

func (a Alignment) String() string {
	switch a {
	case AlignLeft:
		return "left"
	case AlignCenter:
		return "center"
	case AlignRight:
		return "right"
	default:
		return "left"
	}
}

// Cell represents an individual cell within a table row.
type Cell struct {
	Text    string // Trimmed text content of the cell
	Width   int    // Visual terminal column width
	RawText string // Untrimmed raw cell text
}

// NewCell creates a Cell and computes its visual width using go-runewidth.
func NewCell(raw string) Cell {
	trimmed := strings.TrimSpace(raw)
	return Cell{
		Text:    trimmed,
		Width:   runewidth.StringWidth(trimmed),
		RawText: raw,
	}
}

// Row represents a row in a table containing multiple cells.
type Row struct {
	Cells       []Cell
	IsHeader    bool
	IsDelimiter bool
}

// Table represents a parsed Markdown table block and its metadata.
type Table struct {
	StartLine  int
	EndLine    int
	Alignments []Alignment
	Header     Row
	Rows       []Row
	ColWidths  []int
}

// ColumnCount returns the number of columns in the table.
func (t *Table) ColumnCount() int {
	count := len(t.Header.Cells)
	if len(t.Alignments) > count {
		count = len(t.Alignments)
	}
	for _, r := range t.Rows {
		if len(r.Cells) > count {
			count = len(r.Cells)
		}
	}
	return count
}
