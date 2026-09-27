package buffer_test

import (
	"errors"
	"testing"

	"md-notes/internal/buffer"
)

func TestBuffer_ReplaceRangeAndLineEndings(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		start, end buffer.Cursor
		text       string
		want       string
		cursor     buffer.Cursor
	}{
		{"cross lines", buffer.Cursor{Line: 0, Col: 3}, buffer.Cursor{Line: 2, Col: 3}, "uno\r\ndos", "preuno\r\ndossuf", buffer.Cursor{Line: 1, Col: 3}},
		{"insert lines", buffer.Cursor{Line: 0, Col: 3}, buffer.Cursor{Line: 0, Col: 3}, "a\rb", "prea\r\nbABCDE\r\nmiddle\nXYZsuf", buffer.Cursor{Line: 1, Col: 1}},
		{"same line", buffer.Cursor{Line: 0, Col: 3}, buffer.Cursor{Line: 0, Col: 6}, "é", "preéDE\r\nmiddle\nXYZsuf", buffer.Cursor{Line: 0, Col: 4}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := buffer.NewEmptyBuffer("replace.md")
			buf.Lines = []buffer.Line{
				buffer.NewLine("preABCDE", buffer.EndingCRLF),
				buffer.NewLine("middle", buffer.EndingLF),
				buffer.NewLine("XYZsuf", buffer.EndingNone),
			}
			if err := buf.ReplaceRange(tc.start, tc.end, tc.text); err != nil {
				t.Fatal(err)
			}
			if got := buf.RawText(); got != tc.want || buf.Cursor != tc.cursor || !buf.IsDirty {
				t.Fatalf("replace = %q, cursor=%+v, dirty=%t; want %q, %+v", got, buf.Cursor, buf.IsDirty, tc.want, tc.cursor)
			}
		})
	}
	t.Run("invalid range is atomic", func(t *testing.T) {
		t.Parallel()
		buf := buffer.NewEmptyBuffer("replace.md")
		buf.Cursor = buffer.Cursor{Line: 0, Col: 0}
		before := buf.RawText()
		if err := buf.ReplaceRange(buffer.Cursor{Line: 0, Col: 2}, buffer.Cursor{Line: 0, Col: 0}, "x"); !errors.Is(err, buffer.ErrInvalidRange) {
			t.Fatalf("error = %v, want ErrInvalidRange", err)
		}
		if buf.RawText() != before || buf.IsDirty || buf.Cursor != (buffer.Cursor{Line: 0, Col: 0}) {
			t.Fatal("invalid replacement mutated buffer")
		}
	})
}

// TC02 — Inicialização de Buffer Scratchpad Vazio
func TestBuffer_EmptyAndScratchpad(t *testing.T) {
	buf := buffer.NewScratchpadBuffer()

	if buf.FilePath != "" {
		t.Fatalf("expected empty FilePath, got %q", buf.FilePath)
	}
	if buf.IsDirty {
		t.Fatalf("expected IsDirty == false, got true")
	}
	if !buf.IsNewFile {
		t.Fatalf("expected IsNewFile == true, got false")
	}
	if buf.IsStdinBuffer {
		t.Fatalf("expected IsStdinBuffer == false, got true")
	}
	if buf.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", buf.LineCount())
	}
	line, err := buf.GetLine(0)
	if err != nil {
		t.Fatalf("unexpected error getting line 0: %v", err)
	}
	if line.Length() != 0 {
		t.Fatalf("expected empty line, got length %d", line.Length())
	}
}

// TC03 — Rastreamento de Estado Modificado (Dirty Tracking)
func TestBuffer_DirtyTracking(t *testing.T) {
	buf := buffer.NewEmptyBuffer("teste.md")
	if buf.IsDirty {
		t.Fatalf("expected clean buffer, got dirty")
	}

	// Insert text
	buf.InsertText("Olá")
	if !buf.IsDirty {
		t.Fatalf("expected buffer to be dirty after InsertText")
	}

	// Reset dirty
	buf.SetDirty(false)
	if buf.IsDirty {
		t.Fatalf("expected buffer to be clean after SetDirty(false)")
	}

	// Insert newline
	buf.InsertNewLine()
	if !buf.IsDirty {
		t.Fatalf("expected buffer to be dirty after InsertNewLine")
	}
}

func TestBuffer_LineMutationsAndRawText(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	buf.InsertText("Linha 1\nLinha 2")

	if buf.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", buf.LineCount())
	}

	l1, err := buf.GetLine(0)
	if err != nil || l1.String() != "Linha 1" {
		t.Fatalf("expected 'Linha 1', got %q (err: %v)", l1.String(), err)
	}

	l2, err := buf.GetLine(1)
	if err != nil || l2.String() != "Linha 2" {
		t.Fatalf("expected 'Linha 2', got %q (err: %v)", l2.String(), err)
	}

	// Delete line
	deleted, err := buf.DeleteLine(0)
	if err != nil || deleted.String() != "Linha 1" {
		t.Fatalf("expected deleted 'Linha 1', got %q (err: %v)", deleted.String(), err)
	}

	if buf.LineCount() != 1 {
		t.Fatalf("expected 1 line remaining, got %d", buf.LineCount())
	}
}
