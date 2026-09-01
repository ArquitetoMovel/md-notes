package search

import (
	"fmt"
)

// Match represents a single pattern occurrence within the buffer.
type Match struct {
	Line      int    // 0-indexed line number in buffer
	StartCol  int    // 0-indexed starting rune column in line
	EndCol    int    // 0-indexed ending rune column in line (exclusive)
	Text      string // Matched string content
	IsCurrent bool   // True if this match is the active target under cursor
}

// SearchResult holds the collection of all matches found in the buffer.
type SearchResult struct {
	Matches      []Match // Ordered slice of all matches
	TotalCount   int     // Total number of matches
	CurrentIndex int     // 1-indexed active match position for status display (1..TotalCount)
	Query        string  // Search query string
	IsRegex      bool    // True if query was compiled as regular expression
}

// StatusString returns a formatted status display string like "[2/5]" or "[0/0]".
func (sr *SearchResult) StatusString() string {
	if sr == nil || sr.TotalCount == 0 {
		return "[0/0]"
	}
	idx := sr.CurrentIndex
	if idx < 1 {
		idx = 1
	}
	if idx > sr.TotalCount {
		idx = sr.TotalCount
	}
	return fmt.Sprintf("[%d/%d]", idx, sr.TotalCount)
}

// GetCurrentMatch returns the currently selected Match pointer, or nil if none.
func (sr *SearchResult) GetCurrentMatch() *Match {
	if sr == nil || len(sr.Matches) == 0 {
		return nil
	}
	idx := sr.CurrentIndex - 1
	if idx < 0 || idx >= len(sr.Matches) {
		idx = 0
	}
	return &sr.Matches[idx]
}
