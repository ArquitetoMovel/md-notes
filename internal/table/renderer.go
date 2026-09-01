package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/markdown"
	"md-notes/internal/theme"
)

// Renderer provides rendering capabilities for Tables with Unicode box borders or ASCII fallback.
type Renderer struct {
	theme *theme.CompiledTheme
	lexer *markdown.Lexer
}

// NewRenderer creates a new Renderer with the given theme.
func NewRenderer(th *theme.CompiledTheme) *Renderer {
	var lex *markdown.Lexer
	if th != nil {
		lex = markdown.NewLexer(th)
	}
	return &Renderer{
		theme: th,
		lexer: lex,
	}
}

// RenderUnicode renders a table with aesthetic Unicode box-drawing borders and Lipgloss styles.
func (r *Renderer) RenderUnicode(t *Table) []string {
	cols := t.ColumnCount()
	if cols == 0 {
		return nil
	}

	widths := CalculateColumnWidths(t)
	t.ColWidths = widths

	borderStyle := lipgloss.NewStyle()
	headerStyle := lipgloss.NewStyle().Bold(true)
	cellStyle := lipgloss.NewStyle()

	if r.theme != nil {
		borderStyle = r.theme.TableBorder
		headerStyle = r.theme.TableHeader
		cellStyle = r.theme.TableCell
	}

	var lines []string

	// 1. Top border: ┌───┬───┐
	var topParts []string
	for i := 0; i < cols; i++ {
		topParts = append(topParts, strings.Repeat("─", widths[i]+2))
	}
	topLine := borderStyle.Render("┌" + strings.Join(topParts, "┬") + "┐")
	lines = append(lines, topLine)

	// 2. Header row
	var headerCells []string
	for i := 0; i < cols; i++ {
		text := ""
		if i < len(t.Header.Cells) {
			text = t.Header.Cells[i].Text
		}
		align := AlignLeft
		if i < len(t.Alignments) {
			align = t.Alignments[i]
		}
		padded := padCell(text, widths[i], align)
		headerCells = append(headerCells, " "+headerStyle.Render(padded)+" ")
	}
	sep := borderStyle.Render("│")
	lines = append(lines, sep+strings.Join(headerCells, sep)+sep)

	// 3. Delimiter / middle border: ├───┼───┤
	var midParts []string
	for i := 0; i < cols; i++ {
		midParts = append(midParts, strings.Repeat("─", widths[i]+2))
	}
	midLine := borderStyle.Render("├" + strings.Join(midParts, "┼") + "┤")
	lines = append(lines, midLine)

	// 4. Data rows
	for _, row := range t.Rows {
		var rowCells []string
		for i := 0; i < cols; i++ {
			text := ""
			if i < len(row.Cells) {
				text = row.Cells[i].Text
			}
			align := AlignLeft
			if i < len(t.Alignments) {
				align = t.Alignments[i]
			}
			renderedCell := r.renderCellContent(text, widths[i], align, cellStyle)
			rowCells = append(rowCells, " "+renderedCell+" ")
		}
		lines = append(lines, sep+strings.Join(rowCells, sep)+sep)
	}

	// 5. Bottom border: └───┴───┘
	var botParts []string
	for i := 0; i < cols; i++ {
		botParts = append(botParts, strings.Repeat("─", widths[i]+2))
	}
	botLine := borderStyle.Render("└" + strings.Join(botParts, "┴") + "┘")
	lines = append(lines, botLine)

	return lines
}

// RenderASCII renders a table using standard ASCII characters (| and -).
func (r *Renderer) RenderASCII(t *Table) []string {
	return FormatMarkdownTable(t)
}

func (r *Renderer) renderCellContent(text string, width int, align Alignment, defaultStyle lipgloss.Style) string {
	if r.lexer == nil || !strings.ContainsAny(text, "*_`~[") {
		padded := padCell(text, width, align)
		return defaultStyle.Render(padded)
	}

	// Tokenize inline Markdown within the cell
	spans := r.lexer.TokenizeInline(text, defaultStyle)
	var renderedContent strings.Builder
	for _, s := range spans {
		renderedContent.WriteString(s.Style.Render(s.Text))
	}

	// Calculate remaining width padding
	cellW := NewCell(text).Width
	if cellW >= width {
		return renderedContent.String()
	}
	diff := width - cellW

	switch align {
	case AlignRight:
		return strings.Repeat(" ", diff) + renderedContent.String()
	case AlignCenter:
		leftPad := diff / 2
		rightPad := diff - leftPad
		return strings.Repeat(" ", leftPad) + renderedContent.String() + strings.Repeat(" ", rightPad)
	case AlignLeft:
		fallthrough
	default:
		return renderedContent.String() + strings.Repeat(" ", diff)
	}
}

// RenderUnicode is a convenience helper on Table to render directly with theme.
func (t *Table) RenderUnicode(th *theme.CompiledTheme) []string {
	renderer := NewRenderer(th)
	return renderer.RenderUnicode(t)
}

// RenderASCII is a convenience helper on Table to render directly in ASCII.
func (t *Table) RenderASCII(th *theme.CompiledTheme) []string {
	renderer := NewRenderer(th)
	return renderer.RenderASCII(t)
}
