package ui

import (
	"md-notes/internal/buffer"
)

// VisualLine represents a single rendered line on screen (either an entire logical line or a soft-wrapped segment).
type VisualLine struct {
	LogicalLine int    // 0-indexed line index in the buffer
	WrapIndex   int    // 0-indexed wrap segment (0 = head, 1+ = wrapped continuation)
	Text        string // Text content of this visual line
	StartCol    int    // Starting rune offset in the logical line
	EndCol      int    // Ending rune offset in the logical line
}

// VisualCoordinate represents the visual screen position of a cursor.
type VisualCoordinate struct {
	ScreenY int // 0-indexed visual line row on screen
	ScreenX int // 0-indexed visual column on screen (including gutter)
}

// Viewport handles window dimensions, scrolling, soft-wrap, and visual line rendering.
type Viewport struct {
	Width       int
	Height      int
	TopLine     int
	Scrolloff   int
	RelativeNum bool
	SoftWrap    bool
}

// NewViewport creates a new Viewport with standard defaults.
func NewViewport(w, h int) *Viewport {
	return &Viewport{
		Width:       w,
		Height:      h,
		TopLine:     0,
		Scrolloff:   4,
		RelativeNum: false,
		SoftWrap:    true,
	}
}

// SetDimensions updates the viewport width and height.
func (v *Viewport) SetDimensions(w, h int) {
	v.Width = w
	v.Height = h
	if v.Height < 1 {
		v.Height = 1
	}
	if v.Width < 1 {
		v.Width = 1
	}
}

// AdjustScroll updates TopLine based on cursor position and scrolloff margin.
func (v *Viewport) AdjustScroll(cursor buffer.Cursor, totalLines int) {
	if totalLines <= 0 {
		v.TopLine = 0
		return
	}

	curLine := cursor.Line
	if curLine < 0 {
		curLine = 0
	}
	if curLine >= totalLines {
		curLine = totalLines - 1
	}

	scrolloff := v.Scrolloff
	if scrolloff*2 >= v.Height {
		scrolloff = v.Height / 4
	}

	// Scroll up if cursor is within top scrolloff margin
	if curLine < v.TopLine+scrolloff {
		v.TopLine = curLine - scrolloff
		if v.TopLine < 0 {
			v.TopLine = 0
		}
	}

	// Scroll down if cursor is within bottom scrolloff margin
	bottomLimit := v.TopLine + v.Height - 1 - scrolloff
	if curLine > bottomLimit {
		v.TopLine = curLine - (v.Height - 1 - scrolloff)
		maxTop := totalLines - 1
		if v.TopLine > maxTop {
			v.TopLine = maxTop
		}
		if v.TopLine < 0 {
			v.TopLine = 0
		}
	}
}

// WrapLogicalLine breaks a single buffer line into VisualLine segments fitting within contentWidth.
func WrapLogicalLine(logicalIdx int, line buffer.Line, contentWidth int) []VisualLine {
	runes := line.Runes
	if len(runes) == 0 {
		return []VisualLine{
			{
				LogicalLine: logicalIdx,
				WrapIndex:   0,
				Text:        "",
				StartCol:    0,
				EndCol:      0,
			},
		}
	}

	if contentWidth <= 0 {
		contentWidth = 80
	}

	var visualLines []VisualLine
	totalRunes := len(runes)
	wrapIdx := 0

	for start := 0; start < totalRunes; start += contentWidth {
		end := start + contentWidth
		if end > totalRunes {
			end = totalRunes
		}
		visualLines = append(visualLines, VisualLine{
			LogicalLine: logicalIdx,
			WrapIndex:   wrapIdx,
			Text:        string(runes[start:end]),
			StartCol:    start,
			EndCol:      end,
		})
		wrapIdx++
	}

	return visualLines
}

// LogicalToVisual maps logical buffer coordinates (Line, Col) to screen coordinates (ScreenY, ScreenX).
func LogicalToVisual(cursor buffer.Cursor, visualLines []VisualLine, gutterWidth int) VisualCoordinate {
	for screenY, vLine := range visualLines {
		if vLine.LogicalLine == cursor.Line {
			if cursor.Col >= vLine.StartCol && (cursor.Col < vLine.EndCol || (cursor.Col == vLine.EndCol && vLine.WrapIndex == 0)) {
				screenX := (cursor.Col - vLine.StartCol) + gutterWidth
				return VisualCoordinate{ScreenY: screenY, ScreenX: screenX}
			}
		}
	}
	return VisualCoordinate{ScreenY: 0, ScreenX: gutterWidth}
}

// GetVisibleLines computes the visual lines to render for the current viewport view.
func (v *Viewport) GetVisibleLines(buf *buffer.Buffer) []VisualLine {
	if buf == nil || buf.LineCount() == 0 {
		return nil
	}

	totalLines := buf.LineCount()
	gutterW := CalculateGutterWidth(totalLines)
	// Viewport width minus gutter and scrollbar (1 column)
	contentW := v.Width - gutterW - 1
	if contentW < 10 {
		contentW = 10
	}

	var allVisualLines []VisualLine
	for i := 0; i < totalLines; i++ {
		line, _ := buf.GetLine(i)
		if v.SoftWrap {
			chunks := WrapLogicalLine(i, line, contentW)
			allVisualLines = append(allVisualLines, chunks...)
		} else {
			raw := line.String()
			allVisualLines = append(allVisualLines, VisualLine{
				LogicalLine: i,
				WrapIndex:   0,
				Text:        raw,
				StartCol:    0,
				EndCol:      len(line.Runes),
			})
		}
	}

	// Slice from TopLine up to Height
	start := v.TopLine
	if start < 0 {
		start = 0
	}
	if start >= len(allVisualLines) {
		start = len(allVisualLines) - 1
		if start < 0 {
			start = 0
		}
	}

	end := start + v.Height
	if end > len(allVisualLines) {
		end = len(allVisualLines)
	}

	return allVisualLines[start:end]
}
