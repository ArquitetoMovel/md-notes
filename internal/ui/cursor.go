package ui

import (
	"fmt"

	"md-notes/internal/vim"
)

// Standard DECSCUSR ANSI escape sequences for terminal cursor shapes
const (
	CursorBlock      = "\x1b[2 q" // Solid block (Normal / Command)
	CursorBeam       = "\x1b[6 q" // Vertical bar (Insert)
	CursorUnderline  = "\x1b[4 q" // Underline (Visual modes)
	CursorRestore    = "\x1b[0 q" // Restore user terminal default cursor
)

// GetCursorShapeSequence returns the ANSI escape sequence to set the cursor shape for the given Vim mode.
func GetCursorShapeSequence(mode vim.Mode) string {
	switch mode {
	case vim.ModeInsert:
		return CursorBeam
	case vim.ModeVisual, vim.ModeVisualLine, vim.ModeVisualBlock:
		return CursorUnderline
	case vim.ModeNormal, vim.ModeCommand, vim.ModeSearch:
		fallthrough
	default:
		return CursorBlock
	}
}

// GetRestoreCursorSequence returns the ANSI escape sequence to restore the default cursor.
func GetRestoreCursorSequence() string {
	return CursorRestore
}

// GetCursorPositionSequence returns the ANSI escape sequence to move the terminal hardware cursor to (screenY, screenX) (0-indexed).
func GetCursorPositionSequence(screenY, screenX int) string {
	if screenY < 0 {
		screenY = 0
	}
	if screenX < 0 {
		screenX = 0
	}
	return fmt.Sprintf("\x1b[%d;%dH", screenY+1, screenX+1)
}

