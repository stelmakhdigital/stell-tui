package editor

import (
	"testing"

	"stell/tui/keys"
)

func TestEditorJumpForward(t *testing.T) {
	ed := NewEditor()
	ed.SetValue("abc.def")
	ed.cursor = 0
	ed.HandleInput("\x1d") // jump forward mode
	ed.HandleInput(".")
	if ed.cursor != 3 {
		t.Fatalf("cursor=%d want 3", ed.cursor)
	}
}

func TestEditorJumpBackward(t *testing.T) {
	ed := NewEditor()
	ed.SetValue("abc.def")
	ed.cursor = len(ed.value)
	ed.HandleInput("\x1b\x1d")
	ed.HandleInput(".")
	if ed.cursor != 3 {
		t.Fatalf("cursor=%d want 3", ed.cursor)
	}
}

func TestKeybindingsManagerMatches(t *testing.T) {
	kb := keys.NewKeybindingsManager(keys.DefaultTUIKeybindings(), nil)
	if !kb.Matches("ctrl+]", "tui.editor.jumpForward") {
		t.Fatal("jumpForward binding")
	}
	if !kb.Matches("ctrl+alt+]", "tui.editor.jumpBackward") {
		t.Fatal("jumpBackward binding")
	}
}
