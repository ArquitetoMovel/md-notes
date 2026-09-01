package vim

// Mode represents the editing mode in Vim emulation.
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisualChar
	ModeVisualLine
	ModeVisualBlock
	ModeCommand
)

// String returns the user-facing string representation of the mode.
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeVisualChar:
		return "VISUAL"
	case ModeVisualLine:
		return "VISUAL LINE"
	case ModeVisualBlock:
		return "VISUAL BLOCK"
	case ModeCommand:
		return "COMMAND"
	default:
		return "NORMAL"
	}
}

// Position represents a 0-indexed coordinate in the buffer (Line, Col).
type Position struct {
	Line int
	Col  int
}

// Selection holds the start and end coordinates of a visual selection.
type Selection struct {
	Start Position
	End   Position
	Type  Mode
}

// Normalized returns the bounding coordinates sorted from top-left to bottom-right.
func (s *Selection) Normalized() (start Position, end Position) {
	if s.Start.Line < s.End.Line || (s.Start.Line == s.End.Line && s.Start.Col <= s.End.Col) {
		return s.Start, s.End
	}
	return s.End, s.Start
}

// Register stores copied or deleted text with its orientation.
type Register struct {
	Content    string
	IsLineWise bool
}

// ModalState represents the exposed modal engine state for viewport and status bar rendering.
type ModalState struct {
	CurrentMode  Mode
	Selection    *Selection
	CommandInput string
	StatusMsg    string
	Count        int
}
