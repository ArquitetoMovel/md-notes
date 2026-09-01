package search

import (
	"errors"
	"regexp"
	"strings"

	"md-notes/internal/buffer"
)

var (
	// ErrInvalidReplaceSyntax is returned when a replace command has invalid syntax.
	ErrInvalidReplaceSyntax = errors.New("E488: Caracteres adicionais ou sintaxe de substituição inválida")
)

// ReplaceCommand encapsulates the parsed parameters of an Ex substitution command (:s or :%s).
type ReplaceCommand struct {
	Pattern           string
	Replacement       string
	IsGlobalDoc       bool // true for :%s, false for :s
	IsGlobalLine      bool // true if 'g' flag is present
	IsInteractive     bool // true if 'c' flag is present
	IsCaseInsensitive bool // true if 'i' flag is present
}

// ParseReplaceCommand parses commands like :s/old/new/g or :%s/old/new/gc.
func ParseReplaceCommand(cmdStr string) (*ReplaceCommand, error) {
	s := strings.TrimSpace(cmdStr)
	if strings.HasPrefix(s, ":") {
		s = s[1:]
	}

	isGlobalDoc := false
	if strings.HasPrefix(s, "%s") {
		isGlobalDoc = true
		s = s[2:]
	} else if strings.HasPrefix(s, "s") {
		isGlobalDoc = false
		s = s[1:]
	} else {
		return nil, ErrInvalidReplaceSyntax
	}

	if len(s) == 0 {
		return nil, ErrInvalidReplaceSyntax
	}

	delim := rune(s[0])
	if delim == ' ' || delim == '\t' || delim == '\n' || delim == '\r' {
		return nil, ErrInvalidReplaceSyntax
	}

	// Split by delimiter considering escapes
	parts := splitByDelimiter(s[1:], delim)
	if len(parts) < 2 {
		return nil, ErrInvalidReplaceSyntax
	}

	pattern := parts[0]
	replacement := parts[1]
	flags := ""
	if len(parts) >= 3 {
		flags = parts[2]
	}

	if pattern == "" {
		return nil, ErrInvalidReplaceSyntax
	}

	// Convert \1, \2, \3 to $1, $2, $3 for Go regex compatibility
	normalizedReplacement := normalizeReplacementGroups(replacement)

	isGlobalLine := false
	isInteractive := false
	isCaseInsensitive := false

	for _, flag := range flags {
		switch flag {
		case 'g':
			isGlobalLine = true
		case 'c':
			isInteractive = true
		case 'i':
			isCaseInsensitive = true
		case 'I':
			isCaseInsensitive = false
		default:
			// ignore or accept standard flags
		}
	}

	return &ReplaceCommand{
		Pattern:           pattern,
		Replacement:       normalizedReplacement,
		IsGlobalDoc:       isGlobalDoc,
		IsGlobalLine:      isGlobalLine,
		IsInteractive:     isInteractive,
		IsCaseInsensitive: isCaseInsensitive,
	}, nil
}

func splitByDelimiter(s string, delim rune) []string {
	var parts []string
	var cur strings.Builder
	escaped := false

	for _, r := range s {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			cur.WriteRune(r)
			continue
		}
		if r == delim {
			parts = append(parts, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteRune(r)
	}
	parts = append(parts, cur.String())
	return parts
}

func normalizeReplacementGroups(repl string) string {
	// Converts \1..\9 to ${1}..${9}
	reBackslash := regexp.MustCompile(`\\([0-9]+)`)
	repl = reBackslash.ReplaceAllString(repl, "$${$1}")

	// Converts bare $1..$9 (not already enclosed in {}) to ${1}..${9}
	reDollar := regexp.MustCompile(`\$([0-9]+)`)
	repl = reDollar.ReplaceAllString(repl, "$${$1}")
	return repl
}

// ExecuteReplace executes the substitution command on the given buffer.
// Returns the count of replaced occurrences.
func ExecuteReplace(buf *buffer.Buffer, cmd *ReplaceCommand, curLine int) (int, error) {
	if buf == nil || cmd == nil {
		return 0, nil
	}

	patternStr := cmd.Pattern
	if cmd.IsCaseInsensitive && !strings.HasPrefix(patternStr, "(?i)") {
		patternStr = "(?i)" + patternStr
	}

	re, err := regexp.Compile(patternStr)
	if err != nil {
		// Fallback to literal matching
		return executeLiteralReplace(buf, cmd, curLine), nil
	}

	startLine := 0
	endLine := buf.LineCount() - 1
	if !cmd.IsGlobalDoc {
		if curLine < 0 || curLine >= buf.LineCount() {
			return 0, nil
		}
		startLine = curLine
		endLine = curLine
	}

	totalReplaced := 0
	for lineIdx := startLine; lineIdx <= endLine; lineIdx++ {
		line, err := buf.GetLine(lineIdx)
		if err != nil {
			continue
		}
		raw := line.String()
		matches := re.FindAllStringIndex(raw, -1)
		if len(matches) == 0 {
			continue
		}

		var newRaw string
		if cmd.IsGlobalLine {
			newRaw = re.ReplaceAllString(raw, cmd.Replacement)
			totalReplaced += len(matches)
		} else {
			// Replace only first occurrence on the line
			loc := matches[0]
			matchedSub := raw[loc[0]:loc[1]]
			replacedSub := re.ReplaceAllString(matchedSub, cmd.Replacement)
			newRaw = raw[:loc[0]] + replacedSub + raw[loc[1]:]
			totalReplaced++
		}

		_ = buf.SetLine(lineIdx, buffer.NewLine(newRaw, line.Ending))
	}

	return totalReplaced, nil
}

func executeLiteralReplace(buf *buffer.Buffer, cmd *ReplaceCommand, curLine int) int {
	startLine := 0
	endLine := buf.LineCount() - 1
	if !cmd.IsGlobalDoc {
		if curLine < 0 || curLine >= buf.LineCount() {
			return 0
		}
		startLine = curLine
		endLine = curLine
	}

	totalReplaced := 0
	for lineIdx := startLine; lineIdx <= endLine; lineIdx++ {
		line, err := buf.GetLine(lineIdx)
		if err != nil {
			continue
		}
		raw := line.String()
		count := strings.Count(raw, cmd.Pattern)
		if count == 0 {
			continue
		}

		var newRaw string
		if cmd.IsGlobalLine {
			newRaw = strings.ReplaceAll(raw, cmd.Pattern, cmd.Replacement)
			totalReplaced += count
		} else {
			newRaw = strings.Replace(raw, cmd.Pattern, cmd.Replacement, 1)
			totalReplaced++
		}

		_ = buf.SetLine(lineIdx, buffer.NewLine(newRaw, line.Ending))
	}

	return totalReplaced
}
