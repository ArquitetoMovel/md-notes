package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TC13: Hot-Reload Dinâmico de Tema via fsnotify (Full Scope)
func TestWatcher_HotReload_ThemeChange(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	initialTOML := `theme = "dracula"

[editor]
tab_size = 4
`
	if err := os.WriteFile(configPath, []byte(initialTOML), 0644); err != nil {
		t.Fatalf("Falha ao criar config inicial: %v", err)
	}

	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher() falhou: %v", err)
	}
	defer w.Close()

	var receivedMsg ThemeReloadedMsg
	var msgMu sync.Mutex
	msgReceived := make(chan struct{}, 5)

	sendCallback := func(msg tea.Msg) {
		if trMsg, ok := msg.(ThemeReloadedMsg); ok {
			msgMu.Lock()
			receivedMsg = trMsg
			msgMu.Unlock()
			select {
			case msgReceived <- struct{}{}:
			default:
			}
		}
	}

	if err := w.Start(configPath, sendCallback); err != nil {
		t.Fatalf("w.Start() falhou: %v", err)
	}

	// Wait briefly for watcher setup
	time.Sleep(50 * time.Millisecond)

	// Step 1: Overwrite config.toml changing theme to "nord"
	updatedTOML := `theme = "nord"

[editor]
tab_size = 2
`
	if err := os.WriteFile(configPath, []byte(updatedTOML), 0644); err != nil {
		t.Fatalf("Falha ao atualizar config: %v", err)
	}

	// Step 2: Await ThemeReloadedMsg dispatched after debounce
	select {
	case <-msgReceived:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout aguardando evento ThemeReloadedMsg")
	}

	// Step 3: Inspect active CompiledTheme and Config
	msgMu.Lock()
	defer msgMu.Unlock()

	if receivedMsg.Config == nil {
		t.Fatal("Config recebido é nil")
	}
	if receivedMsg.Config.Theme != "nord" {
		t.Errorf("Esperado tema 'nord', obtido '%s'", receivedMsg.Config.Theme)
	}
	if receivedMsg.CompiledTheme == nil {
		t.Fatal("CompiledTheme recebido é nil")
	}
	if receivedMsg.CompiledTheme.Palette.Name != "nord" {
		t.Errorf("Esperado CompiledTheme com paleta 'nord', obtido '%s'", receivedMsg.CompiledTheme.Palette.Name)
	}
}

func TestWatcher_PauseAndResume(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	if err := os.WriteFile(configPath, []byte("theme = \"dracula\"\n"), 0644); err != nil {
		t.Fatalf("Falha ao gravar arquivo inicial: %v", err)
	}

	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher() falhou: %v", err)
	}
	defer w.Close()

	callCount := 0
	var mu sync.Mutex
	sendCallback := func(msg tea.Msg) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	if err := w.Start(configPath, sendCallback); err != nil {
		t.Fatalf("w.Start() falhou: %v", err)
	}

	w.Pause()

	// Modify file while paused
	if err := os.WriteFile(configPath, []byte("theme = \"monokai\"\n"), 0644); err != nil {
		t.Fatalf("Falha ao modificar arquivo pausado: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	if callCount != 0 {
		t.Errorf("Watcher disparou callback enquanto pausado: %d chamadas", callCount)
	}
	mu.Unlock()

	w.Resume()

	// Modify file after resume
	if err := os.WriteFile(configPath, []byte("theme = \"catppuccin-mocha\"\n"), 0644); err != nil {
		t.Fatalf("Falha ao modificar arquivo resumido: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if callCount == 0 {
		t.Error("Watcher não disparou callback após retomar (Resume)")
	}
	mu.Unlock()
}
