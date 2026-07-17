package editor

import (
	"strings"
	"testing"
)

func TestEditorPasteMarker(t *testing.T) {
	e := NewEditor()
	e.pasteThresh = 3
	e.HandleInput("a\nb\nc\nd\n")
	if !strings.Contains(e.Value(), "[paste #") {
		t.Fatalf("expected paste marker, got %q", e.Value())
	}
}
