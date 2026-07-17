package overlay

import "testing"

func TestClampOverlayHeightPct(t *testing.T) {
	lines := []string{"1", "2", "3", "4", "5", "6"}
	got := ClampOverlayLines(lines, 50, 10)
	if len(got) != 5 {
		t.Fatalf("len=%d want 5", len(got))
	}
}

func TestCompositeOverlayCenter(t *testing.T) {
	chat := []string{"a", "b"}
	ol := []string{"X", "Y"}
	out := CompositeOverlayLines(chat, ol, OverlayOptions{Anchor: OverlayAnchorCenter, MaxHeightPct: 80}, 6)
	if len(out) != 6 {
		t.Fatalf("len=%d", len(out))
	}
	if out[2] != "X" || out[3] != "Y" {
		t.Fatalf("center placement: %v", out)
	}
}

func TestResolveDimPercent(t *testing.T) {
	w, ok := resolveDim("50%", 100)
	if !ok || w != 50 {
		t.Fatalf("got %d ok=%v", w, ok)
	}
	w, ok = resolveDim(40, 100)
	if !ok || w != 40 {
		t.Fatalf("got %d ok=%v", w, ok)
	}
}

func TestCompositeOverlayFrameBottomRight(t *testing.T) {
	base := []string{"........", "........", "........", "........"}
	ol := []string{"XY"}
	out := CompositeFrame(base, ol, OverlayOptions{
		Anchor: OverlayAnchorBottomRight,
		Width:  2,
	}, 8, 4)
	if len(out) != 4 {
		t.Fatalf("len=%d", len(out))
	}
	last := out[3]
	if !stringsContains(last, "XY") {
		t.Fatalf("expected XY in bottom row: %q", last)
	}
}

func stringsContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
