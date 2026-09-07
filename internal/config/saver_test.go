package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TC14: Gravação atômica em config.toml e integridade de conteúdo
func TestSaveConfigFile_Success(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	cfg := &Config{
		Theme: "nord",
		Editor: EditorConfig{
			TabSize:             2,
			LineNumbers:         false,
			RelativeLineNumbers: true,
			Scrolloff:           6,
			WordWrap:            false,
		},
		Colors: ColorOverrides{
			H1:     "#88C0D0",
			CodeBg: "#2E3440",
		},
	}

	err := SaveConfigFile(configPath, cfg)
	if err != nil {
		t.Fatalf("SaveConfigFile falhou: %v", err)
	}

	// 1. Verifica se o arquivo de destino existe
	info, statErr := os.Stat(configPath)
	if statErr != nil {
		t.Fatalf("Arquivo de configuração não encontrado em disco: %v", statErr)
	}

	// 2. Verifica permissões no Unix/macOS (0644)
	if runtime.GOOS != "windows" {
		perm := info.Mode().Perm()
		if perm != 0644 {
			t.Errorf("Permissões esperadas 0644, obtido: %04o", perm)
		}
	}

	// 3. Verifica se não há arquivos temporários residuais
	entries, readErr := os.ReadDir(tempDir)
	if readErr != nil {
		t.Fatalf("Falha ao ler diretório temporário: %v", readErr)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp.config-") {
			t.Errorf("Arquivo temporário residual encontrado: %s", entry.Name())
		}
	}

	// 4. Recarrega o arquivo e valida integridade dos dados
	loadedCfg, warnings, loadErr := LoadConfigFile(configPath)
	if loadErr != nil {
		t.Fatalf("Falha ao recarregar configuração salva: %v", loadErr)
	}
	if len(warnings) > 0 {
		t.Errorf("Avisos inesperados ao carregar configuração: %v", warnings)
	}

	if loadedCfg.Theme != "nord" {
		t.Errorf("Esperado tema 'nord', obtido '%s'", loadedCfg.Theme)
	}
	if loadedCfg.Editor.TabSize != 2 {
		t.Errorf("Esperado TabSize 2, obtido %d", loadedCfg.Editor.TabSize)
	}
	if loadedCfg.Editor.LineNumbers {
		t.Errorf("Esperado LineNumbers false, obtido true")
	}
	if !loadedCfg.Editor.RelativeLineNumbers {
		t.Errorf("Esperado RelativeLineNumbers true, obtido false")
	}
	if loadedCfg.Editor.Scrolloff != 6 {
		t.Errorf("Esperado Scrolloff 6, obtido %d", loadedCfg.Editor.Scrolloff)
	}
	if loadedCfg.Editor.WordWrap {
		t.Errorf("Esperado WordWrap false, obtido true")
	}
	if loadedCfg.Colors.H1 != "#88C0D0" {
		t.Errorf("Esperado Colors.H1 '#88C0D0', obtido '%s'", loadedCfg.Colors.H1)
	}
	if loadedCfg.Colors.CodeBg != "#2E3440" {
		t.Errorf("Esperado Colors.CodeBg '#2E3440', obtido '%s'", loadedCfg.Colors.CodeBg)
	}
}

// Criação automática de diretório pai inexistente com 0755
func TestSaveConfigFile_CreateParentDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "subdir1", "subdir2", "config.toml")
	cfg := DefaultConfig()

	err := SaveConfigFile(nestedPath, cfg)
	if err != nil {
		t.Fatalf("SaveConfigFile falhou ao criar diretórios pais: %v", err)
	}

	if _, statErr := os.Stat(nestedPath); statErr != nil {
		t.Fatalf("Arquivo não foi criado no caminho aninhado: %v", statErr)
	}

	loadedCfg, _, loadErr := LoadConfigFile(nestedPath)
	if loadErr != nil {
		t.Fatalf("Falha ao recarregar configuração de caminho aninhado: %v", loadErr)
	}
	if loadedCfg.Theme != "default-dark" {
		t.Errorf("Esperado tema padrão 'default-dark', obtido '%s'", loadedCfg.Theme)
	}
}

// Tratamento de erro de permissão negada
func TestSaveConfigFile_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Teste de permissão POSIX não aplicável ao Windows")
	}

	tempDir := t.TempDir()
	roDir := filepath.Join(tempDir, "readonly")
	if err := os.MkdirAll(roDir, 0755); err != nil {
		t.Fatalf("Falha ao criar diretório para teste: %v", err)
	}

	// Altera permissões para somente leitura (sem permissão de escrita)
	if err := os.Chmod(roDir, 0555); err != nil {
		t.Fatalf("Falha ao configurar permissão somente leitura: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(roDir, 0755)
	})

	cfg := DefaultConfig()
	targetPath := filepath.Join(roDir, "config.toml")
	err := SaveConfigFile(targetPath, cfg)
	if err == nil {
		t.Fatal("Esperava erro ao salvar em diretório somente leitura, mas retornou nil")
	}

	// Garante que nenhum arquivo foi persistido
	if _, statErr := os.Stat(targetPath); !os.IsNotExist(statErr) {
		t.Errorf("Arquivo de configuração não deveria existir em diretório sem escrita")
	}
}

// Sobrescrita atômica segura de configuração existente
func TestSaveConfigFile_OverwriteExisting(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	initialCfg := &Config{
		Theme:  "default-dark",
		Editor: EditorConfig{TabSize: 4},
	}
	if err := SaveConfigFile(configPath, initialCfg); err != nil {
		t.Fatalf("Falha na gravação inicial: %v", err)
	}

	updatedCfg := &Config{
		Theme:  "dracula",
		Editor: EditorConfig{TabSize: 8},
	}
	if err := SaveConfigFile(configPath, updatedCfg); err != nil {
		t.Fatalf("Falha na sobrescrita: %v", err)
	}

	loaded, _, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("Falha ao ler arquivo sobrescrito: %v", err)
	}
	if loaded.Theme != "dracula" || loaded.Editor.TabSize != 8 {
		t.Errorf("Valores atualizados não correspondem: theme=%s, tab_size=%d", loaded.Theme, loaded.Editor.TabSize)
	}
}

// Validação de parâmetros inválidos (nil config e path vazio)
func TestSaveConfigFile_InvalidParameters(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	if err := SaveConfigFile(configPath, nil); err == nil {
		t.Error("Esperava erro para cfg nil, retornou nil")
	}

	if err := SaveConfigFile("", DefaultConfig()); err == nil {
		t.Error("Esperava erro para filePath vazio, retornou nil")
	}
}
