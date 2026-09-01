package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"md-notes/internal/buffer"
	"md-notes/internal/config"
	"md-notes/internal/search"
	"md-notes/internal/theme"
	"md-notes/internal/ui"
	"md-notes/internal/vim"
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
	Vim           *vim.Engine
	Config        *config.Config
	Theme         *theme.CompiledTheme
	ConfigWatcher *config.Watcher
	Viewport      *ui.Viewport
	SearchEngine  *search.SearchEngine
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
		Buffer:       buf,
		Watcher:      w,
		StatusMsg:    status,
		Quitting:     false,
		Width:        80,
		Height:       24,
		SearchEngine: search.NewSearchEngine(),
	}

	m.Viewport = ui.NewViewport(80, 22)

	m.Vim = vim.NewEngine(buf)
	m.Vim.SaveCallback = func(targetPath string, force bool) tea.Cmd {
		return func() tea.Msg {
			return SaveFileMsg{TargetFilePath: targetPath, Force: force}
		}
	}
	m.Vim.QuitCallback = func(force bool) tea.Cmd {
		return func() tea.Msg {
			return QuitMsg{Force: force}
		}
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.Config == nil {
		m.Config = config.DefaultConfig()
	}

	if m.Viewport != nil && m.Config != nil {
		m.Viewport.RelativeNum = m.Config.Editor.RelativeLineNumbers
		m.Viewport.Scrolloff = m.Config.Editor.Scrolloff
		m.Viewport.SoftWrap = m.Config.Editor.WordWrap
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
		vpHeight := msg.Height - 2
		if vpHeight < 1 {
			vpHeight = 1
		}
		if m.Viewport != nil {
			m.Viewport.SetDimensions(msg.Width, vpHeight)
		}
		return m, nil

	case config.ThemeReloadedMsg:
		if msg.Config != nil {
			m.Config = msg.Config
			if m.Viewport != nil {
				m.Viewport.RelativeNum = m.Config.Editor.RelativeLineNumbers
				m.Viewport.Scrolloff = m.Config.Editor.Scrolloff
				m.Viewport.SoftWrap = m.Config.Editor.WordWrap
			}
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
		targetPath := msg.TargetFilePath
		if targetPath == "" {
			targetPath = m.Buffer.FilePath
		}
		if targetPath == "" {
			m.StatusMsg = "Erro: Nenhum nome de arquivo definido. Use :w <caminho>"
			if m.Watcher != nil {
				m.Watcher.Resume()
			}
			return m, func() tea.Msg { return SaveResultMsg{Err: fmt.Errorf("nenhum nome de arquivo definido")} }
		}

		err := buffer.AtomicSave(m.Buffer, targetPath)
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
				if m.Vim != nil {
					m.Vim.Buffer = reloaded
				}
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

		if m.Vim != nil {
			cmd, statusMsg := m.Vim.HandleKey(msg)
			if statusMsg != "" {
				m.StatusMsg = statusMsg
			}
			if cmd != nil {
				return m, cmd
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the complete visual terminal UI.
func (m *Model) View() string {
	if m.Quitting {
		return ui.GetRestoreCursorSequence()
	}

	if m.Viewport == nil {
		m.Viewport = ui.NewViewport(m.Width, m.Height-2)
	}

	totalLines := m.Buffer.LineCount()
	m.Viewport.AdjustScroll(m.Buffer.Cursor, totalLines)

	visibleLines := m.Viewport.GetVisibleLines(m.Buffer)
	vpHeight := m.Viewport.Height
	if vpHeight < 1 {
		vpHeight = 1
	}

	scrollbar := ui.RenderScrollbar(totalLines, vpHeight, m.Viewport.TopLine, m.Theme)
	gutterW := ui.CalculateGutterWidth(totalLines)

	var screenLines []string
	for rowIdx := 0; rowIdx < vpHeight; rowIdx++ {
		var gutterCell string
		var contentCell string

		if rowIdx < len(visibleLines) {
			vLine := visibleLines[rowIdx]
			if vLine.WrapIndex == 0 {
				gutterCell = ui.RenderGutterLine(vLine.LogicalLine, m.Buffer.Cursor.Line, totalLines, m.Viewport.RelativeNum, m.Theme)
			} else {
				gutterCell = ui.RenderEmptyGutterLine(totalLines, m.Theme)
			}
			contentCell = vLine.Text
		} else {
			// Empty area below document
			if m.Theme != nil {
				gutterCell = m.Theme.Muted.Render(fmt.Sprintf(" %*s ", gutterW-2, "~"))
			} else {
				gutterCell = fmt.Sprintf(" %*s ", gutterW-2, "~")
			}
			contentCell = ""
		}

		sbCell := " "
		if rowIdx < len(scrollbar) {
			sbCell = scrollbar[rowIdx]
		}

		contentWidthAvailable := m.Width - gutterW - 1
		if contentWidthAvailable < 0 {
			contentWidthAvailable = 0
		}
		contentVisualW := runewidth.StringWidth(contentCell)
		pad := contentWidthAvailable - contentVisualW
		if pad < 0 {
			pad = 0
		}

		lineStr := gutterCell + contentCell + strings.Repeat(" ", pad) + sbCell
		screenLines = append(screenLines, lineStr)
	}

	// Modal & status information
	currentMode := vim.ModeNormal
	commandInput := ""
	if m.Vim != nil {
		currentMode = m.Vim.State.CurrentMode
		commandInput = m.Vim.State.CommandInput
	}

	searchCur := 0
	searchTot := 0
	if m.SearchEngine != nil && m.SearchEngine.Result.TotalCount > 0 {
		searchCur = m.SearchEngine.Result.CurrentIndex
		searchTot = m.SearchEngine.Result.TotalCount
	}

	lineEnding := buffer.EndingLF
	if totalLines > 0 {
		firstLine, _ := m.Buffer.GetLine(0)
		lineEnding = firstLine.Ending
	}

	statusMsg := m.StatusMsg
	if currentMode == vim.ModeInsert && (statusMsg == "" || statusMsg == "[Novo Arquivo]" || statusMsg == "[Stdin Buffer]") {
		statusMsg = "-- INSERT --"
	} else if currentMode == vim.ModeVisualChar && (statusMsg == "" || statusMsg == "[Novo Arquivo]" || statusMsg == "[Stdin Buffer]") {
		statusMsg = "-- VISUAL --"
	} else if currentMode == vim.ModeVisualLine && (statusMsg == "" || statusMsg == "[Novo Arquivo]" || statusMsg == "[Stdin Buffer]") {
		statusMsg = "-- VISUAL LINE --"
	} else if currentMode == vim.ModeVisualBlock && (statusMsg == "" || statusMsg == "[Novo Arquivo]" || statusMsg == "[Stdin Buffer]") {
		statusMsg = "-- VISUAL BLOCK --"
	}

	statusState := ui.StatusState{
		Mode:           currentMode,
		FilePath:       m.Buffer.FilePath,
		IsDirty:        m.Buffer.IsDirty,
		CursorLine:     m.Buffer.Cursor.Line + 1,
		CursorCol:      m.Buffer.Cursor.Col + 1,
		TotalLines:     totalLines,
		LineEnding:     lineEnding,
		SearchMatchCur: searchCur,
		SearchMatchTot: searchTot,
		StatusMessage:  statusMsg,
		CommandInput:   commandInput,
	}

	statusLine := ui.RenderStatusLine(statusState, m.Theme, m.Width)
	cmdLine := ui.RenderCommandLine(statusState, m.Theme, m.Width)
	cursorSeq := ui.GetCursorShapeSequence(currentMode)

	return strings.Join(screenLines, "\n") + "\n" + statusLine + "\n" + cmdLine + cursorSeq
}
