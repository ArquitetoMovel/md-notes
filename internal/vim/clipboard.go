package vim

import (
	"strings"
	"sync"

	"github.com/atotto/clipboard"
)

// ClipboardProvider supplies plain text to and from the system clipboard.
type ClipboardProvider interface {
	ReadAll() (string, error)
	WriteAll(string) error
}

type systemClipboard struct{}

func (systemClipboard) ReadAll() (string, error)   { return clipboard.ReadAll() }
func (systemClipboard) WriteAll(text string) error { return clipboard.WriteAll(text) }

// Clipboard encapsulates system clipboard integration and internal fallback register.
type Clipboard struct {
	mu       sync.RWMutex
	register Register
	provider ClipboardProvider
}

// NewClipboard initializes a new Clipboard instance.
func NewClipboard() *Clipboard {
	return NewClipboardWithProvider(systemClipboard{})
}

// NewClipboardWithProvider creates a clipboard with an injectable provider.
func NewClipboardWithProvider(provider ClipboardProvider) *Clipboard {
	if provider == nil {
		provider = systemClipboard{}
	}
	return &Clipboard{
		register: Register{
			Content:    "",
			IsLineWise: false,
		},
		provider: provider,
	}
}

// WriteStrict writes text and updates the register only after the provider succeeds.
func (c *Clipboard) WriteStrict(text string, isLineWise bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.provider.WriteAll(text); err != nil {
		return err
	}
	c.register = Register{Content: text, IsLineWise: isLineWise}
	return nil
}

// ReadStrict reads text without falling back to the internal Vim register.
func (c *Clipboard) ReadStrict() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.provider.ReadAll()
}

// Set saves the text and linewise flag into the internal register and the OS clipboard.
func (c *Clipboard) Set(text string, isLineWise bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.register = Register{
		Content:    text,
		IsLineWise: isLineWise,
	}

	// Legacy Vim commands keep the register even when the system clipboard fails.
	_ = c.provider.WriteAll(text)
}

// Get retrieves text and linewise status from the OS clipboard or internal register fallback.
func (c *Clipboard) Get() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sysText, err := c.provider.ReadAll()
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
