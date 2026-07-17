package terminal

import (
	"bytes"
	"strings"
	"testing"
)

func TestProcessTerminalHelpers(t *testing.T) {
	var buf bytes.Buffer
	pt := NewProcessTerminal(strings.NewReader(""), &buf)
	pt.HideCursor()
	pt.ShowCursor()
	pt.ClearLine()
	pt.SetTitle("test")
	if buf.Len() == 0 {
		t.Fatal("expected writes")
	}
	if pt.Columns() < 1 || pt.Rows() < 1 {
		t.Fatal("expected positive size")
	}
}

func TestEnableTerminalFeaturesWriter(t *testing.T) {
	var buf bytes.Buffer
	restore := EnableTerminalFeaturesWriter(&buf)
	if !strings.Contains(buf.String(), seqBracketedPasteOn) {
		t.Fatalf("missing paste on: %q", buf.String())
	}
	restore()
	if !strings.Contains(buf.String(), seqBracketedPasteOff) {
		t.Fatalf("missing paste off: %q", buf.String())
	}
}

func TestEncodeKittyImage(t *testing.T) {
	seq := EncodeTerminalImage(ImageKitty, "image/png", []byte{1, 2, 3, 4}, ImageRenderOptions{ImageID: 7})
	if !strings.Contains(seq, "\x1b_G") || !strings.Contains(seq, "i=7") {
		t.Fatalf("%q", seq)
	}
}
