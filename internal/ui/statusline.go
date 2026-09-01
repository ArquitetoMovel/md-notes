package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"md-notes/internal/buffer"
	"md-notes/internal/theme"
	"md-notes/internal/vim"
)

// StatusState holds all data required to render the status bar and bottom line.
type StatusState struct {
	Mode           vim.Mode
	FilePath       string
	IsDirty        bool
	CursorLine     int // 1-indexed
	CursorCol      int // 1-indexed
	TotalLines     int
	LineEnding     buffer.LineEnding
	SearchMatchCur int // 1-indexed, 0 if inactive
	SearchMatchTot int // 0 if inactive
	StatusMessage  string
	CommandInput   string
}

// RenderStatusLine composes the primary status line using segmented Lipgloss blocks.
func RenderStatusLine(state StatusState, th *theme.CompiledTheme, width int) string {
	if width <= 0 {
		return ""
	}

	modeStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	fileStyle := lipgloss.NewStyle().Padding(0, 1)
	dirtyStyle := lipgloss.NewStyle().Bold(true).Padding(0, 0)
	barStyle := lipgloss.NewStyle()
	infoStyle := lipgloss.NewStyle().Padding(0, 1)
	searchStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)

	if th != nil {
		barStyle = th.StatusBar
		modeStyle = modeStyle.Background(lipgloss.Color(th.Palette.H1)).Foreground(lipgloss.Color(th.Palette.StatusBarBg))
		fileStyle = fileStyle.Background(lipgloss.Color(th.Palette.StatusBarBg)).Foreground(lipgloss.Color(th.Palette.StatusBarFg)).Bold(true)
		dirtyStyle = dirtyStyle.Foreground(lipgloss.Color(th.Palette.H2))
		infoStyle = infoStyle.Background(lipgloss.Color(th.Palette.StatusBarBg)).Foreground(lipgloss.Color(th.Palette.Muted))
		searchStyle = searchStyle.Background(lipgloss.Color(th.Palette.SearchMatchBg)).Foreground(lipgloss.Color(th.Palette.SearchMatchFg))

		// Customize mode background based on mode
		switch state.Mode {
		case vim.ModeInsert:
			modeStyle = modeStyle.Background(lipgloss.Color(th.Palette.H4)).Foreground(lipgloss.Color(th.Palette.StatusBarBg))
		case vim.ModeVisual, vim.ModeVisualLine, vim.ModeVisualBlock:
			modeStyle = modeStyle.Background(lipgloss.Color(th.Palette.H3)).Foreground(lipgloss.Color(th.Palette.StatusBarBg))
		case vim.ModeCommand:
			modeStyle = modeStyle.Background(lipgloss.Color(th.Palette.H2)).Foreground(lipgloss.Color(th.Palette.StatusBarBg))
		}
	}

	// 1. Left segment: Mode
	modeText := state.Mode.String()
	modeBlock := modeStyle.Render(modeText)

	// 2. Left segment: File + Dirty
	fileName := state.FilePath
	if fileName == "" {
		fileName = "[Novo Arquivo]"
	}
	fileBlock := fileStyle.Render(fileName)
	if state.IsDirty {
		fileBlock += " " + dirtyStyle.Render("[+]")
	}

	leftSide := modeBlock + " " + fileBlock

	// 3. Search indicator (if active)
	if state.SearchMatchTot > 0 {
		cur := state.SearchMatchCur
		if cur < 1 {
			cur = 1
		}
		searchBlock := searchStyle.Render(fmt.Sprintf("[%d/%d]", cur, state.SearchMatchTot))
		leftSide += " " + searchBlock
	}

	// 4. Right segments: Encoding, Cursor Pos, Progress Percentage
	encodingStr := "UTF-8"
	endingStr := "LF"
	if state.LineEnding == buffer.EndingCRLF {
		endingStr = "CRLF"
	}
	encodingBlock := infoStyle.Render(fmt.Sprintf("%s | %s", encodingStr, endingStr))

	cursorLine := state.CursorLine
	if cursorLine < 1 {
		cursorLine = 1
	}
	cursorCol := state.CursorCol
	if cursorCol < 1 {
		cursorCol = 1
	}
	posBlock := infoStyle.Render(fmt.Sprintf("Ln %d, Col %d", cursorLine, cursorCol))

	percentStr := "Top"
	if state.TotalLines <= 1 || cursorLine <= 1 {
		percentStr = "Top"
	} else if cursorLine >= state.TotalLines {
		percentStr = "Bot"
	} else {
		pct := (cursorLine * 100) / state.TotalLines
		percentStr = fmt.Sprintf("%d%%", pct)
	}
	percentBlock := infoStyle.Render(percentStr)

	rightSide := encodingBlock + "  " + posBlock + "  " + percentBlock

	leftW := runewidth.StringWidth(lipgloss.NewStyle().Render(leftSide))
	rightW := runewidth.StringWidth(lipgloss.NewStyle().Render(rightSide))

	gap := width - leftW - rightW
	if gap < 1 {
		gap = 1
	}

	middle := strings.Repeat(" ", gap)
	if th != nil {
		middle = barStyle.Render(middle)
	}

	return leftSide + middle + rightSide
}

// RenderCommandLine renders the bottom prompt line (command input or status message).
func RenderCommandLine(state StatusState, th *theme.CompiledTheme, width int) string {
	if state.Mode == vim.ModeCommand {
		prompt := ":" + state.CommandInput
		if th != nil {
			return th.StatusBar.Render(prompt)
		}
		return prompt
	}

	if state.StatusMessage != "" {
		if th != nil {
			if strings.HasPrefix(state.StatusMessage, "E") {
				return lipgloss.NewStyle().Foreground(lipgloss.Color(th.Palette.H2)).Bold(true).Render(state.StatusMessage)
			}
			return th.Muted.Render(state.StatusMessage)
		}
		return state.StatusMessage
	}

	return ""
}
