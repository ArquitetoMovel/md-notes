package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/buffer"
	"md-notes/internal/watcher"
)

// Msg types for internal coordination
type QuitMsg struct {
	Force bool
}

type SaveFileMsg struct {
	TargetFilePath string
	Force          bool
}

type SaveResultMsg struct {
	Err error
}

// Model represents the root Bubble Tea application model for F01.
type Model struct {
	Buffer    *buffer.Buffer
	Watcher   *watcher.Watcher
	StatusMsg string
	Quitting  bool
	Width     int
	Height    int
}

// NewModel creates a new Model instance.
func NewModel(buf *buffer.Buffer, w *watcher.Watcher) *Model {
	status := ""
	if buf.IsNewFile && buf.FilePath != "" {
		status = "[Novo Arquivo]"
	} else if buf.IsStdinBuffer {
		status = "[Stdin Buffer]"
	}

	return &Model{
		Buffer:    buf,
		Watcher:   w,
		StatusMsg: status,
		Quitting:  false,
		Width:     80,
		Height:    24,
	}
}

// Init initializes the Bubble Tea program.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates state accordingly.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case QuitMsg:
		if !msg.Force && m.Buffer.IsDirty {
			m.StatusMsg = "E37: Alterações não salvas. Use :w para salvar ou :q! para forçar a saída"
			return m, nil
		}
		m.Quitting = true
		return m, tea.Quit

	case SaveFileMsg:
		if m.Watcher != nil {
			m.Watcher.Pause()
		}
		err := buffer.AtomicSave(m.Buffer, msg.TargetFilePath)
		if m.Watcher != nil {
			m.Watcher.Resume()
		}

		if err != nil {
			m.StatusMsg = err.Error()
			return m, func() tea.Msg { return SaveResultMsg{Err: err} }
		}
		m.StatusMsg = ""
		return m, func() tea.Msg { return SaveResultMsg{Err: nil} }

	case watcher.FileModifiedMsg:
		if !m.Buffer.IsDirty {
			reloaded, err := buffer.LoadFromFile(msg.Path)
			if err == nil {
				m.Buffer = reloaded
				m.StatusMsg = "Arquivo recarregado do disco."
			}
		} else {
			m.StatusMsg = "Aviso: Arquivo modificado externamente no disco."
		}
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "ctrl+c" {
			if m.Buffer.IsDirty {
				m.StatusMsg = "E37: Alterações não salvas. Use :w para salvar ou :q! para forçar a saída"
				return m, nil
			}
			m.Quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the terminal output for the model.
func (m *Model) View() string {
	if m.Quitting {
		return ""
	}

	var sb strings.Builder
	for i := 0; i < m.Buffer.LineCount(); i++ {
		line, _ := m.Buffer.GetLine(i)
		sb.WriteString(line.String())
		if i < m.Buffer.LineCount()-1 {
			sb.WriteString("\n")
		}
	}

	if m.StatusMsg != "" {
		sb.WriteString(fmt.Sprintf("\n%s", m.StatusMsg))
	}

	return sb.String()
}
