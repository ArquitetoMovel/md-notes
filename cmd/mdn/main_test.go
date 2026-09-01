package main_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	main "md-notes/cmd/mdn"
)

func runWithInput(args []string, stdin io.Reader, inputBytes []byte) int {
	pr, pw := io.Pipe()
	go func() {
		time.Sleep(10 * time.Millisecond)
		_, _ = pw.Write(inputBytes)
		time.Sleep(10 * time.Millisecond)
		_ = pw.Close()
	}()

	return main.Execute(
		args,
		stdin,
		io.Discard,
		io.Discard,
		tea.WithInput(pr),
		tea.WithoutRenderer(),
	)
}

// TC13 — Invocação CLI com Arquivo Existente
func TestCLI_ExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "doc.md")
	_ = os.WriteFile(filePath, []byte("# Header\n"), 0644)

	code := runWithInput([]string{"mdn", filePath}, nil, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TC14 — Invocação CLI com Novo Arquivo
func TestCLI_NewFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "novo.md")

	code := runWithInput([]string{"mdn", filePath}, nil, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TC15 — Invocação CLI sem Argumentos (Scratchpad)
func TestCLI_Scratchpad(t *testing.T) {
	code := runWithInput([]string{"mdn"}, nil, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TC16 — Invocação CLI via Pipe Stdin
func TestCLI_PipeStdin(t *testing.T) {
	pipeContent := strings.NewReader("# Pipe Header\nLinha 2\n")
	code := runWithInput([]string{"mdn"}, pipeContent, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// Erro de permissão encerra imediatamente antes de entrar no modo TUI com código de saída 1
func TestCLI_PermissionDenied_ExitsCode1(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sem_permissao.md")
	_ = os.WriteFile(filePath, []byte("conteudo secreto\n"), 0000)
	defer os.Chmod(filePath, 0644)

	var stderr bytes.Buffer
	code := main.Execute(
		[]string{"mdn", filePath},
		nil,
		io.Discard,
		&stderr,
		tea.WithoutRenderer(),
	)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(stderr.String(), "Permissão negada") {
		t.Fatalf("expected 'Permissão negada' on stderr, got: %q", stderr.String())
	}
}
