package ui

import (
	"md-notes/internal/theme"
)

// RenderScrollbar calculates and returns a vertical scrollbar slice of characters.
func RenderScrollbar(totalLines, viewportHeight, topIndex int, th *theme.CompiledTheme) []string {
	if viewportHeight <= 0 {
		return nil
	}

	result := make([]string, viewportHeight)
	if totalLines <= viewportHeight {
		// Document fits entirely within viewport
		for i := 0; i < viewportHeight; i++ {
			result[i] = " "
		}
		return result
	}

	// Calculate thumb height
	thumbHeight := (viewportHeight * viewportHeight) / totalLines
	if thumbHeight < 1 {
		thumbHeight = 1
	}
	if thumbHeight > viewportHeight {
		thumbHeight = viewportHeight
	}

	// Calculate thumb top position
	maxScroll := totalLines - viewportHeight
	if maxScroll < 1 {
		maxScroll = 1
	}
	maxTrack := viewportHeight - thumbHeight
	if maxTrack < 0 {
		maxTrack = 0
	}

	thumbTop := (topIndex * maxTrack) / maxScroll
	if thumbTop < 0 {
		thumbTop = 0
	}
	if thumbTop+thumbHeight > viewportHeight {
		thumbTop = viewportHeight - thumbHeight
	}

	trackChar := "│"
	thumbChar := "█"

	trackStyle := ""
	thumbStyle := ""
	if th != nil {
		trackChar = th.ScrollbarTrack.Render("│")
		thumbChar = th.ScrollbarThumb.Render("█")
	}
	_ = trackStyle
	_ = thumbStyle

	for i := 0; i < viewportHeight; i++ {
		if i >= thumbTop && i < thumbTop+thumbHeight {
			result[i] = thumbChar
		} else {
			result[i] = trackChar
		}
	}

	return result
}
