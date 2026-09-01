package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"md-notes/internal/app"
	"md-notes/internal/buffer"
	"md-notes/internal/config"
	"md-notes/internal/theme"
	"md-notes/internal/watcher"
)

// TC10 — Bloqueio de Saída com Alterações Não Salvas (`:q`)
func TestApp_QuitDirtyBuffer_Blocked(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.InsertText("Texto alterado")

	m := app.NewModel(buf, nil)
	_, cmd := m.Update(app.QuitMsg{Force: false})

	// cmd should be nil (not tea.Quit)
	if cmd != nil {
		t.Fatalf("expected nil cmd when quitting dirty buffer without force, got %v", cmd)
	}

	if !strings.Contains(m.StatusMsg, "E37: Alterações não salvas") {
		t.Fatalf("expected E37 warning in status msg, got %q", m.StatusMsg)
	}
}

// TC11 — Forçar Saída com Alterações Pendentes (`:q!`)
func TestApp_QuitForceDirtyBuffer(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	buf.InsertText("Texto alterado")

	m := app.NewModel(buf, nil)
	_, cmd := m.Update(app.QuitMsg{Force: true})

	// When forced, cmd should be tea.Quit
	if cmd == nil {
		t.Fatalf("expected tea.Quit cmd when force quitting dirty buffer, got nil")
	}

	if !m.Quitting {
		t.Fatalf("expected Quitting == true")
	}
}

func TestApp_QuitCleanBuffer(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)

	_, cmd := m.Update(app.QuitMsg{Force: false})
	if cmd == nil {
		t.Fatalf("expected tea.Quit cmd when quitting clean buffer, got nil")
	}
}

func TestApp_ExternalReload_CleanBuffer(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "reload_test.md")
	_ = os.WriteFile(filePath, []byte("Linha original\n"), 0644)

	buf, _ := buffer.LoadFromFile(filePath)
	m := app.NewModel(buf, nil)

	// Modify file on disk
	_ = os.WriteFile(filePath, []byte("Linha modificada externamente\n"), 0644)

	_, _ = m.Update(watcher.FileModifiedMsg{Path: filePath, ModTime: time.Now()})

	line, _ := m.Buffer.GetLine(0)
	if line.String() != "Linha modificada externamente" {
		t.Fatalf("expected buffer reloaded with 'Linha modificada externamente', got %q", line.String())
	}
}

func TestApp_ThemeReloadedMsg(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	m := app.NewModel(buf, nil)

	newCfg := &config.Config{
		Theme: "nord",
		Editor: config.EditorConfig{
			TabSize:     2,
			LineNumbers: false,
		},
	}
	pal, _ := theme.GetPalette("nord")
	newCompiled := theme.CompileTheme(pal, newCfg.Colors)

	_, _ = m.Update(config.ThemeReloadedMsg{
		Config:        newCfg,
		CompiledTheme: newCompiled,
		Warning:       "Aviso de teste",
	})

	if m.Config.Theme != "nord" {
		t.Errorf("Esperado tema 'nord', obtido: '%s'", m.Config.Theme)
	}
	if m.Theme.Palette.Name != "nord" {
		t.Errorf("Esperado paleta 'nord', obtido: '%s'", m.Theme.Palette.Name)
	}
	if m.StatusMsg != "Aviso de teste" {
		t.Errorf("Esperado StatusMsg 'Aviso de teste', obtido: '%s'", m.StatusMsg)
	}
}
