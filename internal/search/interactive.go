package search

import (
	"regexp"
	"strings"

	"md-notes/internal/buffer"
)

// InteractiveSession tracks the state of a step-by-step confirmation replacement (:s/.../.../gc).
type InteractiveSession struct {
	Command       *ReplaceCommand
	Matches       []Match
	CurrentIndex  int
	ReplacedCount int
	IsDone        bool
	regex         *regexp.Regexp
}

// NewInteractiveSession initializes an interactive confirmation session.
func NewInteractiveSession(buf *buffer.Buffer, cmd *ReplaceCommand, curLine int) *InteractiveSession {
	if buf == nil || cmd == nil {
		return &InteractiveSession{IsDone: true}
	}

	patternStr := cmd.Pattern
	if cmd.IsCaseInsensitive && !strings.HasPrefix(patternStr, "(?i)") {
		patternStr = "(?i)" + patternStr
	}

	re, _ := regexp.Compile(patternStr)

	// Collect matches in scope
	lines := buf.Lines
	startLine := 0
	endLine := len(lines) - 1

	if !cmd.IsGlobalDoc {
		if curLine < 0 || curLine >= len(lines) {
			return &InteractiveSession{IsDone: true}
		}
		startLine = curLine
		endLine = curLine
	}

	var matches []Match
	for lineIdx := startLine; lineIdx <= endLine; lineIdx++ {
		line := lines[lineIdx]
		raw := line.String()
		var intervals [][]int
		if re != nil {
			intervals = byteRangesToRuneRanges(raw, re.FindAllStringIndex(raw, -1))
		} else {
			p, _ := CompilePattern(cmd.Pattern, !cmd.IsCaseInsensitive)
			intervals = p.FindInLine(raw)
		}

		runes := line.Runes
		for _, inv := range intervals {
			matchText := ""
			if inv[0] >= 0 && inv[1] <= len(runes) {
				matchText = string(runes[inv[0]:inv[1]])
			}
			matches = append(matches, Match{
				Line:     lineIdx,
				StartCol: inv[0],
				EndCol:   inv[1],
				Text:     matchText,
			})
		}
	}

	if len(matches) == 0 {
		return &InteractiveSession{IsDone: true}
	}

	matches[0].IsCurrent = true

	return &InteractiveSession{
		Command:      cmd,
		Matches:      matches,
		CurrentIndex: 0,
		IsDone:       false,
		regex:        re,
	}
}

// CurrentMatch returns the match currently awaiting confirmation.
func (is *InteractiveSession) CurrentMatch() *Match {
	if is == nil || is.IsDone || is.CurrentIndex < 0 || is.CurrentIndex >= len(is.Matches) {
		return nil
	}
	return &is.Matches[is.CurrentIndex]
}

// HandleAction processes a user confirmation choice:
// 'y' (yes), 'n' (no/skip), 'a' (all), 'q' (quit/cancel).
func (is *InteractiveSession) HandleAction(buf *buffer.Buffer, action rune) (done bool, nextMatch *Match) {
	if is == nil || is.IsDone || is.CurrentIndex >= len(is.Matches) {
		return true, nil
	}

	switch action {
	case 'y', 'Y':
		// Replace single match
		is.replaceMatch(buf, is.CurrentIndex)
		is.ReplacedCount++
		is.advance()

	case 'n', 'N':
		// Skip single match
		is.advance()

	case 'a', 'A':
		// Replace all remaining matches
		for idx := is.CurrentIndex; idx < len(is.Matches); idx++ {
			is.replaceMatch(buf, idx)
			is.ReplacedCount++
		}
		is.IsDone = true
		return true, nil

	case 'q', 'Q', 27: // 'q' or Esc
		// Cancel session
		is.IsDone = true
		return true, nil

	default:
		// Unknown action, stay on current match
		return false, is.CurrentMatch()
	}

	if is.CurrentIndex >= len(is.Matches) {
		is.IsDone = true
		return true, nil
	}

	return false, is.CurrentMatch()
}

func (is *InteractiveSession) replaceMatch(buf *buffer.Buffer, matchIdx int) {
	if matchIdx < 0 || matchIdx >= len(is.Matches) {
		return
	}
	m := is.Matches[matchIdx]
	line, err := buf.GetLine(m.Line)
	if err != nil {
		return
	}

	var newRaw string
	runes := line.Runes
	if m.StartCol >= 0 && m.EndCol <= len(runes) && m.StartCol <= m.EndCol {
		matchedStr := string(runes[m.StartCol:m.EndCol])
		replacementStr := is.Command.Replacement
		if is.regex != nil {
			replacementStr = is.regex.ReplaceAllString(matchedStr, is.Command.Replacement)
		}
		newRunes := append(append([]rune{}, runes[:m.StartCol]...), append([]rune(replacementStr), runes[m.EndCol:]...)...)
		newRaw = string(newRunes)

		// Shift subsequent matches on the same line
		colDiff := len([]rune(replacementStr)) - (m.EndCol - m.StartCol)
		for i := matchIdx + 1; i < len(is.Matches); i++ {
			if is.Matches[i].Line == m.Line {
				is.Matches[i].StartCol += colDiff
				is.Matches[i].EndCol += colDiff
			}
		}
	} else {
		return
	}

	_ = buf.SetLine(m.Line, buffer.NewLine(newRaw, line.Ending))
}

func (is *InteractiveSession) advance() {
	if is.CurrentIndex < len(is.Matches) {
		is.Matches[is.CurrentIndex].IsCurrent = false
	}
	is.CurrentIndex++
	if is.CurrentIndex < len(is.Matches) {
		is.Matches[is.CurrentIndex].IsCurrent = true
	} else {
		is.IsDone = true
	}
}
