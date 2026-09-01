package app_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"md-notes/internal/app"
	"md-notes/internal/buffer"
)

// TC11
func TestModel_WindowSizeMsg(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	for i := 0; i < 50; i++ {
		lineStr := fmt.Sprintf("Linha de teste %d", i)
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(lineStr, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(lineStr, buffer.EndingLF))
		}
	}

	m := app.NewModel(buf, nil)

	// Send WindowSizeMsg: 120 width, 40 height
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if m.Width != 120 || m.Height != 40 {
		t.Fatalf("expected dimensions 120x40, got %dx%d", m.Width, m.Height)
	}
	if m.Viewport.Width != 120 || m.Viewport.Height != 38 {
		t.Fatalf("expected viewport dimensions 120x38, got %dx%d", m.Viewport.Width, m.Viewport.Height)
	}

	view := m.View()
	lines := strings.Split(view, "\n")
	// 38 viewport lines + 1 status line + 1 command line = 40 lines
	if len(lines) < 40 {
		t.Errorf("expected at least 40 lines in rendered view, got %d", len(lines))
	}
}

// TC13
func TestModel_FullViewComposition(t *testing.T) {
	buf := buffer.NewEmptyBuffer("artigo.md")
	buf.InsertText("# Título Principal\n\nTexto de parágrafo.")

	m := app.NewModel(buf, nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()

	// 1. Should contain Gutter line numbers
	if !strings.Contains(view, "1") {
		t.Errorf("view missing line 1 in gutter")
	}

	// 2. Should contain Markdown text
	if !strings.Contains(view, "# Título Principal") {
		t.Errorf("view missing text content")
	}

	// 3. Should contain Status line
	if !strings.Contains(view, "NORMAL") || !strings.Contains(view, "artigo.md") {
		t.Errorf("view missing status line blocks: %s", view)
	}

	// 4. Should contain ANSI cursor sequence
	if !strings.Contains(view, "\x1b[2 q") {
		t.Errorf("view missing normal mode ANSI block cursor sequence")
	}
}

// TC17
func TestModel_RenderPerformance60FPS(t *testing.T) {
	numLines := 50000
	buf := buffer.NewEmptyBuffer("grande.md")
	for i := 0; i < numLines; i++ {
		lineStr := fmt.Sprintf("Linha de teste número %d com algum conteúdo para renderizar no terminal", i)
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(lineStr, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(lineStr, buffer.EndingLF))
		}
	}

	m := app.NewModel(buf, nil)
	m.Update(tea.WindowSizeMsg{Width: 160, Height: 50})

	// Benchmark 500 frames of View()
	start := time.Now()
	iterations := 500
	for i := 0; i < iterations; i++ {
		_ = m.View()
	}
	totalElapsed := time.Since(start)
	avgPerFrame := totalElapsed / time.Duration(iterations)

	t.Logf("Rendered %d frames on 50k lines in %v (average %v per frame)", iterations, totalElapsed, avgPerFrame)

	// PRD Requirement: 60 FPS corresponds to < 16.6 ms per frame
	if avgPerFrame > 16*time.Millisecond {
		t.Errorf("average frame time exceeded 16ms: %v", avgPerFrame)
	}
}
