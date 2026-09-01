package theme

import (
	"os"
	"testing"
)

// TC05: Detecção de Modo Claro/Escuro do Terminal (COLORFGBG)
func TestDetect_COLORFGBG_Dark(t *testing.T) {
	orig := os.Getenv("COLORFGBG")
	defer os.Setenv("COLORFGBG", orig)

	os.Setenv("COLORFGBG", "15;0")
	resolved := ResolveThemeName("auto")
	if resolved != "default-dark" {
		t.Errorf("Esperado 'default-dark' para COLORFGBG='15;0', obtido: '%s'", resolved)
	}

	if !IsDarkBackground() {
		t.Error("Esperado IsDarkBackground() == true para COLORFGBG='15;0'")
	}
}

// TC05: Detecção de Modo Claro/Escuro do Terminal (COLORFGBG = 0;15)
func TestDetect_COLORFGBG_Light(t *testing.T) {
	orig := os.Getenv("COLORFGBG")
	defer os.Setenv("COLORFGBG", orig)

	os.Setenv("COLORFGBG", "0;15")
	resolved := ResolveThemeName("auto")
	if resolved != "default-light" {
		t.Errorf("Esperado 'default-light' para COLORFGBG='0;15', obtido: '%s'", resolved)
	}

	if IsDarkBackground() {
		t.Error("Esperado IsDarkBackground() == false para COLORFGBG='0;15'")
	}

	os.Setenv("COLORFGBG", "0;7")
	resolved7 := ResolveThemeName("auto")
	if resolved7 != "default-light" {
		t.Errorf("Esperado 'default-light' para COLORFGBG='0;7', obtido: '%s'", resolved7)
	}
}

func TestDetect_ExplicitThemeNotOverridden(t *testing.T) {
	orig := os.Getenv("COLORFGBG")
	defer os.Setenv("COLORFGBG", orig)

	os.Setenv("COLORFGBG", "0;15")
	resolved := ResolveThemeName("dracula")
	if resolved != "dracula" {
		t.Errorf("Esperado 'dracula' mantido, obtido: '%s'", resolved)
	}
}
