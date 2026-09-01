package vim

import (
	"unicode"

	"md-notes/internal/buffer"
)

type charClass int

const (
	classSpace charClass = iota
	classWord
	classPunct
)

func getCharClass(r rune) charClass {
	if unicode.IsSpace(r) {
		return classSpace
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
		return classWord
	}
	return classPunct
}

// ClampCursor ensures the cursor is within the valid line and column bounds of the buffer.
func ClampCursor(buf *buffer.Buffer, isInsertMode bool) {
	if buf == nil || len(buf.Lines) == 0 {
		return
	}

	if buf.Cursor.Line < 0 {
		buf.Cursor.Line = 0
	}
	if buf.Cursor.Line >= len(buf.Lines) {
		buf.Cursor.Line = len(buf.Lines) - 1
	}

	lineLen := buf.Lines[buf.Cursor.Line].Length()
	maxCol := lineLen - 1
	if isInsertMode || maxCol < 0 {
		maxCol = lineLen
	}
	if maxCol < 0 {
		maxCol = 0
	}

	if buf.Cursor.Col < 0 {
		buf.Cursor.Col = 0
	}
	if buf.Cursor.Col > maxCol {
		buf.Cursor.Col = maxCol
	}
}

// MoveLeft moves the cursor to the left by count columns.
func MoveLeft(buf *buffer.Buffer, count int) {
	if count <= 0 {
		count = 1
	}
	buf.Cursor.Col -= count
	if buf.Cursor.Col < 0 {
		buf.Cursor.Col = 0
	}
}

// MoveRight moves the cursor to the right by count columns.
func MoveRight(buf *buffer.Buffer, count int, isInsertMode bool) {
	if count <= 0 {
		count = 1
	}
	if buf.Cursor.Line < 0 || buf.Cursor.Line >= len(buf.Lines) {
		return
	}

	lineLen := buf.Lines[buf.Cursor.Line].Length()
	maxCol := lineLen - 1
	if isInsertMode || maxCol < 0 {
		maxCol = lineLen
	}
	if maxCol < 0 {
		maxCol = 0
	}

	buf.Cursor.Col += count
	if buf.Cursor.Col > maxCol {
		buf.Cursor.Col = maxCol
	}
}

// MoveDown moves the cursor down by count lines.
func MoveDown(buf *buffer.Buffer, count int, isInsertMode bool) {
	if count <= 0 {
		count = 1
	}
	buf.Cursor.Line += count
	if buf.Cursor.Line >= len(buf.Lines) {
		buf.Cursor.Line = len(buf.Lines) - 1
	}
	ClampCursor(buf, isInsertMode)
}

// MoveUp moves the cursor up by count lines.
func MoveUp(buf *buffer.Buffer, count int, isInsertMode bool) {
	if count <= 0 {
		count = 1
	}
	buf.Cursor.Line -= count
	if buf.Cursor.Line < 0 {
		buf.Cursor.Line = 0
	}
	ClampCursor(buf, isInsertMode)
}

// LineStart moves the cursor to column 0.
func LineStart(buf *buffer.Buffer) {
	buf.Cursor.Col = 0
}

// LineStartNonBlank moves the cursor to the first non-whitespace character in the line.
func LineStartNonBlank(buf *buffer.Buffer) {
	if buf.Cursor.Line < 0 || buf.Cursor.Line >= len(buf.Lines) {
		buf.Cursor.Col = 0
		return
	}
	runes := buf.Lines[buf.Cursor.Line].Runes
	for i, r := range runes {
		if !unicode.IsSpace(r) {
			buf.Cursor.Col = i
			return
		}
	}
	if len(runes) > 0 {
		buf.Cursor.Col = len(runes) - 1
	} else {
		buf.Cursor.Col = 0
	}
}

// LineEnd moves the cursor to the last character of the current line.
func LineEnd(buf *buffer.Buffer, isInsertMode bool) {
	if buf.Cursor.Line < 0 || buf.Cursor.Line >= len(buf.Lines) {
		return
	}
	lineLen := buf.Lines[buf.Cursor.Line].Length()
	if isInsertMode {
		buf.Cursor.Col = lineLen
	} else if lineLen > 0 {
		buf.Cursor.Col = lineLen - 1
	} else {
		buf.Cursor.Col = 0
	}
}

// DocStart moves the cursor to the first line (line 0, col 0).
func DocStart(buf *buffer.Buffer, isInsertMode bool) {
	buf.Cursor.Line = 0
	buf.Cursor.Col = 0
	ClampCursor(buf, isInsertMode)
}

// DocEnd moves the cursor to the last line of the buffer.
func DocEnd(buf *buffer.Buffer, isInsertMode bool) {
	if len(buf.Lines) > 0 {
		buf.Cursor.Line = len(buf.Lines) - 1
	} else {
		buf.Cursor.Line = 0
	}
	buf.Cursor.Col = 0
	ClampCursor(buf, isInsertMode)
}

// NextWordStart moves the cursor to the start of the next word token count times.
func NextWordStart(buf *buffer.Buffer, count int) {
	if count <= 0 {
		count = 1
	}

	for c := 0; c < count; c++ {
		if buf.Cursor.Line >= len(buf.Lines) {
			return
		}

		line := buf.Cursor.Line
		col := buf.Cursor.Col
		runes := buf.Lines[line].Runes

		if col >= len(runes) {
			// At end of line, advance to next line
			if line+1 < len(buf.Lines) {
				buf.Cursor.Line++
				buf.Cursor.Col = 0
				// Skip leading whitespace on new line
				newLineRunes := buf.Lines[buf.Cursor.Line].Runes
				for buf.Cursor.Col < len(newLineRunes) && unicode.IsSpace(newLineRunes[buf.Cursor.Col]) {
					buf.Cursor.Col++
				}
			}
			continue
		}

		startClass := getCharClass(runes[col])

		if startClass != classSpace {
			// Skip current word / punct category
			for col < len(runes) && getCharClass(runes[col]) == startClass {
				col++
			}
		}

		// Skip following whitespace
		for col < len(runes) && unicode.IsSpace(runes[col]) {
			col++
		}

		if col < len(runes) {
			buf.Cursor.Col = col
		} else {
			// Cross into next line
			if line+1 < len(buf.Lines) {
				buf.Cursor.Line++
				buf.Cursor.Col = 0
				newLineRunes := buf.Lines[buf.Cursor.Line].Runes
				for buf.Cursor.Col < len(newLineRunes) && unicode.IsSpace(newLineRunes[buf.Cursor.Col]) {
					buf.Cursor.Col++
				}
			} else {
				buf.Cursor.Col = len(runes) - 1
				if buf.Cursor.Col < 0 {
					buf.Cursor.Col = 0
				}
			}
		}
	}
	ClampCursor(buf, false)
}

// PrevWordStart moves the cursor to the beginning of the previous word token count times.
func PrevWordStart(buf *buffer.Buffer, count int) {
	if count <= 0 {
		count = 1
	}

	for c := 0; c < count; c++ {
		line := buf.Cursor.Line
		col := buf.Cursor.Col

		if line < 0 || line >= len(buf.Lines) {
			return
		}

		runes := buf.Lines[line].Runes

		// If at column 0, jump to previous line end
		if col <= 0 {
			if line > 0 {
				buf.Cursor.Line--
				prevRunes := buf.Lines[buf.Cursor.Line].Runes
				buf.Cursor.Col = len(prevRunes)
				if buf.Cursor.Col > 0 {
					buf.Cursor.Col--
				}
			}
			continue
		}

		col--

		// Skip spaces going backwards
		for col > 0 && unicode.IsSpace(runes[col]) {
			col--
		}

		targetClass := getCharClass(runes[col])
		// Walk backwards to the start of the token
		for col > 0 && getCharClass(runes[col-1]) == targetClass {
			col--
		}

		buf.Cursor.Col = col
	}
	ClampCursor(buf, false)
}

// WordEnd moves the cursor to the end of the current or next word token.
func WordEnd(buf *buffer.Buffer, count int) {
	if count <= 0 {
		count = 1
	}

	for c := 0; c < count; c++ {
		if buf.Cursor.Line >= len(buf.Lines) {
			return
		}

		line := buf.Cursor.Line
		col := buf.Cursor.Col
		runes := buf.Lines[line].Runes

		if col+1 < len(runes) {
			col++
		} else if line+1 < len(buf.Lines) {
			line++
			buf.Cursor.Line = line
			col = 0
			runes = buf.Lines[line].Runes
		} else {
			return
		}

		// Skip spaces
		for col < len(runes) && unicode.IsSpace(runes[col]) {
			col++
		}

		if col >= len(runes) && line+1 < len(buf.Lines) {
			line++
			buf.Cursor.Line = line
			col = 0
			runes = buf.Lines[line].Runes
			for col < len(runes) && unicode.IsSpace(runes[col]) {
				col++
			}
		}

		if col < len(runes) {
			targetClass := getCharClass(runes[col])
			for col+1 < len(runes) && getCharClass(runes[col+1]) == targetClass {
				col++
			}
			buf.Cursor.Col = col
		}
	}
	ClampCursor(buf, false)
}

// FindCharForward moves to the count-th occurrence of target rune to the right on the current line.
func FindCharForward(buf *buffer.Buffer, target rune, count int) bool {
	if count <= 0 {
		count = 1
	}
	if buf.Cursor.Line < 0 || buf.Cursor.Line >= len(buf.Lines) {
		return false
	}

	runes := buf.Lines[buf.Cursor.Line].Runes
	matched := 0

	for i := buf.Cursor.Col + 1; i < len(runes); i++ {
		if runes[i] == target {
			matched++
			if matched == count {
				buf.Cursor.Col = i
				return true
			}
		}
	}
	return false
}

// FindCharBackward moves to the count-th occurrence of target rune to the left on the current line.
func FindCharBackward(buf *buffer.Buffer, target rune, count int) bool {
	if count <= 0 {
		count = 1
	}
	if buf.Cursor.Line < 0 || buf.Cursor.Line >= len(buf.Lines) {
		return false
	}

	runes := buf.Lines[buf.Cursor.Line].Runes
	matched := 0

	for i := buf.Cursor.Col - 1; i >= 0; i-- {
		if runes[i] == target {
			matched++
			if matched == count {
				buf.Cursor.Col = i
				return true
			}
		}
	}
	return false
}

// TillCharForward moves to before the count-th occurrence of target rune to the right.
func TillCharForward(buf *buffer.Buffer, target rune, count int) bool {
	if count <= 0 {
		count = 1
	}
	if buf.Cursor.Line < 0 || buf.Cursor.Line >= len(buf.Lines) {
		return false
	}

	runes := buf.Lines[buf.Cursor.Line].Runes
	matched := 0

	for i := buf.Cursor.Col + 1; i < len(runes); i++ {
		if runes[i] == target {
			matched++
			if matched == count {
				if i > 0 {
					buf.Cursor.Col = i - 1
				} else {
					buf.Cursor.Col = 0
				}
				return true
			}
		}
	}
	return false
}

// TillCharBackward moves to after the count-th occurrence of target rune to the left.
func TillCharBackward(buf *buffer.Buffer, target rune, count int) bool {
	if count <= 0 {
		count = 1
	}
	if buf.Cursor.Line < 0 || buf.Cursor.Line >= len(buf.Lines) {
		return false
	}

	runes := buf.Lines[buf.Cursor.Line].Runes
	matched := 0

	for i := buf.Cursor.Col - 1; i >= 0; i-- {
		if runes[i] == target {
			matched++
			if matched == count {
				if i+1 < len(runes) {
					buf.Cursor.Col = i + 1
				} else {
					buf.Cursor.Col = len(runes) - 1
				}
				return true
			}
		}
	}
	return false
}
