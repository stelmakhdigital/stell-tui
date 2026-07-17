package editor

import "testing"

func TestKillRingAccumulate(t *testing.T) {
	k := NewKillRing(8)
	k.Push("a", true, false)
	k.Push("b", true, true)
	if k.Peek() != "ba" {
		t.Fatalf("peek=%q", k.Peek())
	}
}

func TestEditorUndo(t *testing.T) {
	ed := NewEditor()
	ed.SetValue("")
	ed.HandleInput("ab")
	before := ed.Value()
	ed.HandleInput("\x7f")
	ed.Undo()
	if ed.Value() != before {
		t.Fatalf("undo: got %q want %q", ed.Value(), before)
	}
}
