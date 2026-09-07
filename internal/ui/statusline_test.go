package ui

import (
	"strings"
	"testing"

	"md-notes/internal/buffer"
	"md-notes/internal/theme"
	"md-notes/internal/vim"
)

// TC08
func TestStatusLine_ModeSegments(t *testing.T) {
	th := getTestTheme()

	// Mode Normal
	stateNormal := StatusState{
		Mode:       vim.ModeNormal,
		FilePath:   "docs/PRD.md",
		IsDirty:    false,
		CursorLine: 45,
		CursorCol:  12,
		TotalLines: 100,
		LineEnding: buffer.EndingLF,
	}
	sNormal := RenderStatusLine(stateNormal, th, 120)
	if !strings.Contains(sNormal, "NORMAL") || !strings.Contains(sNormal, "docs/PRD.md") || !strings.Contains(sNormal, "Ln 45, Col 12") || !strings.Contains(sNormal, "UTF-8 | LF") {
		t.Errorf("status line missing expected segments: %s", sNormal)
	}

	// Mode Insert
	stateInsert := stateNormal
	stateInsert.Mode = vim.ModeInsert
	sInsert := RenderStatusLine(stateInsert, th, 120)
	if !strings.Contains(sInsert, "INSERT") {
		t.Errorf("status line missing INSERT mode: %s", sInsert)
	}

	// Mode Visual
	stateVisual := stateNormal
	stateVisual.Mode = vim.ModeVisual
	sVisual := RenderStatusLine(stateVisual, th, 120)
	if !strings.Contains(sVisual, "VISUAL") {
		t.Errorf("status line missing VISUAL mode: %s", sVisual)
	}
}

// TC09
func TestStatusLine_DirtyIndicator(t *testing.T) {
	th := getTestTheme()

	stateClean := StatusState{
		FilePath: "notas.md",
		IsDirty:  false,
	}
	sClean := RenderStatusLine(stateClean, th, 100)
	if strings.Contains(sClean, "[+]") {
		t.Errorf("clean state should not contain [+], got %s", sClean)
	}

	stateDirty := StatusState{
		FilePath: "notas.md",
		IsDirty:  true,
	}
	sDirty := RenderStatusLine(stateDirty, th, 100)
	if !strings.Contains(sDirty, "[+]") {
		t.Errorf("dirty state should contain [+], got %s", sDirty)
	}
}

// TC10
func TestStatusLine_PercentageCalculation(t *testing.T) {
	th := getTestTheme()

	// Top
	sTop := RenderStatusLine(StatusState{CursorLine: 1, TotalLines: 100}, th, 100)
	if !strings.Contains(sTop, "Top") {
		t.Errorf("line 1 should show Top, got %s", sTop)
	}

	// 50%
	s50 := RenderStatusLine(StatusState{CursorLine: 50, TotalLines: 100}, th, 100)
	if !strings.Contains(s50, "50%") {
		t.Errorf("line 50 should show 50%%, got %s", s50)
	}

	// Bot
	sBot := RenderStatusLine(StatusState{CursorLine: 100, TotalLines: 100}, th, 100)
	if !strings.Contains(sBot, "Bot") {
		t.Errorf("line 100 should show Bot, got %s", sBot)
	}
}

// TC12
func TestStatusLine_SearchMatches(t *testing.T) {
	th := getTestTheme()

	stateSearch := StatusState{
		SearchMatchCur: 3,
		SearchMatchTot: 15,
		TotalLines:     100,
	}
	s := RenderStatusLine(stateSearch, th, 120)
	if !strings.Contains(s, "[3/15]") {
		t.Errorf("status line missing search count [3/15]: %s", s)
	}
}

// TC14
func TestStatusLine_DynamicTheme(t *testing.T) {
	palNord, _ := theme.GetPalette("nord")
	thNord := theme.CompileTheme(palNord, theme.ColorOverrides{})

	palDracula, _ := theme.GetPalette("dracula")
	thDracula := theme.CompileTheme(palDracula, theme.ColorOverrides{})

	state := StatusState{
		Mode:       vim.ModeNormal,
		FilePath:   "exemplo.md",
		TotalLines: 10,
	}

	sNord := RenderStatusLine(state, thNord, 100)
	sDracula := RenderStatusLine(state, thDracula, 100)

	if sNord == "" || sDracula == "" {
		t.Fatalf("rendered status lines should not be empty")
	}

	// Check command line error styling
	errState := StatusState{StatusMessage: "E37: Não foram gravadas as alterações"}
	cmdLine := RenderCommandLine(errState, thDracula, 100)
	if !strings.Contains(cmdLine, "E37") {
		t.Errorf("command line missing error message: %s", cmdLine)
	}
}

func TestStatusLine_SearchModeCommandLine(t *testing.T) {
	stateForward := StatusState{
		Mode:            vim.ModeSearch,
		CommandInput:    "pesquisa",
		SearchIsReverse: false,
	}
	outForward := RenderCommandLine(stateForward, nil, 80)
	if !strings.Contains(outForward, "/pesquisa") {
		t.Errorf("Esperado '/pesquisa', obtido: '%s'", outForward)
	}

	stateReverse := StatusState{
		Mode:            vim.ModeSearch,
		CommandInput:    "reverso",
		SearchIsReverse: true,
	}
	outReverse := RenderCommandLine(stateReverse, nil, 80)
	if !strings.Contains(outReverse, "?reverso") {
		t.Errorf("Esperado '?reverso', obtido: '%s'", outReverse)
	}

	// Status line mode block
	pal, _ := theme.GetPalette("default-dark")
	th := theme.CompileTheme(pal, theme.ColorOverrides{})
	sl := RenderStatusLine(stateForward, th, 80)
	if !strings.Contains(sl, "SEARCH") {
		t.Errorf("Status line esperada conter 'SEARCH', obtido: '%s'", sl)
	}
}
