package wrap

import "testing"

func TestVisibleLenIgnoresCursorMarker(t *testing.T) {
	base := "ab"
	with := "a" + CursorMarker + "b"
	if VisibleLen(with) != VisibleLen(base) {
		t.Fatalf("VisibleLen with marker=%d want %d", VisibleLen(with), VisibleLen(base))
	}
	if StripANSI(with) != base {
		t.Fatalf("StripANSI=%q want %q", StripANSI(with), base)
	}
}

func TestVisibleLenIgnoresKittyAPC(t *testing.T) {
	kitty := "\x1b_Ga=T,f=100;data\x07"
	s := "x" + kitty + "y"
	if VisibleLen(s) != 2 {
		t.Fatalf("VisibleLen=%d want 2", VisibleLen(s))
	}
	if StripANSI(s) != "xy" {
		t.Fatalf("StripANSI=%q want xy", StripANSI(s))
	}
}

func TestSkipANSIAPCWithST(t *testing.T) {
	s := "a\x1b_stell:c\x1b\\b"
	skip, next := SkipANSI(s, 1)
	if !skip || next != len(s)-1 {
		t.Fatalf("skip=%v next=%d want %d", skip, next, len(s)-1)
	}
	if VisibleLen(s) != 2 {
		t.Fatalf("VisibleLen=%d want 2", VisibleLen(s))
	}
}
