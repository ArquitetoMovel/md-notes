package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/buffer"
	"md-notes/internal/vim"
)

func makeTestKey(str string) tea.KeyMsg {
	switch str {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	default:
		runes := []rune(str)
		if len(runes) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
		}
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
	}
}

func TestModel_InteractiveSearch_FullFlow(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Primeira linha de texto", buffer.EndingLF),
		buffer.NewLine("Segunda com markdown notas", buffer.EndingLF),
		buffer.NewLine("Terceira linha comum", buffer.EndingLF),
		buffer.NewLine("Quarta com markdown editor", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	m := NewModel(buf, nil)

	// 1. Press '/' to enter search mode
	m.Update(makeTestKey("/"))
	if m.Vim.State.CurrentMode != vim.ModeSearch {
		t.Fatalf("Esperado ModeSearch, obtido %s", m.Vim.State.CurrentMode)
	}

	// 2. Type "markdown"
	for _, ch := range "markdown" {
		m.Update(makeTestKey(string(ch)))
	}

	// Real-time incremental search must find 2 matches
	if m.SearchEngine.Result.TotalCount != 2 {
		t.Fatalf("Esperado 2 matches para 'markdown', obtido %d", m.SearchEngine.Result.TotalCount)
	}
	// Cursor should jump to first match on Line 1, Col 12 ("markdown")
	if buf.Cursor.Line != 1 || buf.Cursor.Col != 12 {
		t.Errorf("Esperado cursor no primeiro match (1, 12), obtido (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}

	// 3. Confirm with Enter
	m.Update(makeTestKey("enter"))
	if m.Vim.State.CurrentMode != vim.ModeNormal {
		t.Fatalf("Esperado retorno para ModeNormal após Enter, obtido %s", m.Vim.State.CurrentMode)
	}
	if buf.Cursor.Line != 1 || buf.Cursor.Col != 12 {
		t.Errorf("Cursor deve permanecer no primeiro match após Enter: (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}

	// 4. Press 'n' to go to next match (Line 3, Col 11)
	m.Update(makeTestKey("n"))
	if buf.Cursor.Line != 3 || buf.Cursor.Col != 11 {
		t.Errorf("Esperado cursor no segundo match (3, 11), obtido (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}

	// 5. Press 'n' again to trigger wrap-around back to Line 1
	m.Update(makeTestKey("n"))
	if buf.Cursor.Line != 1 || buf.Cursor.Col != 12 {
		t.Errorf("Esperado wrap-around para Linha 1, obtido (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}
	if !strings.Contains(m.StatusMsg, "Busca atingiu o fim do arquivo, continuando do início") {
		t.Errorf("Aviso de wrap-around esperado na StatusMsg, obtido '%s'", m.StatusMsg)
	}

	// 6. Press 'N' to go backward (wrap-around to Line 3)
	m.Update(makeTestKey("N"))
	if buf.Cursor.Line != 3 || buf.Cursor.Col != 11 {
		t.Errorf("Esperado retorno ao match da Linha 3 com 'N', obtido (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}
}

func TestModel_InteractiveSearch_CancelWithEsc(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Linha 0 inicial", buffer.EndingLF),
		buffer.NewLine("Linha 1 com alvo", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 5}

	m := NewModel(buf, nil)

	// Press '/'
	m.Update(makeTestKey("/"))

	// Type "alvo"
	for _, ch := range "alvo" {
		m.Update(makeTestKey(string(ch)))
	}

	// Cursor jumped to match
	if buf.Cursor.Line != 1 {
		t.Errorf("Esperado cursor saltar para linha 1 durante digitação, obtido linha %d", buf.Cursor.Line)
	}

	// Cancel with Esc
	m.Update(makeTestKey("esc"))

	if m.Vim.State.CurrentMode != vim.ModeNormal {
		t.Errorf("Esperado retorno a ModeNormal após Esc, obtido %s", m.Vim.State.CurrentMode)
	}
	if buf.Cursor.Line != 0 || buf.Cursor.Col != 5 {
		t.Errorf("Esperado cursor restaurado para posição inicial (0, 5), obtido (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}
	if m.SearchEngine.Result.TotalCount != 0 {
		t.Errorf("Esperado SearchEngine limpo após cancelamento, obtido %d matches", m.SearchEngine.Result.TotalCount)
	}
}

func TestModel_InteractiveSearch_NotFound(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Linha única de teste", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	m := NewModel(buf, nil)

	m.Update(makeTestKey("/"))
	for _, ch := range "inexistente" {
		m.Update(makeTestKey(string(ch)))
	}
	m.Update(makeTestKey("enter"))

	if !strings.Contains(m.StatusMsg, "E486: Padrão não encontrado: inexistente") {
		t.Errorf("Esperado erro E486 na StatusMsg, obtido '%s'", m.StatusMsg)
	}
}

func TestModel_ExReplaceCommand(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("palavra antiga 1", buffer.EndingLF),
		buffer.NewLine("palavra antiga 2", buffer.EndingLF),
	}

	m := NewModel(buf, nil)

	// Enter command mode and type :%s/antiga/nova/g
	m.Update(makeTestKey(":"))
	for _, ch := range "%s/antiga/nova/g" {
		m.Update(makeTestKey(string(ch)))
	}
	m.Update(makeTestKey("enter"))

	l0, _ := buf.GetLine(0)
	l1, _ := buf.GetLine(1)
	if l0.String() != "palavra nova 1" || l1.String() != "palavra nova 2" {
		t.Errorf("Substituição global falhou: '%s', '%s'", l0.String(), l1.String())
	}
	if !strings.Contains(m.StatusMsg, "2 substituições realizadas") {
		t.Errorf("Esperado '2 substituições realizadas', obtido '%s'", m.StatusMsg)
	}
}
