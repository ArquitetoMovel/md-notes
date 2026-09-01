package ui

import (
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
	case vim.ModeNormal, vim.ModeCommand:
		fallthrough
	default:
		return CursorBlock
	}
}

// GetRestoreCursorSequence returns the ANSI escape sequence to restore the default cursor.
func GetRestoreCursorSequence() string {
	return CursorRestore
}
