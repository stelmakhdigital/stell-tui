package diff

import (
	"strings"
	"testing"

	"stell/tui/wrap"
)

func TestExtractCursorMarkerFullSequence(t *testing.T) {
	lines := []string{
		"hello",
		"ab" + wrap.CursorMarker + "cd",
		"tail",
	}
	row, col, cleaned, found := extractCursorMarker(lines)
	if !found {
		t.Fatal("expected marker")
	}
	if row != 1 {
		t.Fatalf("row=%d want 1", row)
	}
	if col != 2 {
		t.Fatalf("col=%d want 2", col)
	}
	if cleaned[1] != "abcd" {
		t.Fatalf("cleaned=%q want abcd", cleaned[1])
	}
	if strings.Contains(cleaned[1], wrap.CursorMarker) {
		t.Fatal("marker should be removed")
	}
}

func TestExtractCursorMarkerWithANSIPrefix(t *testing.T) {
	prefix := "\x1b[31mab\x1b[0m"
	line := prefix + wrap.CursorMarker + "x"
	row, col, cleaned, found := extractCursorMarker([]string{line})
	if !found || row != 0 {
		t.Fatalf("found=%v row=%d", found, row)
	}
	if col != 2 {
		t.Fatalf("col=%d want 2 (visible 'ab')", col)
	}
	if cleaned[0] != prefix+"x" {
		t.Fatalf("cleaned=%q", cleaned[0])
	}
}
