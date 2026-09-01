package watcher_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/watcher"
)

// TC12 — Monitoramento Externo e Auto-Recarregamento com fsnotify (Full Scope)
func TestWatcher_ExternalModification(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "watch_test.md")

	if err := os.WriteFile(filePath, []byte("versao 1\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	w, err := watcher.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	var mu sync.Mutex
	var receivedMsg *watcher.FileModifiedMsg
	msgChan := make(chan struct{}, 1)

	err = w.Start(filePath, func(msg tea.Msg) {
		if modMsg, ok := msg.(watcher.FileModifiedMsg); ok {
			mu.Lock()
			receivedMsg = &modMsg
			mu.Unlock()
			select {
			case msgChan <- struct{}{}:
			default:
			}
		}
	})
	if err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Give watcher a moment to register in OS
	time.Sleep(50 * time.Millisecond)

	// Modify file externally
	if err := os.WriteFile(filePath, []byte("versao 2 modificada\n"), 0644); err != nil {
		t.Fatalf("failed to write external modification: %v", err)
	}

	select {
	case <-msgChan:
		mu.Lock()
		defer mu.Unlock()
		if receivedMsg == nil {
			t.Fatalf("expected FileModifiedMsg, got nil")
		}
		if receivedMsg.Path != filePath {
			t.Fatalf("expected path %q, got %q", filePath, receivedMsg.Path)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for FileModifiedMsg from watcher")
	}
}

func TestWatcher_PauseAndResume(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "pause_test.md")

	if err := os.WriteFile(filePath, []byte("versao 1\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	w, err := watcher.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	var received bool
	var mu sync.Mutex

	_ = w.Start(filePath, func(msg tea.Msg) {
		if _, ok := msg.(watcher.FileModifiedMsg); ok {
			mu.Lock()
			received = true
			mu.Unlock()
		}
	})

	time.Sleep(50 * time.Millisecond)

	// Pause watcher and write file
	w.Pause()
	_ = os.WriteFile(filePath, []byte("versao paused\n"), 0644)
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if received {
		t.Fatalf("received event while watcher was paused")
	}
	mu.Unlock()
}
