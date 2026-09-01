package ui

import (
	"strings"
	"testing"

	"md-notes/internal/theme"
)

func getTestTheme() *theme.CompiledTheme {
	palette, _ := theme.GetPalette("default-dark")
	return theme.CompileTheme(palette, theme.ColorOverrides{})
}

// TC03
func TestGutter_AbsoluteNumbering(t *testing.T) {
	th := getTestTheme()

	// 10 lines in file, cursor at line 4 (line 5 in 1-based)
	totalLines := 10
	cursorLine := 4

	for i := 0; i < 10; i++ {
		rendered := RenderGutterLine(i, cursorLine, totalLines, false, th)
		expectedNum := strings.TrimSpace(rendered)
		expectedStr := strconvFormat(i + 1)
		if expectedNum != expectedStr {
			t.Errorf("line %d: expected gutter number %q, got %q (raw: %q)", i, expectedStr, expectedNum, rendered)
		}
	}
}

// TC04
func TestGutter_RelativeNumbering(t *testing.T) {
	th := getTestTheme()

	// Cursor on line 4 (0-indexed -> line 5)
	totalLines := 20
	cursorLine := 4

	// Line 4 (current) should display "5"
	currLine := RenderGutterLine(4, cursorLine, totalLines, true, th)
	if strings.TrimSpace(currLine) != "5" {
		t.Errorf("current line expected '5', got %q", strings.TrimSpace(currLine))
	}

	// Line 3 (distance 1 above) should display "1"
	line3 := RenderGutterLine(3, cursorLine, totalLines, true, th)
	if strings.TrimSpace(line3) != "1" {
		t.Errorf("line 3 expected '1', got %q", strings.TrimSpace(line3))
	}

	// Line 2 (distance 2 above) should display "2"
	line2 := RenderGutterLine(2, cursorLine, totalLines, true, th)
	if strings.TrimSpace(line2) != "2" {
		t.Errorf("line 2 expected '2', got %q", strings.TrimSpace(line2))
	}

	// Line 5 (distance 1 below) should display "1"
	line5 := RenderGutterLine(5, cursorLine, totalLines, true, th)
	if strings.TrimSpace(line5) != "1" {
		t.Errorf("line 5 expected '1', got %q", strings.TrimSpace(line5))
	}

	// Line 6 (distance 2 below) should display "2"
	line6 := RenderGutterLine(6, cursorLine, totalLines, true, th)
	if strings.TrimSpace(line6) != "2" {
		t.Errorf("line 6 expected '2', got %q", strings.TrimSpace(line6))
	}
}

// TC05
func TestGutter_DynamicWidth(t *testing.T) {
	// 5 lines -> digits 1 + 2 = 3 -> min 4
	if w := CalculateGutterWidth(5); w != 4 {
		t.Errorf("5 lines expected width 4, got %d", w)
	}

	// 1000 lines -> digits 4 + 2 = 6
	if w := CalculateGutterWidth(1000); w != 6 {
		t.Errorf("1000 lines expected width 6, got %d", w)
	}

	// 12500 lines -> digits 5 + 2 = 7
	if w := CalculateGutterWidth(12500); w != 7 {
		t.Errorf("12500 lines expected width 7, got %d", w)
	}
}

func strconvFormat(n int) string {
	if n < 10 {
		return string('0' + rune(n))
	}
	return "10"
}
