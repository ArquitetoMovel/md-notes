package config

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TC01: Resolução de Diretórios Padrão XDG / AppData
func TestConfig_XDGPathResolution(t *testing.T) {
	path, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("GetConfigFilePath() retornou erro inesperado: %v", err)
	}

	if runtime.GOOS == "windows" {
		expectedSuffix := filepath.Join("md-notes", "config.toml")
		if !strings.HasSuffix(path, expectedSuffix) {
			t.Errorf("Caminho esperado com sufixo '%s', obtido: '%s'", expectedSuffix, path)
		}
	} else {
		expectedSuffix := filepath.Join(".config", "md-notes", "config.toml")
		if !strings.HasSuffix(path, expectedSuffix) {
			t.Errorf("Caminho esperado com sufixo '%s', obtido: '%s'", expectedSuffix, path)
		}
	}
}

// TC02: Validação de Estrutura e Valores Padrão da Configuração
func TestConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() retornou nil")
	}

	if cfg.Theme != "default-dark" {
		t.Errorf("Esperado tema 'default-dark', obtido: '%s'", cfg.Theme)
	}

	if cfg.Editor.TabSize != 4 {
		t.Errorf("Esperado TabSize 4, obtido: %d", cfg.Editor.TabSize)
	}

	if !cfg.Editor.LineNumbers {
		t.Errorf("Esperado LineNumbers == true, obtido: %v", cfg.Editor.LineNumbers)
	}
}
