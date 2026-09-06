package vim

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/buffer"
	"md-notes/internal/search"
)

// CommandResult represents the outcome of executing an Ex command.
type CommandResult struct {
	Cmd       tea.Cmd
	StatusMsg string
	CloseMode bool
}

// ExecuteExCommand parses and executes a command line string (after ':').
func ExecuteExCommand(
	input string,
	buf *buffer.Buffer,
	saveFunc func(targetPath string, force bool) tea.Cmd,
	quitFunc func(force bool) tea.Cmd,
) CommandResult {
	cmdStr := strings.TrimSpace(input)
	if cmdStr == "" {
		return CommandResult{CloseMode: true}
	}

	// Check for search replace (:s or :%s)
	if strings.HasPrefix(cmdStr, "s") || strings.HasPrefix(cmdStr, "%s") {
		repCmd, err := search.ParseReplaceCommand(cmdStr)
		if err != nil {
			return CommandResult{
				StatusMsg: err.Error(),
				CloseMode: true,
			}
		}
		curLine := 0
		if buf != nil {
			curLine = buf.Cursor.Line
		}
		count, err := search.ExecuteReplace(buf, repCmd, curLine)
		if err != nil {
			return CommandResult{
				StatusMsg: err.Error(),
				CloseMode: true,
			}
		}
		if buf != nil && count > 0 {
			buf.SetDirty(true)
		}
		return CommandResult{
			StatusMsg: fmt.Sprintf("%d substituições realizadas", count),
			CloseMode: true,
		}
	}

	// Check for direct line number jump (e.g. ":15")
	if lineNum, err := strconv.Atoi(cmdStr); err == nil {
		if lineNum <= 1 {
			buf.Cursor.Line = 0
		} else if lineNum > len(buf.Lines) {
			buf.Cursor.Line = len(buf.Lines) - 1
		} else {
			buf.Cursor.Line = lineNum - 1
		}
		buf.Cursor.Col = 0
		ClampCursor(buf, false)
		return CommandResult{CloseMode: true}
	}

	parts := strings.Fields(cmdStr)
	cmdName := parts[0]

	switch cmdName {
	case "w":
		targetPath := ""
		if len(parts) > 1 {
			targetPath = parts[1]
		}
		var cmd tea.Cmd
		if saveFunc != nil {
			cmd = saveFunc(targetPath, false)
		}
		return CommandResult{Cmd: cmd, CloseMode: true}

	case "q":
		var cmd tea.Cmd
		if quitFunc != nil {
			cmd = quitFunc(false)
		}
		return CommandResult{Cmd: cmd, CloseMode: true}

	case "q!":
		var cmd tea.Cmd
		if quitFunc != nil {
			cmd = quitFunc(true)
		}
		return CommandResult{Cmd: cmd, CloseMode: true}

	case "wq", "x":
		targetPath := ""
		if len(parts) > 1 {
			targetPath = parts[1]
		}
		var cmd tea.Cmd
		if saveFunc != nil && quitFunc != nil {
			cmd = tea.Sequence(saveFunc(targetPath, false), quitFunc(false))
		} else if saveFunc != nil {
			cmd = saveFunc(targetPath, false)
		}
		return CommandResult{Cmd: cmd, CloseMode: true}

	default:
		return CommandResult{
			StatusMsg: fmt.Sprintf("E492: Não é um comando de editor: %s", cmdName),
			CloseMode: true,
		}
	}
}
