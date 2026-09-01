package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/buffer"
	"md-notes/internal/config"
	"md-notes/internal/theme"
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

// ModelOption allows optional configuration of the app Model.
type ModelOption func(*Model)

// WithConfig sets the initial Config.
func WithConfig(cfg *config.Config) ModelOption {
	return func(m *Model) {
		m.Config = cfg
	}
}

// WithTheme sets the initial CompiledTheme.
func WithTheme(th *theme.CompiledTheme) ModelOption {
	return func(m *Model) {
		m.Theme = th
	}
}

// WithConfigWatcher sets the configuration file watcher.
func WithConfigWatcher(cw *config.Watcher) ModelOption {
	return func(m *Model) {
		m.ConfigWatcher = cw
	}
}

// Model represents the root Bubble Tea application model.
type Model struct {
	Buffer        *buffer.Buffer
	Watcher       *watcher.Watcher
	Config        *config.Config
	Theme         *theme.CompiledTheme
	ConfigWatcher *config.Watcher
	StatusMsg     string
	Quitting      bool
	Width         int
	Height        int
}

// NewModel creates a new Model instance with sensible defaults and optional parameters.
func NewModel(buf *buffer.Buffer, w *watcher.Watcher, opts ...ModelOption) *Model {
	status := ""
	if buf.IsNewFile && buf.FilePath != "" {
		status = "[Novo Arquivo]"
	} else if buf.IsStdinBuffer {
		status = "[Stdin Buffer]"
	}

	m := &Model{
		Buffer:    buf,
		Watcher:   w,
		StatusMsg: status,
		Quitting:  false,
		Width:     80,
		Height:    24,
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.Config == nil {
		m.Config = config.DefaultConfig()
	}
	if m.Theme == nil {
		resolvedTheme := theme.ResolveThemeName(m.Config.Theme)
		palette, _ := theme.GetPalette(resolvedTheme)
		m.Theme = theme.CompileTheme(palette, m.Config.Colors)
	}

	return m
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

	case config.ThemeReloadedMsg:
		if msg.Config != nil {
			m.Config = msg.Config
		}
		if msg.CompiledTheme != nil {
			m.Theme = msg.CompiledTheme
		}
		if msg.Warning != "" {
			m.StatusMsg = msg.Warning
		}
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
