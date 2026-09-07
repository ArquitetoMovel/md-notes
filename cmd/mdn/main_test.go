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

// TC13 (F01) — Invocação CLI com Arquivo Existente
func TestCLI_ExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "doc.md")
	_ = os.WriteFile(filePath, []byte("# Header\n"), 0644)

	code := runWithInput([]string{"mdn", filePath}, nil, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TC14 (F01) — Invocação CLI com Novo Arquivo
func TestCLI_NewFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "novo.md")

	code := runWithInput([]string{"mdn", filePath}, nil, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TC15 (F01) — Invocação CLI sem Argumentos (Scratchpad)
func TestCLI_Scratchpad(t *testing.T) {
	code := runWithInput([]string{"mdn"}, nil, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TC16 (F01) — Invocação CLI via Pipe Stdin
func TestCLI_PipeStdin(t *testing.T) {
	pipeContent := strings.NewReader("# Pipe Header\nLinha 2\n")
	code := runWithInput([]string{"mdn"}, pipeContent, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TC14 (F02) — Inicialização E2E com Geração de Configuração
func TestCLI_E2E_ConfigGenerated(t *testing.T) {
	tempDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", tempDir)

	code := runWithInput([]string{"mdn"}, nil, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	expectedConfigPath := filepath.Join(tempDir, "md-notes", "config.toml")
	data, err := os.ReadFile(expectedConfigPath)
	if err != nil {
		t.Fatalf("Arquivo de configuração não foi gerado automaticamente em '%s': %v", expectedConfigPath, err)
	}

	if !strings.Contains(string(data), "theme = \"default-dark\"") {
		t.Errorf("Arquivo de configuração gerado não contém theme padrão: %s", string(data))
	}
}

// TC15 (F02) — Inicialização E2E com TOML Corrompido
func TestCLI_E2E_CorruptedConfig_Fallback(t *testing.T) {
	tempDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Join(tempDir, "md-notes")
	_ = os.MkdirAll(configDir, 0755)
	corruptedFile := filepath.Join(configDir, "config.toml")
	_ = os.WriteFile(corruptedFile, []byte("theme = [invalido_sem_fechar\n"), 0644)

	code := runWithInput([]string{"mdn"}, nil, []byte{3})
	if code != 0 {
		t.Fatalf("expected exit code 0 com fallback resiliente, got %d", code)
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

func TestCLI_VersionFlag(t *testing.T) {
	for _, flag := range []string{"-v", "--version", "version"} {
		var stdout bytes.Buffer
		code := main.Execute(
			[]string{"mdn", flag},
			nil,
			&stdout,
			io.Discard,
		)

		if code != 0 {
			t.Fatalf("expected exit code 0 for flag %s, got %d", flag, code)
		}

		if !strings.Contains(stdout.String(), "mdn version") {
			t.Errorf("expected version output for flag %s, got %q", flag, stdout.String())
		}
	}
}

func TestCLI_HelpFlag(t *testing.T) {
	for _, flag := range []string{"-h", "--help", "help"} {
		var stdout bytes.Buffer
		code := main.Execute(
			[]string{"mdn", flag},
			nil,
			&stdout,
			io.Discard,
		)

		if code != 0 {
			t.Fatalf("expected exit code 0 for flag %s, got %d", flag, code)
		}

		if !strings.Contains(stdout.String(), "Usage:") {
			t.Errorf("expected help output for flag %s, got %q", flag, stdout.String())
		}
	}
}
