package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TC07: Carregamento de Arquivo config.toml Existente
func TestLoader_ParseValidTOML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	tomlContent := `theme = "nord"

[editor]
tab_size = 2
line_numbers = false

[colors]
h1 = "#88C0D0"
code_bg = "#2E3440"
`
	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("Falha ao preparar arquivo TOML de teste: %v", err)
	}

	cfg, warnings, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile retornou erro inesperado: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("Não esperava warnings para TOML válido, obteve: %v", warnings)
	}

	if cfg.Theme != "nord" {
		t.Errorf("Esperado tema 'nord', obtido: '%s'", cfg.Theme)
	}
	if cfg.Editor.TabSize != 2 {
		t.Errorf("Esperado TabSize 2, obtido: %d", cfg.Editor.TabSize)
	}
	if cfg.Editor.LineNumbers {
		t.Errorf("Esperado LineNumbers == false, obtido: %v", cfg.Editor.LineNumbers)
	}
	if cfg.Colors.H1 != "#88C0D0" {
		t.Errorf("Esperado Colors.H1 == '#88C0D0', obtido: '%s'", cfg.Colors.H1)
	}
	if cfg.Colors.CodeBg != "#2E3440" {
		t.Errorf("Esperado Colors.CodeBg == '#2E3440', obtido: '%s'", cfg.Colors.CodeBg)
	}
}

// TC08: Criação Automática do Arquivo Padrão Documentado
func TestLoader_EnsureDefaultFileCreated(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", "config.toml")

	err := EnsureDefaultConfigFile(configPath)
	if err != nil {
		t.Fatalf("EnsureDefaultConfigFile retornou erro inesperado: %v", err)
	}

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("Arquivo não foi criado em disco: %v", readErr)
	}

	content := string(data)
	if !strings.Contains(content, "theme = \"default-dark\"") {
		t.Errorf("Conteúdo padrão não contém 'theme = \"default-dark\"': %s", content)
	}
	if !strings.Contains(content, "tab_size = 4") {
		t.Errorf("Conteúdo padrão não contém 'tab_size = 4': %s", content)
	}
	if !strings.Contains(content, "dracula") || !strings.Contains(content, "catppuccin-mocha") {
		t.Errorf("Conteúdo padrão não documenta os temas disponíveis: %s", content)
	}

	// Calling again on existing file should be a no-op
	if err := EnsureDefaultConfigFile(configPath); err != nil {
		t.Errorf("Segunda chamada a EnsureDefaultConfigFile falhou: %v", err)
	}
}

// TC11: Resiliência a Erros de Sintaxe TOML com Fallback Seguro
func TestLoader_ParseSyntaxError_Fallback(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "corrupted_config.toml")

	corruptedTOML := `theme = [invalido_sem_fechar
[editor
tab_size = "não é número"
`
	if err := os.WriteFile(configPath, []byte(corruptedTOML), 0644); err != nil {
		t.Fatalf("Falha ao gravar TOML corrompido: %v", err)
	}

	cfg, warnings, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile não deveria retornar erro fatal para TOML corrompido, retornou: %v", err)
	}

	if cfg == nil {
		t.Fatal("Config retornado é nil")
	}

	if cfg.Theme != "default-dark" {
		t.Errorf("Esperado fallback para 'default-dark', obtido: '%s'", cfg.Theme)
	}
	if cfg.Editor.TabSize != 4 {
		t.Errorf("Esperado fallback TabSize 4, obtido: %d", cfg.Editor.TabSize)
	}

	if len(warnings) == 0 {
		t.Error("Esperava emissão de aviso não bloqueante para erro de sintaxe")
	} else if !strings.Contains(warnings[0], "Aviso: Erro de sintaxe") {
		t.Errorf("Mensagem de warning não esperada: %s", warnings[0])
	}
}
