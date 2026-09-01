package vim

import (
	"strings"
	"sync"

	"github.com/atotto/clipboard"
)

// Clipboard encapsulates system clipboard integration and internal fallback register.
type Clipboard struct {
	mu       sync.RWMutex
	register Register
}

// NewClipboard initializes a new Clipboard instance.
func NewClipboard() *Clipboard {
	return &Clipboard{
		register: Register{
			Content:    "",
			IsLineWise: false,
		},
	}
}

// Set saves the text and linewise flag into the internal register and the OS clipboard.
func (c *Clipboard) Set(text string, isLineWise bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.register = Register{
		Content:    text,
		IsLineWise: isLineWise,
	}

	_ = clipboard.WriteAll(text)
}

// Get retrieves text and linewise status from the OS clipboard or internal register fallback.
func (c *Clipboard) Get() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sysText, err := clipboard.ReadAll()
	if err == nil && sysText != "" {
		isLineWise := c.register.IsLineWise && (sysText == c.register.Content || strings.HasSuffix(sysText, "\n"))
		return sysText, isLineWise
	}

	return c.register.Content, c.register.IsLineWise
}

// GetRegister returns a copy of the current internal register.
func (c *Clipboard) GetRegister() Register {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.register
}
