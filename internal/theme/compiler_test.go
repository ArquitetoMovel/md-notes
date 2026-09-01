package theme

import (
	"strings"
	"testing"
	"time"

	"md-notes/internal/config"
)

// TC04 & TC10: Compilação e Sobreposição de Cores Granulares (Hex / ANSI)
func TestCompiler_ApplyColorOverrides(t *testing.T) {
	palette, found := GetPalette("dracula")
	if !found {
		t.Fatal("Tema dracula não encontrado")
	}

	overrides := config.ColorOverrides{
		H1:          "#FF0000",
		TableBorder: "#00FF00",
		CodeBg:      "#112233",
	}

	compiled := CompileTheme(palette, overrides)
	if compiled == nil {
		t.Fatal("CompileTheme retornou nil")
	}

	if compiled.Palette.H1 != "#FF0000" {
		t.Errorf("Esperado Palette.H1 == '#FF0000', obtido: '%s'", compiled.Palette.H1)
	}
	if compiled.Palette.TableBorder != "#00FF00" {
		t.Errorf("Esperado Palette.TableBorder == '#00FF00', obtido: '%s'", compiled.Palette.TableBorder)
	}
	if compiled.Palette.CodeBg != "#112233" {
		t.Errorf("Esperado Palette.CodeBg == '#112233', obtido: '%s'", compiled.Palette.CodeBg)
	}
	// Non-overridden colors must remain Dracula defaults
	if compiled.Palette.H2 != palette.H2 {
		t.Errorf("Esperado Palette.H2 mantido '%s', obtido: '%s'", palette.H2, compiled.Palette.H2)
	}
}

// TC09 & TC12: Validação de Contrato de Estilos e Renderização com Escape ANSI
func TestCompiler_CompileTheme_Success(t *testing.T) {
	palette, _ := GetPalette("monokai")
	compiled := CompileTheme(palette, config.ColorOverrides{})

	h1Rendered := compiled.H1.Render("# Título 1")
	if !strings.Contains(h1Rendered, "Título 1") {
		t.Errorf("Renderização de H1 incorreta: %s", h1Rendered)
	}

	codeRendered := compiled.CodeBlock.Render("fmt.Println()")
	if !strings.Contains(codeRendered, "fmt.Println()") {
		t.Errorf("Renderização de CodeBlock incorreta: %s", codeRendered)
	}

	statusRendered := compiled.StatusBar.Render("NORMAL")
	if !strings.Contains(statusRendered, "NORMAL") {
		t.Errorf("Renderização de StatusBar incorreta: %s", statusRendered)
	}
}

// TC06: Performance de Compilação de Estilos (< 5 ms)
func TestCompiler_Performance(t *testing.T) {
	palette, _ := GetPalette("catppuccin-mocha")
	overrides := config.ColorOverrides{
		H1:     "#FF1122",
		CodeBg: "#001122",
	}

	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = CompileTheme(palette, overrides)
	}
	totalDuration := time.Since(start)

	avgDuration := totalDuration / time.Duration(iterations)
	if avgDuration > 5*time.Millisecond {
		t.Errorf("Tempo médio de compilação excedeu 5ms: %v (total: %v)", avgDuration, totalDuration)
	}
}
