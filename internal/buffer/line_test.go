package buffer_test

import (
	"testing"

	"md-notes/internal/buffer"
)

// TC01 — Manipulação de Linha com Caracteres UTF-8 Multibyte
func TestLine_CreationAndRunes(t *testing.T) {
	content := "Olá, Mundo! 🚀\n"
	line := buffer.NewLine(content, buffer.EndingLF)

	// UTF-8 rune count ("Olá, Mundo! 🚀" = 13 runes)
	expectedRunes := 13
	if line.Length() != expectedRunes {
		t.Fatalf("expected length %d runes, got %d", expectedRunes, line.Length())
	}

	// Insert rune at index 4 ('✨')
	line.InsertAt(4, '✨')
	if line.Length() != expectedRunes+1 {
		t.Fatalf("expected length %d after insert, got %d", expectedRunes+1, line.Length())
	}

	// The string should have '✨' at rune index 4
	if string(line.Runes[4]) != "✨" {
		t.Fatalf("expected rune '✨' at index 4, got %c", line.Runes[4])
	}

	// Delete rune at the end (rocket emoji)
	lastIdx := line.Length() - 1
	deleted := line.DeleteAt(lastIdx)
	if deleted != '🚀' {
		t.Fatalf("expected deleted rune '🚀', got %c", deleted)
	}

	if line.Length() != expectedRunes {
		t.Fatalf("expected length %d after delete, got %d", expectedRunes, line.Length())
	}
}

func TestLine_LineEndingsAndBytes(t *testing.T) {
	crlfLine := buffer.NewLine("Texto com CRLF", buffer.EndingCRLF)
	if crlfLine.Ending != buffer.EndingCRLF {
		t.Fatalf("expected CRLF ending, got %s", crlfLine.Ending)
	}

	raw := crlfLine.RawString()
	if raw != "Texto com CRLF\r\n" {
		t.Fatalf("expected 'Texto com CRLF\\r\\n', got %q", raw)
	}

	bytes := crlfLine.Bytes()
	if string(bytes) != "Texto com CRLF\r\n" {
		t.Fatalf("expected byte content 'Texto com CRLF\\r\\n', got %q", string(bytes))
	}
}

func TestLine_SplitAt(t *testing.T) {
	line := buffer.NewLine("PrimeiraSegunda", buffer.EndingLF)
	remainder := line.SplitAt(8)

	if line.String() != "Primeira" {
		t.Fatalf("expected first part 'Primeira', got %q", line.String())
	}
	if remainder.String() != "Segunda" {
		t.Fatalf("expected remainder 'Segunda', got %q", remainder.String())
	}
}
