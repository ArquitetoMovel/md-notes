package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// FileModifiedMsg is sent to the Bubble Tea program when the watched file is modified externally.
type FileModifiedMsg struct {
	Path    string
	ModTime time.Time
}

// FileDeletedMsg is sent to the Bubble Tea program when the watched file is deleted or renamed.
type FileDeletedMsg struct {
	Path string
}

// Watcher monitors file changes on disk using OS notifications.
type Watcher struct {
	fw       *fsnotify.Watcher
	mu       sync.Mutex
	paused   bool
	filePath string
	stopChan chan struct{}
}

// NewWatcher creates an instance of Watcher.
func NewWatcher() (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		fw:       fw,
		stopChan: make(chan struct{}),
	}, nil
}

// Start watching a file and dispatch tea.Msg to the provided Bubble Tea program or msg sender.
func (w *Watcher) Start(filePath string, sendMsg func(tea.Msg)) error {
	if filePath == "" || filePath == "[Stdin Buffer]" {
		return nil
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	dir := filepath.Dir(absPath)
	if err := w.fw.Add(dir); err != nil {
		return err
	}

	w.mu.Lock()
	w.filePath = absPath
	w.mu.Unlock()

	go func() {
		for {
			select {
			case <-w.stopChan:
				return
			case event, ok := <-w.fw.Events:
				if !ok {
					return
				}

				w.mu.Lock()
				paused := w.paused
				watchedFile := w.filePath
				w.mu.Unlock()

				if paused {
					continue
				}

				eventAbs, _ := filepath.Abs(event.Name)
				if eventAbs != watchedFile {
					continue
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					info, err := os.Stat(watchedFile)
					if err == nil && sendMsg != nil {
						sendMsg(FileModifiedMsg{
							Path:    watchedFile,
							ModTime: info.ModTime(),
						})
					}
				} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					if sendMsg != nil {
						sendMsg(FileDeletedMsg{Path: watchedFile})
					}
				}

			case _, ok := <-w.fw.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return nil
}

// Pause temporarily suppresses event emission (used during internal saves).
func (w *Watcher) Pause() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.paused = true
}

// Resume re-enables event emission after a brief delay to consume lingering save events.
func (w *Watcher) Resume() {
	go func() {
		time.Sleep(100 * time.Millisecond)
		w.mu.Lock()
		w.paused = false
		w.mu.Unlock()
	}()
}

// Close terminates the watcher and releases OS resources.
func (w *Watcher) Close() error {
	close(w.stopChan)
	return w.fw.Close()
}
