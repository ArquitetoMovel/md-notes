package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// SaveConfigFile serializes cfg into TOML format and atomically persists it to filePath.
// It ensures that parent directories are created with 0755 permissions, writes the content
// to a temporary file in the destination directory, synchronizes buffers with fsync,
// sets 0644 permissions, and performs an atomic rename to filePath.
// If any step fails, the temporary file is safely cleaned up.
func SaveConfigFile(filePath string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuração não pode ser nula")
	}
	if filePath == "" {
		return fmt.Errorf("caminho do arquivo não pode ser vazio")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("falha ao criar diretório '%s': %w", dir, err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("falha ao serializar configuração para TOML: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp.config-*")
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo temporário em '%s': %w", dir, err)
	}

	tmpPath := tmpFile.Name()
	success := false
	defer func() {
		if !success {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("falha ao escrever no arquivo temporário '%s': %w", tmpPath, err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("falha ao sincronizar arquivo temporário '%s': %w", tmpPath, err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("falha ao fechar arquivo temporário '%s': %w", tmpPath, err)
	}

	if err := os.Chmod(tmpPath, 0644); err != nil {
		return fmt.Errorf("falha ao definir permissões no arquivo temporário '%s': %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("falha ao renomear arquivo temporário para '%s': %w", filePath, err)
	}

	success = true
	return nil
}
