package keys

import "testing"

func TestStdinBufferSplitEscapeSequence(t *testing.T) {
	var out []string
	buf := NewStdinBuffer(StdinBufferOptions{})
	buf.OnData = func(s string) { out = append(out, s) }
	buf.Process("\x1b")
	buf.Process("[27;1;27~")
	if len(out) != 1 || out[0] != "\x1b[27;1;27~" {
		t.Fatalf("got %v want single modifyOtherKeys esc", out)
	}
}

func TestStdinBufferPlainText(t *testing.T) {
	var out []string
	buf := NewStdinBuffer(StdinBufferOptions{})
	buf.OnData = func(s string) { out = append(out, s) }
	buf.Process("hello")
	if len(out) != 5 {
		t.Fatalf("got %d sequences want 5 chars", len(out))
	}
}

func TestStdinBufferBracketedPaste(t *testing.T) {
	var pasted string
	buf := NewStdinBuffer(StdinBufferOptions{})
	buf.OnPaste = func(s string) { pasted = s }
	buf.Process("\x1b[200~paste\x1b[201~")
	if pasted != "paste" {
		t.Fatalf("got %q", pasted)
	}
}
