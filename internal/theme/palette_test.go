package theme

import (
	"reflect"
	"testing"
)

// TC03: Integridade das 7 Paletas de Cores Embutidas
func TestPalette_BuiltinThemesExist(t *testing.T) {
	expectedThemes := []string{
		"default-dark",
		"default-light",
		"dracula",
		"nord",
		"catppuccin-mocha",
		"catppuccin-macchiato",
		"monokai",
	}

	for _, themeName := range expectedThemes {
		palette, found := GetPalette(themeName)
		if !found {
			t.Errorf("Tema '%s' não foi encontrado no registro de paletas embutidas", themeName)
		}
		if palette.Name != themeName {
			t.Errorf("Nome da paleta mismatch: esperado '%s', obtido '%s'", themeName, palette.Name)
		}
	}
}

// TC03: Validação de que nenhum token de cor está vazio em nenhum dos 7 temas
func TestPalette_ColorTokensComplete(t *testing.T) {
	themes := BuiltinThemes()
	if len(themes) != 7 {
		t.Fatalf("Esperado 7 temas embutidos, encontrados %d", len(themes))
	}

	for _, themeName := range themes {
		p, found := GetPalette(themeName)
		if !found {
			t.Fatalf("Tema '%s' não encontrado", themeName)
		}

		val := reflect.ValueOf(p)
		typ := reflect.TypeOf(p)

		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldName := typ.Field(i).Name

			if field.Kind() == reflect.String {
				strVal := field.String()
				if strVal == "" {
					t.Errorf("Tema '%s': campo de cor '%s' está vazio", themeName, fieldName)
				}
			}
		}
	}
}

func TestPalette_UnknownFallback(t *testing.T) {
	p, found := GetPalette("tema-inexistente")
	if found {
		t.Error("GetPalette deveria retornar found == false para tema inexistente")
	}
	if p.Name != "default-dark" {
		t.Errorf("Esperado fallback para 'default-dark', obtido '%s'", p.Name)
	}
}
