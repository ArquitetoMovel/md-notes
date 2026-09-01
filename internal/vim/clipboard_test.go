package vim

import (
	"testing"
)

// TC11: Sincronização com o Clipboard do Sistema Operacional e Registrador Local
func TestClipboard_SetAndGet(t *testing.T) {
	cb := NewClipboard()
	if cb == nil {
		t.Fatal("NewClipboard() retornou nil")
	}

	testText := "Texto para cópia e colagem\n"
	cb.Set(testText, true)

	content, isLineWise := cb.Get()
	if content != testText {
		t.Errorf("Esperado '%s', obtido '%s'", testText, content)
	}
	if !isLineWise {
		t.Error("Esperado isLineWise == true")
	}

	reg := cb.GetRegister()
	if reg.Content != testText || !reg.IsLineWise {
		t.Errorf("Registrador local inconsistente: %+v", reg)
	}
}
