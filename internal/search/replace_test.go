package search

import (
	"testing"

	"md-notes/internal/buffer"
)

// TC07
func TestReplace_ParseCommands(t *testing.T) {
	// 1. :%s/antigo/novo/g
	cmd1, err := ParseReplaceCommand(":%s/antigo/novo/g")
	if err != nil {
		t.Fatalf("unexpected error parsing cmd1: %v", err)
	}
	if cmd1.Pattern != "antigo" || cmd1.Replacement != "novo" || !cmd1.IsGlobalDoc || !cmd1.IsGlobalLine || cmd1.IsInteractive {
		t.Errorf("cmd1 parsed incorrectly: %+v", cmd1)
	}

	// 2. :s/foo/bar/
	cmd2, err := ParseReplaceCommand(":s/foo/bar/")
	if err != nil {
		t.Fatalf("unexpected error parsing cmd2: %v", err)
	}
	if cmd2.Pattern != "foo" || cmd2.Replacement != "bar" || cmd2.IsGlobalDoc || cmd2.IsGlobalLine {
		t.Errorf("cmd2 parsed incorrectly: %+v", cmd2)
	}

	// 3. :%s/item/element/gc
	cmd3, err := ParseReplaceCommand(":%s/item/element/gc")
	if err != nil {
		t.Fatalf("unexpected error parsing cmd3: %v", err)
	}
	if !cmd3.IsGlobalDoc || !cmd3.IsGlobalLine || !cmd3.IsInteractive {
		t.Errorf("cmd3 parsed incorrectly: %+v", cmd3)
	}

	// 4. Custom delimiter :%s#a/b#c/d#g
	cmd4, err := ParseReplaceCommand(":%s#a/b#c/d#g")
	if err != nil {
		t.Fatalf("unexpected error parsing cmd4: %v", err)
	}
	if cmd4.Pattern != "a/b" || cmd4.Replacement != "c/d" || !cmd4.IsGlobalDoc || !cmd4.IsGlobalLine {
		t.Errorf("cmd4 parsed incorrectly: %+v", cmd4)
	}

	// 5. Invalid command syntax
	_, errInvalid := ParseReplaceCommand("invalid_cmd")
	if errInvalid == nil {
		t.Errorf("expected error on invalid command, got nil")
	}
}

// TC08
func TestReplace_CaptureGroups(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	_ = buf.SetLine(0, buffer.NewLine("item_123, item_456 e item_789", buffer.EndingLF))

	// Replace with group: :s/item_(\d+)/elemento_$1/g
	cmd, err := ParseReplaceCommand(`:s/item_(\d+)/elemento_$1/g`)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	count, err := ExecuteReplace(buf, cmd, 0)
	if err != nil {
		t.Fatalf("unexpected replace error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 replacements, got %d", count)
	}

	line, _ := buf.GetLine(0)
	expected := "elemento_123, elemento_456 e elemento_789"
	if line.String() != expected {
		t.Errorf("line content mismatch: got %q, want %q", line.String(), expected)
	}

	// Test \1 format conversion
	buf2 := buffer.NewEmptyBuffer("doc.md")
	_ = buf2.SetLine(0, buffer.NewLine("foo_bar", buffer.EndingLF))
	cmdBackslash, _ := ParseReplaceCommand(`:s/(\w+)_(\w+)/\2_\1/g`)
	_, _ = ExecuteReplace(buf2, cmdBackslash, 0)
	line2, _ := buf2.GetLine(0)
	if line2.String() != "bar_foo" {
		t.Errorf("backslash capture groups mismatch: got %q, want 'bar_foo'", line2.String())
	}
}

// TC10
func TestReplace_GlobalDocument(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	for i := 0; i < 20; i++ {
		lineStr := "codigo legado com padrão legado"
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(lineStr, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(lineStr, buffer.EndingLF))
		}
	}

	cmd, err := ParseReplaceCommand(":%s/legado/moderno/g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 20 lines * 2 occurrences = 40 occurrences
	count, err := ExecuteReplace(buf, cmd, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 40 {
		t.Errorf("expected 40 replacements, got %d", count)
	}

	raw := buf.RawText()
	if countLegado := len(searchLiteral(raw, "legado")); countLegado != 0 {
		t.Errorf("still found 'legado' in buffer after global replace")
	}
}

func searchLiteral(s, target string) []int {
	p, _ := CompilePattern(target, false)
	intervals := p.FindInLine(s)
	var res []int
	for _, inv := range intervals {
		res = append(res, inv[0])
	}
	return res
}
