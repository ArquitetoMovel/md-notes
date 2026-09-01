package buffer

import (
	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrLineIndexOutOfBounds = errors.New("line index out of bounds")
)

// Cursor represents the cursor coordinate within the buffer (0-indexed internally).
type Cursor struct {
	Line int
	Col  int
}

// Buffer holds the text lines in memory and tracks state metadata.
type Buffer struct {
	mu              sync.RWMutex
	Lines           []Line
	Cursor          Cursor
	FilePath        string
	IsDirty         bool
	IsNewFile       bool
	IsStdinBuffer   bool
	OriginalModTime time.Time
	OriginalMode    os.FileMode
}

// NewEmptyBuffer creates a new buffer associated with a file path.
func NewEmptyBuffer(filePath string) *Buffer {
	return &Buffer{
		Lines: []Line{
			NewLine("", EndingLF),
		},
		Cursor:          Cursor{Line: 0, Col: 0},
		FilePath:        filePath,
		IsDirty:         false,
		IsNewFile:       true,
		IsStdinBuffer:   false,
		OriginalModTime: time.Now(),
		OriginalMode:    0644,
	}
}

// NewScratchpadBuffer creates a new unnamed buffer for quick notes.
func NewScratchpadBuffer() *Buffer {
	return &Buffer{
		Lines: []Line{
			NewLine("", EndingLF),
		},
		Cursor:          Cursor{Line: 0, Col: 0},
		FilePath:        "",
		IsDirty:         false,
		IsNewFile:       true,
		IsStdinBuffer:   false,
		OriginalModTime: time.Now(),
		OriginalMode:    0644,
	}
}

// LineCount returns the number of lines currently in the buffer.
func (b *Buffer) LineCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.Lines)
}

// GetLine retrieves a copy of the line at the given 0-indexed position.
func (b *Buffer) GetLine(index int) (Line, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if index < 0 || index >= len(b.Lines) {
		return Line{}, ErrLineIndexOutOfBounds
	}
	return b.Lines[index], nil
}

// SetLine replaces the line at the given 0-indexed position.
func (b *Buffer) SetLine(index int, line Line) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if index < 0 || index >= len(b.Lines) {
		return ErrLineIndexOutOfBounds
	}
	b.Lines[index] = line
	b.IsDirty = true
	return nil
}

// InsertLine inserts a line at the given index.
func (b *Buffer) InsertLine(index int, line Line) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if index < 0 {
		index = 0
	}
	if index >= len(b.Lines) {
		b.Lines = append(b.Lines, line)
	} else {
		b.Lines = append(b.Lines[:index], append([]Line{line}, b.Lines[index:]...)...)
	}
	b.IsDirty = true
}

// DeleteLine removes and returns the line at the given index.
func (b *Buffer) DeleteLine(index int) (Line, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if index < 0 || index >= len(b.Lines) {
		return Line{}, ErrLineIndexOutOfBounds
	}
	deleted := b.Lines[index]
	b.Lines = append(b.Lines[:index], b.Lines[index+1:]...)
	if len(b.Lines) == 0 {
		b.Lines = []Line{NewLine("", EndingLF)}
	}
	b.IsDirty = true
	return deleted, nil
}

// SetDirty manually sets the dirty flag.
func (b *Buffer) SetDirty(dirty bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.IsDirty = dirty
}

// SetFilePath updates the file path associated with the buffer.
func (b *Buffer) SetFilePath(path string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.FilePath = path
	b.IsStdinBuffer = false
}

// InsertRune inserts a rune at the current cursor position.
func (b *Buffer) InsertRune(r rune) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Cursor.Line >= len(b.Lines) {
		b.Cursor.Line = len(b.Lines) - 1
	}
	if b.Cursor.Line < 0 {
		b.Lines = []Line{NewLine("", EndingLF)}
		b.Cursor.Line = 0
		b.Cursor.Col = 0
	}

	b.Lines[b.Cursor.Line].InsertAt(b.Cursor.Col, r)
	b.Cursor.Col++
	b.IsDirty = true
}

// DeleteRune deletes the rune before the cursor (backspace behavior).
func (b *Buffer) DeleteRune() rune {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Cursor.Line < 0 || b.Cursor.Line >= len(b.Lines) {
		return 0
	}

	if b.Cursor.Col > 0 {
		r := b.Lines[b.Cursor.Line].DeleteAt(b.Cursor.Col - 1)
		b.Cursor.Col--
		b.IsDirty = true
		return r
	}

	// Join with previous line if at beginning of line
	if b.Cursor.Line > 0 {
		prevIdx := b.Cursor.Line - 1
		currLine := b.Lines[b.Cursor.Line]
		prevCol := len(b.Lines[prevIdx].Runes)

		b.Lines[prevIdx].Runes = append(b.Lines[prevIdx].Runes, currLine.Runes...)
		b.Lines = append(b.Lines[:b.Cursor.Line], b.Lines[b.Cursor.Line+1:]...)

		b.Cursor.Line = prevIdx
		b.Cursor.Col = prevCol
		b.IsDirty = true
		return '\n'
	}

	return 0
}

// InsertNewLine splits the line at cursor and inserts a newline.
func (b *Buffer) InsertNewLine() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Cursor.Line < 0 || b.Cursor.Line >= len(b.Lines) {
		b.Lines = append(b.Lines, NewLine("", EndingLF))
		b.Cursor.Line = len(b.Lines) - 1
		b.Cursor.Col = 0
		b.IsDirty = true
		return
	}

	newLine := b.Lines[b.Cursor.Line].SplitAt(b.Cursor.Col)
	b.Lines = append(b.Lines[:b.Cursor.Line+1], append([]Line{newLine}, b.Lines[b.Cursor.Line+1:]...)...)
	b.Cursor.Line++
	b.Cursor.Col = 0
	b.IsDirty = true
}

// InsertText inserts arbitrary text at the current cursor position.
func (b *Buffer) InsertText(text string) {
	for _, r := range text {
		if r == '\n' {
			b.InsertNewLine()
		} else if r != '\r' {
			b.InsertRune(r)
		}
	}
}

// RawText returns the entire buffer content serialized as a single string.
func (b *Buffer) RawText() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var sb strings.Builder
	for _, l := range b.Lines {
		sb.WriteString(l.RawString())
	}
	return sb.String()
}
