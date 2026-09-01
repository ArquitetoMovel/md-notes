package theme

import (
	"os"
	"strconv"
	"strings"
)

// IsDarkBackground inspects environment variables (such as COLORFGBG) to determine
// if the current terminal has a dark background.
func IsDarkBackground() bool {
	colorfgbg := os.Getenv("COLORFGBG")
	if colorfgbg == "" {
		return true // Default assumption for modern developer terminals
	}

	parts := strings.Split(colorfgbg, ";")
	if len(parts) < 2 {
		return true
	}

	bgStr := strings.TrimSpace(parts[len(parts)-1])
	bgInt, err := strconv.Atoi(bgStr)
	if err != nil {
		return true
	}

	// In standard ANSI terminals:
	// 0-6: Dark colors (Black, Red, Green, Yellow, Blue, Magenta, Cyan)
	// 8: Dark gray / bright black
	// 7: Light gray (Light mode)
	// 15: Bright white (Light mode)
	if bgInt == 7 || bgInt == 15 || (bgInt > 8 && bgInt <= 15) {
		return false
	}

	return true
}

// ResolveThemeName resolves a configured theme name.
// If the theme is "auto", it dynamically detects whether to use "default-dark" or "default-light".
func ResolveThemeName(configuredTheme string) string {
	norm := strings.ToLower(strings.TrimSpace(configuredTheme))
	if norm == "auto" {
		if IsDarkBackground() {
			return "default-dark"
		}
		return "default-light"
	}
	if norm == "" {
		return "default-dark"
	}
	return norm
}
