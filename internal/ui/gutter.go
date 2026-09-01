package ui

import (
	"fmt"
	"strconv"

	"md-notes/internal/theme"
)

// CalculateGutterWidth returns the number of columns required for the line number gutter.
func CalculateGutterWidth(totalLines int) int {
	if totalLines <= 0 {
		totalLines = 1
	}
	digits := len(strconv.Itoa(totalLines))
	// digits + 2 padding columns (1 leading, 1 trailing separator)
	width := digits + 2
	if width < 4 {
		width = 4
	}
	return width
}

// RenderGutterLine formats a single line number for the gutter.
func RenderGutterLine(lineIdx, cursorLine, totalLines int, relative bool, th *theme.CompiledTheme) string {
	width := CalculateGutterWidth(totalLines)
	numWidth := width - 2 // available width for digits

	var numStr string
	isCurrent := lineIdx == cursorLine

	if relative {
		if isCurrent {
			numStr = strconv.Itoa(lineIdx + 1)
		} else {
			dist := lineIdx - cursorLine
			if dist < 0 {
				dist = -dist
			}
			numStr = strconv.Itoa(dist)
		}
	} else {
		numStr = strconv.Itoa(lineIdx + 1)
	}

	formatted := fmt.Sprintf(" %*s ", numWidth, numStr)

	if th == nil {
		return formatted
	}

	if isCurrent {
		return th.GutterCurrent.Render(formatted)
	}
	return th.GutterNormal.Render(formatted)
}

// RenderEmptyGutterLine formats an empty gutter cell (e.g. for soft-wrapped continuation lines or ~ padding).
func RenderEmptyGutterLine(totalLines int, th *theme.CompiledTheme) string {
	width := CalculateGutterWidth(totalLines)
	empty := fmt.Sprintf("%*s", width, " ")
	if th != nil {
		return th.GutterNormal.Render(empty)
	}
	return empty
}
