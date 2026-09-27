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
	writeErr error
	writes   int
}

func (p *appClipboardProvider) ReadAll() (string, error) { return p.text, nil }
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
