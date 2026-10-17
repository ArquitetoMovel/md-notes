package app_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/app"
	"md-notes/internal/buffer"
	"md-notes/internal/markdown"
	"md-notes/internal/ui"
	"md-notes/internal/vim"
)

type appClipboardProvider struct {
	text     string
	readErr  error
	writeErr error
	writes   int
}

func (p *appClipboardProvider) ReadAll() (string, error) { return p.text, p.readErr }
func (p *appClipboardProvider) WriteAll(text string) error {
	p.writes++
	if p.writeErr != nil {
		return p.writeErr
	}
	p.text = text
	return nil
}

func TestApp_CopyShortcutAndQuitFallback(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		selection  bool
		dirty      bool
		writeErr   error
		wantStatus string
		wantQuit   bool
	}{
		{"copy selection", true, false, nil, "Texto copiado", false},
		{"copy failure", true, false, errors.New("unavailable"), "Não foi possível copiar para a área de transferência", false},
		{"quit clean", false, false, nil, "", true},
		{"protect dirty", false, true, nil, "E37", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := buffer.NewEmptyBuffer("copy.md")
			buf.Lines = []buffer.Line{
				buffer.NewLine("# **é**", buffer.EndingCRLF),
				buffer.NewLine("| 🙂 |", buffer.EndingNone),
			}
			buf.IsDirty = tc.dirty
			m := app.NewModel(buf, nil)
			provider := &appClipboardProvider{writeErr: tc.writeErr}
			m.Vim.Clipboard = vim.NewClipboardWithProvider(provider)
			if tc.selection {
				m.Vim.State.CurrentMode = vim.ModeVisualChar
				m.Vim.State.Selection = &vim.Selection{
					Start: vim.Position{Line: 0, Col: 0},
					End:   vim.Position{Line: 1, Col: 4},
					Type:  vim.ModeVisualChar,
				}
			}
			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			if m.Quitting != tc.wantQuit || (cmd != nil) != tc.wantQuit {
				t.Fatalf("quit = %t, cmd = %v; want quit %t", m.Quitting, cmd != nil, tc.wantQuit)
			}
			if !strings.Contains(m.StatusMsg, tc.wantStatus) {
				t.Fatalf("status = %q, want %q", m.StatusMsg, tc.wantStatus)
			}
			if tc.selection {
				if provider.writes != 1 || buf.IsDirty != tc.dirty || buf.Lines[0].String() != "# **é**" {
					t.Fatalf("copy changed document or write count: writes=%d dirty=%t", provider.writes, buf.IsDirty)
				}
				if tc.writeErr == nil {
					if provider.text != "# **é**\r\n| 🙂 |" || m.Vim.State.Selection != nil {
						t.Fatalf("copy content/selection mismatch: %q, %+v", provider.text, m.Vim.State.Selection)
					}
				} else if m.Vim.State.Selection == nil || provider.text != "" {
					t.Fatal("failed copy lost selection or reported content")
				}
			}
		})
	}
}

func TestApp_PasteRouting(t *testing.T) {
	t.Parallel()
	t.Run("bracketed paste takes priority and undoes once", func(t *testing.T) {
		t.Parallel()
		buf := buffer.NewEmptyBuffer("paste.md")
		m := app.NewModel(buf, nil)
		m.Update(makeKeyMsg("i"))
		payload := "# title\n:q dd\x1b"
		m.Update(tea.KeyMsg{Type: tea.KeyCtrlC, Runes: []rune(payload), Paste: true})
		if !buf.IsDirty || !strings.HasPrefix(buf.RawText(), payload) || m.Quitting || m.Vim.History.Count() != 2 {
			t.Fatalf("paste route: text=%q dirty=%t quit=%t history=%d", buf.RawText(), buf.IsDirty, m.Quitting, m.Vim.History.Count())
		}
		m.Update(makeKeyMsg("esc"))
		m.Update(makeKeyMsg("u"))
		if buf.RawText() != "\n" {
			t.Fatalf("undo = %q, want original empty line", buf.RawText())
		}
	})
	for _, tc := range []struct {
		name       string
		text       string
		readErr    error
		bracketed  bool
		wantStatus string
	}{
		{"empty clipboard", "", nil, false, "Área de transferência sem texto para colar"},
		{"clipboard read failure", "", errors.New("unavailable"), false, "Não foi possível ler a área de transferência"},
		{"clipboard limit", strings.Repeat("a", 10001), nil, false, "Texto para colar excede 10.000 caracteres ou 200 linhas"},
		{"bracketed line limit", strings.Repeat("a\n", 200), nil, true, "Texto para colar excede 10.000 caracteres ou 200 linhas"},
		{"invalid UTF-8", "", nil, true, "Texto para colar contém UTF-8 inválido"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := buffer.NewEmptyBuffer("paste.md")
			buf.Lines[0] = buffer.NewLine("ABCDE", buffer.EndingNone)
			m := app.NewModel(buf, nil)
			provider := &appClipboardProvider{text: tc.text, readErr: tc.readErr}
			m.Vim.Clipboard = vim.NewClipboardWithProvider(provider)
			m.Vim.State.CurrentMode = vim.ModeVisualChar
			m.Vim.State.Selection = &vim.Selection{Start: vim.Position{Line: 0, Col: 1}, End: vim.Position{Line: 0, Col: 2}, Type: vim.ModeVisualChar}
			beforeCursor := buf.Cursor
			beforeHistory := m.Vim.History.Count()
			if tc.bracketed {
				runes := []rune(tc.text)
				if tc.name == "invalid UTF-8" {
					runes = []rune{0xfffd}
				}
				m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: runes, Paste: true})
			} else {
				m.Update(tea.KeyMsg{Type: tea.KeyCtrlV})
			}
			if m.StatusMsg != tc.wantStatus || buf.RawText() != "ABCDE" || buf.IsDirty || buf.Cursor != beforeCursor || m.Vim.History.Count() != beforeHistory || m.Vim.State.Selection == nil {
				t.Fatalf("failure changed state: status=%q text=%q dirty=%t cursor=%+v history=%d selection=%+v", m.StatusMsg, buf.RawText(), buf.IsDirty, buf.Cursor, m.Vim.History.Count(), m.Vim.State.Selection)
			}
		})
	}
	t.Run("clipboard replaces selection", func(t *testing.T) {
		t.Parallel()
		buf := buffer.NewEmptyBuffer("paste.md")
		buf.Lines[0] = buffer.NewLine("ABCDE", buffer.EndingNone)
		m := app.NewModel(buf, nil)
		m.Vim.Clipboard = vim.NewClipboardWithProvider(&appClipboardProvider{text: "x\ny"})
		m.Vim.State.CurrentMode = vim.ModeVisualChar
		m.Vim.State.Selection = &vim.Selection{Start: vim.Position{Line: 0, Col: 1}, End: vim.Position{Line: 0, Col: 2}, Type: vim.ModeVisualChar}
		m.Update(tea.KeyMsg{Type: tea.KeyCtrlV})
		if buf.RawText() != "Ax\nyDE" || m.Vim.State.Selection != nil || !buf.IsDirty {
			t.Fatalf("clipboard paste = %q, selection=%+v", buf.RawText(), m.Vim.State.Selection)
		}
	})
	t.Run("insert clipboard and normal visual block", func(t *testing.T) {
		t.Parallel()
		buf := buffer.NewEmptyBuffer("paste.md")
		m := app.NewModel(buf, nil)
		m.Vim.Clipboard = vim.NewClipboardWithProvider(&appClipboardProvider{text: "**é**"})
		m.Update(makeKeyMsg("i"))
		m.Update(tea.KeyMsg{Type: tea.KeyCtrlV})
		if !strings.HasPrefix(buf.RawText(), "**é**") || !buf.IsDirty {
			t.Fatalf("insert clipboard paste = %q", buf.RawText())
		}
		m.Update(makeKeyMsg("esc"))
		m.Update(tea.KeyMsg{Type: tea.KeyCtrlV})
		if m.Vim.State.CurrentMode != vim.ModeVisualBlock {
			t.Fatalf("normal Ctrl+V mode = %v", m.Vim.State.CurrentMode)
		}
	})
}

func TestApp_SettingsBlocksTransfer(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		mode  vim.Mode
		modal bool
	}{
		{"settings", vim.ModeInsert, true},
		{"command", vim.ModeCommand, false},
		{"search", vim.ModeSearch, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := buffer.NewEmptyBuffer("paste.md")
			m := app.NewModel(buf, nil)
			m.Vim.SetMode(tc.mode)
			if tc.modal {
				m.Update(app.OpenSettingsMsg{})
			}
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dd\n:q"), Paste: true})
			m.Update(tea.KeyMsg{Type: tea.KeyCtrlV})
			if buf.IsDirty || buf.RawText() != "\n" || m.Vim.History.Count() != 1 {
				t.Fatalf("blocked mode changed document: mode=%v modal=%t text=%q", tc.mode, tc.modal, buf.RawText())
			}
		})
	}
}

func makeKeyMsg(str string) tea.KeyMsg {
	if str == "esc" {
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	if str == "enter" {
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	if str == "backspace" {
		return tea.KeyMsg{Type: tea.KeyBackspace}
	}
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(str),
	}
}

func TestApp_MouseSelectionLifecycle(t *testing.T) {
	t.Parallel()
	buf := buffer.NewEmptyBuffer("mouse.md")
	for i, text := range []string{"first", "second", "third", "fourth", "fifth", "sixth"} {
		if i == 0 {
			if err := buf.SetLine(0, buffer.NewLine(text, buffer.EndingLF)); err != nil {
				t.Fatal(err)
			}
		} else {
			buf.InsertLine(i, buffer.NewLine(text, buffer.EndingLF))
		}
	}
	buf.IsDirty = false
	m := app.NewModel(buf, nil)
	m.Viewport.SetDimensions(20, 3)
	m.Viewport.Scrolloff = 0
	gutter := ui.CalculateGutterWidth(buf.LineCount())
	m.Update(tea.MouseMsg{X: gutter + 1, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m.Update(tea.MouseMsg{X: gutter + 2, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m.Update(tea.MouseMsg{X: gutter + 2, Y: 1, Action: tea.MouseActionRelease})
	if m.Vim.State.Selection == nil || m.Vim.State.Selection.Start != (vim.Position{Line: 0, Col: 1}) || m.Vim.State.Selection.End != (vim.Position{Line: 1, Col: 2}) {
		t.Fatalf("unexpected mouse selection: %+v", m.Vim.State.Selection)
	}
	if buf.IsDirty || buf.Lines[0].String() != "first" {
		t.Fatal("selection modified buffer")
	}
	m.Update(tea.MouseMsg{X: gutter + 3, Y: 2, Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	if m.Viewport.TopLine != 3 || m.Vim.State.Selection == nil {
		t.Fatalf("wheel lost selection or scroll: top=%d selection=%+v", m.Viewport.TopLine, m.Vim.State.Selection)
	}
	m.Update(tea.MouseMsg{X: gutter, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m.Update(tea.MouseMsg{X: gutter + 2, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m.Update(tea.MouseMsg{X: gutter + 2, Y: 1, Action: tea.MouseActionRelease})
	if m.Vim.State.Selection == nil || m.Vim.State.Selection.Start.Line != 3 || m.Vim.State.Selection.End.Line != 4 {
		t.Fatalf("scroll adjusted selection incorrectly: %+v", m.Vim.State.Selection)
	}
	provider := &appClipboardProvider{}
	m.Vim.Clipboard = vim.NewClipboardWithProvider(provider)
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if provider.text != "fourth\nfif" || m.Vim.State.Selection != nil || buf.IsDirty {
		t.Fatalf("mouse copy changed state: text=%q selection=%+v dirty=%t", provider.text, m.Vim.State.Selection, buf.IsDirty)
	}
	m.Update(tea.MouseMsg{X: gutter, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m.Update(tea.MouseMsg{X: gutter + 2, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m.Update(tea.MouseMsg{X: gutter + 2, Y: 1, Action: tea.MouseActionRelease})
	m.Update(makeKeyMsg("esc"))
	if m.Vim.State.Selection != nil || buf.IsDirty {
		t.Fatal("Esc did not cancel selection cleanly")
	}
	m.Update(tea.MouseMsg{X: gutter, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m.Update(tea.MouseMsg{X: gutter, Y: 0, Action: tea.MouseActionRelease})
	if m.Vim.State.Selection != nil || m.Vim.State.CurrentMode != vim.ModeNormal {
		t.Fatal("simple click left a visual selection")
	}
}

func TestApp_RightClickClipboard(t *testing.T) {
	t.Parallel()
	t.Run("selected text is copied once", func(t *testing.T) {
		t.Parallel()
		buf := buffer.NewEmptyBuffer("mouse.md")
		buf.Lines[0] = buffer.NewLine("# é🙂", buffer.EndingNone)
		m := app.NewModel(buf, nil)
		m.Viewport.SetDimensions(20, 4)
		provider := &appClipboardProvider{}
		m.Vim.Clipboard = vim.NewClipboardWithProvider(provider)
		gutter := ui.CalculateGutterWidth(buf.LineCount())
		m.Update(tea.MouseMsg{X: gutter + 2, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
		m.Update(tea.MouseMsg{X: gutter + 3, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
		m.Update(tea.MouseMsg{X: gutter + 3, Y: 0, Action: tea.MouseActionRelease})
		if m.Vim.State.Selection == nil {
			t.Fatal("mouse drag did not create selection")
		}
		beforeCursor := buf.Cursor
		m.Update(tea.MouseMsg{X: 19, Y: 0, Button: tea.MouseButtonRight, Action: tea.MouseActionPress})
		m.Update(tea.MouseMsg{X: 19, Y: 0, Button: tea.MouseButtonRight, Action: tea.MouseActionRelease})
		if provider.text != "é🙂" || provider.writes != 1 || m.Vim.State.Selection != nil || m.Vim.State.CurrentMode != vim.ModeNormal || m.StatusMsg != "Texto copiado" || buf.IsDirty || buf.Cursor != beforeCursor {
			t.Fatalf("right copy: text=%q writes=%d selection=%+v status=%q dirty=%t cursor=%+v", provider.text, provider.writes, m.Vim.State.Selection, m.StatusMsg, buf.IsDirty, buf.Cursor)
		}
	})
	t.Run("empty area pastes at clicked point and undoes once", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			name string
			x    int
			y    int
			want string
		}{
			{"inside line", ui.CalculateGutterWidth(1) + 1, 0, "a# é\n🙂bc"},
			{"after line text", 19, 0, "abc# é\n🙂"},
			{"below document", 19, 2, "abc# é\n🙂"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				buf := buffer.NewEmptyBuffer("mouse.md")
				buf.Lines[0] = buffer.NewLine("abc", buffer.EndingNone)
				m := app.NewModel(buf, nil)
				m.Viewport.SetDimensions(20, 4)
				m.Vim.Clipboard = vim.NewClipboardWithProvider(&appClipboardProvider{text: "# é\n🙂"})
				m.Update(tea.MouseMsg{X: tc.x, Y: tc.y, Button: tea.MouseButtonRight, Action: tea.MouseActionPress})
				if got := buf.RawText(); got != tc.want || !buf.IsDirty || buf.Cursor != (buffer.Cursor{Line: 1, Col: 1}) || m.Vim.History.Count() != 2 || m.Vim.State.CurrentMode != vim.ModeNormal {
					t.Fatalf("right paste: text=%q dirty=%t cursor=%+v history=%d mode=%v", got, buf.IsDirty, buf.Cursor, m.Vim.History.Count(), m.Vim.State.CurrentMode)
				}
				m.Update(makeKeyMsg("u"))
				if buf.RawText() != "abc" || buf.IsDirty {
					t.Fatalf("undo = %q, dirty=%t", buf.RawText(), buf.IsDirty)
				}
			})
		}
	})
	t.Run("clipboard failures preserve document and selection", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			name       string
			selection  bool
			provider   appClipboardProvider
			wantStatus string
		}{
			{"copy unavailable", true, appClipboardProvider{writeErr: errors.New("unavailable")}, "Não foi possível copiar para a área de transferência"},
			{"paste unavailable", false, appClipboardProvider{readErr: errors.New("unavailable")}, "Não foi possível ler a área de transferência"},
			{"clipboard empty", false, appClipboardProvider{}, "Área de transferência sem texto para colar"},
			{"paste too large", false, appClipboardProvider{text: strings.Repeat("x", 10001)}, "Texto para colar excede 10.000 caracteres ou 200 linhas"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				buf := buffer.NewEmptyBuffer("mouse.md")
				buf.Lines[0] = buffer.NewLine("abc", buffer.EndingNone)
				m := app.NewModel(buf, nil)
				m.Viewport.SetDimensions(20, 4)
				provider := tc.provider
				m.Vim.Clipboard = vim.NewClipboardWithProvider(&provider)
				if tc.selection {
					m.Vim.State.CurrentMode = vim.ModeVisualChar
					m.Vim.State.Selection = &vim.Selection{Start: vim.Position{Line: 0, Col: 0}, End: vim.Position{Line: 0, Col: 1}, Type: vim.ModeVisualChar}
				}
				beforeCursor := buf.Cursor
				m.Update(tea.MouseMsg{X: 19, Y: 0, Button: tea.MouseButtonRight, Action: tea.MouseActionPress})
				if buf.RawText() != "abc" || buf.IsDirty || buf.Cursor != beforeCursor || m.Vim.History.Count() != 1 || (m.Vim.State.Selection != nil) != tc.selection || m.StatusMsg != tc.wantStatus {
					t.Fatalf("failure changed state: text=%q dirty=%t cursor=%+v history=%d selection=%+v status=%q", buf.RawText(), buf.IsDirty, buf.Cursor, m.Vim.History.Count(), m.Vim.State.Selection, m.StatusMsg)
				}
			})
		}
	})
	t.Run("settings and status row ignore right click", func(t *testing.T) {
		t.Parallel()
		buf := buffer.NewEmptyBuffer("mouse.md")
		m := app.NewModel(buf, nil)
		m.Viewport.SetDimensions(20, 4)
		provider := &appClipboardProvider{text: "paste"}
		m.Vim.Clipboard = vim.NewClipboardWithProvider(provider)
		m.Update(tea.MouseMsg{X: 5, Y: 4, Button: tea.MouseButtonRight, Action: tea.MouseActionPress})
		m.SettingsOpen = true
		m.Update(tea.MouseMsg{X: 5, Y: 0, Button: tea.MouseButtonRight, Action: tea.MouseActionPress})
		if buf.IsDirty || buf.RawText() != "\n" || m.Vim.History.Count() != 1 || provider.writes != 0 {
			t.Fatal("right click outside document changed buffer")
		}
	})
}

func TestApp_MouseClipboardPasteUTF8CRLF(t *testing.T) {
	t.Parallel()
	buf := buffer.NewEmptyBuffer("integration.md")
	buf.Lines = []buffer.Line{
		buffer.NewLine("# 界🙂 **forte**", buffer.EndingCRLF),
		buffer.NewLine("| célula |", buffer.EndingNone),
	}
	original := buf.RawText()
	m := app.NewModel(buf, nil)
	m.Viewport.SetDimensions(18, 6)
	m.Viewport.Scrolloff = 0
	provider := &appClipboardProvider{}
	m.Vim.Clipboard = vim.NewClipboardWithProvider(provider)
	gutter := ui.CalculateGutterWidth(buf.LineCount())
	selectText := func() {
		visible := m.Viewport.GetVisibleLines(buf)
		start := ui.LogicalToVisual(buffer.Cursor{Line: 0, Col: 2}, visible, gutter)
		end := ui.LogicalToVisual(buffer.Cursor{Line: 1, Col: 4}, visible, gutter)
		m.Update(tea.MouseMsg{X: start.ScreenX + 1, Y: start.ScreenY, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
		m.Update(tea.MouseMsg{X: end.ScreenX, Y: end.ScreenY, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
		m.Update(tea.MouseMsg{X: end.ScreenX, Y: end.ScreenY, Action: tea.MouseActionRelease})
	}
	selectText()
	if m.Vim.State.Selection == nil || m.Vim.State.Selection.Start != (vim.Position{Line: 0, Col: 2}) || m.Vim.State.Selection.End != (vim.Position{Line: 1, Col: 4}) {
		t.Fatalf("Unicode mouse selection = %+v", m.Vim.State.Selection)
	}
	selectedView := m.View()
	if !strings.Contains(selectedView, "-- VISUAL --") || !strings.Contains(selectedView, m.Theme.Selection.Render("界")) {
		t.Fatal("selection highlight or status missing after release")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if provider.text != "界🙂 **forte**\r\n| cél" || m.Vim.State.Selection != nil || buf.IsDirty || !strings.Contains(m.View(), "Texto copiado") {
		t.Fatalf("copy state: clipboard=%q selection=%+v dirty=%t status=%q", provider.text, m.Vim.State.Selection, buf.IsDirty, m.StatusMsg)
	}
	selectText()
	if m.StatusMsg != "" {
		t.Fatalf("new selection retained stale status: %q", m.StatusMsg)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("é\r\n# novo"), Paste: true})
	if got := buf.RawText(); got != "# é\r\n# novoula |" || !buf.IsDirty || m.Vim.State.Selection != nil || buf.Cursor != (buffer.Cursor{Line: 1, Col: 6}) {
		t.Fatalf("paste state: text=%q cursor=%+v dirty=%t selection=%+v", got, buf.Cursor, buf.IsDirty, m.Vim.State.Selection)
	}
	if view := m.View(); !strings.Contains(view, "[+]") || strings.Contains(view, "-- VISUAL --") {
		t.Fatal("paste did not refresh dirty indicator or selection status")
	}
	m.Update(makeKeyMsg("u"))
	if buf.RawText() != original || buf.IsDirty {
		t.Fatalf("undo = %q, dirty=%t", buf.RawText(), buf.IsDirty)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if buf.RawText() != "# é\r\n# novoula |" || !buf.IsDirty {
		t.Fatalf("redo = %q, dirty=%t", buf.RawText(), buf.IsDirty)
	}
}

// Test Bubble Tea key delegation to Vim Engine (Insert Mode, View indicator)
func TestApp_VimKeyDelegation_InsertMode(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)

	// Press 'i' to enter insert mode
	m.Update(makeKeyMsg("i"))
	view := m.View()
	if !strings.Contains(view, "-- INSERT --") {
		t.Errorf("View esperado com '-- INSERT --', obtido:\n%s", view)
	}

	// Type text
	for _, ch := range "Hello Vim" {
		m.Update(makeKeyMsg(string(ch)))
	}

	line0, _ := buf.GetLine(0)
	if line0.String() != "Hello Vim" {
		t.Errorf("Buffer esperado 'Hello Vim', obtido '%s'", line0.String())
	}

	// Press 'Esc' to return to normal mode
	m.Update(makeKeyMsg("esc"))
	if m.Vim.State.CurrentMode != vim.ModeNormal {
		t.Errorf("Esperado retorno para ModeNormal, obtido %s", m.Vim.State.CurrentMode)
	}
}

// Test Ex Command prompt rendering and execution in Bubble Tea
func TestApp_VimCommandMode_SaveAndQuit(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "notas_salvas.md")

	buf := buffer.NewEmptyBuffer(filePath)
	buf.InsertText("Linha gravada via :w")
	m := app.NewModel(buf, nil)

	// Press ':' to enter command mode
	m.Update(makeKeyMsg(":"))
	view := m.View()
	if !strings.Contains(view, "\n:") {
		t.Errorf("View esperado com prompt ':', obtido:\n%s", view)
	}

	// Type "w" and press Enter
	m.Update(makeKeyMsg("w"))
	_, cmd := m.Update(makeKeyMsg("enter"))

	if cmd == nil {
		t.Fatal("Esperado tea.Cmd retornado para :w")
	}

	// Execute the returned cmd (which dispatches SaveFileMsg)
	saveMsg := cmd()
	_, saveResultCmd := m.Update(saveMsg)

	if saveResultCmd == nil {
		t.Fatal("Esperado SaveResultMsg disparado após salvar")
	}
	saveResult := saveResultCmd()
	m.Update(saveResult)

	if buf.IsDirty {
		t.Error("Buffer não deveria estar dirty após :w")
	}
}

// Test Ex command unknown error (E492)
func TestApp_VimCommandMode_UnknownCommand_E492(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)

	m.Update(makeKeyMsg(":"))
	for _, ch := range "comando_desconhecido" {
		m.Update(makeKeyMsg(string(ch)))
	}
	m.Update(makeKeyMsg("enter"))

	if !strings.Contains(m.StatusMsg, "E492: Não é um comando de editor: comando_desconhecido") {
		t.Errorf("StatusMsg esperado com erro E492, obtido: %q", m.StatusMsg)
	}
}

// Test cursor highlight during Vim navigation (h, j, k, l)
func TestApp_VimNavigation_CursorHighlighted(t *testing.T) {
	buf := buffer.NewEmptyBuffer("navegacao.md")
	buf.InsertText("Linha 1\nLinha 2\n")
	buf.Cursor = buffer.Cursor{Line: 0, Col: 0}

	m := app.NewModel(buf, nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Initial position (Line 0, Col 0) -> status indicates Ln 1, Col 1 and shape sequence \x1b[2 q
	view0 := m.View()
	if !strings.Contains(view0, "Ln 1, Col 1") {
		t.Errorf("view missing status pos 'Ln 1, Col 1', got:\n%s", view0)
	}
	if !strings.Contains(view0, "\x1b[2 q") {
		t.Errorf("view missing normal mode block cursor shape sequence, got:\n%s", view0)
	}

	// Move right with 'l'
	m.Update(makeKeyMsg("l"))
	if buf.Cursor.Col != 1 {
		t.Fatalf("expected cursor Col 1, got %d", buf.Cursor.Col)
	}
	view1 := m.View()
	if !strings.Contains(view1, "Ln 1, Col 2") {
		t.Errorf("view missing status pos 'Ln 1, Col 2', got:\n%s", view1)
	}

	// Move down with 'j'
	m.Update(makeKeyMsg("j"))
	if buf.Cursor.Line != 1 {
		t.Fatalf("expected cursor Line 1, got %d", buf.Cursor.Line)
	}
	view2 := m.View()
	if !strings.Contains(view2, "Ln 2, Col 2") {
		t.Errorf("view missing status pos 'Ln 2, Col 2', got:\n%s", view2)
	}
}

// Test markdown headings pre-formatting in View()
func TestApp_MarkdownHeadings_Formatted(t *testing.T) {
	buf := buffer.NewEmptyBuffer("headings.md")
	buf.InsertText("# Título H1\n## Subtítulo H2\nTexto normal\n")

	m := app.NewModel(buf, nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()

	// Should contain the title text
	if !strings.Contains(view, "Título H1") {
		t.Errorf("view missing 'Título H1'")
	}
	if !strings.Contains(view, "Subtítulo H2") {
		t.Errorf("view missing 'Subtítulo H2'")
	}
	if !strings.Contains(view, "Texto normal") {
		t.Errorf("view missing 'Texto normal'")
	}

	// Verify that the markdown parser parsed the headings
	orig0, _ := buf.GetLine(0)
	var inCode bool
	var lang string
	tl := m.MarkdownParser.ParseLine(orig0, &inCode, &lang)
	if !tl.IsHeading || tl.HeadingLvl != 1 {
		t.Errorf("expected parsed line 0 to be Heading level 1")
	}
	if len(tl.Spans) < 2 || tl.Spans[0].Type != markdown.TokenMarker || tl.Spans[1].Type != markdown.TokenHeading {
		t.Errorf("expected marker and heading spans, got: %+v", tl.Spans)
	}
}
