package buffer

import "strings"

// LineEnding defines the newline character sequence for a line.
type LineEnding string

const (
	EndingLF   LineEnding = "\n"
	EndingCRLF LineEnding = "\r\n"
	EndingNone LineEnding = ""
)

// Line represents a single line of text stored as runes to properly support UTF-8 characters.
type Line struct {
	Runes  []rune
	Ending LineEnding
}

// NewLine creates a Line instance from a string content and line ending.
func NewLine(content string, ending LineEnding) Line {
	// Strip trailing newline characters from content if present
	clean := strings.TrimRight(content, "\r\n")
	return Line{
		Runes:  []rune(clean),
		Ending: ending,
	}
}

// Length returns the number of Unicode runes in the line.
func (l *Line) Length() int {
	return len(l.Runes)
}

// String returns the string representation of the line without its ending.
func (l *Line) String() string {
	return string(l.Runes)
}

// RawString returns the string representation of the line including its line ending.
func (l *Line) RawString() string {
	return string(l.Runes) + string(l.Ending)
}

// Bytes returns the UTF-8 byte representation of the line including its line ending.
func (l *Line) Bytes() []byte {
	return []byte(l.RawString())
}

// InsertAt inserts a rune at the specified column index (0-indexed).
// If col is beyond length, it appends at the end.
func (l *Line) InsertAt(col int, r rune) {
	if col < 0 {
		col = 0
	}
	if col >= len(l.Runes) {
		l.Runes = append(l.Runes, r)
		return
	}
	l.Runes = append(l.Runes[:col], append([]rune{r}, l.Runes[col:]...)...)
}

// DeleteAt removes and returns the rune at the specified column index (0-indexed).
// If col is invalid, it returns 0.
func (l *Line) DeleteAt(col int) rune {
	if col < 0 || col >= len(l.Runes) {
		return 0
	}
	r := l.Runes[col]
	l.Runes = append(l.Runes[:col], l.Runes[col+1:]...)
	return r
}

// SplitAt splits the line at the specified column index, returning a new Line with the remainder.
func (l *Line) SplitAt(col int) Line {
	if col < 0 {
		col = 0
	}
	if col > len(l.Runes) {
		col = len(l.Runes)
	}

	remainder := l.Runes[col:]
	l.Runes = l.Runes[:col]

	newLine := Line{
		Runes:  make([]rune, len(remainder)),
		Ending: l.Ending,
	}
	copy(newLine.Runes, remainder)
	return newLine
}
