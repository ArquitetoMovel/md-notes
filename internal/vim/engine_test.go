package vim

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/buffer"
)

func makeKey(str string) tea.KeyMsg {
	if str == "esc" {
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	if str == "enter" {
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	if str == "backspace" {
		return tea.KeyMsg{Type: tea.KeyBackspace}
	}
	if str == "ctrl+r" {
		return tea.KeyMsg{Type: tea.KeyCtrlR}
	}
	if str == "ctrl+v" {
		return tea.KeyMsg{Type: tea.KeyCtrlV}
	}
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(str),
	}
}

// TC01: Inicialização no Modo Normal
func TestEngine_Initialization(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	engine := NewEngine(buf)

	if engine.State.CurrentMode != ModeNormal {
		t.Errorf("Esperado ModeNormal, obtido %s", engine.State.CurrentMode)
	}
	if buf.Cursor.Line != 0 || buf.Cursor.Col != 0 {
		t.Errorf("Esperado cursor em (0, 0), obtido (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}
}

// TC05 & TC06: Operadores de Entrada no Modo Insert (i, a, I, A, o, O) e retorno com Esc
func TestEngine_ModeTransitions(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Texto", buffer.EndingLF),
	}
	engine := NewEngine(buf)

	// Test 'i' on col 2
	buf.Cursor = buffer.Cursor{Line: 0, Col: 2}
	engine.HandleKey(makeKey("i"))
	if engine.State.CurrentMode != ModeInsert || buf.Cursor.Col != 2 {
		t.Errorf("'i': esperado ModeInsert na col 2, obtido %s col %d", engine.State.CurrentMode, buf.Cursor.Col)
	}
	engine.HandleKey(makeKey("esc"))
	if engine.State.CurrentMode != ModeNormal {
		t.Errorf("'esc': esperado ModeNormal, obtido %s", engine.State.CurrentMode)
	}

	// Test 'a' on col 2 -> should advance to col 3
	buf.Cursor = buffer.Cursor{Line: 0, Col: 2}
	engine.HandleKey(makeKey("a"))
	if engine.State.CurrentMode != ModeInsert || buf.Cursor.Col != 3 {
		t.Errorf("'a': esperado ModeInsert na col 3, obtido %s col %d", engine.State.CurrentMode, buf.Cursor.Col)
	}
	engine.HandleKey(makeKey("esc"))

	// Test 'I' -> col 0
	buf.Cursor = buffer.Cursor{Line: 0, Col: 3}
	engine.HandleKey(makeKey("I"))
	if engine.State.CurrentMode != ModeInsert || buf.Cursor.Col != 0 {
		t.Errorf("'I': esperado ModeInsert na col 0, obtido %s col %d", engine.State.CurrentMode, buf.Cursor.Col)
	}
	engine.HandleKey(makeKey("esc"))

	// Test 'A' -> end of line (col 5 for insert)
	buf.Cursor = buffer.Cursor{Line: 0, Col: 1}
	engine.HandleKey(makeKey("A"))
	if engine.State.CurrentMode != ModeInsert || buf.Cursor.Col != 5 {
		t.Errorf("'A': esperado ModeInsert na col 5, obtido %s col %d", engine.State.CurrentMode, buf.Cursor.Col)
	}
	engine.HandleKey(makeKey("esc"))

	// Test 'o' -> line below
	buf.Cursor = buffer.Cursor{Line: 0, Col: 2}
	engine.HandleKey(makeKey("o"))
	if engine.State.CurrentMode != ModeInsert || buf.Cursor.Line != 1 || buf.Cursor.Col != 0 {
		t.Errorf("'o': esperado ModeInsert em (1, 0), obtido %s em (%d, %d)", engine.State.CurrentMode, buf.Cursor.Line, buf.Cursor.Col)
	}
	engine.HandleKey(makeKey("esc"))

	// Test 'O' -> line above
	buf.Cursor = buffer.Cursor{Line: 1, Col: 0}
	engine.HandleKey(makeKey("O"))
	if engine.State.CurrentMode != ModeInsert || buf.Cursor.Line != 1 || buf.Cursor.Col != 0 {
		t.Errorf("'O': esperado ModeInsert em (1, 0), obtido %s em (%d, %d)", engine.State.CurrentMode, buf.Cursor.Line, buf.Cursor.Col)
	}
	engine.HandleKey(makeKey("esc"))
}

// TC07: Operação de Deleção de Caractere (x)
func TestEngine_DeleteChar(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Teste", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 1} // 'e'
	engine := NewEngine(buf)

	engine.HandleKey(makeKey("x"))

	line, _ := buf.GetLine(0)
	if line.String() != "Tste" {
		t.Errorf("Esperado 'Tste', obtido '%s'", line.String())
	}
	clipText, _ := engine.Clipboard.Get()
	if clipText != "e" {
		t.Errorf("Clipboard esperado 'e', obtido '%s'", clipText)
	}
}

// TC08 & TC09: Deleção de Linha (dd), Palavra (dw), Cópia (yy) e Colagem (p, P)
func TestEngine_DeleteAndYank(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Primeira", buffer.EndingLF),
		buffer.NewLine("Segunda linha", buffer.EndingLF),
		buffer.NewLine("Terceira", buffer.EndingLF),
	}
	engine := NewEngine(buf)

	// dd on line 1 ("Segunda linha")
	buf.Cursor = buffer.Cursor{Line: 1, Col: 0}
	engine.HandleKey(makeKey("d"))
	engine.HandleKey(makeKey("d"))

	if buf.LineCount() != 2 {
		t.Fatalf("Esperado 2 linhas após dd, obtido %d", buf.LineCount())
	}
	clipText, isLineWise := engine.Clipboard.Get()
	if !strings.Contains(clipText, "Segunda linha") || !isLineWise {
		t.Errorf("Clipboard após dd inválido: text=%q, linewise=%v", clipText, isLineWise)
	}

	// p (paste line below line 0)
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}
	engine.HandleKey(makeKey("p"))

	line1, _ := buf.GetLine(1)
	if line1.String() != "Segunda linha" {
		t.Errorf("Esperado 'Segunda linha' na linha 1 após p, obtido '%s'", line1.String())
	}

	// dw on line 1 ("Segunda linha" -> delete "Segunda ")
	buf.Cursor = buffer.Cursor{Line: 1, Col: 0}
	engine.HandleKey(makeKey("d"))
	engine.HandleKey(makeKey("w"))

	line1AfterDw, _ := buf.GetLine(1)
	if line1AfterDw.String() != "linha" {
		t.Errorf("Esperado 'linha' após dw, obtido '%s'", line1AfterDw.String())
	}
}

// TC12: Desfazer Atômico de Sessão Contínua de Inserção
func TestEngine_AtomicInsertSessionUndo(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	engine := NewEngine(buf)

	// Enter insert mode and type
	engine.HandleKey(makeKey("i"))
	for _, ch := range "Texto digitado" {
		engine.HandleKey(makeKey(string(ch)))
	}
	engine.HandleKey(makeKey("enter"))
	for _, ch := range "Segunda linha" {
		engine.HandleKey(makeKey(string(ch)))
	}
	engine.HandleKey(makeKey("esc"))

	if buf.LineCount() != 2 {
		t.Fatalf("Esperado 2 linhas após digitação, obtido %d", buf.LineCount())
	}

	// Single 'u' must undo the entire session back to empty
	engine.HandleKey(makeKey("u"))

	if buf.LineCount() != 1 {
		t.Fatalf("Esperado 1 linha após undo atômico, obtido %d", buf.LineCount())
	}
	line0, _ := buf.GetLine(0)
	if line0.String() != "" {
		t.Errorf("Esperado buffer vazio após undo atômico, obtido '%s'", line0.String())
	}
}

// TC14: Entrada no Modo Command (:)
func TestEngine_CommandInput(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	engine := NewEngine(buf)

	engine.HandleKey(makeKey(":"))
	if engine.State.CurrentMode != ModeCommand {
		t.Errorf("Esperado ModeCommand, obtido %s", engine.State.CurrentMode)
	}

	for _, ch := range "w notas.md" {
		engine.HandleKey(makeKey(string(ch)))
	}

	if engine.State.CommandInput != "w notas.md" {
		t.Errorf("Esperado CommandInput 'w notas.md', obtido '%s'", engine.State.CommandInput)
	}

	// Backspace
	engine.HandleKey(makeKey("backspace"))
	if engine.State.CommandInput != "w notas.m" {
		t.Errorf("Esperado 'w notas.m' após backspace, obtido '%s'", engine.State.CommandInput)
	}

	engine.HandleKey(makeKey("esc"))
	if engine.State.CurrentMode != ModeNormal {
		t.Errorf("Esperado ModeNormal após esc no command mode, obtido %s", engine.State.CurrentMode)
	}
}

// TC18: Contadores Numéricos (3dd)
func TestEngine_RepetitionCounts(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	for i := 0; i < 10; i++ {
		buf.Lines = append(buf.Lines, buffer.NewLine("Linha", buffer.EndingLF))
	}
	engine := NewEngine(buf)

	buf.Cursor = buffer.Cursor{Line: 2, Col: 0}
	engine.HandleKey(makeKey("3"))
	engine.HandleKey(makeKey("d"))
	engine.HandleKey(makeKey("d"))

	// 1 initial + 10 = 11 lines. 11 - 3 = 8 lines.
	if buf.LineCount() != 8 {
		t.Errorf("Esperado 8 linhas após 3dd, obtido %d", buf.LineCount())
	}
}

// TC20: Modos Visuais por Caractere (v), Linha (V) e Bloco (Ctrl+v)
func TestEngine_VisualSelection(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Linha 1", buffer.EndingLF),
		buffer.NewLine("Linha 2", buffer.EndingLF),
		buffer.NewLine("Linha 3", buffer.EndingLF),
	}
	engine := NewEngine(buf)

	// Test 'v' + motion + 'y'
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}
	engine.HandleKey(makeKey("v"))
	if engine.State.CurrentMode != ModeVisualChar {
		t.Fatalf("Esperado ModeVisualChar, obtido %s", engine.State.CurrentMode)
	}
	engine.HandleKey(makeKey("l"))
	engine.HandleKey(makeKey("l"))
	engine.HandleKey(makeKey("l"))
	engine.HandleKey(makeKey("y"))

	if engine.State.CurrentMode != ModeNormal {
		t.Errorf("Esperado retorno para ModeNormal após yank visual, obtido %s", engine.State.CurrentMode)
	}
	clipText, _ := engine.Clipboard.Get()
	if clipText != "Linh" {
		t.Errorf("Esperado 'Linh' copiado no modo visual char, obtido '%s'", clipText)
	}

	// Test 'V' + motion + 'y'
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}
	engine.HandleKey(makeKey("V"))
	engine.HandleKey(makeKey("j"))
	engine.HandleKey(makeKey("y"))

	clipText, isLineWise := engine.Clipboard.Get()
	if !isLineWise || !strings.Contains(clipText, "Linha 1\nLinha 2\n") {
		t.Errorf("Modo Visual Line yank incorreto: %q, isLineWise=%v", clipText, isLineWise)
	}

	// Test 'v' + 'd' (delete selection)
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}
	engine.HandleKey(makeKey("v"))
	engine.HandleKey(makeKey("l"))
	engine.HandleKey(makeKey("l"))
	engine.HandleKey(makeKey("d"))

	line0, _ := buf.GetLine(0)
	if line0.String() != "ha 1" {
		t.Errorf("Esperado 'ha 1' após deleção visual, obtido '%s'", line0.String())
	}
}

// Testes de Modo de Busca (F06)
func TestEngine_SearchMode_ForwardAndBackward(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Primeira linha", buffer.EndingLF),
		buffer.NewLine("Segunda linha", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 1, Col: 4}
	engine := NewEngine(buf)

	// Test '/' enters ModeSearch forward
	engine.HandleKey(makeKey("/"))
	if engine.State.CurrentMode != ModeSearch {
		t.Fatalf("Esperado ModeSearch, obtido %s", engine.State.CurrentMode)
	}
	if engine.State.SearchIsReverse {
		t.Errorf("Esperado SearchIsReverse false para '/', obtido true")
	}
	if engine.State.SearchInitPos.Line != 1 || engine.State.SearchInitPos.Col != 4 {
		t.Errorf("Esperado SearchInitPos (1, 4), obtido (%d, %d)", engine.State.SearchInitPos.Line, engine.State.SearchInitPos.Col)
	}

	// Exit with esc
	engine.HandleKey(makeKey("esc"))
	if engine.State.CurrentMode != ModeNormal {
		t.Fatalf("Esperado retorno para ModeNormal, obtido %s", engine.State.CurrentMode)
	}

	// Test '?' enters ModeSearch reverse
	engine.HandleKey(makeKey("?"))
	if engine.State.CurrentMode != ModeSearch {
		t.Fatalf("Esperado ModeSearch para '?', obtido %s", engine.State.CurrentMode)
	}
	if !engine.State.SearchIsReverse {
		t.Errorf("Esperado SearchIsReverse true para '?', obtido false")
	}
}

func TestEngine_SearchMode_TypingAndBackspace(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	engine := NewEngine(buf)

	var lastQuery string
	engine.SearchQueryCallback = func(query string, isReverse bool) tea.Cmd {
		lastQuery = query
		return nil
	}

	engine.HandleKey(makeKey("/"))
	for _, ch := range "markdown" {
		engine.HandleKey(makeKey(string(ch)))
	}

	if engine.State.SearchQuery != "markdown" {
		t.Errorf("Esperado SearchQuery 'markdown', obtido '%s'", engine.State.SearchQuery)
	}
	if lastQuery != "markdown" {
		t.Errorf("Esperado callback com 'markdown', obtido '%s'", lastQuery)
	}

	// Backspace once
	engine.HandleKey(makeKey("backspace"))
	if engine.State.SearchQuery != "markdow" {
		t.Errorf("Esperado 'markdow' após backspace, obtido '%s'", engine.State.SearchQuery)
	}
}

func TestEngine_SearchMode_EscCancelRollback(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Linha 0", buffer.EndingLF),
		buffer.NewLine("Linha 1", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 2}
	engine := NewEngine(buf)

	cancelCalled := false
	engine.SearchCancelCallback = func(initPos Position) tea.Cmd {
		cancelCalled = true
		return nil
	}

	engine.HandleKey(makeKey("/"))
	// Simulate cursor moved by live search
	buf.Cursor = buffer.Cursor{Line: 1, Col: 5}

	engine.HandleKey(makeKey("esc"))

	if engine.State.CurrentMode != ModeNormal {
		t.Errorf("Esperado ModeNormal após esc, obtido %s", engine.State.CurrentMode)
	}
	if buf.Cursor.Line != 0 || buf.Cursor.Col != 2 {
		t.Errorf("Esperado cursor restaurado para (0, 2), obtido (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}
	if !cancelCalled {
		t.Errorf("Esperado SearchCancelCallback ter sido chamado")
	}
}

func TestEngine_SearchMode_EnterConfirm(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	engine := NewEngine(buf)

	confirmedQuery := ""
	engine.SearchConfirmCallback = func(query string, isReverse bool) tea.Cmd {
		confirmedQuery = query
		return nil
	}

	engine.HandleKey(makeKey("/"))
	for _, ch := range "busca" {
		engine.HandleKey(makeKey(string(ch)))
	}
	engine.HandleKey(makeKey("enter"))

	if engine.State.CurrentMode != ModeNormal {
		t.Errorf("Esperado ModeNormal após enter, obtido %s", engine.State.CurrentMode)
	}
	if confirmedQuery != "busca" {
		t.Errorf("Esperado confirmedQuery 'busca', obtido '%s'", confirmedQuery)
	}
	if engine.LastSearch != "busca" {
		t.Errorf("Esperado LastSearch 'busca', obtido '%s'", engine.LastSearch)
	}
}

func TestEngine_SearchMode_NavigationKeys(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	engine := NewEngine(buf)

	navDirection := ""
	engine.SearchNavCallback = func(forward bool) tea.Cmd {
		if forward {
			navDirection = "forward"
		} else {
			navDirection = "backward"
		}
		return nil
	}

	engine.HandleKey(makeKey("n"))
	if navDirection != "forward" {
		t.Errorf("Esperado navDirection 'forward' para 'n', obtido '%s'", navDirection)
	}

	engine.HandleKey(makeKey("N"))
	if navDirection != "backward" {
		t.Errorf("Esperado navDirection 'backward' para 'N', obtido '%s'", navDirection)
	}
}

func TestEngine_PageScrollingKeys(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	for i := 0; i < 100; i++ {
		buf.Lines = append(buf.Lines, buffer.NewLine("Linha de teste", buffer.EndingLF))
	}
	engine := NewEngine(buf)
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	// ctrl+d moves down 12 lines
	engine.HandleKey(makeKey("ctrl+d"))
	if buf.Cursor.Line != 12 {
		t.Errorf("Esperado Cursor.Line 12 após ctrl+d, obtido %d", buf.Cursor.Line)
	}

	// pagedown moves down 12 lines
	engine.HandleKey(makeKey("pagedown"))
	if buf.Cursor.Line != 24 {
		t.Errorf("Esperado Cursor.Line 24 após pagedown, obtido %d", buf.Cursor.Line)
	}

	// ctrl+u moves up 12 lines
	engine.HandleKey(makeKey("ctrl+u"))
	if buf.Cursor.Line != 12 {
		t.Errorf("Esperado Cursor.Line 12 após ctrl+u, obtido %d", buf.Cursor.Line)
	}

	// pageup moves up 12 lines
	engine.HandleKey(makeKey("pageup"))
	if buf.Cursor.Line != 0 {
		t.Errorf("Esperado Cursor.Line 0 após pageup, obtido %d", buf.Cursor.Line)
	}
}
