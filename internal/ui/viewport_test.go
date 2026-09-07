package ui

import (
	"strings"
	"testing"

	"md-notes/internal/buffer"
)

// TC01
func TestViewport_VisibleLinesRendering(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	for i := 0; i < 100; i++ {
		lineText := strings.Repeat("a", 20)
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(lineText, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(lineText, buffer.EndingLF))
		}
	}

	vp := NewViewport(80, 20)
	vp.TopLine = 10

	visible := vp.GetVisibleLines(buf)

	if len(visible) != 20 {
		t.Fatalf("expected 20 visible lines, got %d", len(visible))
	}

	if visible[0].LogicalLine != 10 {
		t.Errorf("first visible line expected logical 10, got %d", visible[0].LogicalLine)
	}
	if visible[19].LogicalLine != 29 {
		t.Errorf("last visible line expected logical 29, got %d", visible[19].LogicalLine)
	}
}

// TC02
func TestViewport_ScrolloffMargin(t *testing.T) {
	vp := NewViewport(80, 20)
	vp.Scrolloff = 4
	vp.TopLine = 0

	totalLines := 100

	// Cursor at line 18 (near bottom of initial 0..19 window)
	// With height 20 and scrolloff 4, bottom limit is TopLine + 20 - 1 - 4 = 15.
	// Line 18 > 15 -> TopLine should scroll down to 18 - (20 - 1 - 4) = 3!
	vp.AdjustScroll(buffer.Cursor{Line: 18, Col: 0}, totalLines)

	if vp.TopLine < 3 {
		t.Errorf("expected TopLine >= 3 after scrolloff adjustment, got %d", vp.TopLine)
	}

	// Cursor back up at line 2
	// TopLine is 3. Cursor 2 < 3 + 4 = 7. TopLine scrolls up to 2 - 4 = -2 -> clamped to 0!
	vp.AdjustScroll(buffer.Cursor{Line: 2, Col: 0}, totalLines)
	if vp.TopLine != 0 {
		t.Errorf("expected TopLine 0 when moving cursor near top, got %d", vp.TopLine)
	}
}

// TC06
func TestViewport_SoftWrapLineBreaks(t *testing.T) {
	// 150 characters on 1 line
	lineStr := strings.Repeat("x", 150)
	line := buffer.NewLine(lineStr, buffer.EndingLF)

	// Content width 60
	chunks := WrapLogicalLine(0, line, 60)

	// 150 / 60 = 2 chunks of 60 + 1 chunk of 30 = 3 visual lines!
	if len(chunks) != 3 {
		t.Fatalf("expected 3 visual chunks, got %d", len(chunks))
	}

	if len(chunks[0].Text) != 60 || len(chunks[1].Text) != 60 || len(chunks[2].Text) != 30 {
		t.Errorf("chunk lengths mismatch: %d, %d, %d", len(chunks[0].Text), len(chunks[1].Text), len(chunks[2].Text))
	}

	// Buffer line should remain intact
	if line.Length() != 150 {
		t.Errorf("original buffer line modified: len %d", line.Length())
	}
}

// TC07
func TestViewport_CoordinateMapping(t *testing.T) {
	lineStr := strings.Repeat("x", 150)
	line := buffer.NewLine(lineStr, buffer.EndingLF)
	chunks := WrapLogicalLine(0, line, 60)

	gutterWidth := 4

	// Cursor at logical column 80 (falls in chunk 1, which spans cols 60..120)
	// Visual offset = 80 - 60 = 20. ScreenX = 20 + 4 = 24. ScreenY = 1.
	cursor := buffer.Cursor{Line: 0, Col: 80}
	coord := LogicalToVisual(cursor, chunks, gutterWidth)

	if coord.ScreenY != 1 {
		t.Errorf("expected ScreenY 1, got %d", coord.ScreenY)
	}
	if coord.ScreenX != 24 {
		t.Errorf("expected ScreenX 24, got %d", coord.ScreenX)
	}
}

func TestViewport_LargeDocumentWithSoftWrap_EndRendered(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	lineText := strings.Repeat("a", 150)
	for i := 0; i < 100; i++ {
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(lineText, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(lineText, buffer.EndingLF))
		}
	}

	vp := NewViewport(60, 20)
	vp.SoftWrap = true

	// Move cursor to line 99 using AdjustScrollWithBuffer
	vp.AdjustScrollWithBuffer(buffer.Cursor{Line: 99, Col: 0}, buf)

	visible := vp.GetVisibleLines(buf)
	found := false
	for _, vl := range visible {
		if vl.LogicalLine == 99 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Line 99 is NOT rendered! Visible lines only range from logical line %d to %d", visible[0].LogicalLine, visible[len(visible)-1].LogicalLine)
	}
}

func TestViewport_MouseScrollDownAndUp(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	for i := 0; i < 50; i++ {
		lineText := strings.Repeat("x", 20)
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(lineText, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(lineText, buffer.EndingLF))
		}
	}

	vp := NewViewport(80, 10)
	vp.TopLine = 0
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	// Scroll down 5 lines
	vp.ScrollDown(5, buf)
	if vp.TopLine != 5 {
		t.Errorf("expected TopLine 5 after ScrollDown(5), got %d", vp.TopLine)
	}
	// Cursor should have been brought into view (at least at TopLine)
	if buf.Cursor.Line < 5 {
		t.Errorf("expected Cursor.Line >= 5, got %d", buf.Cursor.Line)
	}

	// Scroll up 3 lines
	vp.ScrollUp(3, buf)
	if vp.TopLine != 2 {
		t.Errorf("expected TopLine 2 after ScrollUp(3), got %d", vp.TopLine)
	}
}
