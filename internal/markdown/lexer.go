package markdown

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/theme"
)

// InlineStyleState tracks active nested inline styles during lexing.
type InlineStyleState struct {
	Bold          bool
	Italic        bool
	Strikethrough bool
	IsLink        bool
	LastOpened    TokenType
	BaseStyle     lipgloss.Style
}

// Lexer provides lexical scanning and tokenization for inline Markdown syntax.
type Lexer struct {
	theme *theme.CompiledTheme
}

// NewLexer creates a new inline Lexer with the given compiled theme.
func NewLexer(th *theme.CompiledTheme) *Lexer {
	return &Lexer{theme: th}
}

// SetTheme updates the theme used by the lexer.
func (l *Lexer) SetTheme(th *theme.CompiledTheme) {
	l.theme = th
}

// TokenizeInline parses a raw string and returns the slice of StyledSpans with base styling.
func (l *Lexer) TokenizeInline(text string, baseStyle lipgloss.Style) []StyledSpan {
	state := InlineStyleState{
		BaseStyle: baseStyle,
	}
	return l.TokenizeInlineWithState(text, state)
}

// TokenizeInlineWithState parses raw text using a specific initial style state.
func (l *Lexer) TokenizeInlineWithState(text string, state InlineStyleState) []StyledSpan {
	if text == "" {
		return nil
	}
	var spans []StyledSpan
	l.lexInlineRecursive(text, state, &spans)
	return spans
}

func (l *Lexer) lexInlineRecursive(text string, state InlineStyleState, spans *[]StyledSpan) {
	if text == "" {
		return
	}

	mutedStyle := lipgloss.NewStyle()
	if l.theme != nil {
		mutedStyle = l.theme.Muted
	}
	boldMarkerStyle := mutedStyle.Bold(true)
	italicMarkerStyle := mutedStyle.Italic(true)

	i := 0
	plainStart := 0

	flushPlain := func(end int) {
		if end > plainStart {
			chunk := text[plainStart:end]
			tokenType := TokenText
			style := state.BaseStyle

			if state.IsLink {
				tokenType = TokenLink
				if l.theme != nil {
					style = l.theme.H2.Copy().Underline(true)
				} else {
					style = style.Underline(true)
				}
			} else if state.LastOpened == TokenHeading {
				tokenType = TokenHeading
			} else if state.LastOpened == TokenQuote {
				tokenType = TokenQuote
			} else if state.LastOpened == TokenItalic || (state.Italic && !state.Bold) {
				tokenType = TokenItalic
				if l.theme != nil {
					style = l.theme.Italic
					if state.Bold {
						style = style.Bold(true)
					}
				} else {
					style = style.Italic(true)
					if state.Bold {
						style = style.Bold(true)
					}
				}
			} else if state.LastOpened == TokenBold || state.Bold {
				tokenType = TokenBold
				if l.theme != nil {
					style = l.theme.Bold
					if state.Italic {
						style = style.Italic(true)
					}
				} else {
					style = style.Bold(true)
					if state.Italic {
						style = style.Italic(true)
					}
				}
			} else if state.Strikethrough {
				tokenType = TokenStrike
				style = style.Strikethrough(true)
			}

			if state.Strikethrough && !state.IsLink && tokenType != TokenStrike {
				style = style.Strikethrough(true)
			}

			*spans = append(*spans, StyledSpan{
				Text:  chunk,
				Style: style,
				Type:  tokenType,
			})
			plainStart = end
		}
	}

	n := len(text)
	for i < n {
		// 1. Inline code: `code`
		if text[i] == '`' {
			// Find closing backtick
			closeIdx := strings.IndexByte(text[i+1:], '`')
			if closeIdx != -1 {
				closeIdx += i + 1
				flushPlain(i)

				// Opening ` marker
				*spans = append(*spans, StyledSpan{
					Text:  "`",
					Style: mutedStyle,
					Type:  TokenMarker,
				})

				codeContent := text[i+1 : closeIdx]
				codeStyle := lipgloss.NewStyle()
				if l.theme != nil {
					codeStyle = l.theme.CodeInline
				}
				if state.Bold {
					codeStyle = codeStyle.Bold(true)
				}
				if state.Italic {
					codeStyle = codeStyle.Italic(true)
				}
				if state.Strikethrough {
					codeStyle = codeStyle.Strikethrough(true)
				}

				*spans = append(*spans, StyledSpan{
					Text:  codeContent,
					Style: codeStyle,
					Type:  TokenCodeInline,
				})

				// Closing ` marker
				*spans = append(*spans, StyledSpan{
					Text:  "`",
					Style: mutedStyle,
					Type:  TokenMarker,
				})

				i = closeIdx + 1
				plainStart = i
				continue
			}
		}

		// 2. Markdown Link: [text](url)
		if text[i] == '[' {
			closeBracket := strings.IndexByte(text[i+1:], ']')
			if closeBracket != -1 {
				closeBracket += i + 1
				if closeBracket+1 < n && text[closeBracket+1] == '(' {
					closeParen := strings.IndexByte(text[closeBracket+2:], ')')
					if closeParen != -1 {
						closeParen += closeBracket + 2
						flushPlain(i)

						linkText := text[i+1 : closeBracket]
						linkURL := text[closeBracket+2 : closeParen]

						// Marker '['
						*spans = append(*spans, StyledSpan{
							Text:  "[",
							Style: mutedStyle,
							Type:  TokenMarker,
						})

						// Link text (can be nested or styled)
						subState := state
						subState.IsLink = true
						subState.LastOpened = TokenLink
						l.lexInlineRecursive(linkText, subState, spans)

						// Marker ']'
						*spans = append(*spans, StyledSpan{
							Text:  "]",
							Style: mutedStyle,
							Type:  TokenMarker,
						})

						// Marker '('
						*spans = append(*spans, StyledSpan{
							Text:  "(",
							Style: mutedStyle,
							Type:  TokenMarker,
						})

						// URL content
						*spans = append(*spans, StyledSpan{
							Text:  linkURL,
							Style: mutedStyle,
							Type:  TokenMarker,
						})

						// Marker ')'
						*spans = append(*spans, StyledSpan{
							Text:  ")",
							Style: mutedStyle,
							Type:  TokenMarker,
						})

						i = closeParen + 1
						plainStart = i
						continue
					}
				}
			}
		}

		// 3. Raw URL: http:// or https://
		if strings.HasPrefix(text[i:], "http://") || strings.HasPrefix(text[i:], "https://") {
			flushPlain(i)
			endURL := i
			for endURL < n {
				r := rune(text[endURL])
				if unicode.IsSpace(r) || r == ')' || r == ']' || r == '>' || r == '"' || r == '\'' {
					break
				}
				endURL++
			}
			urlStr := text[i:endURL]
			urlStyle := mutedStyle
			if l.theme != nil {
				urlStyle = l.theme.H2.Copy().Underline(true)
			} else {
				urlStyle = urlStyle.Underline(true)
			}

			*spans = append(*spans, StyledSpan{
				Text:  urlStr,
				Style: urlStyle,
				Type:  TokenLink,
			})

			i = endURL
			plainStart = i
			continue
		}

		// 4. Bold: **text**
		if i+1 < n && text[i] == '*' && text[i+1] == '*' {
			closeIdx := strings.Index(text[i+2:], "**")
			if closeIdx != -1 {
				closeIdx += i + 2
				flushPlain(i)

				*spans = append(*spans, StyledSpan{
					Text:  "**",
					Style: boldMarkerStyle,
					Type:  TokenMarker,
				})

				inner := text[i+2 : closeIdx]
				subState := state
				subState.Bold = true
				subState.LastOpened = TokenBold
				l.lexInlineRecursive(inner, subState, spans)

				*spans = append(*spans, StyledSpan{
					Text:  "**",
					Style: boldMarkerStyle,
					Type:  TokenMarker,
				})

				i = closeIdx + 2
				plainStart = i
				continue
			}
		}

		// 5. Bold: __text__
		if i+1 < n && text[i] == '_' && text[i+1] == '_' {
			closeIdx := strings.Index(text[i+2:], "__")
			if closeIdx != -1 {
				closeIdx += i + 2
				flushPlain(i)

				*spans = append(*spans, StyledSpan{
					Text:  "__",
					Style: boldMarkerStyle,
					Type:  TokenMarker,
				})

				inner := text[i+2 : closeIdx]
				subState := state
				subState.Bold = true
				subState.LastOpened = TokenBold
				l.lexInlineRecursive(inner, subState, spans)

				*spans = append(*spans, StyledSpan{
					Text:  "__",
					Style: boldMarkerStyle,
					Type:  TokenMarker,
				})

				i = closeIdx + 2
				plainStart = i
				continue
			}
		}

		// 6. Strikethrough: ~~text~~
		if i+1 < n && text[i] == '~' && text[i+1] == '~' {
			closeIdx := strings.Index(text[i+2:], "~~")
			if closeIdx != -1 {
				closeIdx += i + 2
				flushPlain(i)

				*spans = append(*spans, StyledSpan{
					Text:  "~~",
					Style: mutedStyle,
					Type:  TokenMarker,
				})

				inner := text[i+2 : closeIdx]
				subState := state
				subState.Strikethrough = true
				subState.LastOpened = TokenStrike
				l.lexInlineRecursive(inner, subState, spans)

				*spans = append(*spans, StyledSpan{
					Text:  "~~",
					Style: mutedStyle,
					Type:  TokenMarker,
				})

				i = closeIdx + 2
				plainStart = i
				continue
			}
		}

		// 7. Italic: *text*
		if text[i] == '*' {
			closeIdx := strings.IndexByte(text[i+1:], '*')
			if closeIdx != -1 {
				closeIdx += i + 1
				flushPlain(i)

				*spans = append(*spans, StyledSpan{
					Text:  "*",
					Style: italicMarkerStyle,
					Type:  TokenMarker,
				})

				inner := text[i+1 : closeIdx]
				subState := state
				subState.Italic = true
				subState.LastOpened = TokenItalic
				l.lexInlineRecursive(inner, subState, spans)

				*spans = append(*spans, StyledSpan{
					Text:  "*",
					Style: italicMarkerStyle,
					Type:  TokenMarker,
				})

				i = closeIdx + 1
				plainStart = i
				continue
			}
		}

		// 8. Italic: _text_
		if text[i] == '_' {
			// Check if this is an emphasis delimiter and not within a snake_case_identifier
			closeIdx := strings.IndexByte(text[i+1:], '_')
			if closeIdx != -1 {
				closeIdx += i + 1
				flushPlain(i)

				*spans = append(*spans, StyledSpan{
					Text:  "_",
					Style: italicMarkerStyle,
					Type:  TokenMarker,
				})

				inner := text[i+1 : closeIdx]
				subState := state
				subState.Italic = true
				subState.LastOpened = TokenItalic
				l.lexInlineRecursive(inner, subState, spans)

				*spans = append(*spans, StyledSpan{
					Text:  "_",
					Style: italicMarkerStyle,
					Type:  TokenMarker,
				})

				i = closeIdx + 1
				plainStart = i
				continue
			}
		}

		i++
	}

	flushPlain(n)
}
