package table

import (
	"testing"
)

// TC01
func TestParser_DetectValidTable(t *testing.T) {
	lines := []string{
		"# Document Title",
		"",
		"| Col A | Col B | Col C |",
		"|:---|:---:|---:|",
		"| Data 1 | Data 2 | Data 3 |",
		"| 10 | 20 | 30 |",
		"",
		"Paragraph after table",
	}

	// Test cursor at header line (index 2)
	table, ok := DetectTable(lines, 2)
	if !ok || table == nil {
		t.Fatalf("expected table to be detected at line 2")
	}

	if table.StartLine != 2 {
		t.Errorf("expected StartLine 2, got %d", table.StartLine)
	}
	if table.EndLine != 5 {
		t.Errorf("expected EndLine 5, got %d", table.EndLine)
	}
	if table.ColumnCount() != 3 {
		t.Errorf("expected 3 columns, got %d", table.ColumnCount())
	}
	if len(table.Rows) != 2 {
		t.Errorf("expected 2 data rows, got %d", len(table.Rows))
	}

	// Check alignments: Left, Center, Right
	if len(table.Alignments) != 3 {
		t.Fatalf("expected 3 alignments, got %d", len(table.Alignments))
	}
	if table.Alignments[0] != AlignLeft {
		t.Errorf("column 0 expected AlignLeft, got %v", table.Alignments[0])
	}
	if table.Alignments[1] != AlignCenter {
		t.Errorf("column 1 expected AlignCenter, got %v", table.Alignments[1])
	}
	if table.Alignments[2] != AlignRight {
		t.Errorf("column 2 expected AlignRight, got %v", table.Alignments[2])
	}

	// Test cursor at data line (index 4)
	table2, ok2 := DetectTable(lines, 4)
	if !ok2 || table2 == nil {
		t.Fatalf("expected table to be detected at line 4")
	}
	if table2.StartLine != 2 || table2.EndLine != 5 {
		t.Errorf("detected table boundaries mismatch at line 4: got %d..%d", table2.StartLine, table2.EndLine)
	}
}

// TC02
func TestParser_IgnoreNonTablePipes(t *testing.T) {
	lines := []string{
		"Execute o comando1 | comando2 no terminal",
		"Outra linha normal sem tabela",
	}

	table, ok := DetectTable(lines, 0)
	if ok || table != nil {
		t.Errorf("expected non-table line to NOT be detected as table, got %+v", table)
	}

	// Single line with pipes without delimiter row
	singleLineWithPipes := []string{
		"| Not | A | Table |",
	}
	t2, ok2 := DetectTable(singleLineWithPipes, 0)
	if ok2 || t2 != nil {
		t.Errorf("expected pipe line without delimiter to NOT be detected as table")
	}
}

func TestParser_AlignmentsExtraction(t *testing.T) {
	tests := []struct {
		line    string
		want    []Alignment
		isValid bool
	}{
		{
			line:    "|:---|:---:|---:|",
			want:    []Alignment{AlignLeft, AlignCenter, AlignRight},
			isValid: true,
		},
		{
			line:    "|---|---|",
			want:    []Alignment{AlignLeft, AlignLeft},
			isValid: true,
		},
		{
			line:    "| :--- | ---: |",
			want:    []Alignment{AlignLeft, AlignRight},
			isValid: true,
		},
		{
			line:    "| invalid | row |",
			want:    nil,
			isValid: false,
		},
	}

	for _, tt := range tests {
		alignments, ok := IsDelimiterLine(tt.line)
		if ok != tt.isValid {
			t.Errorf("IsDelimiterLine(%q) validity = %v, want %v", tt.line, ok, tt.isValid)
			continue
		}
		if tt.isValid {
			if len(alignments) != len(tt.want) {
				t.Errorf("IsDelimiterLine(%q) len = %d, want %d", tt.line, len(alignments), len(tt.want))
				continue
			}
			for i := range alignments {
				if alignments[i] != tt.want[i] {
					t.Errorf("IsDelimiterLine(%q) col %d = %v, want %v", tt.line, i, alignments[i], tt.want[i])
				}
			}
		}
	}
}

func TestTable_AlignmentString(t *testing.T) {
	if AlignLeft.String() != "left" {
		t.Errorf("AlignLeft.String() = %q", AlignLeft.String())
	}
	if AlignCenter.String() != "center" {
		t.Errorf("AlignCenter.String() = %q", AlignCenter.String())
	}
	if AlignRight.String() != "right" {
		t.Errorf("AlignRight.String() = %q", AlignRight.String())
	}
	if Alignment(99).String() != "left" {
		t.Errorf("unknown Alignment.String() = %q", Alignment(99).String())
	}
}
