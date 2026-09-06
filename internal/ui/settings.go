package ui

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/config"
	"md-notes/internal/theme"
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// SettingsTab identifies the active tab in the settings dialog.
type SettingsTab int

const (
	TabThemes SettingsTab = iota
	TabEditor
	TabColors
)

// SettingsAction represents state transition signals emitted by the settings panel.
type SettingsAction int

const (
	ActionNone SettingsAction = iota
	ActionLivePreviewUpdated
	ActionSaveAndClose
	ActionCancelAndClose
)

// EditorFieldType represents the type of control for an editor preference.
type EditorFieldType int

const (
	FieldToggleBool EditorFieldType = iota // Toggle true/false
	FieldTabSize                           // Cycle: 2 -> 4 -> 8 -> 2
	FieldScrolloff                         // Numeric bound: 0 to 10
)

// EditorOption represents a configurable editor preference item.
type EditorOption struct {
	Key         string
	Label       string
	Type        EditorFieldType
	Description string
}

// ColorTokenItem represents a theme color token that can be customized.
type ColorTokenItem struct {
	TokenKey string
	Label    string
	HexValue string
}

// SettingsState holds the interactive state of the settings dialog.
type SettingsState struct {
	ActiveTab        SettingsTab
	ActiveThemeIndex int
	ActiveEditorRow  int
	ActiveColorRow   int

	AvailableThemes []string
	EditorOptions   []EditorOption
	ColorTokens     []ColorTokenItem

	WorkingConfig config.Config

	IsEditingHex   bool
	HexInputBuffer string
	HexInputError  string

	PersistenceError string
	FocusArea        int  // 0: tabs, 1: content list, 2: save button
	FocusSaveButton  bool // true when the save button has focus
}

// NewSettingsState initializes a new settings state cloned from the provided configuration.
func NewSettingsState(cfg *config.Config) *SettingsState {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	state := &SettingsState{
		ActiveTab:        TabThemes,
		ActiveThemeIndex: 0,
		ActiveEditorRow:  0,
		ActiveColorRow:   0,
		AvailableThemes:  theme.BuiltinThemes(),
		WorkingConfig:    *cfg,
		FocusArea:        1,
		FocusSaveButton:  false,
	}

	// Match active theme index
	for i, name := range state.AvailableThemes {
		if name == state.WorkingConfig.Theme {
			state.ActiveThemeIndex = i
			break
		}
	}

	// Initialize editor options
	state.EditorOptions = []EditorOption{
		{
			Key:         "line_numbers",
			Label:       "Numeração de Linhas",
			Type:        FieldToggleBool,
			Description: "Exibe números de linha na margem esquerda",
		},
		{
			Key:         "relative_line_numbers",
			Label:       "Numeração Relativa",
			Type:        FieldToggleBool,
			Description: "Exibe números de linha relativos à linha do cursor",
		},
		{
			Key:         "tab_size",
			Label:       "Tamanho do Tab",
			Type:        FieldTabSize,
			Description: "Largura da tabulação em espaços (2, 4, 8)",
		},
		{
			Key:         "scrolloff",
			Label:       "Margem de Rolagem (Scrolloff)",
			Type:        FieldScrolloff,
			Description: "Quantidade de linhas visíveis mantidas acima/abaixo do cursor (0 a 10)",
		},
		{
			Key:         "word_wrap",
			Label:       "Quebra de Linha (Word Wrap)",
			Type:        FieldToggleBool,
			Description: "Quebra linhas longas automaticamente na largura da tela",
		},
	}

	// Initialize color tokens supported by theme.ColorOverrides
	palette, _ := theme.GetPalette(state.WorkingConfig.Theme)
	tokenDefs := []struct {
		key   string
		label string
	}{
		{"h1", "Título H1"},
		{"h2", "Título H2"},
		{"h3", "Título H3"},
		{"h4", "Título H4"},
		{"h5", "Título H5"},
		{"h6", "Título H6"},
		{"muted", "Texto Atenuado"},
		{"bold", "Texto em Negrito"},
		{"italic", "Texto em Itálico"},
		{"code_bg", "Fundo de Bloco de Código"},
		{"code_fg", "Texto de Bloco de Código"},
		{"table_border", "Borda de Tabelas"},
		{"table_header", "Cabeçalho de Tabelas"},
		{"status_bar_bg", "Fundo da Barra de Status"},
		{"status_bar_fg", "Texto da Barra de Status"},
		{"search_match_bg", "Destaque de Busca (Fundo)"},
		{"search_match_fg", "Destaque de Busca (Texto)"},
	}

	state.ColorTokens = make([]ColorTokenItem, len(tokenDefs))
	for i, def := range tokenDefs {
		hexVal := state.getOverrideValue(def.key)
		if hexVal == "" {
			hexVal = state.getPaletteValue(palette, def.key)
		}
		state.ColorTokens[i] = ColorTokenItem{
			TokenKey: def.key,
			Label:    def.label,
			HexValue: hexVal,
		}
	}

	return state
}

// IncrementScrolloff increments the scrolloff setting up to the maximum limit of 10.
func (s *SettingsState) IncrementScrolloff() bool {
	if s.WorkingConfig.Editor.Scrolloff < 10 {
		s.WorkingConfig.Editor.Scrolloff++
		return true
	}
	return false
}

// DecrementScrolloff decrements the scrolloff setting down to the minimum limit of 0.
func (s *SettingsState) DecrementScrolloff() bool {
	if s.WorkingConfig.Editor.Scrolloff > 0 {
		s.WorkingConfig.Editor.Scrolloff--
		return true
	}
	return false
}

// SetSaveButtonFocused sets whether the save button is currently focused.
func (s *SettingsState) SetSaveButtonFocused(focused bool) {
	s.FocusSaveButton = focused
	if focused {
		s.FocusArea = 2
	} else {
		s.FocusArea = 1
	}
}

// HandleKey processes keyboard input on the settings modal and returns the resulting action.
func (s *SettingsState) HandleKey(msg tea.KeyMsg) SettingsAction {
	keyStr := strings.ToLower(msg.String())

	// 1. Inline Hex Color Editing Mode
	if s.IsEditingHex {
		switch {
		case msg.Type == tea.KeyEsc || keyStr == "esc":
			s.IsEditingHex = false
			s.HexInputBuffer = ""
			s.HexInputError = ""
			return ActionNone

		case msg.Type == tea.KeyBackspace || keyStr == "backspace":
			runes := []rune(s.HexInputBuffer)
			if len(runes) > 0 {
				s.HexInputBuffer = string(runes[:len(runes)-1])
			}
			s.HexInputError = ""
			return ActionNone

		case msg.Type == tea.KeyEnter || keyStr == "enter":
			if hexColorRegex.MatchString(s.HexInputBuffer) {
				s.applyHexColor(s.ActiveColorRow, s.HexInputBuffer)
				s.IsEditingHex = false
				s.HexInputBuffer = ""
				s.HexInputError = ""
				return ActionLivePreviewUpdated
			}
			s.HexInputError = "Código de cor inválido: use formato #RRGGBB"
			return ActionNone

		default:
			runes := msg.Runes
			if len(runes) == 0 && len(keyStr) == 1 {
				runes = []rune(keyStr)
			}
			for _, r := range runes {
				if (r == '#' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) && len(s.HexInputBuffer) < 7 {
					s.HexInputBuffer += string(r)
					s.HexInputError = ""
				}
			}
			return ActionNone
		}
	}

	// 2. Normal Modal Navigation & Control
	// Close / Cancel modal
	if msg.Type == tea.KeyEsc || keyStr == "esc" || keyStr == "q" {
		return ActionCancelAndClose
	}

	// Direct shortcut to save
	if keyStr == "ctrl+s" || msg.Type == tea.KeyCtrlS {
		return ActionSaveAndClose
	}

	// Tab switching: Tab, Shift+Tab
	if msg.Type == tea.KeyTab || keyStr == "tab" {
		s.ActiveTab = (s.ActiveTab + 1) % 3
		s.FocusSaveButton = false
		s.FocusArea = 1
		s.HexInputError = ""
		return ActionNone
	}

	if msg.Type == tea.KeyShiftTab || keyStr == "shift+tab" {
		s.ActiveTab = (s.ActiveTab + 2) % 3
		s.FocusSaveButton = false
		s.FocusArea = 1
		s.HexInputError = ""
		return ActionNone
	}

	// Scrolloff row special handling for horizontal adjustments
	isScrolloffRow := s.ActiveTab == TabEditor && !s.FocusSaveButton && s.ActiveEditorRow == 3
	if isScrolloffRow {
		if keyStr == "+" || keyStr == "=" || keyStr == "]" {
			s.IncrementScrolloff()
			return ActionLivePreviewUpdated
		}
		if keyStr == "-" || keyStr == "_" || keyStr == "[" {
			s.DecrementScrolloff()
			return ActionLivePreviewUpdated
		}
		if keyStr == "right" || msg.Type == tea.KeyRight || keyStr == "l" {
			s.IncrementScrolloff()
			return ActionLivePreviewUpdated
		}
		if keyStr == "left" || msg.Type == tea.KeyLeft || keyStr == "h" {
			s.DecrementScrolloff()
			return ActionLivePreviewUpdated
		}
	}

	// Horizontal navigation between tabs (when not adjusting scrolloff)
	if keyStr == "right" || msg.Type == tea.KeyRight || keyStr == "l" {
		s.ActiveTab = (s.ActiveTab + 1) % 3
		s.FocusSaveButton = false
		s.FocusArea = 1
		s.HexInputError = ""
		return ActionNone
	}

	if keyStr == "left" || msg.Type == tea.KeyLeft || keyStr == "h" {
		s.ActiveTab = (s.ActiveTab + 2) % 3
		s.FocusSaveButton = false
		s.FocusArea = 1
		s.HexInputError = ""
		return ActionNone
	}

	// Vertical navigation
	if keyStr == "up" || msg.Type == tea.KeyUp || keyStr == "k" {
		if s.FocusSaveButton {
			s.FocusSaveButton = false
			s.FocusArea = 1
			return ActionNone
		}
		switch s.ActiveTab {
		case TabThemes:
			if s.ActiveThemeIndex > 0 {
				s.ActiveThemeIndex--
			}
		case TabEditor:
			if s.ActiveEditorRow > 0 {
				s.ActiveEditorRow--
			}
		case TabColors:
			if s.ActiveColorRow > 0 {
				s.ActiveColorRow--
			}
		}
		return ActionNone
	}

	if keyStr == "down" || msg.Type == tea.KeyDown || keyStr == "j" {
		if s.FocusSaveButton {
			return ActionNone
		}
		switch s.ActiveTab {
		case TabThemes:
			if s.ActiveThemeIndex < len(s.AvailableThemes)-1 {
				s.ActiveThemeIndex++
			}
		case TabEditor:
			if s.ActiveEditorRow < len(s.EditorOptions)-1 {
				s.ActiveEditorRow++
			}
		case TabColors:
			if s.ActiveColorRow < len(s.ColorTokens)-1 {
				s.ActiveColorRow++
			}
		}
		return ActionNone
	}

	// Focus save button key
	if keyStr == "s" {
		s.FocusSaveButton = true
		s.FocusArea = 2
		return ActionNone
	}

	// Selection and Activation: Enter or Space
	isEnterOrSpace := keyStr == "enter" || msg.Type == tea.KeyEnter || keyStr == " " || keyStr == "space" || msg.Type == tea.KeySpace

	if s.FocusSaveButton {
		if isEnterOrSpace {
			return ActionSaveAndClose
		}
		return ActionNone
	}

	switch s.ActiveTab {
	case TabThemes:
		if isEnterOrSpace {
			if s.ActiveThemeIndex >= 0 && s.ActiveThemeIndex < len(s.AvailableThemes) {
				s.WorkingConfig.Theme = s.AvailableThemes[s.ActiveThemeIndex]
				s.syncColorTokensFromTheme(s.WorkingConfig.Theme)
				return ActionLivePreviewUpdated
			}
		}

	case TabEditor:
		if isEnterOrSpace {
			if s.ActiveEditorRow >= 0 && s.ActiveEditorRow < len(s.EditorOptions) {
				opt := s.EditorOptions[s.ActiveEditorRow]
				switch opt.Type {
				case FieldToggleBool:
					switch opt.Key {
					case "line_numbers":
						s.WorkingConfig.Editor.LineNumbers = !s.WorkingConfig.Editor.LineNumbers
					case "relative_line_numbers":
						s.WorkingConfig.Editor.RelativeLineNumbers = !s.WorkingConfig.Editor.RelativeLineNumbers
					case "word_wrap":
						s.WorkingConfig.Editor.WordWrap = !s.WorkingConfig.Editor.WordWrap
					}
					return ActionLivePreviewUpdated

				case FieldTabSize:
					switch s.WorkingConfig.Editor.TabSize {
					case 2:
						s.WorkingConfig.Editor.TabSize = 4
					case 4:
						s.WorkingConfig.Editor.TabSize = 8
					case 8:
						s.WorkingConfig.Editor.TabSize = 2
					default:
						s.WorkingConfig.Editor.TabSize = 4
					}
					return ActionLivePreviewUpdated

				case FieldScrolloff:
					s.IncrementScrolloff()
					return ActionLivePreviewUpdated
				}
			}
		}

	case TabColors:
		if keyStr == "e" || isEnterOrSpace {
			if s.ActiveColorRow >= 0 && s.ActiveColorRow < len(s.ColorTokens) {
				s.IsEditingHex = true
				s.HexInputBuffer = s.ColorTokens[s.ActiveColorRow].HexValue
				s.HexInputError = ""
				return ActionNone
			}
		}
	}

	return ActionNone
}

func (s *SettingsState) applyHexColor(tokenIdx int, hexValue string) {
	if tokenIdx < 0 || tokenIdx >= len(s.ColorTokens) {
		return
	}
	s.ColorTokens[tokenIdx].HexValue = hexValue
	key := s.ColorTokens[tokenIdx].TokenKey
	switch key {
	case "h1":
		s.WorkingConfig.Colors.H1 = hexValue
	case "h2":
		s.WorkingConfig.Colors.H2 = hexValue
	case "h3":
		s.WorkingConfig.Colors.H3 = hexValue
	case "h4":
		s.WorkingConfig.Colors.H4 = hexValue
	case "h5":
		s.WorkingConfig.Colors.H5 = hexValue
	case "h6":
		s.WorkingConfig.Colors.H6 = hexValue
	case "muted":
		s.WorkingConfig.Colors.Muted = hexValue
	case "bold":
		s.WorkingConfig.Colors.Bold = hexValue
	case "italic":
		s.WorkingConfig.Colors.Italic = hexValue
	case "code_bg":
		s.WorkingConfig.Colors.CodeBg = hexValue
	case "code_fg":
		s.WorkingConfig.Colors.CodeFg = hexValue
	case "table_border":
		s.WorkingConfig.Colors.TableBorder = hexValue
	case "table_header":
		s.WorkingConfig.Colors.TableHeader = hexValue
	case "status_bar_bg":
		s.WorkingConfig.Colors.StatusBarBg = hexValue
	case "status_bar_fg":
		s.WorkingConfig.Colors.StatusBarFg = hexValue
	case "search_match_bg":
		s.WorkingConfig.Colors.SearchMatchBg = hexValue
	case "search_match_fg":
		s.WorkingConfig.Colors.SearchMatchFg = hexValue
	}
}

func (s *SettingsState) syncColorTokensFromTheme(themeName string) {
	palette, ok := theme.GetPalette(themeName)
	if !ok {
		return
	}
	for i := range s.ColorTokens {
		token := &s.ColorTokens[i]
		override := s.getOverrideValue(token.TokenKey)
		if override != "" {
			token.HexValue = override
		} else {
			token.HexValue = s.getPaletteValue(palette, token.TokenKey)
		}
	}
}

func (s *SettingsState) getOverrideValue(key string) string {
	c := s.WorkingConfig.Colors
	switch key {
	case "h1":
		return c.H1
	case "h2":
		return c.H2
	case "h3":
		return c.H3
	case "h4":
		return c.H4
	case "h5":
		return c.H5
	case "h6":
		return c.H6
	case "muted":
		return c.Muted
	case "bold":
		return c.Bold
	case "italic":
		return c.Italic
	case "code_bg":
		return c.CodeBg
	case "code_fg":
		return c.CodeFg
	case "table_border":
		return c.TableBorder
	case "table_header":
		return c.TableHeader
	case "status_bar_bg":
		return c.StatusBarBg
	case "status_bar_fg":
		return c.StatusBarFg
	case "search_match_bg":
		return c.SearchMatchBg
	case "search_match_fg":
		return c.SearchMatchFg
	default:
		return ""
	}
}

func (s *SettingsState) getPaletteValue(p theme.Palette, key string) string {
	switch key {
	case "h1":
		return p.H1
	case "h2":
		return p.H2
	case "h3":
		return p.H3
	case "h4":
		return p.H4
	case "h5":
		return p.H5
	case "h6":
		return p.H6
	case "muted":
		return p.Muted
	case "bold":
		return p.Bold
	case "italic":
		return p.Italic
	case "code_bg":
		return p.CodeBg
	case "code_fg":
		return p.CodeFg
	case "table_border":
		return p.TableBorder
	case "table_header":
		return p.TableHeader
	case "status_bar_bg":
		return p.StatusBarBg
	case "status_bar_fg":
		return p.StatusBarFg
	case "search_match_bg":
		return p.SearchMatchBg
	case "search_match_fg":
		return p.SearchMatchFg
	default:
		return ""
	}
}
