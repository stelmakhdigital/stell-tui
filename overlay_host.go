package tui

import (
	"github.com/stelmakhdigital/stell-tui/component"
	"github.com/stelmakhdigital/stell-tui/overlay"
)

// overlayEntry — один уровень стека оверлеев TUI.
type overlayEntry struct {
	comp     component.Component
	preFocus component.Focusable
	opts     overlay.OverlayOptions
	hidden   bool
	handle   *OverlayHandle
}

// OverlayHandle управляет показанным оверлеем.
type OverlayHandle struct {
	tui   *TUI
	entry *overlayEntry
}

// UnfocusOptions настраивает OverlayHandle.Unfocus.
type UnfocusOptions struct {
	Target component.Focusable
	Clear  bool
}

// Hide навсегда убирает оверлей из стека.
func (h *OverlayHandle) Hide() {
	if h == nil || h.tui == nil || h.entry == nil {
		return
	}
	h.tui.removeOverlayEntry(h.entry)
}

// SetHidden временно скрывает или показывает оверлей, не удаляя его.
func (h *OverlayHandle) SetHidden(hidden bool) {
	if h == nil || h.entry == nil || h.tui == nil {
		return
	}
	h.tui.mu.Lock()
	h.entry.hidden = hidden
	h.tui.dirty = true
	h.tui.mu.Unlock()
}

// IsHidden сообщает, временно ли скрыт оверлей.
func (h *OverlayHandle) IsHidden() bool {
	if h == nil || h.entry == nil || h.tui == nil {
		return true
	}
	h.tui.mu.Lock()
	defer h.tui.mu.Unlock()
	return h.entry.hidden
}

// Focus фокусирует оверлей и поднимает его визуально наверх.
func (h *OverlayHandle) Focus() {
	if h == nil || h.tui == nil || h.entry == nil {
		return
	}
	h.tui.focusOverlayEntry(h.entry)
}

// Unfocus снимает фокус с оверлея.
func (h *OverlayHandle) Unfocus(opts ...UnfocusOptions) {
	if h == nil || h.tui == nil || h.entry == nil {
		return
	}
	var o UnfocusOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	h.tui.unfocusOverlayEntry(h.entry, o)
}

// IsFocused сообщает, есть ли сейчас фокус у этого оверлея.
func (h *OverlayHandle) IsFocused() bool {
	if h == nil || h.tui == nil || h.entry == nil {
		return false
	}
	h.tui.mu.Lock()
	defer h.tui.mu.Unlock()
	if f, ok := h.entry.comp.(component.Focusable); ok {
		return h.tui.focus == f
	}
	return false
}

// ShowOverlay показывает компонент как оверлей.
func (t *TUI) ShowOverlay(c component.Component, opts ...overlay.OverlayOptions) *OverlayHandle {
	var o overlay.OverlayOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e := &overlayEntry{comp: c, preFocus: t.focus, opts: o}
	h := &OverlayHandle{tui: t, entry: e}
	e.handle = h
	t.overlayStack = append(t.overlayStack, e)
	t.overlay = c
	if !o.NonCapturing {
		if f, ok := c.(component.Focusable); ok {
			if t.focus != nil {
				t.focus.SetFocused(false)
			}
			t.focus = f
			f.SetFocused(true)
		}
	}
	t.dirty = true
	return h
}

// HideOverlay скрывает верхний оверлей.
func (t *TUI) HideOverlay() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.overlayStack) == 0 {
		t.overlay = nil
		t.dirty = true
		return
	}
	top := t.overlayStack[len(t.overlayStack)-1]
	t.overlayStack = t.overlayStack[:len(t.overlayStack)-1]
	if t.focus != nil {
		t.focus.SetFocused(false)
	}
	t.restoreFocusAfterHide(top)
	t.dirty = true
}

func (t *TUI) restoreFocusAfterHide(top *overlayEntry) {
	if len(t.overlayStack) > 0 {
		next := t.overlayStack[len(t.overlayStack)-1]
		t.overlay = next.comp
		if !next.opts.NonCapturing && !next.hidden {
			if f, ok := next.comp.(component.Focusable); ok {
				t.focus = f
				f.SetFocused(true)
				return
			}
		}
		if top != nil && top.preFocus != nil {
			t.focus = top.preFocus
			t.focus.SetFocused(true)
		} else {
			t.focus = nil
		}
		return
	}
	t.overlay = nil
	if top != nil && top.preFocus != nil {
		t.focus = top.preFocus
		t.focus.SetFocused(true)
	} else {
		t.focus = nil
	}
}

// SetOverlay заменяет верхний оверлей (или показывает новый).
func (t *TUI) SetOverlay(c component.Component) {
	if c == nil {
		t.HideOverlay()
		return
	}
	t.ShowOverlay(c)
}

// ClearOverlay удаляет все оверлеи.
func (t *TUI) ClearOverlay() {
	t.mu.Lock()
	t.overlayStack = nil
	t.overlay = nil
	t.dirty = true
	t.mu.Unlock()
}

// HasOverlay сообщает, есть ли видимый активный оверлей.
func (t *TUI) HasOverlay() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, e := range t.overlayStack {
		if e != nil && !e.hidden && e.comp != nil {
			return true
		}
	}
	return t.overlay != nil
}

func (t *TUI) removeOverlayEntry(target *overlayEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	idx := -1
	for i, e := range t.overlayStack {
		if e == target {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	top := t.overlayStack[idx]
	t.overlayStack = append(t.overlayStack[:idx], t.overlayStack[idx+1:]...)
	if t.focus != nil {
		if f, ok := top.comp.(component.Focusable); ok && t.focus == f {
			t.focus.SetFocused(false)
			t.restoreFocusAfterHide(top)
		}
	}
	if len(t.overlayStack) == 0 {
		t.overlay = nil
	} else {
		t.overlay = t.overlayStack[len(t.overlayStack)-1].comp
	}
	t.dirty = true
}

func (t *TUI) focusOverlayEntry(target *overlayEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	idx := -1
	for i, e := range t.overlayStack {
		if e == target {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	e := t.overlayStack[idx]
	e.hidden = false
	t.overlayStack = append(append(t.overlayStack[:idx], t.overlayStack[idx+1:]...), e)
	t.overlay = e.comp
	if f, ok := e.comp.(component.Focusable); ok {
		if t.focus != nil {
			t.focus.SetFocused(false)
		}
		t.focus = f
		f.SetFocused(true)
	}
	t.dirty = true
}

func (t *TUI) unfocusOverlayEntry(target *overlayEntry, opts UnfocusOptions) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if opts.Clear {
		if t.focus != nil {
			t.focus.SetFocused(false)
		}
		t.focus = nil
		t.dirty = true
		return
	}
	if opts.Target != nil {
		if t.focus != nil {
			t.focus.SetFocused(false)
		}
		t.focus = opts.Target
		opts.Target.SetFocused(true)
		t.dirty = true
		return
	}
	if t.focus != nil {
		t.focus.SetFocused(false)
	}
	if target != nil && target.preFocus != nil {
		t.focus = target.preFocus
		t.focus.SetFocused(true)
	} else {
		t.focus = nil
	}
	t.dirty = true
}
