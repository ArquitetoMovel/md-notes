package ui

import (
	"strings"
	"testing"
)

// TC15
func TestScrollbar_Render(t *testing.T) {
	th := getTestTheme()

	// Case 1: totalLines <= height -> empty track
	sb1 := RenderScrollbar(10, 20, 0, th)
	if len(sb1) != 20 {
		t.Fatalf("expected 20 rows, got %d", len(sb1))
	}
	for i, r := range sb1 {
		if strings.TrimSpace(r) != "" {
			t.Errorf("row %d expected space, got %q", i, r)
		}
	}

	// Case 2: totalLines = 200, height = 20, topIndex = 100 (middle)
	sb2 := RenderScrollbar(200, 20, 100, th)
	if len(sb2) != 20 {
		t.Fatalf("expected 20 rows, got %d", len(sb2))
	}

	hasThumb := false
	hasTrack := false
	for _, r := range sb2 {
		if strings.Contains(r, "█") {
			hasThumb = true
		}
		if strings.Contains(r, "│") {
			hasTrack = true
		}
	}

	if !hasThumb {
		t.Errorf("scrollbar missing thumb '█'")
	}
	if !hasTrack {
		t.Errorf("scrollbar missing track '│'")
	}
}
