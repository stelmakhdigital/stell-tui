package diff

import (
	"bytes"
	"strings"
	"testing"
)

func TestDiffEngineFirstAndPatch(t *testing.T) {
	var buf bytes.Buffer
	d := NewDiffEngine(&buf, false)
	d.Resize(40, 10)
	if err := d.Render([]string{"hello", "world"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "\x1b[?2026h") || !strings.Contains(out, "hello") {
		t.Fatalf("first render: %q", out)
	}
	buf.Reset()
	if err := d.Render([]string{"hello", "WORLD"}); err != nil {
		t.Fatal(err)
	}
	out = buf.String()
	if !strings.Contains(out, "WORLD") {
		t.Fatalf("patch: %q", out)
	}
}

func TestDiffScrollStrategy(t *testing.T) {
	var buf strings.Builder
	d := NewDiffEngine(&buf, false)
	d.SetStrategy(DiffScroll)
	_ = d.Render([]string{"a", "b", "c"})
	buf.Reset()
	_ = d.Render([]string{"a", "X", "c"})
	out := buf.String()
	if !strings.Contains(out, "X") {
		t.Fatalf("missing patch: %q", out)
	}
}
