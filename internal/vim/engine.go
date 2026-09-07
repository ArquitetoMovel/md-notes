package vim

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/buffer"
)

// Engine is the central Vim modal editing state machine.
type Engine struct {
	Buffer                *buffer.Buffer
	History               *History
	Clipboard             *Clipboard
	State                 ModalState
	pendingOp             rune
	countAccum            int
	SaveCallback          func(targetPath string, force bool) tea.Cmd
	QuitCallback          func(force bool) tea.Cmd
	ConfigCallback        func() tea.Cmd
	SearchQueryCallback   func(query string, isReverse bool) tea.Cmd
	SearchNavCallback     func(forward bool) tea.Cmd
	SearchCancelCallback  func(initPos Position) tea.Cmd
	SearchConfirmCallback func(query string, isReverse bool) tea.Cmd
	insertChanged         bool
	LastSearch            string
	LastSearchReverse     bool
}

// NewEngine creates and initializes a new Vim Engine for the given buffer.
func NewEngine(buf *buffer.Buffer) *Engine {
	h := NewHistory(200)
	if buf != nil {
		h.Push(buf)
	}

	return &Engine{
		Buffer:    buf,
		History:   h,
		Clipboard: NewClipboard(),
		State: ModalState{
			CurrentMode:  ModeNormal,
			Selection:    nil,
			CommandInput: "",
			StatusMsg:    "",
			Count:        0,
		},
	}
}

// GetModalState returns a copy of the current modal state.
func (e *Engine) GetModalState() ModalState {
	return e.State
}

// SetMode forces the active editing mode.
func (e *Engine) SetMode(m Mode) {
	e.State.CurrentMode = m
	if m != ModeVisualChar && m != ModeVisualLine && m != ModeVisualBlock {
		e.State.Selection = nil
	}
}

// HandleKey processes an incoming keyboard event according to the active mode.
func (e *Engine) HandleKey(msg tea.KeyMsg) (tea.Cmd, string) {
	e.State.StatusMsg = ""

	switch e.State.CurrentMode {
	case ModeNormal:
		return e.handleNormalKey(msg)
	case ModeInsert:
		return e.handleInsertKey(msg)
	case ModeVisualChar, ModeVisualLine, ModeVisualBlock:
		return e.handleVisualKey(msg)
	case ModeCommand:
		return e.handleCommandKey(msg)
	case ModeSearch:
		return e.handleSearchKey(msg)
	default:
		e.State.CurrentMode = ModeNormal
		return nil, ""
	}
}

func (e *Engine) handleNormalKey(msg tea.KeyMsg) (tea.Cmd, string) {
	keyStr := msg.String()

	// Handle Escape / Cancel pending state
	if msg.Type == tea.KeyEsc || keyStr == "esc" {
		e.pendingOp = 0
		e.countAccum = 0
		e.State.Count = 0
		return nil, ""
	}

	// Accumulate numeric multiplier counts (1-9 to start, 0-9 subsequent)
	if len(keyStr) == 1 && unicode.IsDigit(rune(keyStr[0])) && e.pendingOp != 'f' && e.pendingOp != 'F' && e.pendingOp != 't' && e.pendingOp != 'T' {
		digit := int(keyStr[0] - '0')
		if e.countAccum > 0 || digit > 0 {
			e.countAccum = e.countAccum*10 + digit
			e.State.Count = e.countAccum
			return nil, ""
		}
	}

	count := 1
	if e.countAccum > 0 {
		count = e.countAccum
	}

	// Handle pending multi-key operators (g, f, F, t, T, d, y)
	if e.pendingOp != 0 {
		op := e.pendingOp
		e.pendingOp = 0
		e.countAccum = 0
		e.State.Count = 0

		switch op {
		case 'g':
			if keyStr == "g" {
				DocStart(e.Buffer, false)
			}
			return nil, ""

		case 'f':
			if len([]rune(keyStr)) > 0 {
				r := []rune(keyStr)[0]
				FindCharForward(e.Buffer, r, count)
			}
			return nil, ""

		case 'F':
			if len([]rune(keyStr)) > 0 {
				r := []rune(keyStr)[0]
				FindCharBackward(e.Buffer, r, count)
			}
			return nil, ""

		case 't':
			if len([]rune(keyStr)) > 0 {
				r := []rune(keyStr)[0]
				TillCharForward(e.Buffer, r, count)
			}
			return nil, ""

		case 'T':
			if len([]rune(keyStr)) > 0 {
				r := []rune(keyStr)[0]
				TillCharBackward(e.Buffer, r, count)
			}
			return nil, ""

		case 'd':
			if keyStr == "d" { // dd
				e.deleteLines(count)
				return nil, ""
			} else if keyStr == "w" { // dw
				e.deleteWord(count)
				return nil, ""
			} else if keyStr == "$" {
				e.deleteToEndOfLine()
				return nil, ""
			} else if keyStr == "0" {
				e.deleteToStartOfLine()
				return nil, ""
			}
			return nil, ""

		case 'y':
			if keyStr == "y" { // yy
				e.yankLines(count)
				return nil, ""
			} else if keyStr == "w" { // yw
				e.yankWord(count)
				return nil, ""
			} else if keyStr == "$" {
				e.yankToEndOfLine()
				return nil, ""
			}
			return nil, ""
		}
	}

	// Single key normal operations
	e.countAccum = 0
	e.State.Count = 0

	switch keyStr {
	case "h", "left":
		MoveLeft(e.Buffer, count)
	case "l", "right":
		MoveRight(e.Buffer, count, false)
	case "j", "down":
		MoveDown(e.Buffer, count, false)
	case "k", "up":
		MoveUp(e.Buffer, count, false)
	case "ctrl+d", "pagedown", "pgdown":
		MoveDown(e.Buffer, 12*count, false)
	case "ctrl+u", "pageup", "pgup":
		MoveUp(e.Buffer, 12*count, false)
	case "ctrl+f":
		MoveDown(e.Buffer, 24*count, false)
	case "ctrl+b":
		MoveUp(e.Buffer, 24*count, false)
	case "ctrl+e":
		MoveDown(e.Buffer, count, false)
	case "ctrl+y":
		MoveUp(e.Buffer, count, false)
	case "w":
		NextWordStart(e.Buffer, count)
	case "b":
		PrevWordStart(e.Buffer, count)
	case "e":
		WordEnd(e.Buffer, count)
	case "0", "home":
		LineStart(e.Buffer)
	case "^":
		LineStartNonBlank(e.Buffer)
	case "$", "end":
		LineEnd(e.Buffer, false)
	case "G":
		if count > 1 {
			targetLine := count - 1
			if targetLine >= len(e.Buffer.Lines) {
				targetLine = len(e.Buffer.Lines) - 1
			}
			e.Buffer.Cursor.Line = targetLine
			LineStart(e.Buffer)
			ClampCursor(e.Buffer, false)
		} else {
			DocEnd(e.Buffer, false)
		}
	case "g":
		e.pendingOp = 'g'
		e.countAccum = count
	case "f":
		e.pendingOp = 'f'
		e.countAccum = count
	case "F":
		e.pendingOp = 'F'
		e.countAccum = count
	case "t":
		e.pendingOp = 't'
		e.countAccum = count
	case "T":
		e.pendingOp = 'T'
		e.countAccum = count
	case "d":
		e.pendingOp = 'd'
		e.countAccum = count
	case "y":
		e.pendingOp = 'y'
		e.countAccum = count

	case "x":
		e.deleteChar(count)

	case "p":
		e.paste(false)
	case "P":
		e.paste(true)

	case "u":
		statusMsg, _ := e.History.Undo(e.Buffer)
		return nil, statusMsg

	case "ctrl+r":
		statusMsg, _ := e.History.Redo(e.Buffer)
		return nil, statusMsg

	case "i":
		e.State.CurrentMode = ModeInsert
		e.insertChanged = false
	case "a":
		if e.Buffer.Cursor.Line < len(e.Buffer.Lines) {
			lineLen := e.Buffer.Lines[e.Buffer.Cursor.Line].Length()
			if lineLen > 0 {
				e.Buffer.Cursor.Col++
			}
		}
		e.State.CurrentMode = ModeInsert
		e.insertChanged = false
	case "I":
		LineStartNonBlank(e.Buffer)
		e.State.CurrentMode = ModeInsert
		e.insertChanged = false
	case "A":
		LineEnd(e.Buffer, true)
		e.State.CurrentMode = ModeInsert
		e.insertChanged = false
	case "o":
		e.openLineBelow()
		e.State.CurrentMode = ModeInsert
		e.insertChanged = false
	case "O":
		e.openLineAbove()
		e.State.CurrentMode = ModeInsert
		e.insertChanged = false

	case "v":
		e.State.CurrentMode = ModeVisualChar
		e.State.Selection = &Selection{
			Start: Position{Line: e.Buffer.Cursor.Line, Col: e.Buffer.Cursor.Col},
			End:   Position{Line: e.Buffer.Cursor.Line, Col: e.Buffer.Cursor.Col},
			Type:  ModeVisualChar,
		}
	case "V":
		e.State.CurrentMode = ModeVisualLine
		e.State.Selection = &Selection{
			Start: Position{Line: e.Buffer.Cursor.Line, Col: 0},
			End:   Position{Line: e.Buffer.Cursor.Line, Col: 0},
			Type:  ModeVisualLine,
		}
	case "ctrl+v":
		e.State.CurrentMode = ModeVisualBlock
		e.State.Selection = &Selection{
			Start: Position{Line: e.Buffer.Cursor.Line, Col: e.Buffer.Cursor.Col},
			End:   Position{Line: e.Buffer.Cursor.Line, Col: e.Buffer.Cursor.Col},
			Type:  ModeVisualBlock,
		}

	case ":":
		e.State.CurrentMode = ModeCommand
		e.State.CommandInput = ""

	case "/":
		initPos := Position{Line: 0, Col: 0}
		if e.Buffer != nil {
			initPos = Position{Line: e.Buffer.Cursor.Line, Col: e.Buffer.Cursor.Col}
		}
		e.State.CurrentMode = ModeSearch
		e.State.SearchQuery = ""
		e.State.SearchIsReverse = false
		e.State.SearchInitPos = initPos
		e.State.CommandInput = ""
		if e.SearchQueryCallback != nil {
			return e.SearchQueryCallback("", false), ""
		}

	case "?":
		initPos := Position{Line: 0, Col: 0}
		if e.Buffer != nil {
			initPos = Position{Line: e.Buffer.Cursor.Line, Col: e.Buffer.Cursor.Col}
		}
		e.State.CurrentMode = ModeSearch
		e.State.SearchQuery = ""
		e.State.SearchIsReverse = true
		e.State.SearchInitPos = initPos
		e.State.CommandInput = ""
		if e.SearchQueryCallback != nil {
			return e.SearchQueryCallback("", true), ""
		}

	case "n":
		if e.SearchNavCallback != nil {
			return e.SearchNavCallback(true), ""
		}

	case "N":
		if e.SearchNavCallback != nil {
			return e.SearchNavCallback(false), ""
		}
	}

	return nil, ""
}

func (e *Engine) handleInsertKey(msg tea.KeyMsg) (tea.Cmd, string) {
	switch msg.Type {
	case tea.KeyEsc:
		e.State.CurrentMode = ModeNormal
		ClampCursor(e.Buffer, false)
		if e.insertChanged {
			e.History.Push(e.Buffer)
			e.insertChanged = false
		}
		return nil, ""

	case tea.KeyEnter:
		e.Buffer.InsertNewLine()
		e.insertChanged = true
		return nil, ""

	case tea.KeyBackspace, tea.KeyDelete:
		e.Buffer.DeleteRune()
		e.insertChanged = true
		return nil, ""

	case tea.KeyRunes:
		for _, r := range msg.Runes {
			e.Buffer.InsertRune(r)
		}
		e.insertChanged = true
		return nil, ""

	case tea.KeySpace:
		e.Buffer.InsertRune(' ')
		e.insertChanged = true
		return nil, ""

	case tea.KeyTab:
		for i := 0; i < 4; i++ {
			e.Buffer.InsertRune(' ')
		}
		e.insertChanged = true
		return nil, ""

	default:
		if len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				e.Buffer.InsertRune(r)
			}
			e.insertChanged = true
		}
		return nil, ""
	}
}

func (e *Engine) handleVisualKey(msg tea.KeyMsg) (tea.Cmd, string) {
	keyStr := msg.String()

	if msg.Type == tea.KeyEsc || keyStr == "esc" {
		e.State.CurrentMode = ModeNormal
		e.State.Selection = nil
		return nil, ""
	}

	switch keyStr {
	case "h", "left":
		MoveLeft(e.Buffer, 1)
	case "l", "right":
		MoveRight(e.Buffer, 1, false)
	case "j", "down":
		MoveDown(e.Buffer, 1, false)
	case "k", "up":
		MoveUp(e.Buffer, 1, false)
	case "ctrl+d", "pagedown", "pgdown":
		MoveDown(e.Buffer, 12, false)
	case "ctrl+u", "pageup", "pgup":
		MoveUp(e.Buffer, 12, false)
	case "w":
		NextWordStart(e.Buffer, 1)
	case "b":
		PrevWordStart(e.Buffer, 1)
	case "0":
		LineStart(e.Buffer)
	case "$":
		LineEnd(e.Buffer, false)
	case "G":
		DocEnd(e.Buffer, false)
	case "y":
		e.yankVisualSelection()
		e.State.CurrentMode = ModeNormal
		e.State.Selection = nil
		return nil, ""
	case "d", "x":
		e.deleteVisualSelection()
		e.State.CurrentMode = ModeNormal
		e.State.Selection = nil
		return nil, ""
	}

	if e.State.Selection != nil {
		e.State.Selection.End = Position{
			Line: e.Buffer.Cursor.Line,
			Col:  e.Buffer.Cursor.Col,
		}
	}

	return nil, ""
}

func (e *Engine) handleCommandKey(msg tea.KeyMsg) (tea.Cmd, string) {
	switch msg.Type {
	case tea.KeyEsc:
		e.State.CurrentMode = ModeNormal
		e.State.CommandInput = ""
		return nil, ""

	case tea.KeyEnter:
		res := ExecuteExCommand(e.State.CommandInput, e.Buffer, e.SaveCallback, e.QuitCallback, e.ConfigCallback)
		if res.CloseMode {
			e.State.CurrentMode = ModeNormal
			e.State.CommandInput = ""
		}
		return res.Cmd, res.StatusMsg

	case tea.KeyBackspace, tea.KeyDelete:
		runes := []rune(e.State.CommandInput)
		if len(runes) > 0 {
			e.State.CommandInput = string(runes[:len(runes)-1])
		} else {
			e.State.CurrentMode = ModeNormal
			e.State.CommandInput = ""
		}
		return nil, ""

	default:
		if len(msg.Runes) > 0 {
			e.State.CommandInput += string(msg.Runes)
		} else if msg.String() == "space" {
			e.State.CommandInput += " "
		}
		return nil, ""
	}
}

func (e *Engine) handleSearchKey(msg tea.KeyMsg) (tea.Cmd, string) {
	switch msg.Type {
	case tea.KeyEsc:
		if e.Buffer != nil {
			e.Buffer.Cursor.Line = e.State.SearchInitPos.Line
			e.Buffer.Cursor.Col = e.State.SearchInitPos.Col
			ClampCursor(e.Buffer, false)
		}
		e.State.CurrentMode = ModeNormal
		e.State.SearchQuery = ""
		e.State.CommandInput = ""
		var cmd tea.Cmd
		if e.SearchCancelCallback != nil {
			cmd = e.SearchCancelCallback(e.State.SearchInitPos)
		}
		return cmd, ""

	case tea.KeyEnter:
		e.State.CurrentMode = ModeNormal
		query := e.State.SearchQuery
		isReverse := e.State.SearchIsReverse
		if query != "" {
			e.LastSearch = query
			e.LastSearchReverse = isReverse
		}
		e.State.CommandInput = ""
		var cmd tea.Cmd
		if e.SearchConfirmCallback != nil {
			cmd = e.SearchConfirmCallback(query, isReverse)
		}
		return cmd, ""

	case tea.KeyBackspace, tea.KeyDelete:
		runes := []rune(e.State.SearchQuery)
		if len(runes) > 0 {
			e.State.SearchQuery = string(runes[:len(runes)-1])
			e.State.CommandInput = e.State.SearchQuery
			var cmd tea.Cmd
			if e.SearchQueryCallback != nil {
				cmd = e.SearchQueryCallback(e.State.SearchQuery, e.State.SearchIsReverse)
			}
			return cmd, ""
		}
		if e.Buffer != nil {
			e.Buffer.Cursor.Line = e.State.SearchInitPos.Line
			e.Buffer.Cursor.Col = e.State.SearchInitPos.Col
			ClampCursor(e.Buffer, false)
		}
		e.State.CurrentMode = ModeNormal
		e.State.SearchQuery = ""
		e.State.CommandInput = ""
		var cmd tea.Cmd
		if e.SearchCancelCallback != nil {
			cmd = e.SearchCancelCallback(e.State.SearchInitPos)
		}
		return cmd, ""

	default:
		added := ""
		if len(msg.Runes) > 0 {
			added = string(msg.Runes)
		} else if msg.String() == "space" {
			added = " "
		}
		if added != "" {
			e.State.SearchQuery += added
			e.State.CommandInput = e.State.SearchQuery
			var cmd tea.Cmd
			if e.SearchQueryCallback != nil {
				cmd = e.SearchQueryCallback(e.State.SearchQuery, e.State.SearchIsReverse)
			}
			return cmd, ""
		}
		return nil, ""
	}
}

// Editing helpers

func (e *Engine) deleteChar(count int) {
	if count <= 0 {
		count = 1
	}
	if e.Buffer.Cursor.Line < 0 || e.Buffer.Cursor.Line >= len(e.Buffer.Lines) {
		return
	}

	line := &e.Buffer.Lines[e.Buffer.Cursor.Line]
	if line.Length() == 0 {
		return
	}

	var deletedRunes []rune
	for i := 0; i < count && e.Buffer.Cursor.Col < line.Length(); i++ {
		r := line.DeleteAt(e.Buffer.Cursor.Col)
		deletedRunes = append(deletedRunes, r)
	}

	e.Clipboard.Set(string(deletedRunes), false)
	e.Buffer.SetDirty(true)
	ClampCursor(e.Buffer, false)
	e.History.Push(e.Buffer)
}

func (e *Engine) deleteLines(count int) {
	if count <= 0 {
		count = 1
	}
	startLine := e.Buffer.Cursor.Line
	if startLine < 0 || startLine >= len(e.Buffer.Lines) {
		return
	}

	var deletedText strings.Builder
	for i := 0; i < count && startLine < len(e.Buffer.Lines); i++ {
		deleted, err := e.Buffer.DeleteLine(startLine)
		if err == nil {
			deletedText.WriteString(deleted.String() + "\n")
		}
	}

	e.Clipboard.Set(deletedText.String(), true)
	ClampCursor(e.Buffer, false)
	e.History.Push(e.Buffer)
}

func (e *Engine) yankLines(count int) {
	if count <= 0 {
		count = 1
	}
	startLine := e.Buffer.Cursor.Line
	if startLine < 0 || startLine >= len(e.Buffer.Lines) {
		return
	}

	var sb strings.Builder
	for i := 0; i < count && startLine+i < len(e.Buffer.Lines); i++ {
		line, _ := e.Buffer.GetLine(startLine + i)
		sb.WriteString(line.String() + "\n")
	}

	e.Clipboard.Set(sb.String(), true)
}

func (e *Engine) deleteWord(count int) {
	origCol := e.Buffer.Cursor.Col
	origLine := e.Buffer.Cursor.Line

	NextWordStart(e.Buffer, count)
	targetCol := e.Buffer.Cursor.Col
	targetLine := e.Buffer.Cursor.Line

	e.Buffer.Cursor.Line = origLine
	e.Buffer.Cursor.Col = origCol

	if origLine == targetLine && origLine < len(e.Buffer.Lines) {
		line := &e.Buffer.Lines[origLine]
		numToDelete := targetCol - origCol
		var deletedRunes []rune
		for i := 0; i < numToDelete && origCol < line.Length(); i++ {
			r := line.DeleteAt(origCol)
			deletedRunes = append(deletedRunes, r)
		}
		e.Clipboard.Set(string(deletedRunes), false)
		e.Buffer.SetDirty(true)
		ClampCursor(e.Buffer, false)
		e.History.Push(e.Buffer)
	}
}

func (e *Engine) yankWord(count int) {
	origCol := e.Buffer.Cursor.Col
	origLine := e.Buffer.Cursor.Line

	NextWordStart(e.Buffer, count)
	targetCol := e.Buffer.Cursor.Col
	targetLine := e.Buffer.Cursor.Line

	e.Buffer.Cursor.Line = origLine
	e.Buffer.Cursor.Col = origCol

	if origLine == targetLine && origLine < len(e.Buffer.Lines) {
		runes := e.Buffer.Lines[origLine].Runes
		if origCol < len(runes) {
			end := targetCol
			if end > len(runes) {
				end = len(runes)
			}
			e.Clipboard.Set(string(runes[origCol:end]), false)
		}
	}
}

func (e *Engine) deleteToEndOfLine() {
	if e.Buffer.Cursor.Line >= len(e.Buffer.Lines) {
		return
	}
	line := &e.Buffer.Lines[e.Buffer.Cursor.Line]
	col := e.Buffer.Cursor.Col
	if col < line.Length() {
		deleted := line.Runes[col:]
		line.Runes = line.Runes[:col]
		e.Clipboard.Set(string(deleted), false)
		e.Buffer.SetDirty(true)
		ClampCursor(e.Buffer, false)
		e.History.Push(e.Buffer)
	}
}

func (e *Engine) deleteToStartOfLine() {
	if e.Buffer.Cursor.Line >= len(e.Buffer.Lines) {
		return
	}
	line := &e.Buffer.Lines[e.Buffer.Cursor.Line]
	col := e.Buffer.Cursor.Col
	if col > 0 && col <= line.Length() {
		deleted := line.Runes[:col]
		line.Runes = line.Runes[col:]
		e.Buffer.Cursor.Col = 0
		e.Clipboard.Set(string(deleted), false)
		e.Buffer.SetDirty(true)
		ClampCursor(e.Buffer, false)
		e.History.Push(e.Buffer)
	}
}

func (e *Engine) yankToEndOfLine() {
	if e.Buffer.Cursor.Line >= len(e.Buffer.Lines) {
		return
	}
	line := e.Buffer.Lines[e.Buffer.Cursor.Line]
	col := e.Buffer.Cursor.Col
	if col < line.Length() {
		e.Clipboard.Set(string(line.Runes[col:]), false)
	}
}

func (e *Engine) paste(before bool) {
	text, isLineWise := e.Clipboard.Get()
	if text == "" {
		return
	}

	if isLineWise {
		lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
		insertIdx := e.Buffer.Cursor.Line
		if !before {
			insertIdx++
		}
		for i, l := range lines {
			e.Buffer.InsertLine(insertIdx+i, buffer.NewLine(l, buffer.EndingLF))
		}
		e.Buffer.Cursor.Line = insertIdx
		e.Buffer.Cursor.Col = 0
	} else {
		if !before && e.Buffer.Cursor.Line < len(e.Buffer.Lines) {
			lineLen := e.Buffer.Lines[e.Buffer.Cursor.Line].Length()
			if lineLen > 0 && e.Buffer.Cursor.Col < lineLen {
				e.Buffer.Cursor.Col++
			}
		}
		e.Buffer.InsertText(text)
	}

	ClampCursor(e.Buffer, false)
	e.History.Push(e.Buffer)
}

func (e *Engine) openLineBelow() {
	insertIdx := e.Buffer.Cursor.Line + 1
	e.Buffer.InsertLine(insertIdx, buffer.NewLine("", buffer.EndingLF))
	e.Buffer.Cursor.Line = insertIdx
	e.Buffer.Cursor.Col = 0
}

func (e *Engine) openLineAbove() {
	insertIdx := e.Buffer.Cursor.Line
	e.Buffer.InsertLine(insertIdx, buffer.NewLine("", buffer.EndingLF))
	e.Buffer.Cursor.Col = 0
}

func (e *Engine) yankVisualSelection() {
	if e.State.Selection == nil {
		return
	}
	text, isLineWise := e.getVisualText()
	e.Clipboard.Set(text, isLineWise)
}

func (e *Engine) deleteVisualSelection() {
	if e.State.Selection == nil {
		return
	}
	text, isLineWise := e.getVisualText()
	e.Clipboard.Set(text, isLineWise)

	sel := e.State.Selection
	start, end := sel.Normalized()

	switch sel.Type {
	case ModeVisualLine:
		for i := 0; i <= (end.Line - start.Line); i++ {
			_, _ = e.Buffer.DeleteLine(start.Line)
		}
		e.Buffer.Cursor.Line = start.Line
		e.Buffer.Cursor.Col = 0

	case ModeVisualChar:
		if start.Line == end.Line {
			line := &e.Buffer.Lines[start.Line]
			num := end.Col - start.Col + 1
			for i := 0; i < num && start.Col < line.Length(); i++ {
				line.DeleteAt(start.Col)
			}
			e.Buffer.Cursor.Line = start.Line
			e.Buffer.Cursor.Col = start.Col
		} else {
			// First line: delete from start.Col to end
			firstLine := &e.Buffer.Lines[start.Line]
			if start.Col < firstLine.Length() {
				firstLine.Runes = firstLine.Runes[:start.Col]
			}
			// Middle lines: delete entirely
			for i := start.Line + 1; i < end.Line; i++ {
				_, _ = e.Buffer.DeleteLine(start.Line + 1)
			}
			// Last line: delete from 0 to end.Col and join with first line
			lastLineIdx := start.Line + 1
			if lastLineIdx < len(e.Buffer.Lines) {
				lastLine := e.Buffer.Lines[lastLineIdx]
				var remainder []rune
				if end.Col+1 < len(lastLine.Runes) {
					remainder = lastLine.Runes[end.Col+1:]
				}
				_, _ = e.Buffer.DeleteLine(lastLineIdx)
				firstLine.Runes = append(firstLine.Runes, remainder...)
			}
			e.Buffer.Cursor.Line = start.Line
			e.Buffer.Cursor.Col = start.Col
		}

	case ModeVisualBlock:
		minCol := min(start.Col, end.Col)
		maxCol := max(start.Col, end.Col)
		minLine := min(start.Line, end.Line)
		maxLine := max(start.Line, end.Line)

		for l := minLine; l <= maxLine && l < len(e.Buffer.Lines); l++ {
			line := &e.Buffer.Lines[l]
			if minCol < line.Length() {
				deleteCount := maxCol - minCol + 1
				for i := 0; i < deleteCount && minCol < line.Length(); i++ {
					line.DeleteAt(minCol)
				}
			}
		}
		e.Buffer.Cursor.Line = minLine
		e.Buffer.Cursor.Col = minCol
	}

	ClampCursor(e.Buffer, false)
	e.History.Push(e.Buffer)
}

func (e *Engine) getVisualText() (string, bool) {
	sel := e.State.Selection
	if sel == nil {
		return "", false
	}

	start, end := sel.Normalized()

	switch sel.Type {
	case ModeVisualLine:
		var sb strings.Builder
		for l := start.Line; l <= end.Line && l < len(e.Buffer.Lines); l++ {
			line, _ := e.Buffer.GetLine(l)
			sb.WriteString(line.String() + "\n")
		}
		return sb.String(), true

	case ModeVisualChar:
		if start.Line == end.Line {
			line, err := e.Buffer.GetLine(start.Line)
			if err != nil {
				return "", false
			}
			runes := line.Runes
			if start.Col >= len(runes) {
				return "", false
			}
			endCol := end.Col + 1
			if endCol > len(runes) {
				endCol = len(runes)
			}
			return string(runes[start.Col:endCol]), false
		}

		var sb strings.Builder
		for l := start.Line; l <= end.Line && l < len(e.Buffer.Lines); l++ {
			line, _ := e.Buffer.GetLine(l)
			runes := line.Runes
			if l == start.Line {
				if start.Col < len(runes) {
					sb.WriteString(string(runes[start.Col:]) + "\n")
				} else {
					sb.WriteString("\n")
				}
			} else if l == end.Line {
				endCol := end.Col + 1
				if endCol > len(runes) {
					endCol = len(runes)
				}
				sb.WriteString(string(runes[:endCol]))
			} else {
				sb.WriteString(line.String() + "\n")
			}
		}
		return sb.String(), false

	case ModeVisualBlock:
		minCol := min(start.Col, end.Col)
		maxCol := max(start.Col, end.Col)
		minLine := min(start.Line, end.Line)
		maxLine := max(start.Line, end.Line)

		var sb strings.Builder
		for l := minLine; l <= maxLine && l < len(e.Buffer.Lines); l++ {
			line, _ := e.Buffer.GetLine(l)
			runes := line.Runes
			if minCol < len(runes) {
				end := maxCol + 1
				if end > len(runes) {
					end = len(runes)
				}
				sb.WriteString(string(runes[minCol:end]))
			}
			if l < maxLine {
				sb.WriteString("\n")
			}
		}
		return sb.String(), false
	}

	return "", false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
