package table

import (
	"errors"
	"strings"
)

// IsTableLine checks whether a line has the basic structural markers of a table row.
func IsTableLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "|")
}

// splitCells splits a Markdown table row into individual raw cell strings.
func splitCells(line string) []string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}

	// Strip leading pipe if present
	if strings.HasPrefix(trimmed, "|") {
		trimmed = trimmed[1:]
	}
	// Strip trailing pipe if present (and not escaped)
	if strings.HasSuffix(trimmed, "|") && !strings.HasSuffix(trimmed, `\|`) {
		trimmed = trimmed[:len(trimmed)-1]
	}

	var cells []string
	var cur strings.Builder
	escaped := false

	for i := 0; i < len(trimmed); i++ {
		ch := trimmed[i]
		if escaped {
			cur.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			cur.WriteByte(ch)
			continue
		}
		if ch == '|' {
			cells = append(cells, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(ch)
	}
	cells = append(cells, cur.String())
	return cells
}

// ParseRow converts a table line into a Row structure.
func ParseRow(line string) Row {
	rawCells := splitCells(line)
	cells := make([]Cell, len(rawCells))
	for i, rc := range rawCells {
		cells[i] = NewCell(rc)
	}
	return Row{
		Cells: cells,
	}
}

// IsDelimiterLine checks if a line is a valid Markdown table delimiter row (e.g. |:---|:---:|---:|)
// and returns the extracted column alignments.
func IsDelimiterLine(line string) ([]Alignment, bool) {
	cells := splitCells(line)
	if len(cells) == 0 {
		return nil, false
	}

	alignments := make([]Alignment, len(cells))
	for i, c := range cells {
		t := strings.TrimSpace(c)
		if t == "" {
			return nil, false
		}
		hasHyphen := false
		validChars := true
		for _, r := range t {
			if r == '-' {
				hasHyphen = true
			} else if r == ':' || r == ' ' || r == '\t' {
				// allowed
			} else {
				validChars = false
				break
			}
		}
		if !hasHyphen || !validChars {
			return nil, false
		}

		if strings.HasPrefix(t, ":") && strings.HasSuffix(t, ":") {
			alignments[i] = AlignCenter
		} else if strings.HasSuffix(t, ":") {
			alignments[i] = AlignRight
		} else {
			alignments[i] = AlignLeft
		}
	}
	return alignments, true
}

// DetectTable detects if the given cursorLine is part of a Markdown table in the provided lines slice.
func DetectTable(lines []string, cursorLine int) (*Table, bool) {
	if cursorLine < 0 || cursorLine >= len(lines) {
		return nil, false
	}
	if !IsTableLine(lines[cursorLine]) {
		return nil, false
	}

	// Scan contiguous table lines around cursorLine
	start := cursorLine
	for start > 0 && IsTableLine(lines[start-1]) {
		start--
	}
	end := cursorLine
	for end+1 < len(lines) && IsTableLine(lines[end+1]) {
		end++
	}

	// Search for a delimiter line in this contiguous block
	for i := start; i <= end; i++ {
		alignments, isDelim := IsDelimiterLine(lines[i])
		if isDelim && i > start {
			// Found valid delimiter with header at i-1
			headerIdx := i - 1
			header := ParseRow(lines[headerIdx])
			header.IsHeader = true

			var dataRows []Row
			for k := i + 1; k <= end; k++ {
				dataRows = append(dataRows, ParseRow(lines[k]))
			}

			t := &Table{
				StartLine:  headerIdx,
				EndLine:    end,
				Alignments: alignments,
				Header:     header,
				Rows:       dataRows,
			}

			// Ensure alignments match column count
			cols := t.ColumnCount()
			for len(t.Alignments) < cols {
				t.Alignments = append(t.Alignments, AlignLeft)
			}

			t.ColWidths = CalculateColumnWidths(t)
			return t, true
		}
	}

	return nil, false
}

// ParseTable parses a Markdown table starting at startIdx.
func ParseTable(lines []string, startIdx int) (*Table, int, error) {
	if startIdx < 0 || startIdx >= len(lines) {
		return nil, startIdx, errors.New("startIdx out of range")
	}

	t, ok := DetectTable(lines, startIdx)
	if !ok {
		return nil, startIdx, errors.New("no table found at given index")
	}
	return t, t.EndLine, nil
}
