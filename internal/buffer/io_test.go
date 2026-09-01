package buffer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"md-notes/internal/buffer"
)

// TC04 — Leitura de Arquivo Existente e Detecção de Quebra de Linha
func TestIO_LoadExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "teste_crlf.md")
	content := "Linha 1 com acento: Olá!\r\nLinha 2\r\nLinha 3 sem terminador"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	buf, err := buffer.LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	if buf.IsNewFile {
		t.Fatalf("expected IsNewFile == false")
	}
	if buf.IsDirty {
		t.Fatalf("expected IsDirty == false")
	}
	if buf.LineCount() != 3 {
		t.Fatalf("expected 3 lines, got %d", buf.LineCount())
	}

	l1, _ := buf.GetLine(0)
	if l1.String() != "Linha 1 com acento: Olá!" || l1.Ending != buffer.EndingCRLF {
		t.Fatalf("line 1 mismatch: got %q, ending %s", l1.String(), l1.Ending)
	}

	l3, _ := buf.GetLine(2)
	if l3.String() != "Linha 3 sem terminador" || l3.Ending != buffer.EndingNone {
		t.Fatalf("line 3 mismatch: got %q, ending %s", l3.String(), l3.Ending)
	}
}

// TC05 — Abertura de Caminho Inexistente como Novo Arquivo
func TestIO_LoadNonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "novo_arquivo.md")

	buf, err := buffer.LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error loading non-existent file: %v", err)
	}

	if !buf.IsNewFile {
		t.Fatalf("expected IsNewFile == true")
	}
	if buf.IsDirty {
		t.Fatalf("expected IsDirty == false")
	}
	if buf.FilePath != filePath {
		t.Fatalf("expected FilePath %q, got %q", filePath, buf.FilePath)
	}
	if buf.LineCount() != 1 {
		t.Fatalf("expected 1 empty line, got %d", buf.LineCount())
	}
}

// TC06 — Leitura de Dados via Reader / Pipe Stdin
func TestIO_LoadFromReader(t *testing.T) {
	input := "# Pipe Header\nConteúdo enviado via pipe\n"
	buf, err := buffer.LoadFromReader(strings.NewReader(input), "[Stdin Buffer]")
	if err != nil {
		t.Fatalf("failed to load from reader: %v", err)
	}

	if !buf.IsStdinBuffer {
		t.Fatalf("expected IsStdinBuffer == true")
	}
	if buf.FilePath != "[Stdin Buffer]" {
		t.Fatalf("expected FilePath == '[Stdin Buffer]', got %q", buf.FilePath)
	}
	if buf.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", buf.LineCount())
	}
	if buf.IsDirty {
		t.Fatalf("expected IsDirty == false")
	}
}

// TC07 — Validação de Erro ao Salvar Buffer sem Caminho
func TestIO_SaveWithoutPath_Error(t *testing.T) {
	buf := buffer.NewScratchpadBuffer()
	buf.InsertText("Nota sem nome")

	err := buffer.AtomicSave(buf, "")
	if err == nil {
		t.Fatalf("expected error when saving buffer without path, got nil")
	}
	if !strings.Contains(err.Error(), "Nenhum nome de arquivo definido") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// TC08 — Gravação Atômica e Preservação de Permissões
func TestIO_AtomicSave_PreservesPermissions(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "script.sh")

	// Create initial file with 0755
	if err := os.WriteFile(filePath, []byte("#!/bin/sh\necho hi\n"), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	buf, err := buffer.LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	buf.InsertNewLine()
	buf.InsertText("echo updated")

	if err := buffer.AtomicSave(buf, filePath); err != nil {
		t.Fatalf("failed to atomic save: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat saved file: %v", err)
	}

	// Verify permissions preserved (0755)
	if info.Mode().Perm() != 0755 {
		t.Fatalf("expected file permission 0755, got %o", info.Mode().Perm())
	}

	// Verify content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if !strings.Contains(string(data), "echo updated") {
		t.Fatalf("saved file missing updated content: %q", string(data))
	}
}

// TC09 — Criação Automática de Diretórios Pais ao Salvar (Full Scope)
func TestIO_AtomicSave_CreatesParentDirs(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "sub1", "sub2", "nested_doc.md")

	buf := buffer.NewEmptyBuffer(nestedPath)
	buf.InsertText("# Nested Document\nCreated with auto mkdir")

	if err := buffer.AtomicSave(buf, nestedPath); err != nil {
		t.Fatalf("failed to atomic save with nested dirs: %v", err)
	}

	if _, err := os.Stat(nestedPath); err != nil {
		t.Fatalf("nested file does not exist after save: %v", err)
	}

	if buf.IsDirty {
		t.Fatalf("expected IsDirty == false after save")
	}
}
