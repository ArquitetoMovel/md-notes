package search

import (
	"md-notes/internal/buffer"
)

// SearchEngine manages searching within buffer lines and navigating through matches.
type SearchEngine struct {
	ActivePattern *Pattern
	Result        SearchResult
	IsReverse     bool
	SmartCase     bool
}

// NewSearchEngine creates an initialized SearchEngine instance.
func NewSearchEngine() *SearchEngine {
	return &SearchEngine{
		SmartCase: true,
	}
}

// Search performs a fresh search on the given buffer lines and stores the results.
func (se *SearchEngine) Search(lines []buffer.Line, query string, isReverse bool) *SearchResult {
	se.IsReverse = isReverse
	if query == "" {
		se.Clear()
		return &se.Result
	}

	pat, err := CompilePattern(query, se.SmartCase)
	if err != nil {
		se.Clear()
		return &se.Result
	}

	se.ActivePattern = pat
	res := se.FindAll(lines, pat)
	res.Query = query
	se.Result = *res
	return &se.Result
}

// FindAll scans all buffer lines using the compiled pattern and returns a SearchResult.
func (se *SearchEngine) FindAll(lines []buffer.Line, pattern *Pattern) *SearchResult {
	if pattern == nil || pattern.RawQuery == "" {
		return &SearchResult{}
	}

	var matches []Match
	for lineIdx, line := range lines {
		raw := line.String()
		intervals := pattern.FindInLine(raw)
		runes := line.Runes
		for _, interval := range intervals {
			start := interval[0]
			end := interval[1]
			matchText := ""
			if start >= 0 && end <= len(runes) && start <= end {
				matchText = string(runes[start:end])
			} else if start >= 0 && start <= len(raw) && end <= len(raw) && start <= end {
				matchText = raw[start:end]
			}

			matches = append(matches, Match{
				Line:      lineIdx,
				StartCol:  start,
				EndCol:    end,
				Text:      matchText,
				IsCurrent: false,
			})
		}
	}

	total := len(matches)
	curIdx := 0
	if total > 0 {
		curIdx = 1
		matches[0].IsCurrent = true
	}

	return &SearchResult{
		Matches:      matches,
		TotalCount:   total,
		CurrentIndex: curIdx,
		Query:        pattern.RawQuery,
		IsRegex:      !pattern.IsLiteral && pattern.Regex != nil,
	}
}

// Next jumps to the next match relative to current cursor position (curLine, curCol).
// Returns the target match and whether a wrap-around occurred.
func (se *SearchEngine) Next(curLine, curCol int) (*Match, bool) {
	if se.IsReverse {
		return se.navigate(curLine, curCol, false)
	}
	return se.navigate(curLine, curCol, true)
}

// Prev jumps to the previous match relative to current cursor position (curLine, curCol).
// Returns the target match and whether a wrap-around occurred.
func (se *SearchEngine) Prev(curLine, curCol int) (*Match, bool) {
	if se.IsReverse {
		return se.navigate(curLine, curCol, true)
	}
	return se.navigate(curLine, curCol, false)
}

func (se *SearchEngine) navigate(curLine, curCol int, forward bool) (*Match, bool) {
	matches := se.Result.Matches
	total := len(matches)
	if total == 0 {
		return nil, false
	}

	// Reset all IsCurrent
	for i := range se.Result.Matches {
		se.Result.Matches[i].IsCurrent = false
	}

	var targetIdx int
	wrapAround := false

	if forward {
		// Find first match strictly after (curLine, curCol)
		found := false
		for i, m := range matches {
			if m.Line > curLine || (m.Line == curLine && m.StartCol > curCol) {
				targetIdx = i
				found = true
				break
			}
		}
		if !found {
			// Wrap around to the first match
			targetIdx = 0
			wrapAround = true
		}
	} else {
		// Find last match strictly before (curLine, curCol)
		found := false
		for i := total - 1; i >= 0; i-- {
			m := matches[i]
			if m.Line < curLine || (m.Line == curLine && m.StartCol < curCol) {
				targetIdx = i
				found = true
				break
			}
		}
		if !found {
			// Wrap around to the last match
			targetIdx = total - 1
			wrapAround = true
		}
	}

	se.Result.CurrentIndex = targetIdx + 1
	se.Result.Matches[targetIdx].IsCurrent = true
	return &se.Result.Matches[targetIdx], wrapAround
}

// Clear resets the active search state.
func (se *SearchEngine) Clear() {
	se.ActivePattern = nil
	se.Result = SearchResult{}
}
