package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"md-notes/internal/theme"
)

// ThemeReloadedMsg is dispatched whenever config.toml is modified and reloaded.
type ThemeReloadedMsg struct {
	Config        *Config
	CompiledTheme *theme.CompiledTheme
	Warning       string
}

// Watcher monitors changes in config.toml with debounce and dispatches ThemeReloadedMsg.
type Watcher struct {
	fsWatcher *fsnotify.Watcher
	filePath  string
	sendFunc  func(tea.Msg)
	stopChan  chan struct{}
	wg        sync.WaitGroup
	mu        sync.Mutex
	paused    bool
}

// NewWatcher creates a new configuration file Watcher instance.
func NewWatcher() (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar fsnotify watcher para config: %w", err)
	}

	return &Watcher{
		fsWatcher: fw,
		stopChan:  make(chan struct{}),
	}, nil
}

// Start initiates the background watching on the given config file.
func (w *Watcher) Start(filePath string, send func(tea.Msg)) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	w.filePath = absPath
	w.sendFunc = send

	dir := filepath.Dir(absPath)
	if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
			return fmt.Errorf("falha ao criar diretório para monitoramento '%s': %w", dir, mkdirErr)
		}
	}

	if addErr := w.fsWatcher.Add(dir); addErr != nil {
		return fmt.Errorf("falha ao registrar diretório no watcher '%s': %w", dir, addErr)
	}

	w.wg.Add(1)
	go w.watchLoop()

	return nil
}

// Pause temporarily ignores file changes.
func (w *Watcher) Pause() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.paused = true
}

// Resume re-enables file modification processing.
func (w *Watcher) Resume() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.paused = false
}

// Close gracefully terminates the watcher.
func (w *Watcher) Close() error {
	w.mu.Lock()
	select {
	case <-w.stopChan:
		w.mu.Unlock()
		return nil
	default:
		close(w.stopChan)
	}
	w.mu.Unlock()

	err := w.fsWatcher.Close()
	w.wg.Wait()
	return err
}

func (w *Watcher) watchLoop() {
	defer w.wg.Done()

	var debounceTimer *time.Timer
	var debounceDuration = 50 * time.Millisecond
	targetBase := filepath.Base(w.filePath)

	for {
		select {
		case <-w.stopChan:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			if filepath.Base(event.Name) != targetBase {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			w.mu.Lock()
			if w.paused {
				w.mu.Unlock()
				continue
			}
			w.mu.Unlock()

			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.AfterFunc(debounceDuration, func() {
				w.reloadAndDispatch()
			})

		case _, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *Watcher) reloadAndDispatch() {
	w.mu.Lock()
	filePath := w.filePath
	send := w.sendFunc
	isPaused := w.paused
	w.mu.Unlock()

	if isPaused || send == nil {
		return
	}

	cfg, warnings, _ := LoadConfigFile(filePath)
	resolvedThemeName := theme.ResolveThemeName(cfg.Theme)
	palette, _ := theme.GetPalette(resolvedThemeName)
	compiled := theme.CompileTheme(palette, cfg.Colors)

	var warnStr string
	if len(warnings) > 0 {
		warnStr = strings.Join(warnings, " | ")
	}

	send(ThemeReloadedMsg{
		Config:        cfg,
		CompiledTheme: compiled,
		Warning:       warnStr,
	})
}
