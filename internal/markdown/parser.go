package markdown

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/buffer"
	"md-notes/internal/theme"
)

// Parser handles block-level Markdown structure analysis and orchestrates syntax highlighting.
type Parser struct {
	theme       *theme.CompiledTheme
	lexer       *Lexer
	highlighter *Highlighter
}

// NewParser creates a new Parser instance with the given theme.
func NewParser(th *theme.CompiledTheme) *Parser {
	lexer := NewLexer(th)
	highlighter := NewHighlighter()
	return &Parser{
		theme:       th,
		lexer:       lexer,
		highlighter: highlighter,
	}
}

// SetTheme updates the active theme on the parser, lexer and highlighter.
func (p *Parser) SetTheme(th *theme.CompiledTheme) {
	p.theme = th
	p.lexer.SetTheme(th)
}

// ParseBuffer processes a complete slice of buffer lines, tracking multiline context.
func (p *Parser) ParseBuffer(lines []buffer.Line) []TokenizedLine {
	result := make([]TokenizedLine, len(lines))
	var inCodeBlock bool
	var codeLang string

	for i, line := range lines {
		result[i] = p.ParseLine(line, &inCodeBlock, &codeLang)
	}
	return result
}

// ParseLine tokenizes a single buffer.Line within the current multiline code block context.
func (p *Parser) ParseLine(line buffer.Line, inCodeBlock *bool, codeLang *string) TokenizedLine {
	raw := line.String()
	trimmed := strings.TrimSpace(raw)

	mutedStyle := lipgloss.NewStyle()
	if p.theme != nil {
		mutedStyle = p.theme.Muted
	}

	// 1. Check if we are inside a code block or opening/closing one
	if inCodeBlock != nil && *inCodeBlock {
		// Check for code block closing delimiter: ```
		if strings.HasPrefix(trimmed, "```") {
			*inCodeBlock = false
			if codeLang != nil {
				*codeLang = ""
			}
			return TokenizedLine{
				Spans: []StyledSpan{
					{
						Text:  raw,
						Style: mutedStyle,
						Type:  TokenMarker,
					},
				},
				Original:    line,
				IsCodeBlock: false,
			}
		}

		// Inside code block
		var spans []StyledSpan
		currentLang := ""
		if codeLang != nil {
			currentLang = *codeLang
		}

		if p.highlighter != nil && currentLang != "" {
			spans = p.highlighter.HighlightLine(raw, currentLang, p.theme)
		} else {
			codeStyle := lipgloss.NewStyle()
			if p.theme != nil {
				codeStyle = p.theme.CodeBlock
			}
			spans = []StyledSpan{
				{
					Text:  "  " + raw,
					Style: codeStyle,
					Type:  TokenCodeBlock,
				},
			}
		}

		return TokenizedLine{
			Spans:       spans,
			Original:    line,
			IsCodeBlock: true,
		}
	}

	// Check for code block opening delimiter: ```lang
	if strings.HasPrefix(trimmed, "```") {
		if inCodeBlock != nil {
			*inCodeBlock = true
		}
		lang := strings.TrimPrefix(trimmed, "```")
		lang = strings.TrimSpace(lang)
		if codeLang != nil {
			*codeLang = lang
		}
		return TokenizedLine{
			Spans: []StyledSpan{
				{
					Text:  raw,
					Style: mutedStyle,
					Type:  TokenMarker,
				},
			},
			Original:    line,
			IsCodeBlock: false,
		}
	}

	// 2. Headings: # H1 to ###### H6
	if strings.HasPrefix(trimmed, "#") {
		i := 0
		for i < len(trimmed) && trimmed[i] == '#' {
			i++
		}
		if i >= 1 && i <= 6 && (i == len(trimmed) || unicode.IsSpace(rune(trimmed[i]))) {
			// Valid heading
			lvl := i
			// Find prefix in raw text preserving leading indentation
			rawTrimmedStart := strings.Index(raw, "#")
			indent := raw[:rawTrimmedStart]
			afterHashes := raw[rawTrimmedStart+lvl:]
			spaceLen := 0
			for spaceLen < len(afterHashes) && (afterHashes[spaceLen] == ' ' || afterHashes[spaceLen] == '\t') {
				spaceLen++
			}
			marker := indent + strings.Repeat("#", lvl) + afterHashes[:spaceLen]
			headingContent := afterHashes[spaceLen:]

			headingStyle := p.getHeadingStyle(lvl)
			headingMarkerStyle := mutedStyle.Bold(true)

			spans := []StyledSpan{
				{
					Text:  marker,
					Style: headingMarkerStyle,
					Type:  TokenMarker,
				},
			}

			if headingContent != "" {
				contentSpans := p.lexer.TokenizeInlineWithState(headingContent, InlineStyleState{
					BaseStyle:  headingStyle,
					LastOpened: TokenHeading,
				})
				spans = append(spans, contentSpans...)
			}

			return TokenizedLine{
				Spans:       spans,
				Original:    line,
				IsHeading:   true,
				HeadingLvl:  lvl,
				IsCodeBlock: false,
			}
		}
	}

	// 3. Blockquotes: > Quote text
	if strings.HasPrefix(trimmed, ">") {
		rawTrimmedStart := strings.Index(raw, ">")
		indent := raw[:rawTrimmedStart]
		afterGreater := raw[rawTrimmedStart+1:]
		if len(afterGreater) > 0 && afterGreater[0] == ' ' {
			afterGreater = afterGreater[1:]
		}

		quoteMarkerStyle := mutedStyle
		if p.theme != nil {
			quoteMarkerStyle = p.theme.TableBorder
		}

		quoteTextStyle := lipgloss.NewStyle().Italic(true)
		if p.theme != nil {
			quoteTextStyle = p.theme.Italic
		}

		spans := []StyledSpan{
			{
				Text:  indent + "│ ",
				Style: quoteMarkerStyle,
				Type:  TokenMarker,
			},
		}

		if afterGreater != "" {
			contentSpans := p.lexer.TokenizeInlineWithState(afterGreater, InlineStyleState{
				BaseStyle:  quoteTextStyle,
				LastOpened: TokenQuote,
			})
			spans = append(spans, contentSpans...)
		}

		return TokenizedLine{
			Spans:       spans,
			Original:    line,
			IsCodeBlock: false,
		}
	}

	// 4. Checkboxes: - [ ] or - [x] or * [ ] or * [x]
	if (strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "- [X] ") ||
		strings.HasPrefix(trimmed, "* [ ] ") || strings.HasPrefix(trimmed, "* [x] ") || strings.HasPrefix(trimmed, "* [X] ")) {
		rawStart := strings.IndexAny(raw, "-*")
		indent := raw[:rawStart]
		prefix := raw[rawStart : rawStart+6]
		rest := raw[rawStart+6:]

		chkStyle := mutedStyle
		if strings.Contains(prefix, "x") || strings.Contains(prefix, "X") {
			if p.theme != nil {
				chkStyle = p.theme.H3
			}
		}

		spans := []StyledSpan{
			{
				Text:  indent + prefix,
				Style: chkStyle,
				Type:  TokenCheckbox,
			},
		}

		if rest != "" {
			contentSpans := p.lexer.TokenizeInline(rest, lipgloss.NewStyle())
			spans = append(spans, contentSpans...)
		}

		return TokenizedLine{
			Spans:       spans,
			Original:    line,
			IsCodeBlock: false,
		}
	}

	// 5. Unordered Lists: - item, * item, + item
	if (strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ")) {
		rawStart := strings.IndexAny(raw, "-*+")
		indent := raw[:rawStart]
		bullet := raw[rawStart : rawStart+2]
		rest := raw[rawStart+2:]

		listStyle := mutedStyle
		if p.theme != nil {
			listStyle = p.theme.H4
		}

		spans := []StyledSpan{
			{
				Text:  indent + bullet,
				Style: listStyle,
				Type:  TokenList,
			},
		}

		if rest != "" {
			contentSpans := p.lexer.TokenizeInline(rest, lipgloss.NewStyle())
			spans = append(spans, contentSpans...)
		}

		return TokenizedLine{
			Spans:       spans,
			Original:    line,
			IsCodeBlock: false,
		}
	}

	// 6. Ordered Lists: 1. item, 2. item, etc.
	if len(trimmed) >= 3 && unicode.IsDigit(rune(trimmed[0])) {
		dotIdx := strings.Index(trimmed, ". ")
		if dotIdx != -1 {
			isNumeric := true
			for k := 0; k < dotIdx; k++ {
				if !unicode.IsDigit(rune(trimmed[k])) {
					isNumeric = false
					break
				}
			}
			if isNumeric {
				rawStart := strings.Index(raw, trimmed[:dotIdx+2])
				indent := raw[:rawStart]
				numPrefix := raw[rawStart : rawStart+dotIdx+2]
				rest := raw[rawStart+dotIdx+2:]

				listStyle := mutedStyle
				if p.theme != nil {
					listStyle = p.theme.H4
				}

				spans := []StyledSpan{
					{
						Text:  indent + numPrefix,
						Style: listStyle,
						Type:  TokenList,
					},
				}

				if rest != "" {
					contentSpans := p.lexer.TokenizeInline(rest, lipgloss.NewStyle())
					spans = append(spans, contentSpans...)
				}

				return TokenizedLine{
					Spans:       spans,
					Original:    line,
					IsCodeBlock: false,
				}
			}
		}
	}

	// 7. Regular paragraph / text line
	spans := p.lexer.TokenizeInline(raw, lipgloss.NewStyle())

	return TokenizedLine{
		Spans:       spans,
		Original:    line,
		IsCodeBlock: false,
	}
}

func (p *Parser) getHeadingStyle(level int) lipgloss.Style {
	if p.theme == nil {
		return lipgloss.NewStyle().Bold(true)
	}
	switch level {
	case 1:
		return p.theme.H1
	case 2:
		return p.theme.H2
	case 3:
		return p.theme.H3
	case 4:
		return p.theme.H4
	case 5:
		return p.theme.H5
	case 6:
		return p.theme.H6
	default:
		return p.theme.H1
	}
}
