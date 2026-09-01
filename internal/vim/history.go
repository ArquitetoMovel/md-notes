package vim

import (
	"md-notes/internal/buffer"
)

// Snapshot represents an immutable state of the buffer for undo/redo history.
type Snapshot struct {
	Lines    []buffer.Line
	Cursor   buffer.Cursor
	IsDirty  bool
	FilePath string
}

// CloneLines performs a deep copy of a slice of buffer.Line instances.
func CloneLines(lines []buffer.Line) []buffer.Line {
	cloned := make([]buffer.Line, len(lines))
	for i, l := range lines {
		runesCopy := make([]rune, len(l.Runes))
		copy(runesCopy, l.Runes)
		cloned[i] = buffer.Line{
			Runes:  runesCopy,
			Ending: l.Ending,
		}
	}
	return cloned
}

// History manages the circular undo/redo snapshot stack.
type History struct {
	snapshots []Snapshot
	current   int
	maxSize   int
}

// NewHistory creates a new History manager with the given maximum size (default 200).
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &History{
		snapshots: make([]Snapshot, 0, maxSize),
		current:   -1,
		maxSize:   maxSize,
	}
}

// Push records a new snapshot of the buffer into history, truncating redo history.
func (h *History) Push(buf *buffer.Buffer) {
	if buf == nil {
		return
	}

	snap := Snapshot{
		Lines:    CloneLines(buf.Lines),
		Cursor:   buf.Cursor,
		IsDirty:  buf.IsDirty,
		FilePath: buf.FilePath,
	}

	// Truncate any redo branch if we pushed after an undo
	if h.current >= 0 && h.current < len(h.snapshots)-1 {
		h.snapshots = h.snapshots[:h.current+1]
	}

	h.snapshots = append(h.snapshots, snap)

	// Maintain circular buffer constraint
	if len(h.snapshots) > h.maxSize {
		h.snapshots = h.snapshots[len(h.snapshots)-h.maxSize:]
	}

	h.current = len(h.snapshots) - 1
}

// Undo restores the buffer state to the previous snapshot.
func (h *History) Undo(buf *buffer.Buffer) (string, error) {
	if buf == nil || h.current <= 0 || len(h.snapshots) == 0 {
		return "Já na alteração mais antiga", nil
	}

	h.current--
	snap := h.snapshots[h.current]

	buf.Lines = CloneLines(snap.Lines)
	buf.Cursor = snap.Cursor
	buf.IsDirty = snap.IsDirty
	buf.FilePath = snap.FilePath

	return "", nil
}

// Redo restores the buffer state to the next available snapshot.
func (h *History) Redo(buf *buffer.Buffer) (string, error) {
	if buf == nil || h.current >= len(h.snapshots)-1 || len(h.snapshots) == 0 {
		return "Já na alteração mais recente", nil
	}

	h.current++
	snap := h.snapshots[h.current]

	buf.Lines = CloneLines(snap.Lines)
	buf.Cursor = snap.Cursor
	buf.IsDirty = snap.IsDirty
	buf.FilePath = snap.FilePath

	return "", nil
}

// Count returns the number of snapshots in history.
func (h *History) Count() int {
	return len(h.snapshots)
}

// CurrentIndex returns the active snapshot index.
func (h *History) CurrentIndex() int {
	return h.current
}
