package vim

import (
	"fmt"
	"testing"

	"md-notes/internal/buffer"
)

// TC10: Pilha de Histórico de Desfazer (u) e Refazer (Ctrl+r)
func TestHistory_UndoRedoBasic(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	h := NewHistory(200)

	// Initial baseline snapshot
	h.Push(buf)

	// Mutation 1
	buf.InsertText("Primeira alteração")
	h.Push(buf)

	// Mutation 2
	buf.InsertText(" e Segunda alteração")
	h.Push(buf)

	if h.Count() != 3 {
		t.Fatalf("Esperado 3 snapshots, obtido %d", h.Count())
	}

	// Undo to Mutation 1
	msg, err := h.Undo(buf)
	if err != nil || msg != "" {
		t.Fatalf("Undo falhou: %v, %s", err, msg)
	}
	line, _ := buf.GetLine(0)
	if line.String() != "Primeira alteração" {
		t.Errorf("Esperado 'Primeira alteração', obtido '%s'", line.String())
	}

	// Undo to Initial
	msg, err = h.Undo(buf)
	if err != nil || msg != "" {
		t.Fatalf("Undo falhou: %v, %s", err, msg)
	}
	line, _ = buf.GetLine(0)
	if line.String() != "" {
		t.Errorf("Esperado '', obtido '%s'", line.String())
	}

	// Redo back to Mutation 1
	msg, err = h.Redo(buf)
	if err != nil || msg != "" {
		t.Fatalf("Redo falhou: %v, %s", err, msg)
	}
	line, _ = buf.GetLine(0)
	if line.String() != "Primeira alteração" {
		t.Errorf("Esperado 'Primeira alteração' após redo, obtido '%s'", line.String())
	}

	// Redo back to Mutation 2
	msg, err = h.Redo(buf)
	if err != nil || msg != "" {
		t.Fatalf("Redo falhou: %v, %s", err, msg)
	}
	line, _ = buf.GetLine(0)
	if line.String() != "Primeira alteração e Segunda alteração" {
		t.Errorf("Esperado 'Primeira alteração e Segunda alteração' após redo, obtido '%s'", line.String())
	}
}

// TC17: Notificação de Limites da Pilha de Undo / Redo
func TestHistory_BoundaryLimits(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	h := NewHistory(200)

	h.Push(buf)

	// Undo on oldest change
	msg, _ := h.Undo(buf)
	if msg != "Já na alteração mais antiga" {
		t.Errorf("Esperado aviso 'Já na alteração mais antiga', obtido '%s'", msg)
	}

	// Redo on newest change
	msg, _ = h.Redo(buf)
	if msg != "Já na alteração mais recente" {
		t.Errorf("Esperado aviso 'Já na alteração mais recente', obtido '%s'", msg)
	}
}

// TC13: Descarte Circular de Histórico no Limite de 200 Snapshots
func TestHistory_Max200Limit(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	h := NewHistory(200)

	for i := 0; i < 250; i++ {
		buf.SetLine(0, buffer.NewLine(fmt.Sprintf("Linha %d", i), buffer.EndingLF))
		h.Push(buf)
	}

	if h.Count() != 200 {
		t.Errorf("Esperado exatamente 200 snapshots, obtido %d", h.Count())
	}

	// The newest snapshot should be Linha 249
	line, _ := buf.GetLine(0)
	if line.String() != "Linha 249" {
		t.Errorf("Esperado estado 'Linha 249', obtido '%s'", line.String())
	}

	// Undo 199 times should reach the oldest preserved state (Linha 50)
	for i := 0; i < 199; i++ {
		_, _ = h.Undo(buf)
	}

	line, _ = buf.GetLine(0)
	if line.String() != "Linha 50" {
		t.Errorf("Esperado estado 'Linha 50' após desfazer até o limite, obtido '%s'", line.String())
	}

	// 200th undo should hit boundary
	msg, _ := h.Undo(buf)
	if msg != "Já na alteração mais antiga" {
		t.Errorf("Esperado atingir limite de histórico, obtido '%s'", msg)
	}
}
