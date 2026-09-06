package app_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/app"
	"md-notes/internal/buffer"
	"md-notes/internal/config"
	"md-notes/internal/vim"
)

// TC02: Rejeição de Abertura em Janela Insuficiente (< 50x14)
func TestApp_SettingsSmallTerminal(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)

	// Define dimensões insuficientes (48x12 < 50x14)
	m.Width = 48
	m.Height = 12

	// 1. Tenta abrir via mensagem OpenSettingsMsg
	m.Update(app.OpenSettingsMsg{})
	if m.SettingsOpen {
		t.Error("Painel não deveria abrir com terminal 48x12")
	}
	expectedWarning := "Dimensão insuficiente (mínimo 50x14) para abrir o painel de configurações"
	if m.StatusMsg != expectedWarning {
		t.Errorf("StatusMsg esperado '%s', obtido '%s'", expectedWarning, m.StatusMsg)
	}

	// 2. Tenta abrir via atalho global F2
	m.StatusMsg = ""
	m.Update(tea.KeyMsg{Type: tea.KeyF2})
	if m.SettingsOpen {
		t.Error("Painel não deveria abrir via F2 com terminal insuficiente")
	}
	if m.StatusMsg != expectedWarning {
		t.Errorf("StatusMsg esperado '%s' via F2, obtido '%s'", expectedWarning, m.StatusMsg)
	}
}

// TC03: Alternância de Abertura via Atalho Global F2
func TestApp_SettingsF2Shortcut(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)
	m.Width = 80
	m.Height = 24

	if m.SettingsOpen {
		t.Fatal("SettingsOpen deveria iniciar false")
	}

	// 1º toque em F2 abre o modal
	m.Update(tea.KeyMsg{Type: tea.KeyF2})
	if !m.SettingsOpen {
		t.Fatal("Esperado SettingsOpen == true após 1º toque em F2")
	}

	// 2º toque em F2 fecha o modal
	m.Update(tea.KeyMsg{Type: tea.KeyF2})
	if m.SettingsOpen {
		t.Fatal("Esperado SettingsOpen == false após 2º toque em F2")
	}
}

// TC01: Execução dos Comandos Ex :c e :config no Modelo
func TestApp_SettingsExCommand(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)
	m.Width = 80
	m.Height = 24

	// 1. Testa execução de :c
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}
	if !m.SettingsOpen {
		t.Errorf("Esperado SettingsOpen == true após comando :c")
	}

	// Fecha o painel
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.SettingsOpen {
		t.Fatal("Esperado SettingsOpen == false após Esc")
	}

	// 2. Testa execução de :config
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		msg := cmd()
		m.Update(msg)
	}
	if !m.SettingsOpen {
		t.Errorf("Esperado SettingsOpen == true após comando :config")
	}
}

// TC12 & TC13: Live Preview Dinâmico de Tema e Cancelamento com Restauração Fiel
func TestApp_SettingsLivePreviewAndCancel(t *testing.T) {
	buf := buffer.NewEmptyBuffer("preview.md")
	buf.InsertText("# Título Principal\nLinha de texto normal.")

	cfg := config.DefaultConfig()
	cfg.Theme = "default-dark"
	cfg.Editor.LineNumbers = true

	m := app.NewModel(buf, nil, app.WithConfig(cfg))
	m.Width = 80
	m.Height = 24

	// 1. Abre o painel
	m.Update(app.OpenSettingsMsg{})
	if !m.SettingsOpen {
		t.Fatal("Falha ao abrir o painel de configurações")
	}

	// 2. Navega na lista de temas e seleciona "dracula"
	// Encontra índice de "dracula"
	draculaIdx := -1
	for i, name := range m.SettingsState.AvailableThemes {
		if name == "dracula" {
			draculaIdx = i
			break
		}
	}
	if draculaIdx < 0 {
		t.Fatal("Tema 'dracula' não encontrado entre os disponíveis")
	}

	m.SettingsState.ActiveThemeIndex = draculaIdx
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verifica Live Preview imediato no modelo
	if m.Config.Theme != "dracula" {
		t.Errorf("Esperado m.Config.Theme == 'dracula' após live preview, obtido '%s'", m.Config.Theme)
	}
	if m.Theme.Palette.Name != "dracula" {
		t.Errorf("Esperado m.Theme.Palette.Name == 'dracula', obtido '%s'", m.Theme.Palette.Name)
	}

	// Alterna line_numbers para false na aba Editor
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m.SettingsState.ActiveEditorRow = 0 // line_numbers
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.Config.Editor.LineNumbers {
		t.Errorf("Esperado line_numbers == false após alternar")
	}

	// Renderiza frame com View()
	view := m.View()
	if !strings.Contains(view, "Configurações") {
		t.Errorf("View não contém modal de configurações")
	}

	// 3. Pressiona Esc para cancelar (TC13)
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.SettingsOpen {
		t.Errorf("Esperado SettingsOpen == false após cancelamento")
	}
	if m.Config.Theme != "default-dark" {
		t.Errorf("Esperado restauração de m.Config.Theme para 'default-dark', obtido '%s'", m.Config.Theme)
	}
	if !m.Config.Editor.LineNumbers {
		t.Errorf("Esperado restauração de line_numbers para true")
	}
	if m.Theme.Palette.Name != "default-dark" {
		t.Errorf("Esperado restauração do tema compilado para 'default-dark', obtido '%s'", m.Theme.Palette.Name)
	}
}

// TC14 & TC15: Persistência Atômica no Arquivo config.toml e Coordenação com Watcher
func TestApp_SettingsSaveAndPersist(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	buf := buffer.NewEmptyBuffer("notas.md")
	cfg := config.DefaultConfig()
	cfg.Theme = "default-dark"
	cfg.Editor.TabSize = 4

	cw, watcherErr := config.NewWatcher()
	if watcherErr != nil {
		t.Fatalf("Falha ao criar watcher: %v", watcherErr)
	}
	defer cw.Close()

	m := app.NewModel(buf, nil, app.WithConfig(cfg), app.WithConfigWatcher(cw))
	m.Width = 80
	m.Height = 24

	// Abre o painel
	m.Update(app.OpenSettingsMsg{})

	// Altera tema para "nord"
	m.SettingsState.WorkingConfig.Theme = "nord"
	m.SettingsState.WorkingConfig.Editor.TabSize = 2

	// Dispara salvamento
	m.Update(app.CloseSettingsMsg{Save: true})

	if m.SettingsOpen {
		t.Errorf("Esperado painel fechado após salvar com sucesso")
	}
	if m.StatusMsg != "Configurações salvas em config.toml" {
		t.Errorf("Mensagem esperada 'Configurações salvas em config.toml', obtido '%s'", m.StatusMsg)
	}
	if m.Config.Theme != "nord" {
		t.Errorf("Esperado tema 'nord' ativo no modelo, obtido '%s'", m.Config.Theme)
	}
	if m.Config.Editor.TabSize != 2 {
		t.Errorf("Esperado tab_size 2 ativo no modelo, obtido %d", m.Config.Editor.TabSize)
	}

	// Valida arquivo persistido no disco
	cfgPath, pathErr := config.GetConfigFilePath()
	if pathErr != nil {
		t.Fatalf("Erro ao obter caminho de configuração: %v", pathErr)
	}

	loaded, warnings, loadErr := config.LoadConfigFile(cfgPath)
	if loadErr != nil {
		t.Fatalf("Falha ao carregar arquivo persistido: %v", loadErr)
	}
	if len(warnings) > 0 {
		t.Errorf("Avisos inesperados ao ler arquivo: %v", warnings)
	}
	if loaded.Theme != "nord" {
		t.Errorf("Arquivo salvo esperado theme 'nord', obtido '%s'", loaded.Theme)
	}
	if loaded.Editor.TabSize != 2 {
		t.Errorf("Arquivo salvo esperado tab_size 2, obtido %d", loaded.Editor.TabSize)
	}
}

// TC17: Bloqueio Estrito de Propagação de Teclas para o Buffer
func TestApp_SettingsKeyBlocking(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.InsertText("Linha original 1\nLinha original 2")

	m := app.NewModel(buf, nil)
	m.Width = 80
	m.Height = 24

	// Abre o painel de configurações
	m.Update(app.OpenSettingsMsg{})
	if !m.SettingsOpen {
		t.Fatal("Esperava modal aberto")
	}

	// Envia teclas normais do Vim que modificariam o buffer caso chegassem até ele
	keys := []string{"i", "x", "d", "o", "a"}
	for _, k := range keys {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}

	// Verifica se o texto do buffer permaneceu estritamente inalterado
	l0, _ := buf.GetLine(0)
	l1, _ := buf.GetLine(1)
	if l0.String() != "Linha original 1" || l1.String() != "Linha original 2" {
		t.Errorf("Buffer foi corrompido durante modal aberto: l0='%s', l1='%s'", l0.String(), l1.String())
	}

	// Verifica se o modo Vim permanece Normal (não transitou para Insert)
	if m.Vim.State.CurrentMode != vim.ModeNormal {
		t.Errorf("Modo Vim alterado indevidamente para: %v", m.Vim.State.CurrentMode)
	}
}

// TC16: Tratamento de Erro de Permissão sem Travar a Sessão
func TestApp_SettingsPersistenceError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Teste de permissão POSIX não aplicável ao Windows")
	}

	tempDir := t.TempDir()
	roDir := filepath.Join(tempDir, "readonly_config")
	if err := os.MkdirAll(roDir, 0755); err != nil {
		t.Fatalf("Falha ao criar diretório temporário: %v", err)
	}
	if err := os.Chmod(roDir, 0555); err != nil {
		t.Fatalf("Falha ao configurar permissão 0555: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(roDir, 0755)
	})

	t.Setenv("XDG_CONFIG_HOME", roDir)

	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)
	m.Width = 80
	m.Height = 24

	// Abre o painel
	m.Update(app.OpenSettingsMsg{})

	// Dispara gravação em diretório protegido
	m.Update(app.CloseSettingsMsg{Save: true})

	// O modal deve permanecer aberto exibindo a falha sem pânico
	if !m.SettingsOpen {
		t.Error("Modal deveria permanecer aberto após erro de gravação")
	}
	if m.SettingsState.PersistenceError == "" {
		t.Error("PersistenceError deveria conter a mensagem de erro do disco")
	}

	// Renderiza a tela e verifica se o alerta visual é exibido
	view := m.View()
	if !strings.Contains(view, "Erro ao gravar config.toml:") {
		t.Errorf("View deveria renderizar o alerta 'Erro ao gravar config.toml:', obtido:\n%s", view)
	}
}
