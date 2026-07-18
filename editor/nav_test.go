package editor

import (
	"strings"
	"testing"

	"github.com/stelmakhdigital/stell-tui/wrap"
)

func TestEditorCtrlAELineBounds(t *testing.T) {
	ed := NewEditor()
	ed.SetValue("hello\nworld")
	ed.cursor = len("hello\nwo")
	ed.HandleInput("\x01")
	if ed.cursor != len("hello\n") {
		t.Fatalf("line start cursor=%d", ed.cursor)
	}
	ed.HandleInput("\x05")
	if ed.cursor != len("hello\nworld") {
		t.Fatalf("line end cursor=%d", ed.cursor)
	}
}

func TestEditorHardwareCursorMarker(t *testing.T) {
	ed := NewEditor()
	ed.SetShowHardwareCursor(true)
	ed.SetFocused(true)
	ed.SetValue("ab")
	ed.cursor = 1
	lines := ed.Render(40)
	joined := strings.Join(lines, "")
	if !strings.Contains(joined, wrap.CursorMarker) {
		t.Fatalf("expected wrap.CursorMarker in render: %q", joined)
	}
	if strings.ContainsRune(joined, '█') {
		t.Fatalf("block cursor should be off: %q", joined)
	}
}
