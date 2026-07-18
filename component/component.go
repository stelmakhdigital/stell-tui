package component

import (
	"github.com/stelmakhdigital/stell-tui/wrap"
)

// Component — retained-mode элемент UI. Render возвращает строки ширины width
// (ANSI-стили допускаются). HandleInput опционален через InputHandler.
type Component interface {
	Render(width int) []string
}

// InputHandler принимает сырые данные клавиатуры / paste.
type InputHandler interface {
	HandleInput(data string)
}

// Focusable может владеть фокусом клавиатуры и отдавать маркер курсора для IME.
type Focusable interface {
	Component
	InputHandler
	Focused() bool
	SetFocused(bool)
}

// Invalidatable сбрасывает кэш отрисованных строк.
type Invalidatable interface {
	Invalidate()
}

// Container складывает дочерние компоненты вертикально.
type Container struct {
	children []Component
}

// NewContainer создаёт контейнер с начальными детьми.
func NewContainer(children ...Component) *Container {
	return &Container{children: append([]Component{}, children...)}
}

// Add добавляет дочерний компонент.
func (c *Container) Add(child Component) {
	c.children = append(c.children, child)
}

// SetChildren заменяет список детей.
func (c *Container) SetChildren(children []Component) {
	c.children = append([]Component{}, children...)
}

// Children возвращает срез дочерних компонентов.
func (c *Container) Children() []Component {
	return c.children
}

// Render склеивает вывод всех детей.
func (c *Container) Render(width int) []string {
	var out []string
	for _, ch := range c.children {
		if ch == nil {
			continue
		}
		out = append(out, ch.Render(width)...)
	}
	return out
}

// Invalidate сбрасывает кэш у детей, поддерживающих Invalidatable.
func (c *Container) Invalidate() {
	for _, ch := range c.children {
		if inv, ok := ch.(Invalidatable); ok {
			inv.Invalidate()
		}
	}
}

// Text — статический многострочный текст.
type Text struct {
	Lines []string
}

// Render переносит строки под ширину.
func (t *Text) Render(width int) []string {
	if t == nil {
		return nil
	}
	return wrap.WrapLines(t.Lines, width)
}

// Spacer добавляет пустые строки.
type Spacer struct {
	N int
}

// Render возвращает N пустых строк (минимум 1).
func (s *Spacer) Render(width int) []string {
	n := s.N
	if n < 1 {
		n = 1
	}
	out := make([]string, n)
	return out
}

// TruncatedText показывает не больше MaxLines строк, с «…» при обрезке.
type TruncatedText struct {
	Lines    []string
	MaxLines int
}

// Render обрезает вывод по MaxLines.
func (t *TruncatedText) Render(width int) []string {
	lines := wrap.WrapLines(t.Lines, width)
	max := t.MaxLines
	if max <= 0 || len(lines) <= max {
		return lines
	}
	out := append([]string{}, lines[:max-1]...)
	out = append(out, "…")
	return out
}
