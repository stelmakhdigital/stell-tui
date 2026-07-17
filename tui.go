package tui

import (
	"io"
	"os"
	"sync"
	"time"

	"stell/tui/component"
	"stell/tui/diff"
	"stell/tui/overlay"
	"stell/tui/terminal"
)

// InputListener получает сырой ввод до маршрутизации фокуса.
// Верните true, чтобы поглотить событие (не отдавать компоненту).
type InputListener func(data string) (consumed bool)

// TUI — корневой контейнер с фокусом, оверлеями и циклом отрисовки.
type TUI struct {
	mu           sync.Mutex
	root         *component.Container
	overlay      component.Component
	overlayStack []*overlayEntry
	focus        component.Focusable
	engine       *diff.DiffEngine
	width        int
	height       int
	dirty        bool
	stdin        io.Reader
	stdout       io.Writer
	term         terminal.Terminal
	onResize     func(w, h int)
	showHWCursor bool
	showImages   bool
	cellW        int
	cellH        int

	listeners []InputListener
	started   bool
	stopCh    chan struct{}
	doneCh    chan struct{}
}

// New создаёт hosted-TUI, пишущий в stdout (цикл событий у вызывающего кода).
func New(stdout io.Writer, altScreen bool) *TUI {
	if stdout == nil {
		stdout = os.Stdout
	}
	return &TUI{
		root:   component.NewContainer(),
		engine: diff.NewDiffEngine(stdout, altScreen),
		stdout: stdout,
		stdin:  os.Stdin,
		width:  80,
		height: 24,
		dirty:  true,
	}
}

// NewWithTerminal создаёт TUI, привязанный к Terminal (standalone Start/Stop).
func NewWithTerminal(term terminal.Terminal, altScreen bool) *TUI {
	if term == nil {
		term = terminal.NewProcessTerminal(nil, nil)
	}
	out := io.Writer(os.Stdout)
	if pt, ok := term.(*terminal.ProcessTerminal); ok && pt.Out != nil {
		out = pt.Out
	}
	t := New(out, altScreen)
	t.term = term
	t.width = term.Columns()
	t.height = term.Rows()
	t.engine.Resize(t.width, t.height)
	return t
}

// SetTerminal привязывает Terminal для Start/Stop (hosted-приложениям можно не вызывать).
func (t *TUI) SetTerminal(term terminal.Terminal) {
	t.mu.Lock()
	t.term = term
	if term != nil {
		t.width = term.Columns()
		t.height = term.Rows()
		t.engine.Resize(t.width, t.height)
	}
	t.mu.Unlock()
}

// SetRoot задаёт корневой контейнер.
func (t *TUI) SetRoot(c *component.Container) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = c
	t.dirty = true
}

// Root возвращает корневой контейнер.
func (t *TUI) Root() *component.Container {
	return t.root
}

// AddChild добавляет компонент в корневой контейнер.
func (t *TUI) AddChild(c component.Component) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.root == nil {
		t.root = component.NewContainer()
	}
	t.root.Add(c)
	t.dirty = true
}

// RemoveChild удаляет первое совпадение компонента из корневого контейнера.
func (t *TUI) RemoveChild(c component.Component) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.root == nil {
		return
	}
	children := t.root.Children()
	out := make([]component.Component, 0, len(children))
	for _, ch := range children {
		if ch != c {
			out = append(out, ch)
		}
	}
	t.root.SetChildren(out)
	t.dirty = true
}

// SetFocus устанавливает фокус клавиатуры.
func (t *TUI) SetFocus(f component.Focusable) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.focus != nil {
		t.focus.SetFocused(false)
	}
	t.focus = f
	if f != nil {
		f.SetFocused(true)
	}
	t.dirty = true
}

// RequestRender помечает кадр грязным.
func (t *TUI) RequestRender() {
	t.mu.Lock()
	t.dirty = true
	t.mu.Unlock()
}

// ForceFullRedraw инвалидирует DiffEngine, чтобы следующий кадр был полной перерисовкой.
func (t *TUI) ForceFullRedraw() {
	t.mu.Lock()
	t.engine.Invalidate()
	t.dirty = true
	t.mu.Unlock()
}

// SetSize задаёт размер терминала.
func (t *TUI) SetSize(w, h int) {
	t.mu.Lock()
	t.width, t.height = w, h
	t.engine.Resize(w, h)
	t.dirty = true
	cb := t.onResize
	t.mu.Unlock()
	if cb != nil {
		cb(w, h)
	}
}

// OnResize регистрирует колбэк изменения размера.
func (t *TUI) OnResize(fn func(w, h int)) {
	t.mu.Lock()
	t.onResize = fn
	t.mu.Unlock()
}

// AddInputListener регистрирует обработчик ввода до фокуса.
func (t *TUI) AddInputListener(fn InputListener) {
	if fn == nil {
		return
	}
	t.mu.Lock()
	t.listeners = append(t.listeners, fn)
	t.mu.Unlock()
}

// SetShowHardwareCursor включает позиционирование аппаратного курсора по CursorMarker.
func (t *TUI) SetShowHardwareCursor(on bool) {
	t.mu.Lock()
	t.showHWCursor = on
	t.engine.SetShowHardwareCursor(on)
	t.dirty = true
	t.mu.Unlock()
}

// SetShowImages включает инлайн-изображения терминала.
func (t *TUI) SetShowImages(on bool) {
	t.mu.Lock()
	t.showImages = on
	t.dirty = true
	t.mu.Unlock()
}

// ShowImages сообщает, включены ли инлайн-изображения.
func (t *TUI) ShowImages() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.showImages
}

// SetDiffStrategy выбирает DiffFull / DiffPatch / DiffScroll.
func (t *TUI) SetDiffStrategy(s diff.DiffStrategy) {
	t.mu.Lock()
	t.engine.SetStrategy(s)
	t.dirty = true
	t.mu.Unlock()
}

// SetCellDimensions сохраняет размер ячейки в пикселях (CSI 16 t) для масштаба картинок.
func (t *TUI) SetCellDimensions(w, h int) {
	t.mu.Lock()
	t.cellW, t.cellH = w, h
	t.mu.Unlock()
}

// CellDimensions возвращает последний известный размер ячейки в пикселях (0 если неизвестен).
func (t *TUI) CellDimensions() (w, h int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cellW, t.cellH
}

// HandleInput маршрутизирует ввод слушателям, оверлею или фокусу.
func (t *TUI) HandleInput(data string) {
	t.mu.Lock()
	listeners := append([]InputListener(nil), t.listeners...)
	var ov component.Component
	if n := len(t.overlayStack); n > 0 {
		top := t.overlayStack[n-1]
		if top != nil && !top.hidden && !top.opts.NonCapturing {
			ov = top.comp
		}
	} else {
		ov = t.overlay
	}
	focus := t.focus
	t.mu.Unlock()

	for _, fn := range listeners {
		if fn(data) {
			t.RequestRender()
			return
		}
	}

	if ov != nil {
		if h, ok := ov.(component.InputHandler); ok {
			h.HandleInput(data)
			t.RequestRender()
			return
		}
	}
	if focus != nil {
		focus.HandleInput(data)
		t.RequestRender()
	}
}

// RenderNow принудительно пишет кадр (для хостов со своим циклом событий).
func (t *TUI) RenderNow() error {
	t.mu.Lock()
	t.dirty = true
	t.mu.Unlock()
	return t.renderFrame()
}

// Start запускает standalone-цикл на привязанном Terminal.
// Блокируется до Stop. Для standalone предпочтителен NewWithTerminal.
func (t *TUI) Start() {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return
	}
	term := t.term
	if term == nil {
		term = terminal.NewProcessTerminal(nil, nil)
		t.term = term
	}
	t.started = true
	t.stopCh = make(chan struct{})
	t.doneCh = make(chan struct{})
	stopCh := t.stopCh
	doneCh := t.doneCh
	t.mu.Unlock()

	term.Start(func(data string) {
		t.HandleInput(data)
		_ = t.RenderNow()
	}, func() {
		t.SetSize(term.Columns(), term.Rows())
		_ = t.RenderNow()
	})
	t.SetSize(term.Columns(), term.Rows())
	_ = t.RenderNow()

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	defer close(doneCh)

	for {
		select {
		case <-stopCh:
			term.Stop()
			_ = t.Close()
			t.mu.Lock()
			t.started = false
			t.mu.Unlock()
			return
		case <-ticker.C:
			t.mu.Lock()
			dirty := t.dirty
			t.mu.Unlock()
			if dirty {
				_ = t.renderFrame()
			}
		}
	}
}

// Stop завершает цикл Start и восстанавливает терминал.
func (t *TUI) Stop() {
	t.mu.Lock()
	if !t.started || t.stopCh == nil {
		t.mu.Unlock()
		return
	}
	ch := t.stopCh
	done := t.doneCh
	t.mu.Unlock()
	select {
	case <-ch:
	default:
		close(ch)
	}
	if done != nil {
		<-done
	}
}

// Close восстанавливает терминал (курсор / alt screen).
func (t *TUI) Close() error {
	return t.engine.Close()
}

func (t *TUI) renderFrame() error {
	t.mu.Lock()
	w, h := t.width, t.height
	root := t.root
	ov := t.overlay
	stack := append([]*overlayEntry(nil), t.overlayStack...)
	t.dirty = false
	t.mu.Unlock()

	var lines []string
	if root != nil {
		lines = root.Render(w)
	}

	if len(stack) > 0 {
		for _, e := range stack {
			if e == nil || e.hidden || e.comp == nil {
				continue
			}
			ol := e.comp.Render(overlay.RenderWidth(w, e.opts))
			lines = overlay.CompositeFrame(lines, ol, e.opts, w, h)
		}
	} else if ov != nil {
		ol := ov.Render(w)
		if len(ol) > 0 {
			if len(ol) >= h {
				lines = ol[len(ol)-h:]
			} else if len(lines)+len(ol) > h {
				keep := h - len(ol)
				if keep < 0 {
					keep = 0
				}
				lines = append(lines[len(lines)-keep:], ol...)
			} else {
				lines = append(lines, ol...)
			}
		}
	}
	t.engine.Resize(w, h)
	return t.engine.Render(lines)
}
