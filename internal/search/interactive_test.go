package search

import (
	"strings"
	"testing"

	"md-notes/internal/buffer"
)

// TC12, TC13
func TestInteractive_ConfirmOptions(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	// 4 occurrences of "alvo":
	// Line 0: "alvo 1"
	// Line 1: "alvo 2"
	// Line 2: "alvo 3"
	// Line 3: "alvo 4"
	lines := []string{"alvo 1", "alvo 2", "alvo 3", "alvo 4"}
	for i, l := range lines {
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(l, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(l, buffer.EndingLF))
		}
	}

	cmd, _ := ParseReplaceCommand(":%s/alvo/substituto/gc")
	session := NewInteractiveSession(buf, cmd, 0)

	if len(session.Matches) != 4 {
		t.Fatalf("expected 4 matches in session, got %d", len(session.Matches))
	}

	// 1. Action 'y' on match 0 -> replaces match 0
	done, next := session.HandleAction(buf, 'y')
	if done || next == nil {
		t.Fatalf("expected session to continue with match 1, got done=%v", done)
	}
	l0, _ := buf.GetLine(0)
	if !strings.HasPrefix(l0.String(), "substituto 1") {
		t.Errorf("line 0 not replaced: %s", l0.String())
	}

	// 2. Action 'n' on match 1 -> skips match 1
	done, next = session.HandleAction(buf, 'n')
	if done || next == nil {
		t.Fatalf("expected session to continue with match 2, got done=%v", done)
	}
	l1, _ := buf.GetLine(1)
	if !strings.HasPrefix(l1.String(), "alvo 2") {
		t.Errorf("line 1 should remain 'alvo 2', got %s", l1.String())
	}

	// 3. Action 'a' on match 2 -> replaces match 2 and remaining match 3, then finishes
	done, next = session.HandleAction(buf, 'a')
	if !done || next != nil {
		t.Errorf("expected session to be done after 'a', got done=%v", done)
	}

	l2, _ := buf.GetLine(2)
	l3, _ := buf.GetLine(3)
	if !strings.HasPrefix(l2.String(), "substituto 3") {
		t.Errorf("line 2 not replaced by 'a': %s", l2.String())
	}
	if !strings.HasPrefix(l3.String(), "substituto 4") {
		t.Errorf("line 3 not replaced by 'a': %s", l3.String())
	}

	if session.ReplacedCount != 3 {
		t.Errorf("expected ReplacedCount 3 (matches 0, 2, 3), got %d", session.ReplacedCount)
	}
}

// TC13
func TestInteractive_CancelQuit(t *testing.T) {
	buf := buffer.NewEmptyBuffer("doc.md")
	lines := []string{"teste A", "teste B", "teste C"}
	for i, l := range lines {
		if i == 0 {
			_ = buf.SetLine(0, buffer.NewLine(l, buffer.EndingLF))
		} else {
			buf.InsertLine(i, buffer.NewLine(l, buffer.EndingLF))
		}
	}

	cmd, _ := ParseReplaceCommand(":%s/teste/novo/gc")
	session := NewInteractiveSession(buf, cmd, 0)

	// 'y' on match 0
	session.HandleAction(buf, 'y')

	// 'q' on match 1 -> cancel immediately
	done, next := session.HandleAction(buf, 'q')
	if !done || next != nil {
		t.Errorf("expected session done on 'q', got done=%v, next=%v", done, next)
	}

	l0, _ := buf.GetLine(0)
	l1, _ := buf.GetLine(1)
	l2, _ := buf.GetLine(2)
	if l0.String() != "novo A" {
		t.Errorf("line 0 mismatch: %s", l0.String())
	}
	if l1.String() != "teste B" {
		t.Errorf("line 1 should be unchanged, got %s", l1.String())
	}
	if l2.String() != "teste C" {
		t.Errorf("line 2 should be unchanged, got %s", l2.String())
	}
}
