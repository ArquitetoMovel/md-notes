package vim

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/buffer"
)

// TC15: Execução de Comandos Ex :w, :q, :wq e :q!
func TestCommand_SaveAndQuit(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")

	var saveCalled bool
	var savedPath string
	var savedForce bool

	var quitCalled bool
	var quitForce bool

	saveFunc := func(path string, force bool) tea.Cmd {
		saveCalled = true
		savedPath = path
		savedForce = force
		return nil
	}

	quitFunc := func(force bool) tea.Cmd {
		quitCalled = true
		quitForce = force
		return nil
	}

	// Test :w
	res := ExecuteExCommand("w", buf, saveFunc, quitFunc)
	if !saveCalled || savedPath != "" || savedForce != false {
		t.Errorf("Execução de :w falhou: called=%v, path=%s, force=%v", saveCalled, savedPath, savedForce)
	}

	// Test :w custom_path.md
	saveCalled = false
	res = ExecuteExCommand("w custom_path.md", buf, saveFunc, quitFunc)
	if !saveCalled || savedPath != "custom_path.md" {
		t.Errorf("Execução de :w custom_path.md falhou: called=%v, path=%s", saveCalled, savedPath)
	}

	// Test :q
	res = ExecuteExCommand("q", buf, saveFunc, quitFunc)
	if !quitCalled || quitForce != false {
		t.Errorf("Execução de :q falhou: called=%v, force=%v", quitCalled, quitForce)
	}

	// Test :q!
	quitCalled = false
	res = ExecuteExCommand("q!", buf, saveFunc, quitFunc)
	if !quitCalled || quitForce != true {
		t.Errorf("Execução de :q! falhou: called=%v, force=%v", quitCalled, quitForce)
	}

	// Test :wq
	saveCalled = false
	quitCalled = false
	res = ExecuteExCommand("wq", buf, saveFunc, quitFunc)
	if !saveCalled {
		t.Error("Execução de :wq não chamou salvamento")
	}
	if !res.CloseMode {
		t.Error("Esperado CloseMode == true")
	}
}

// TC16: Tratamento de Comando Ex Inválido (E492)
func TestCommand_UnknownCommand_E492(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	res := ExecuteExCommand("comando_invalido", buf, nil, nil)

	if !strings.Contains(res.StatusMsg, "E492: Não é um comando de editor: comando_invalido") {
		t.Errorf("Esperado erro E492 formatado, obtido: '%s'", res.StatusMsg)
	}
	if !res.CloseMode {
		t.Error("Esperado fechar modo command após comando inválido")
	}
}

func TestCommand_JumpToLine(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	for i := 0; i < 10; i++ {
		buf.Lines = append(buf.Lines, buffer.NewLine("Linha", buffer.EndingLF))
	}
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	res := ExecuteExCommand("5", buf, nil, nil)
	if buf.Cursor.Line != 4 { // 1-indexed 5 -> 0-indexed 4
		t.Errorf("Salto :5 esperado linha 4, obtido %d", buf.Cursor.Line)
	}
	if !res.CloseMode {
		t.Error("Esperado CloseMode == true")
	}
}

func TestCommand_Replace(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("antigo item 1", buffer.EndingLF),
		buffer.NewLine("antigo item 2", buffer.EndingLF),
	}

	// Global replace :%s/antigo/novo/g
	res := ExecuteExCommand("%s/antigo/novo/g", buf, nil, nil)
	if !strings.Contains(res.StatusMsg, "2 substituições realizadas") {
		t.Errorf("Esperado '2 substituições realizadas', obtido '%s'", res.StatusMsg)
	}
	l0, _ := buf.GetLine(0)
	l1, _ := buf.GetLine(1)
	if l0.String() != "novo item 1" || l1.String() != "novo item 2" {
		t.Errorf("Linhas não substituídas corretamente: '%s', '%s'", l0.String(), l1.String())
	}

	// Line replace :s/novo/renovado/g
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}
	res2 := ExecuteExCommand("s/novo/renovado/g", buf, nil, nil)
	if !strings.Contains(res2.StatusMsg, "1 substituições realizadas") {
		t.Errorf("Esperado '1 substituições realizadas', obtido '%s'", res2.StatusMsg)
	}
	l0After, _ := buf.GetLine(0)
	if l0After.String() != "renovado item 1" {
		t.Errorf("Esperado 'renovado item 1', obtido '%s'", l0After.String())
	}
}

