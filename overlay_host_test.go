package tui

import "testing"

func TestEngineOverlayStack(t *testing.T) {
	ui := New(nil, false)
	a := NewSelectList([]string{"a"}, nil)
	b := NewSelectList([]string{"b"}, nil)
	ui.ShowOverlay(a)
	ui.ShowOverlay(b)
	if len(ui.overlayStack) != 2 {
		t.Fatalf("stack=%d", len(ui.overlayStack))
	}
	ui.HideOverlay()
	if ui.overlay != a || len(ui.overlayStack) != 1 {
		t.Fatal("hide should restore previous")
	}
	ui.ClearOverlay()
	if ui.overlay != nil || len(ui.overlayStack) != 0 {
		t.Fatal("clear should empty stack")
	}
}

func TestOverlayHandleHideAndHidden(t *testing.T) {
	ui := New(nil, false)
	ui.SetSize(40, 10)
	a := NewSelectList([]string{"a"}, nil)
	h := ui.ShowOverlay(a, OverlayOptions{Anchor: OverlayAnchorCenter, Width: 20})
	if !ui.HasOverlay() {
		t.Fatal("expected overlay")
	}
	h.SetHidden(true)
	if !h.IsHidden() {
		t.Fatal("expected hidden")
	}
	h.SetHidden(false)
	h.Hide()
	if ui.HasOverlay() {
		t.Fatal("expected no overlay after hide")
	}
}

func TestOverlayNonCapturing(t *testing.T) {
	ui := New(nil, false)
	ed := NewEditor()
	ui.SetFocus(ed)
	list := NewSelectList([]string{"x"}, nil)
	ui.ShowOverlay(list, OverlayOptions{NonCapturing: true})
	if ui.focus != ed {
		t.Fatal("nonCapturing should keep prior focus")
	}
}
