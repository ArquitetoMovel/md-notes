package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/buffer"
	"md-notes/internal/config"
	"md-notes/internal/markdown"
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

// OpenSettingsMsg requests opening the configuration settings modal.
type OpenSettingsMsg struct{}

// CloseSettingsMsg requests closing the settings modal, optionally persisting changes.
type CloseSettingsMsg struct {
	Save bool
}

// ConfigSavedMsg is dispatched when the configuration is successfully saved to disk.
type ConfigSavedMsg struct {
	FilePath string
}

// ConfigSaveErrorMsg is dispatched when saving the configuration to disk fails.
type ConfigSaveErrorMsg struct {
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
	Buffer         *buffer.Buffer
	Watcher        *watcher.Watcher
	Vim            *vim.Engine
	Config         *config.Config
	Theme          *theme.CompiledTheme
	MarkdownParser *markdown.Parser
	ConfigWatcher  *config.Watcher
	Viewport       *ui.Viewport
	SearchEngine   *search.SearchEngine
	StatusMsg      string
	Quitting       bool
	Width          int
	Height         int

	// Settings panel state (F08)
	SettingsOpen  bool
	SettingsState *ui.SettingsState
	ConfigBackup  *config.Config
	ThemeBackup   *theme.CompiledTheme
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
	m.Vim.ConfigCallback = func() tea.Cmd {
		return func() tea.Msg {
			return OpenSettingsMsg{}
		}
	}

	m.Vim.SearchQueryCallback = func(query string, isReverse bool) tea.Cmd {
		if query == "" {
			m.SearchEngine.Clear()
			if m.Buffer != nil {
				m.Buffer.Cursor.Line = m.Vim.State.SearchInitPos.Line
				m.Buffer.Cursor.Col = m.Vim.State.SearchInitPos.Col
				if m.Viewport != nil {
					m.Viewport.AdjustScrollWithBuffer(m.Buffer.Cursor, m.Buffer)
				}
			}
			return nil
		}

		res := m.SearchEngine.Search(m.Buffer.Lines, query, isReverse)
		if len(res.Matches) > 0 {
			targetMatch := res.Matches[0]
			for _, match := range res.Matches {
				if match.Line > m.Vim.State.SearchInitPos.Line ||
					(match.Line == m.Vim.State.SearchInitPos.Line && match.StartCol >= m.Vim.State.SearchInitPos.Col) {
					targetMatch = match
					break
				}
			}
			m.Buffer.Cursor.Line = targetMatch.Line
			m.Buffer.Cursor.Col = targetMatch.StartCol
			if m.Viewport != nil {
				m.Viewport.AdjustScrollWithBuffer(m.Buffer.Cursor, m.Buffer)
			}
		}
		return nil
	}

	m.Vim.SearchCancelCallback = func(initPos vim.Position) tea.Cmd {
		m.SearchEngine.Clear()
		m.StatusMsg = ""
		if m.Buffer != nil {
			m.Buffer.Cursor.Line = initPos.Line
			m.Buffer.Cursor.Col = initPos.Col
			if m.Viewport != nil {
				m.Viewport.AdjustScrollWithBuffer(m.Buffer.Cursor, m.Buffer)
			}
		}
		return nil
	}

	m.Vim.SearchConfirmCallback = func(query string, isReverse bool) tea.Cmd {
		if query == "" && m.Vim.LastSearch != "" {
			query = m.Vim.LastSearch
			isReverse = m.Vim.LastSearchReverse
		}
		if query == "" {
			m.SearchEngine.Clear()
			return nil
		}

		res := m.SearchEngine.Search(m.Buffer.Lines, query, isReverse)
		if res.TotalCount == 0 {
			m.StatusMsg = fmt.Sprintf("E486: Padrão não encontrado: %s", query)
			if m.Buffer != nil {
				m.Buffer.Cursor.Line = m.Vim.State.SearchInitPos.Line
				m.Buffer.Cursor.Col = m.Vim.State.SearchInitPos.Col
				if m.Viewport != nil {
					m.Viewport.AdjustScrollWithBuffer(m.Buffer.Cursor, m.Buffer)
				}
			}
		} else {
			match, _ := m.SearchEngine.Next(m.Vim.State.SearchInitPos.Line, m.Vim.State.SearchInitPos.Col-1)
			if match != nil && m.Buffer != nil {
				m.Buffer.Cursor.Line = match.Line
				m.Buffer.Cursor.Col = match.StartCol
				if m.Viewport != nil {
					m.Viewport.AdjustScrollWithBuffer(m.Buffer.Cursor, m.Buffer)
				}
			}
			m.StatusMsg = ""
		}
		return nil
	}

	m.Vim.SearchNavCallback = func(forward bool) tea.Cmd {
		if m.SearchEngine.ActivePattern == nil {
			if m.Vim.LastSearch != "" {
				m.SearchEngine.Search(m.Buffer.Lines, m.Vim.LastSearch, m.Vim.LastSearchReverse)
			}
		}
		if m.SearchEngine.ActivePattern == nil || m.SearchEngine.Result.TotalCount == 0 {
			m.StatusMsg = "E486: Padrão não encontrado"
			return nil
		}

		var match *search.Match
		var wrapAround bool
		if forward {
			match, wrapAround = m.SearchEngine.Next(m.Buffer.Cursor.Line, m.Buffer.Cursor.Col)
		} else {
			match, wrapAround = m.SearchEngine.Prev(m.Buffer.Cursor.Line, m.Buffer.Cursor.Col)
		}

		if match != nil && m.Buffer != nil {
			m.Buffer.Cursor.Line = match.Line
			m.Buffer.Cursor.Col = match.StartCol
			if m.Viewport != nil {
				m.Viewport.AdjustScrollWithBuffer(m.Buffer.Cursor, m.Buffer)
			}
		}

		if wrapAround {
			if forward {
				m.StatusMsg = "Busca atingiu o fim do arquivo, continuando do início"
			} else {
				m.StatusMsg = "Busca atingiu o início do arquivo, continuando do fim"
			}
		} else {
			m.StatusMsg = ""
		}
		return nil
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

	m.MarkdownParser = markdown.NewParser(m.Theme)

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
			if m.MarkdownParser != nil {
				m.MarkdownParser.SetTheme(m.Theme)
			}
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

	case OpenSettingsMsg:
		if m.Width < 50 || m.Height < 14 {
			m.StatusMsg = "Dimensão insuficiente (mínimo 50x14) para abrir o painel de configurações"
			return m, nil
		}
		if m.Config != nil {
			cfgCopy := *m.Config
			cfgCopy.Editor = m.Config.Editor
			cfgCopy.Colors = m.Config.Colors
			m.ConfigBackup = &cfgCopy
		}
		m.ThemeBackup = m.Theme
		m.SettingsState = ui.NewSettingsState(m.Config)
		m.SettingsOpen = true
		m.StatusMsg = ""
		return m, nil

	case CloseSettingsMsg:
		if !msg.Save {
			if m.ConfigBackup != nil {
				*m.Config = *m.ConfigBackup
				m.ConfigBackup = nil
			}
			if m.ThemeBackup != nil {
				m.Theme = m.ThemeBackup
				m.ThemeBackup = nil
			}
			if m.Theme != nil {
				m.MarkdownParser = markdown.NewParser(m.Theme)
			}
			if m.Viewport != nil && m.Config != nil {
				m.Viewport.RelativeNum = m.Config.Editor.RelativeLineNumbers
				m.Viewport.Scrolloff = m.Config.Editor.Scrolloff
				m.Viewport.SoftWrap = m.Config.Editor.WordWrap
			}
			m.SettingsOpen = false
			m.SettingsState = nil
			return m, nil
		}

		if m.ConfigWatcher != nil {
			m.ConfigWatcher.Pause()
		}
		cfgPath, err := config.GetConfigFilePath()
		var saveErr error
		if err != nil {
			saveErr = err
		} else if m.SettingsState != nil {
			saveErr = config.SaveConfigFile(cfgPath, &m.SettingsState.WorkingConfig)
		}
		if m.ConfigWatcher != nil {
			m.ConfigWatcher.Resume()
		}

		if saveErr != nil {
			if m.SettingsState != nil {
				m.SettingsState.PersistenceError = saveErr.Error()
			}
			return m, nil
		}

		if m.SettingsState != nil && m.Config != nil {
			*m.Config = m.SettingsState.WorkingConfig
		}
		m.SettingsOpen = false
		m.SettingsState = nil
		m.ConfigBackup = nil
		m.ThemeBackup = nil
		m.StatusMsg = "Configurações salvas em config.toml"
		return m, nil

	case tea.MouseMsg:
		if m.SettingsOpen {
			return m, nil
		}
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			if m.Viewport != nil && m.Buffer != nil {
				m.Viewport.ScrollUp(3, m.Buffer)
			}
			return m, nil
		case tea.MouseButtonWheelDown:
			if m.Viewport != nil && m.Buffer != nil {
				m.Viewport.ScrollDown(3, m.Buffer)
			}
			return m, nil
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

		// Global F2 toggle shortcut
		if msg.Type == tea.KeyF2 || msg.String() == "f2" {
			if m.SettingsOpen {
				return m.Update(CloseSettingsMsg{Save: false})
			}
			return m.Update(OpenSettingsMsg{})
		}

		// Intercept exclusively if settings modal is open
		if m.SettingsOpen && m.SettingsState != nil {
			action := m.SettingsState.HandleKey(msg)
			switch action {
			case ui.ActionLivePreviewUpdated:
				m.Config.Theme = m.SettingsState.WorkingConfig.Theme
				m.Config.Editor = m.SettingsState.WorkingConfig.Editor
				m.Config.Colors = m.SettingsState.WorkingConfig.Colors

				resolvedTheme := theme.ResolveThemeName(m.Config.Theme)
				palette, _ := theme.GetPalette(resolvedTheme)
				m.Theme = theme.CompileTheme(palette, m.Config.Colors)
				m.MarkdownParser = markdown.NewParser(m.Theme)

				if m.Viewport != nil {
					m.Viewport.RelativeNum = m.Config.Editor.RelativeLineNumbers
					m.Viewport.Scrolloff = m.Config.Editor.Scrolloff
					m.Viewport.SoftWrap = m.Config.Editor.WordWrap
				}
				return m, nil

			case ui.ActionCancelAndClose:
				return m.Update(CloseSettingsMsg{Save: false})

			case ui.ActionSaveAndClose:
				return m.Update(CloseSettingsMsg{Save: true})

			default:
				return m, nil
			}
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

func isLineInCodeBlock(lines []buffer.Line, targetLineIdx int) (bool, string) {
	inBlock := false
	lang := ""
	for i := 0; i < targetLineIdx && i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i].String())
		if strings.HasPrefix(trimmed, "```") {
			if inBlock {
				inBlock = false
				lang = ""
			} else {
				inBlock = true
				lang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			}
		}
	}
	return inBlock, lang
}

func isRuneSelected(mode vim.Mode, line, col int, selStart, selEnd vim.Position) bool {
	switch mode {
	case vim.ModeVisualLine:
		return line >= selStart.Line && line <= selEnd.Line
	case vim.ModeVisualChar:
		if line < selStart.Line || line > selEnd.Line {
			return false
		}
		if selStart.Line == selEnd.Line {
			return col >= selStart.Col && col <= selEnd.Col
		}
		if line == selStart.Line {
			return col >= selStart.Col
		}
		if line == selEnd.Line {
			return col <= selEnd.Col
		}
		return true
	case vim.ModeVisualBlock:
		if line < selStart.Line || line > selEnd.Line {
			return false
		}
		minCol := selStart.Col
		maxCol := selEnd.Col
		if minCol > maxCol {
			minCol, maxCol = maxCol, minCol
		}
		return col >= minCol && col <= maxCol
	default:
		return false
	}
}

func sliceSpans(spans []markdown.StyledSpan, startCol, endCol int) []markdown.StyledSpan {
	if len(spans) == 0 || startCol >= endCol {
		return nil
	}

	var result []markdown.StyledSpan
	curCol := 0

	for _, span := range spans {
		spanRunes := []rune(span.Text)
		spanLen := len(spanRunes)
		spanEnd := curCol + spanLen

		if spanEnd > startCol && curCol < endCol {
			overlapStart := startCol
			if curCol > overlapStart {
				overlapStart = curCol
			}
			overlapEnd := endCol
			if spanEnd < overlapEnd {
				overlapEnd = spanEnd
			}

			localStart := overlapStart - curCol
			localEnd := overlapEnd - curCol

			if localStart < localEnd && localStart < spanLen {
				result = append(result, markdown.StyledSpan{
					Text:  string(spanRunes[localStart:localEnd]),
					Style: span.Style,
					Type:  span.Type,
				})
			}
		}

		curCol = spanEnd
		if curCol >= endCol {
			break
		}
	}

	return result
}

func applyCursorToSpans(spans []markdown.StyledSpan, cursorRelCol int, cursorStyle lipgloss.Style, isAtEnd bool) []markdown.StyledSpan {
	if isAtEnd {
		return append(spans, markdown.StyledSpan{
			Text:  " ",
			Style: cursorStyle,
			Type:  markdown.TokenText,
		})
	}

	var result []markdown.StyledSpan
	curCol := 0
	cursorApplied := false

	for _, span := range spans {
		spanRunes := []rune(span.Text)
		spanLen := len(spanRunes)
		spanEnd := curCol + spanLen

		if !cursorApplied && cursorRelCol >= curCol && cursorRelCol < spanEnd {
			localIdx := cursorRelCol - curCol
			if localIdx > 0 {
				result = append(result, markdown.StyledSpan{
					Text:  string(spanRunes[:localIdx]),
					Style: span.Style,
					Type:  span.Type,
				})
			}
			charStyle := span.Style.Copy().Reverse(true)
			result = append(result, markdown.StyledSpan{
				Text:  string(spanRunes[localIdx : localIdx+1]),
				Style: charStyle,
				Type:  span.Type,
			})
			if localIdx+1 < spanLen {
				result = append(result, markdown.StyledSpan{
					Text:  string(spanRunes[localIdx+1:]),
					Style: span.Style,
					Type:  span.Type,
				})
			}
			cursorApplied = true
		} else {
			result = append(result, span)
		}
		curCol = spanEnd
	}

	if !cursorApplied && cursorRelCol >= curCol {
		result = append(result, markdown.StyledSpan{
			Text:  " ",
			Style: cursorStyle,
			Type:  markdown.TokenText,
		})
	}

	return result
}

func applyOverlayToSpans(spans []markdown.StyledSpan, overlayStart, overlayEnd int, overlayStyle lipgloss.Style) []markdown.StyledSpan {
	if len(spans) == 0 || overlayStart >= overlayEnd {
		return spans
	}

	var result []markdown.StyledSpan
	curCol := 0

	for _, span := range spans {
		spanRunes := []rune(span.Text)
		spanLen := len(spanRunes)
		spanEnd := curCol + spanLen

		if spanEnd > overlayStart && curCol < overlayEnd {
			relStart := overlayStart - curCol
			if relStart < 0 {
				relStart = 0
			}
			relEnd := overlayEnd - curCol
			if relEnd > spanLen {
				relEnd = spanLen
			}

			if relStart > 0 {
				result = append(result, markdown.StyledSpan{
					Text:  string(spanRunes[:relStart]),
					Style: span.Style,
					Type:  span.Type,
				})
			}
			if relStart < relEnd {
				result = append(result, markdown.StyledSpan{
					Text:  string(spanRunes[relStart:relEnd]),
					Style: overlayStyle,
					Type:  span.Type,
				})
			}
			if relEnd < spanLen {
				result = append(result, markdown.StyledSpan{
					Text:  string(spanRunes[relEnd:]),
					Style: span.Style,
					Type:  span.Type,
				})
			}
		} else {
			result = append(result, span)
		}

		curCol = spanEnd
	}

	return result
}

func (m *Model) renderVisualLine(vLine ui.VisualLine, tl markdown.TokenizedLine, origLine buffer.Line, isCursorLine bool, cursorCol int, currentMode vim.Mode) string {
	origLen := origLine.Length()
	vRunes := []rune(vLine.Text)

	// 1. Slice spans for this visual line segment
	var spans []markdown.StyledSpan
	if len(tl.Spans) == 0 {
		if len(vRunes) > 0 {
			spans = []markdown.StyledSpan{{Text: vLine.Text, Style: lipgloss.NewStyle(), Type: markdown.TokenText}}
		}
	} else if vLine.StartCol == 0 && vLine.EndCol >= origLen {
		spans = tl.Spans
	} else {
		spans = sliceSpans(tl.Spans, vLine.StartCol, vLine.EndCol)
	}

	// 2. Overlay search matches (F06)
	if m.SearchEngine != nil && len(m.SearchEngine.Result.Matches) > 0 {
		for _, match := range m.SearchEngine.Result.Matches {
			if match.Line == vLine.LogicalLine {
				relStart := match.StartCol - vLine.StartCol
				relEnd := match.EndCol - vLine.StartCol
				if relEnd > 0 && relStart < len(vRunes) {
					if m.Theme != nil {
						spans = applyOverlayToSpans(spans, relStart, relEnd, m.Theme.SearchMatch)
					}
				}
			}
		}
	}

	// 3. Overlay visual selection (F03)
	if m.Vim != nil && m.Vim.State.Selection != nil {
		mode := m.Vim.State.CurrentMode
		if mode == vim.ModeVisualChar || mode == vim.ModeVisualLine || mode == vim.ModeVisualBlock {
			start, end := m.Vim.State.Selection.Normalized()
			if vLine.LogicalLine >= start.Line && vLine.LogicalLine <= end.Line {
				selStartCol := 0
				selEndCol := origLen
				if mode == vim.ModeVisualChar {
					if vLine.LogicalLine == start.Line {
						selStartCol = start.Col
					}
					if vLine.LogicalLine == end.Line {
						selEndCol = end.Col + 1
					}
				} else if mode == vim.ModeVisualBlock {
					minC, maxC := start.Col, end.Col
					if minC > maxC {
						minC, maxC = maxC, minC
					}
					selStartCol = minC
					selEndCol = maxC + 1
				}

				relStart := selStartCol - vLine.StartCol
				relEnd := selEndCol - vLine.StartCol
				if relEnd > 0 && relStart < len(vRunes) {
					if m.Theme != nil {
						spans = applyOverlayToSpans(spans, relStart, relEnd, m.Theme.Selection)
					}
				}
			}
		}
	}

	// 4. Cursor highlight during navigation (F03)
	if currentMode != vim.ModeCommand && isCursorLine {
		cursorStyle := lipgloss.NewStyle().Reverse(true)
		if m.Theme != nil {
			cursorStyle = m.Theme.Cursor
		}

		if origLen == 0 {
			if vLine.StartCol == 0 {
				spans = applyCursorToSpans(spans, 0, cursorStyle, true)
			}
		} else if cursorCol >= vLine.StartCol && cursorCol < vLine.EndCol {
			relCol := cursorCol - vLine.StartCol
			spans = applyCursorToSpans(spans, relCol, cursorStyle, false)
		} else if cursorCol >= origLen && vLine.EndCol == origLen {
			spans = applyCursorToSpans(spans, len(vRunes), cursorStyle, true)
		}
	}

	// 5. Render spans
	var sb strings.Builder
	for _, s := range spans {
		sb.WriteString(s.Style.Render(s.Text))
	}
	return sb.String()
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
	m.Viewport.AdjustScrollWithBuffer(m.Buffer.Cursor, m.Buffer)

	visibleLines := m.Viewport.GetVisibleLines(m.Buffer)
	vpHeight := m.Viewport.Height
	if vpHeight < 1 {
		vpHeight = 1
	}

	totalVisualLines := m.Viewport.TotalVisualLines(m.Buffer)
	scrollbar := ui.RenderScrollbar(totalVisualLines, vpHeight, m.Viewport.TopLine, m.Theme)
	gutterW := ui.CalculateGutterWidth(totalLines)

	// Modal & status information
	currentMode := vim.ModeNormal
	commandInput := ""
	searchIsReverse := false
	if m.Vim != nil {
		currentMode = m.Vim.State.CurrentMode
		commandInput = m.Vim.State.CommandInput
		searchIsReverse = m.Vim.State.SearchIsReverse
	}

	// Pre-parse Markdown for visible lines (F04)
	tokenizedCache := make(map[int]markdown.TokenizedLine)
	if len(visibleLines) > 0 {
		startLine := visibleLines[0].LogicalLine
		inCodeBlock, codeLang := isLineInCodeBlock(m.Buffer.Lines, startLine)
		endLine := visibleLines[len(visibleLines)-1].LogicalLine
		for lIdx := startLine; lIdx <= endLine && lIdx < totalLines; lIdx++ {
			line, _ := m.Buffer.GetLine(lIdx)
			if m.MarkdownParser != nil {
				tokenizedCache[lIdx] = m.MarkdownParser.ParseLine(line, &inCodeBlock, &codeLang)
			}
		}
	}

	var screenLines []string
	for rowIdx := 0; rowIdx < vpHeight; rowIdx++ {
		var gutterCell string
		var contentCell string

		if rowIdx < len(visibleLines) {
			vLine := visibleLines[rowIdx]
			isCursorLine := (m.Buffer.Cursor.Line == vLine.LogicalLine)
			if vLine.WrapIndex == 0 {
				gutterCell = ui.RenderGutterLine(vLine.LogicalLine, m.Buffer.Cursor.Line, totalLines, m.Viewport.RelativeNum, m.Theme)
			} else {
				gutterCell = ui.RenderEmptyGutterLine(totalLines, m.Theme)
			}

			tl, found := tokenizedCache[vLine.LogicalLine]
			origLine, _ := m.Buffer.GetLine(vLine.LogicalLine)
			if !found {
				tl = markdown.TokenizedLine{
					Spans:    []markdown.StyledSpan{{Text: vLine.Text, Style: lipgloss.NewStyle(), Type: markdown.TokenText}},
					Original: origLine,
				}
			}

			contentCell = m.renderVisualLine(vLine, tl, origLine, isCursorLine, m.Buffer.Cursor.Col, currentMode)
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
		contentVisualW := lipgloss.Width(contentCell)
		pad := contentWidthAvailable - contentVisualW
		if pad < 0 {
			pad = 0
		}

		lineStr := gutterCell + contentCell + strings.Repeat(" ", pad) + sbCell
		screenLines = append(screenLines, lineStr)
	}

	if m.SettingsOpen && m.SettingsState != nil {
		modalBox := ui.RenderSettingsModal(m.SettingsState, m.Theme, m.Width, vpHeight)
		screenLines = ui.OverlayModal(screenLines, modalBox, m.Width, vpHeight)
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
		Mode:            currentMode,
		FilePath:        m.Buffer.FilePath,
		IsDirty:         m.Buffer.IsDirty,
		CursorLine:      m.Buffer.Cursor.Line + 1,
		CursorCol:       m.Buffer.Cursor.Col + 1,
		TotalLines:      totalLines,
		LineEnding:      lineEnding,
		SearchMatchCur:  searchCur,
		SearchMatchTot:  searchTot,
		SearchIsReverse: searchIsReverse,
		StatusMessage:   statusMsg,
		CommandInput:    commandInput,
	}

	statusLine := ui.RenderStatusLine(statusState, m.Theme, m.Width)
	cmdLine := ui.RenderCommandLine(statusState, m.Theme, m.Width)
	cursorSeq := ui.GetCursorShapeSequence(currentMode)

	return strings.Join(screenLines, "\n") + "\n" + statusLine + "\n" + cmdLine + cursorSeq
}
