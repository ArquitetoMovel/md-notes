package ui

import (
	"testing"

	"md-notes/internal/vim"
)

// TC16, TC18
func TestCursor_ANSISequences(t *testing.T) {
	// ModeNormal -> Block (\x1b[2 q)
	if seq := GetCursorShapeSequence(vim.ModeNormal); seq != "\x1b[2 q" {
		t.Errorf("ModeNormal expected %q, got %q", "\x1b[2 q", seq)
	}

	// ModeInsert -> Beam (\x1b[6 q)
	if seq := GetCursorShapeSequence(vim.ModeInsert); seq != "\x1b[6 q" {
		t.Errorf("ModeInsert expected %q, got %q", "\x1b[6 q", seq)
	}

	// ModeVisual -> Underline (\x1b[4 q)
	if seq := GetCursorShapeSequence(vim.ModeVisual); seq != "\x1b[4 q" {
		t.Errorf("ModeVisual expected %q, got %q", "\x1b[4 q", seq)
	}

	// Restore -> \x1b[0 q
	if seq := GetRestoreCursorSequence(); seq != "\x1b[0 q" {
		t.Errorf("Restore expected %q, got %q", "\x1b[0 q", seq)
	}

	// Position sequence (1-indexed for terminal: row=3, col=5 -> \x1b[4;6H)
	if pos := GetCursorPositionSequence(3, 5); pos != "\x1b[4;6H" {
		t.Errorf("Position expected %q, got %q", "\x1b[4;6H", pos)
	}
}
