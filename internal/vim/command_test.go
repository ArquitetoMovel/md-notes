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
	res := ExecuteExCommand("w", buf, saveFunc, quitFunc, nil)
	if !saveCalled || savedPath != "" || savedForce != false {
		t.Errorf("Execução de :w falhou: called=%v, path=%s, force=%v", saveCalled, savedPath, savedForce)
	}

	// Test :w custom_path.md
	saveCalled = false
	res = ExecuteExCommand("w custom_path.md", buf, saveFunc, quitFunc, nil)
	if !saveCalled || savedPath != "custom_path.md" {
		t.Errorf("Execução de :w custom_path.md falhou: called=%v, path=%s", saveCalled, savedPath)
	}

	// Test :q
	res = ExecuteExCommand("q", buf, saveFunc, quitFunc, nil)
	if !quitCalled || quitForce != false {
		t.Errorf("Execução de :q falhou: called=%v, force=%v", quitCalled, quitForce)
	}

	// Test :q!
	quitCalled = false
	res = ExecuteExCommand("q!", buf, saveFunc, quitFunc, nil)
	if !quitCalled || quitForce != true {
		t.Errorf("Execução de :q! falhou: called=%v, force=%v", quitCalled, quitForce)
	}

	// Test :wq
	saveCalled = false
	quitCalled = false
	res = ExecuteExCommand("wq", buf, saveFunc, quitFunc, nil)
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
	res := ExecuteExCommand("comando_invalido", buf, nil, nil, nil)

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

	res := ExecuteExCommand("5", buf, nil, nil, nil)
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
	res := ExecuteExCommand("%s/antigo/novo/g", buf, nil, nil, nil)
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
	res2 := ExecuteExCommand("s/novo/renovado/g", buf, nil, nil, nil)
	if !strings.Contains(res2.StatusMsg, "1 substituições realizadas") {
		t.Errorf("Esperado '1 substituições realizadas', obtido '%s'", res2.StatusMsg)
	}
	l0After, _ := buf.GetLine(0)
	if l0After.String() != "renovado item 1" {
		t.Errorf("Esperado 'renovado item 1', obtido '%s'", l0After.String())
	}
}

// TC01: Execução dos Comandos Ex :c e :config
func TestExecuteExCommand_Config(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")

	tests := []struct {
		name    string
		command string
	}{
		{name: "short command :c", command: "c"},
		{name: "full command :config", command: "config"},
		{name: "command with spaces :c  ", command: "c  "},
		{name: "command with spaces :config  ", command: "config  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configCalled bool
			configFunc := func() tea.Cmd {
				configCalled = true
				return func() tea.Msg { return "dummy" }
			}

			res := ExecuteExCommand(tt.command, buf, nil, nil, configFunc)
			if !configCalled {
				t.Errorf("Execução de :%s não chamou configFunc", tt.command)
			}
			if !res.CloseMode {
				t.Errorf("Execução de :%s esperava CloseMode == true", tt.command)
			}
			if res.Cmd == nil {
				t.Errorf("Execução de :%s esperava Cmd não nulo", tt.command)
			}
			if res.StatusMsg != "" {
				t.Errorf("Execução de :%s retornou mensagem de erro inesperada: '%s'", tt.command, res.StatusMsg)
			}
		})
	}

	// Teste com configFunc nulo
	t.Run("nil configFunc does not panic and closes mode", func(t *testing.T) {
		res := ExecuteExCommand("c", buf, nil, nil, nil)
		if !res.CloseMode {
			t.Error("Esperado CloseMode == true para configFunc nulo")
		}
		if res.Cmd != nil {
			t.Error("Esperado Cmd == nil quando configFunc é nulo")
		}
		if res.StatusMsg != "" {
			t.Errorf("Mensagem inesperada: '%s'", res.StatusMsg)
		}
	})
}

// TC01: Integração com Engine e ConfigCallback
func TestEngine_ConfigCommandIntegration(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	engine := NewEngine(buf)

	var callbackCalled bool
	engine.ConfigCallback = func() tea.Cmd {
		callbackCalled = true
		return nil
	}

	// Transiciona para modo command via ':'
	engine.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	if engine.State.CurrentMode != ModeCommand {
		t.Fatalf("Esperado modo Command, obtido: %v", engine.State.CurrentMode)
	}

	// Digita 'c'
	engine.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if engine.State.CommandInput != "c" {
		t.Fatalf("Esperado CommandInput 'c', obtido: '%s'", engine.State.CommandInput)
	}

	// Pressiona Enter
	engine.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if !callbackCalled {
		t.Error("ConfigCallback não foi acionado ao pressionar Enter em ':c'")
	}
	if engine.State.CurrentMode != ModeNormal {
		t.Errorf("Esperado retorno ao ModeNormal, obtido: %v", engine.State.CurrentMode)
	}
	if engine.State.CommandInput != "" {
		t.Errorf("Esperado CommandInput vazio após execução, obtido: '%s'", engine.State.CommandInput)
	}
}
