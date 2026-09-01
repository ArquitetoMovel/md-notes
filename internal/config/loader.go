package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// DefaultConfigContent is the commented default configuration written to config.toml on first run.
const DefaultConfigContent = `# md-notes - Arquivo de Configuração do Usuário
# Localização padrão: ~/.config/md-notes/config.toml (Linux/macOS) ou %APPDATA%/md-notes/config.toml (Windows)

# Tema visual ativo.
# Opções disponíveis: "default-dark", "default-light", "dracula", "nord", "catppuccin-mocha", "catppuccin-macchiato", "monokai", "auto"
theme = "default-dark"

[editor]
# Largura da tabulação em espaços (padrão: 4)
tab_size = 4

# Exibição de numeração de linhas na lateral esquerda
line_numbers = true

[colors]
# Sobreposição granular de cores (códigos Hex #RRGGBB ou ANSI 256). Descomente para customizar.
# h1 = "#BD93F9"
# h2 = "#8BE9FD"
# h3 = "#50FA7B"
# muted = "#6272A4"
# bold = "#FF79C6"
# italic = "#F1FA8C"
# code_bg = "#282A36"
# code_fg = "#F8F8F2"
# table_border = "#6272A4"
# table_header = "#BD93F9"
# status_bar_bg = "#44475A"
# status_bar_fg = "#F8F8F2"
# search_match_bg = "#F1FA8C"
# search_match_fg = "#282A36"
`

// EnsureDefaultConfigFile creates the parent directories and default config file if it does not exist.
func EnsureDefaultConfigFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return nil
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("falha ao criar diretório de configuração '%s': %w", dir, err)
	}

	if err := os.WriteFile(filePath, []byte(DefaultConfigContent), 0644); err != nil {
		return fmt.Errorf("falha ao escrever arquivo de configuração padrão '%s': %w", filePath, err)
	}

	return nil
}

// ParseTOML parses the raw TOML data into a Config structure, applying defaults for missing fields.
func ParseTOML(data []byte) (*Config, error) {
	cfg := DefaultConfig()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.Theme == "" {
		cfg.Theme = "default-dark"
	}
	if cfg.Editor.TabSize <= 0 {
		cfg.Editor.TabSize = 4
	}

	return cfg, nil
}

// LoadConfigFile loads and parses a configuration file from the specified path.
// If the file does not exist, it creates the default file.
// If the file contains syntax errors, it returns DefaultConfig() with non-blocking warnings.
func LoadConfigFile(filePath string) (*Config, []string, error) {
	var warnings []string

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if ensureErr := EnsureDefaultConfigFile(filePath); ensureErr != nil {
			warnings = append(warnings, fmt.Sprintf("Aviso: Não foi possível criar arquivo de configuração padrão: %v", ensureErr))
			return DefaultConfig(), warnings, nil
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Aviso: Não foi possível ler arquivo de configuração '%s': %v", filePath, err))
		return DefaultConfig(), warnings, nil
	}

	cfg, parseErr := ParseTOML(data)
	if parseErr != nil {
		warnings = append(warnings, fmt.Sprintf("Aviso: Erro de sintaxe no arquivo de configuração '%s': %v. Usando tema padrão.", filePath, parseErr))
		return DefaultConfig(), warnings, nil
	}

	return cfg, warnings, nil
}

// LoadConfig loads the configuration from standard XDG / AppData path.
func LoadConfig() (*Config, []string, error) {
	path, err := GetConfigFilePath()
	if err != nil {
		return DefaultConfig(), []string{fmt.Sprintf("Aviso: Falha ao obter caminho de configuração: %v", err)}, nil
	}
	return LoadConfigFile(path)
}
