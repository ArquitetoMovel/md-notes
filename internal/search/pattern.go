package search

import (
	"regexp"
	"strings"
	"unicode"
)

// Pattern encapsulates the compiled search query, regex or literal fallback, and case sensitivity.
type Pattern struct {
	RawQuery        string
	Regex           *regexp.Regexp
	IsLiteral       bool
	IsCaseSensitive bool
}

// CompilePattern parses and compiles a search query using Smart-Case with fallback to literal search.
func CompilePattern(query string, smartCase bool) (*Pattern, error) {
	if query == "" {
		return &Pattern{
			RawQuery:        "",
			Regex:           nil,
			IsLiteral:       true,
			IsCaseSensitive: false,
		}, nil
	}

	hasUpper := false
	for _, r := range query {
		if unicode.IsUpper(r) {
			hasUpper = true
			break
		}
	}

	isCaseSensitive := true
	if smartCase && !hasUpper {
		isCaseSensitive = false
	}

	patternStr := query
	if !isCaseSensitive && !strings.HasPrefix(query, "(?i)") {
		patternStr = "(?i)" + query
	}

	re, err := regexp.Compile(patternStr)
	if err != nil {
		// Fallback to literal search
		return &Pattern{
			RawQuery:        query,
			Regex:           nil,
			IsLiteral:       true,
			IsCaseSensitive: isCaseSensitive,
		}, nil
	}

	return &Pattern{
		RawQuery:        query,
		Regex:           re,
		IsLiteral:       false,
		IsCaseSensitive: isCaseSensitive,
	}, nil
}

// FindInLine finds all match ranges [startRune, endRune] in the provided line string.
func (p *Pattern) FindInLine(lineStr string) [][]int {
	if p == nil || p.RawQuery == "" || lineStr == "" {
		return nil
	}

	if p.IsLiteral || p.Regex == nil {
		return p.findLiteralInLine(lineStr)
	}

	byteIndices := p.Regex.FindAllStringIndex(lineStr, -1)
	return byteRangesToRuneRanges(lineStr, byteIndices)
}

func (p *Pattern) findLiteralInLine(lineStr string) [][]int {
	searchIn := lineStr
	target := p.RawQuery
	if !p.IsCaseSensitive {
		searchIn = strings.ToLower(lineStr)
		target = strings.ToLower(p.RawQuery)
	}

	var byteIndices [][]int
	targetLen := len(target)
	if targetLen == 0 {
		return nil
	}

	offset := 0
	for {
		idx := strings.Index(searchIn[offset:], target)
		if idx == -1 {
			break
		}
		start := offset + idx
		end := start + targetLen
		byteIndices = append(byteIndices, []int{start, end})
		offset = end
	}

	return byteRangesToRuneRanges(lineStr, byteIndices)
}

func byteRangesToRuneRanges(s string, byteRanges [][]int) [][]int {
	if len(byteRanges) == 0 {
		return nil
	}

	runeCount := 0
	byteToRune := make([]int, len(s)+1)
	for bIdx := range s {
		byteToRune[bIdx] = runeCount
		runeCount++
	}
	byteToRune[len(s)] = runeCount

	result := make([][]int, len(byteRanges))
	for i, br := range byteRanges {
		bStart := br[0]
		bEnd := br[1]
		if bStart < 0 {
			bStart = 0
		}
		if bEnd > len(s) {
			bEnd = len(s)
		}
		rStart := byteToRune[bStart]
		rEnd := byteToRune[bEnd]
		result[i] = []int{rStart, rEnd}
	}
	return result
}
