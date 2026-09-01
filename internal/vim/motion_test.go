package vim

import (
	"testing"

	"md-notes/internal/buffer"
)

// TC02: Movimentação Direcional Básica (h, j, k, l)
func TestMotion_BasicDirections(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("Linha 1 🚀", buffer.EndingLF),
		buffer.NewLine("Linha 2", buffer.EndingLF),
		buffer.NewLine("Linha 3", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	// Move right across runes including emoji
	for i := 0; i < 20; i++ {
		MoveRight(buf, 1, false)
	}
	// "Linha 1 🚀" has 9 runes -> in normal mode max col is 8
	if buf.Cursor.Col != 8 {
		t.Errorf("Esperado cursor na coluna 8 ('🚀'), obtido: %d", buf.Cursor.Col)
	}

	// Move down to Line 2 ("Linha 2" has 7 runes -> max col is 6)
	MoveDown(buf, 1, false)
	if buf.Cursor.Line != 1 || buf.Cursor.Col != 6 {
		t.Errorf("Esperado cursor em (1, 6), obtido: (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}

	// Move left
	MoveLeft(buf, 2)
	if buf.Cursor.Col != 4 {
		t.Errorf("Esperado cursor na coluna 4, obtido: %d", buf.Cursor.Col)
	}

	// Move up to Line 0
	MoveUp(buf, 1, false)
	if buf.Cursor.Line != 0 || buf.Cursor.Col != 4 {
		t.Errorf("Esperado cursor em (0, 4), obtido: (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}
}

// TC03: Navegação por Palavras (w e b)
func TestMotion_WordNavigation(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("const [count, setCount] = useState(0);", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	// Sequence of token starts expected on 'w':
	// const (0) -> [ (6) -> count (7) -> , (12) -> setCount (14) -> ] (22) -> = (24) -> useState (26) -> ( (34) -> 0 (35) -> ) (36) -> ; (37)
	expectedCols := []int{6, 7, 12, 14, 22, 24, 26, 34, 35, 36, 37}

	for _, expected := range expectedCols {
		NextWordStart(buf, 1)
		if buf.Cursor.Col != expected {
			t.Errorf("NextWordStart: esperado coluna %d, obtido %d", expected, buf.Cursor.Col)
		}
	}

	// Reverse with 'b'
	reverseExpectedCols := []int{36, 35, 34, 26, 24, 22, 14, 12, 7, 6, 0}
	for _, expected := range reverseExpectedCols {
		PrevWordStart(buf, 1)
		if buf.Cursor.Col != expected {
			t.Errorf("PrevWordStart: esperado coluna %d, obtido %d", expected, buf.Cursor.Col)
		}
	}
}

// TC04: Limites de Linha e Documento (0, $, gg, G)
func TestMotion_LineAndDocLimits(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	for i := 0; i < 10; i++ {
		buf.Lines = append(buf.Lines, buffer.NewLine("Exemplo de linha longa", buffer.EndingLF))
	}
	buf.Cursor = buffer.Cursor{Line: 5, Col: 4}

	// 0 -> start of line
	LineStart(buf)
	if buf.Cursor.Col != 0 {
		t.Errorf("LineStart: esperado col 0, obtido %d", buf.Cursor.Col)
	}

	// $ -> end of line
	LineEnd(buf, false)
	expectedEndCol := len([]rune("Exemplo de linha longa")) - 1
	if buf.Cursor.Col != expectedEndCol {
		t.Errorf("LineEnd: esperado col %d, obtido %d", expectedEndCol, buf.Cursor.Col)
	}

	// gg -> start of doc
	DocStart(buf, false)
	if buf.Cursor.Line != 0 || buf.Cursor.Col != 0 {
		t.Errorf("DocStart: esperado (0, 0), obtido (%d, %d)", buf.Cursor.Line, buf.Cursor.Col)
	}

	// G -> end of doc
	DocEnd(buf, false)
	if buf.Cursor.Line != 10 { // 1 initial + 10 appended = 11 lines (index 10)
		t.Errorf("DocEnd: esperado linha %d, obtido %d", len(buf.Lines)-1, buf.Cursor.Line)
	}
}

// TC19: Busca de Caractere na Linha Ativa (f, F, t, T)
func TestMotion_InlineSearch(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("chave = 'valor_configurado'", buffer.EndingLF),
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	// f' -> first single quote at index 8
	if !FindCharForward(buf, '\'', 1) {
		t.Fatal("FindCharForward single quote falhou")
	}
	if buf.Cursor.Col != 8 {
		t.Errorf("Esperado col 8 para f', obtido: %d", buf.Cursor.Col)
	}

	// t_ -> till underscore ('valor_configurado' -> _ is at index 14 -> till is at 13)
	if !TillCharForward(buf, '_', 1) {
		t.Fatal("TillCharForward('_') falhou")
	}
	if buf.Cursor.Col != 13 {
		t.Errorf("Esperado col 13 para t_, obtido: %d", buf.Cursor.Col)
	}

	// F= -> backward to '=' at index 6
	if !FindCharBackward(buf, '=', 1) {
		t.Fatal("FindCharBackward('=') falhou")
	}
	if buf.Cursor.Col != 6 {
		t.Errorf("Esperado col 6 para F=, obtido: %d", buf.Cursor.Col)
	}

	// T= -> backward till '=' -> index 7
	buf.Cursor.Col = 10
	if !TillCharBackward(buf, '=', 1) {
		t.Fatal("TillCharBackward('=') falhou")
	}
	if buf.Cursor.Col != 7 {
		t.Errorf("Esperado col 7 para T=, obtido: %d", buf.Cursor.Col)
	}
}

// TC18: Multiplicadores Numéricos de Repetição (movimento 5j)
func TestMotion_NumericCount(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	for i := 0; i < 20; i++ {
		buf.Lines = append(buf.Lines, buffer.NewLine("Linha", buffer.EndingLF))
	}
	buf.Cursor = buffer.Cursor{Line: 2, Col: 0}

	MoveDown(buf, 5, false)
	if buf.Cursor.Line != 7 {
		t.Errorf("MoveDown count 5: esperado linha 7, obtido %d", buf.Cursor.Line)
	}

	MoveUp(buf, 3, false)
	if buf.Cursor.Line != 4 {
		t.Errorf("MoveUp count 3: esperado linha 4, obtido %d", buf.Cursor.Line)
	}
}
