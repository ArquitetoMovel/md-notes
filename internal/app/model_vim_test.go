package app_test

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/app"
	"md-notes/internal/buffer"
	"md-notes/internal/vim"
)

func makeKeyMsg(str string) tea.KeyMsg {
	if str == "esc" {
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	if str == "enter" {
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	if str == "backspace" {
		return tea.KeyMsg{Type: tea.KeyBackspace}
	}
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(str),
	}
}

// Test Bubble Tea key delegation to Vim Engine (Insert Mode, View indicator)
func TestApp_VimKeyDelegation_InsertMode(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)

	// Press 'i' to enter insert mode
	m.Update(makeKeyMsg("i"))
	view := m.View()
	if !strings.Contains(view, "-- INSERT --") {
		t.Errorf("View esperado com '-- INSERT --', obtido:\n%s", view)
	}

	// Type text
	for _, ch := range "Hello Vim" {
		m.Update(makeKeyMsg(string(ch)))
	}

	line0, _ := buf.GetLine(0)
	if line0.String() != "Hello Vim" {
		t.Errorf("Buffer esperado 'Hello Vim', obtido '%s'", line0.String())
	}

	// Press 'Esc' to return to normal mode
	m.Update(makeKeyMsg("esc"))
	if m.Vim.State.CurrentMode != vim.ModeNormal {
		t.Errorf("Esperado retorno para ModeNormal, obtido %s", m.Vim.State.CurrentMode)
	}
}

// Test Ex Command prompt rendering and execution in Bubble Tea
func TestApp_VimCommandMode_SaveAndQuit(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "notas_salvas.md")

	buf := buffer.NewEmptyBuffer(filePath)
	buf.InsertText("Linha gravada via :w")
	m := app.NewModel(buf, nil)

	// Press ':' to enter command mode
	m.Update(makeKeyMsg(":"))
	view := m.View()
	if !strings.Contains(view, "\n:") {
		t.Errorf("View esperado com prompt ':', obtido:\n%s", view)
	}

	// Type "w" and press Enter
	m.Update(makeKeyMsg("w"))
	_, cmd := m.Update(makeKeyMsg("enter"))

	if cmd == nil {
		t.Fatal("Esperado tea.Cmd retornado para :w")
	}

	// Execute the returned cmd (which dispatches SaveFileMsg)
	saveMsg := cmd()
	_, saveResultCmd := m.Update(saveMsg)

	if saveResultCmd == nil {
		t.Fatal("Esperado SaveResultMsg disparado após salvar")
	}
	saveResult := saveResultCmd()
	m.Update(saveResult)

	if buf.IsDirty {
		t.Error("Buffer não deveria estar dirty após :w")
	}
}

// Test Ex command unknown error (E492)
func TestApp_VimCommandMode_UnknownCommand_E492(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)

	m.Update(makeKeyMsg(":"))
	for _, ch := range "comando_desconhecido" {
		m.Update(makeKeyMsg(string(ch)))
	}
	m.Update(makeKeyMsg("enter"))

	if !strings.Contains(m.StatusMsg, "E492: Não é um comando de editor: comando_desconhecido") {
		t.Errorf("StatusMsg esperado com erro E492, obtido: %q", m.StatusMsg)
	}
}
