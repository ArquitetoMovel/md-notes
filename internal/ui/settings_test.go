package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"md-notes/internal/config"
)

// TC04: Navegação Cíclica Entre Abas (Tab, Shift+Tab, setas e h/l)
func TestSettingsState_TabNavigation(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	if state.ActiveTab != TabThemes {
		t.Fatalf("Esperado aba inicial TabThemes, obtido: %v", state.ActiveTab)
	}

	// 1. Tab transita ordenadamente: TabThemes -> TabEditor -> TabColors -> TabThemes
	state.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
	if state.ActiveTab != TabEditor {
		t.Errorf("Após 1º Tab, esperado TabEditor, obtido: %v", state.ActiveTab)
	}

	state.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
	if state.ActiveTab != TabColors {
		t.Errorf("Após 2º Tab, esperado TabColors, obtido: %v", state.ActiveTab)
	}

	state.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
	if state.ActiveTab != TabThemes {
		t.Errorf("Após 3º Tab, esperado retorno a TabThemes, obtido: %v", state.ActiveTab)
	}

	// 2. Navegação com seta para a direita (Right) e l
	state.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	if state.ActiveTab != TabEditor {
		t.Errorf("Após Right, esperado TabEditor, obtido: %v", state.ActiveTab)
	}

	// 3. Navegação reversa com Shift+Tab, Left e h
	state.HandleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	if state.ActiveTab != TabThemes {
		t.Errorf("Após Shift+Tab, esperado TabThemes, obtido: %v", state.ActiveTab)
	}

	state.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if state.ActiveTab != TabColors {
		t.Errorf("Após Left a partir de TabThemes, esperado TabColors, obtido: %v", state.ActiveTab)
	}
}

// TC05: Navegação Vertical Entre Itens com Limites de Borda (j/k e Up/Down)
func TestSettingsState_VerticalNavigation(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	state.ActiveTab = TabThemes
	state.ActiveThemeIndex = 0

	// 1. k / Up no limite superior (não ultrapassa 0)
	state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if state.ActiveThemeIndex != 0 {
		t.Errorf("Cursor subiu além do limite 0: %d", state.ActiveThemeIndex)
	}
	state.HandleKey(tea.KeyMsg{Type: tea.KeyUp})
	if state.ActiveThemeIndex != 0 {
		t.Errorf("Cursor subiu além do limite 0 via Up: %d", state.ActiveThemeIndex)
	}

	// 2. j / Down 6 vezes até o último tema (índice 6 - monokai)
	for i := 0; i < 6; i++ {
		state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	if state.ActiveThemeIndex != 6 {
		t.Fatalf("Esperado índice 6 após 6 j, obtido: %d", state.ActiveThemeIndex)
	}

	// 3. j mais uma vez (permanece em 6, não ultrapassa)
	state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if state.ActiveThemeIndex != 6 {
		t.Errorf("Cursor desceu além do limite 6: %d", state.ActiveThemeIndex)
	}

	state.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
	if state.ActiveThemeIndex != 6 {
		t.Errorf("Cursor desceu além do limite 6 via Down: %d", state.ActiveThemeIndex)
	}
}

// TC06: Listagem e Seleção dos 7 Temas Incorporados
func TestSettingsState_ThemeSelection(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	expectedThemes := []string{
		"default-dark",
		"default-light",
		"dracula",
		"nord",
		"catppuccin-mocha",
		"catppuccin-macchiato",
		"monokai",
	}

	if len(state.AvailableThemes) != 7 {
		t.Fatalf("Esperado exatamente 7 temas, obtido: %d", len(state.AvailableThemes))
	}
	for i, name := range expectedThemes {
		if state.AvailableThemes[i] != name {
			t.Errorf("Tema no índice %d esperado '%s', obtido '%s'", i, name, state.AvailableThemes[i])
		}
	}

	// Move cursor até "dracula" (índice 2)
	state.ActiveThemeIndex = 2
	action := state.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated, obtido: %v", action)
	}
	if state.WorkingConfig.Theme != "dracula" {
		t.Errorf("Esperado tema 'dracula' selecionado, obtido: '%s'", state.WorkingConfig.Theme)
	}

	// Move cursor até "nord" (índice 3) e confirma com Space
	state.ActiveThemeIndex = 3
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated com Space, obtido: %v", action)
	}
	if state.WorkingConfig.Theme != "nord" {
		t.Errorf("Esperado tema 'nord' selecionado, obtido: '%s'", state.WorkingConfig.Theme)
	}
}

// TC07: Alternância de Preferências Booleanas do Editor
func TestSettingsState_EditorOptionsToggle(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Editor.LineNumbers = true
	cfg.Editor.RelativeLineNumbers = false
	cfg.Editor.WordWrap = true

	state := NewSettingsState(cfg)
	state.ActiveTab = TabEditor
	state.ActiveEditorRow = 0 // line_numbers

	// 1. Alterna line_numbers com Space (true -> false)
	action := state.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated, obtido: %v", action)
	}
	if state.WorkingConfig.Editor.LineNumbers {
		t.Errorf("Esperado line_numbers == false após alternar")
	}

	// 2. Alterna line_numbers novamente com Space (false -> true)
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated, obtido: %v", action)
	}
	if !state.WorkingConfig.Editor.LineNumbers {
		t.Errorf("Esperado line_numbers == true após 2º alternar")
	}

	// 3. Alterna relative_line_numbers (índice 1) com Enter
	state.ActiveEditorRow = 1
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated, obtido: %v", action)
	}
	if !state.WorkingConfig.Editor.RelativeLineNumbers {
		t.Errorf("Esperado relative_line_numbers == true após alternar")
	}

	// 4. Alterna word_wrap (índice 4)
	state.ActiveEditorRow = 4
	state.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if state.WorkingConfig.Editor.WordWrap {
		t.Errorf("Esperado word_wrap == false após alternar")
	}
}

// TC08: Ciclo Sequencial de Tamanho de Tabulação (2 -> 4 -> 8 -> 2)
func TestSettingsState_TabSizeCycle(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Editor.TabSize = 4

	state := NewSettingsState(cfg)
	state.ActiveTab = TabEditor
	state.ActiveEditorRow = 2 // tab_size

	// 4 -> 8
	action := state.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated, obtido: %v", action)
	}
	if state.WorkingConfig.Editor.TabSize != 8 {
		t.Errorf("Esperado TabSize 8, obtido: %d", state.WorkingConfig.Editor.TabSize)
	}

	// 8 -> 2
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated, obtido: %v", action)
	}
	if state.WorkingConfig.Editor.TabSize != 2 {
		t.Errorf("Esperado TabSize 2, obtido: %d", state.WorkingConfig.Editor.TabSize)
	}

	// 2 -> 4
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated, obtido: %v", action)
	}
	if state.WorkingConfig.Editor.TabSize != 4 {
		t.Errorf("Esperado TabSize 4, obtido: %d", state.WorkingConfig.Editor.TabSize)
	}
}

// TC09: Ajuste com Limites Numéricos de Scrolloff (0 a 10)
func TestSettingsState_ScrolloffLimits(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Editor.Scrolloff = 4

	state := NewSettingsState(cfg)
	state.ActiveTab = TabEditor
	state.ActiveEditorRow = 3 // scrolloff

	// 1. Incrementa repetidamente até ultrapassar 10 (deve limitar estritamente em 10)
	for i := 0; i < 15; i++ {
		state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	}
	if state.WorkingConfig.Editor.Scrolloff != 10 {
		t.Errorf("Esperado scrolloff limitado em 10, obtido: %d", state.WorkingConfig.Editor.Scrolloff)
	}

	// 2. Decrementa repetidamente até tentar reduzir abaixo de 0 (deve limitar estritamente em 0)
	for i := 0; i < 20; i++ {
		state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	}
	if state.WorkingConfig.Editor.Scrolloff != 0 {
		t.Errorf("Esperado scrolloff limitado em 0, obtido: %d", state.WorkingConfig.Editor.Scrolloff)
	}

	// 3. Ajuste horizontal com Right e Left na linha de scrolloff
	state.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	if state.WorkingConfig.Editor.Scrolloff != 1 {
		t.Errorf("Esperado scrolloff 1 após Right, obtido: %d", state.WorkingConfig.Editor.Scrolloff)
	}

	state.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if state.WorkingConfig.Editor.Scrolloff != 0 {
		t.Errorf("Esperado scrolloff 0 após Left, obtido: %d", state.WorkingConfig.Editor.Scrolloff)
	}
}

// TC10 & TC11: Validação e Rejeição de Códigos Hexadecimais
func TestSettingsState_ColorHexValidation(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	state.ActiveTab = TabColors
	state.ActiveColorRow = 0 // h1

	// Inicia modo de edição inline via 'e'
	state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if !state.IsEditingHex {
		t.Fatalf("Esperado IsEditingHex == true após pressionar 'e'")
	}

	// Limpa o buffer com Backspace
	for i := 0; i < 10; i++ {
		state.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	if state.HexInputBuffer != "" {
		t.Errorf("Buffer deveria estar vazio após backspaces, obtido: '%s'", state.HexInputBuffer)
	}

	// Digita "#FF79C6"
	for _, r := range "#FF79C6" {
		state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if state.HexInputBuffer != "#FF79C6" {
		t.Errorf("Buffer esperado '#FF79C6', obtido: '%s'", state.HexInputBuffer)
	}

	// Pressiona Enter para confirmar (TC10: valor válido)
	action := state.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != ActionLivePreviewUpdated {
		t.Errorf("Esperado ActionLivePreviewUpdated para hex válido, obtido: %v", action)
	}
	if state.IsEditingHex {
		t.Errorf("Modo de edição deveria ter sido fechado após validação bem-sucedida")
	}
	if state.WorkingConfig.Colors.H1 != "#FF79C6" {
		t.Errorf("Esperado Colors.H1 == '#FF79C6', obtido: '%s'", state.WorkingConfig.Colors.H1)
	}

	// TC11: Rejeição de entradas inválidas
	// Abre edição novamente
	state.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !state.IsEditingHex {
		t.Fatalf("Esperado IsEditingHex == true após Enter")
	}

	// Testa entrada inválida: "#123"
	for i := 0; i < 10; i++ {
		state.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	for _, r := range "#123" {
		state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	action = state.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != ActionNone {
		t.Errorf("Esperado ActionNone para hex inválido, obtido: %v", action)
	}
	if !state.IsEditingHex {
		t.Errorf("Deveria permanecer em modo de edição após erro")
	}
	if state.HexInputError != "Código de cor inválido: use formato #RRGGBB" {
		t.Errorf("Mensagem de erro não esperada: '%s'", state.HexInputError)
	}
	if state.WorkingConfig.Colors.H1 != "#FF79C6" {
		t.Errorf("Cor anterior deveria ter sido preservada, obtido: '%s'", state.WorkingConfig.Colors.H1)
	}

	// Cancela com Esc
	state.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if state.IsEditingHex {
		t.Errorf("Esperado sair do modo de edição após Esc")
	}
}

// Ações de Cancelamento (Esc, q) e Salvamento ([Salvar e Fechar], ctrl+s)
func TestSettingsState_CancelAndSaveActions(t *testing.T) {
	t.Parallel()

	// 1. Esc fora de edição retorna ActionCancelAndClose
	state := NewSettingsState(nil)
	action := state.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if action != ActionCancelAndClose {
		t.Errorf("Esperado ActionCancelAndClose para Esc, obtido: %v", action)
	}

	// 2. 'q' fora de edição retorna ActionCancelAndClose
	state = NewSettingsState(nil)
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if action != ActionCancelAndClose {
		t.Errorf("Esperado ActionCancelAndClose para 'q', obtido: %v", action)
	}

	// 3. 'ctrl+s' retorna ActionSaveAndClose diretamente
	state = NewSettingsState(nil)
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlS})
	if action != ActionSaveAndClose {
		t.Errorf("Esperado ActionSaveAndClose para ctrl+s, obtido: %v", action)
	}

	// 4. 's' foca o botão de salvar e Enter confirma ActionSaveAndClose
	state = NewSettingsState(nil)
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if !state.FocusSaveButton {
		t.Errorf("Esperado FocusSaveButton == true após pressionar 's'")
	}

	// Enter no botão focado dispara ActionSaveAndClose
	action = state.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != ActionSaveAndClose {
		t.Errorf("Esperado ActionSaveAndClose após Enter no botão focado, obtido: %v", action)
	}

	// 4. Teste de desfocar o botão de salvar com 'k' / 'Up'
	state.FocusSaveButton = true
	state.HandleKey(tea.KeyMsg{Type: tea.KeyUp})
	if state.FocusSaveButton {
		t.Errorf("Esperado FocusSaveButton == false após Up")
	}
}

// Renderização da aba de Temas com lista dos 7 temas e indicadores visuais
func TestRenderSettingsModal_Themes(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	state.ActiveTab = TabThemes
	rendered := RenderSettingsModal(state, nil, 80, 24)

	if rendered == "" {
		t.Fatal("RenderSettingsModal retornou string vazia")
	}

	// Verifica título centralizado
	if !strings.Contains(rendered, "Configurações") {
		t.Errorf("Modal não contém o título 'Configurações'")
	}

	// Verifica abas
	if !strings.Contains(rendered, "[1. Temas]") || !strings.Contains(rendered, "[2. Editor]") || !strings.Contains(rendered, "[3. Cores]") {
		t.Errorf("Modal não contém os cabeçalhos de abas esperados")
	}

	// Verifica presença dos 7 temas incorporados
	for _, themeName := range state.AvailableThemes {
		if !strings.Contains(rendered, themeName) {
			t.Errorf("Tema '%s' não encontrado no modal renderizado", themeName)
		}
	}

	// Verifica marcador de tema ativo e botão de salvar
	if !strings.Contains(rendered, "(•)") {
		t.Errorf("Marcador de tema ativo '(•)' não encontrado no modal")
	}
	if !strings.Contains(rendered, "[Salvar e Fechar]") {
		t.Errorf("Botão '[Salvar e Fechar]' não encontrado no modal")
	}
	if !strings.Contains(rendered, "Tab: Abas") {
		t.Errorf("Rodapé com atalhos não encontrado no modal")
	}
}

// Renderização da aba Editor com opções, checkboxes e seletores numéricos
func TestRenderSettingsModal_Editor(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	state.ActiveTab = TabEditor
	rendered := RenderSettingsModal(state, nil, 80, 24)

	if !strings.Contains(rendered, "[2. Editor]") {
		t.Errorf("Aba '[2. Editor]' não identificada no modal")
	}

	// Verifica se as opções do editor são exibidas
	for _, opt := range state.EditorOptions {
		if !strings.Contains(rendered, opt.Label) {
			t.Errorf("Opção de editor '%s' não encontrada no modal", opt.Label)
		}
	}

	// Verifica presença de checkbox e formato numérico
	if !strings.Contains(rendered, "[✓]") && !strings.Contains(rendered, "[ ]") {
		t.Errorf("Checkboxes não encontrados no modal de editor")
	}
	if !strings.Contains(rendered, "< 4 >") {
		t.Errorf("Valor numérico '< 4 >' não encontrado no modal de editor")
	}
}

// Renderização da aba Cores com amostras, códigos hex e modo de edição inline
func TestRenderSettingsModal_Colors(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	state.ActiveTab = TabColors
	rendered := RenderSettingsModal(state, nil, 80, 24)

	if !strings.Contains(rendered, "[3. Cores]") {
		t.Errorf("Aba '[3. Cores]' não identificada no modal")
	}

	// Verifica presença de tokens de cores e swatch
	if !strings.Contains(rendered, "Título H1") {
		t.Errorf("Token 'Título H1' não encontrado na aba de cores")
	}
	if !strings.Contains(rendered, "■") {
		t.Errorf("Amostra visual (swatch) '■' não encontrada na aba de cores")
	}

	// Testa modo de edição hex ativo
	state.IsEditingHex = true
	state.HexInputBuffer = "#BD93F9"
	editRendered := RenderSettingsModal(state, nil, 80, 24)
	if !strings.Contains(editRendered, "Hex: [#BD93F9_]") {
		t.Errorf("Prompt de edição inline 'Hex: [#BD93F9_]' não encontrado no modal")
	}

	// Testa exibição de erro de validação hex
	state.HexInputError = "Código de cor inválido: use formato #RRGGBB"
	errRendered := RenderSettingsModal(state, nil, 80, 24)
	if !strings.Contains(errRendered, "Código de cor inválido: use formato #RRGGBB") {
		t.Errorf("Mensagem de erro de código hex não encontrada no modal")
	}
}

// Renderização de alerta de erro de persistência em disco
func TestRenderSettingsModal_PersistenceError(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	state.PersistenceError = "permissão negada"
	rendered := RenderSettingsModal(state, nil, 80, 24)

	expectedMsg := "Erro ao gravar config.toml: permissão negada"
	if !strings.Contains(rendered, expectedMsg) {
		t.Errorf("Alerta de persistência esperado '%s' não encontrado no modal", expectedMsg)
	}
}

// TC18: Renderização Centralizada e Algoritmo de Overlay Blending
func TestOverlayModal_CenteringAndBlending(t *testing.T) {
	t.Parallel()

	// Cria 20 linhas de fundo preenchidas com 80 caracteres
	termWidth := 80
	termHeight := 24
	bgLines := make([]string, 20)
	for i := 0; i < 20; i++ {
		bgLines[i] = fmt.Sprintf("Line %02d: %s", i, strings.Repeat(".", 71))
	}

	// Gera modal estilizado com width 64 e height 20 (clamp para modalWidth=56, modalHeight=16)
	state := NewSettingsState(nil)
	modalBox := RenderSettingsModal(state, nil, 64, 20)

	modalLines := strings.Split(modalBox, "\n")
	modalHeight := len(modalLines)
	if modalHeight != 16 {
		t.Fatalf("Esperado modalHeight == 16, obtido: %d", modalHeight)
	}
	modalWidth := lipgloss.Width(modalLines[0])
	if modalWidth != 56 {
		t.Fatalf("Esperado modalWidth == 56, obtido: %d", modalWidth)
	}

	blended := OverlayModal(bgLines, modalBox, termWidth, termHeight)
	if len(blended) != len(bgLines) {
		t.Fatalf("Número de linhas mescladas esperado %d, obtido %d", len(bgLines), len(blended))
	}

	topY := (termHeight - modalHeight) / 2 // (24 - 16) / 2 = 4
	leftX := (termWidth - modalWidth) / 2  // (80 - 56) / 2 = 12

	// 1. Linhas acima do modal (0 até topY-1 = 3) devem permanecer idênticas ao fundo
	for y := 0; y < topY; y++ {
		if blended[y] != bgLines[y] {
			t.Errorf("Linha superior %d foi alterada: esperado '%s', obtido '%s'", y, bgLines[y], blended[y])
		}
	}

	// 2. Linhas na região do modal (topY até topY+modalHeight-1 = 4 até 19)
	for y := topY; y < topY+modalHeight; y++ {
		line := blended[y]

		// Margem esquerda (primeiros 12 caracteres) deve preservar caracteres originais
		expectedLeft := bgLines[y][:leftX]
		if !strings.HasPrefix(line, expectedLeft) {
			t.Errorf("Margem esquerda da linha %d corrompida: esperado prefixo '%s'", y, expectedLeft)
		}

		// Margem direita (após leftX + modalWidth = 68) deve preservar caracteres originais
		expectedRight := bgLines[y][leftX+modalWidth:]
		if !strings.HasSuffix(line, expectedRight) {
			t.Errorf("Margem direita da linha %d corrompida: esperado sufixo '%s'", y, expectedRight)
		}
	}

	// 3. Teste de resiliência com entradas vazias
	emptyBlend := OverlayModal(nil, modalBox, termWidth, termHeight)
	if len(emptyBlend) != modalHeight {
		t.Errorf("Esperado fallback com linhas do modal para fundo vazio")
	}

	noopBlend := OverlayModal(bgLines, "", termWidth, termHeight)
	if len(noopBlend) != len(bgLines) {
		t.Errorf("Esperado cópia do fundo para modal vazio")
	}
}

func TestSettingsState_SaveButtonNavigationAndAction(t *testing.T) {
	t.Parallel()

	state := NewSettingsState(nil)
	state.ActiveTab = TabEditor
	state.ActiveEditorRow = len(state.EditorOptions) - 1 // último item (word_wrap)
	state.FocusSaveButton = false

	// Pressionar Down no último item deve focar o botão [Salvar e Fechar]
	state.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
	if !state.FocusSaveButton {
		t.Fatalf("Esperado FocusSaveButton = true após Down no último item")
	}

	// Pressionar Enter no botão focado deve retornar ActionSaveAndClose
	action := state.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != ActionSaveAndClose {
		t.Errorf("Esperado ActionSaveAndClose ao pressionar Enter no botão Salvar, obtido: %v", action)
	}

	// Resetar e testar tecla Up para desfocar
	state.FocusSaveButton = true
	state.HandleKey(tea.KeyMsg{Type: tea.KeyUp})
	if state.FocusSaveButton {
		t.Errorf("Esperado FocusSaveButton = false após Up no botão Salvar")
	}
}

