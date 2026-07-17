package component

import (
	"strings"
	"stell/tui/wrap"
)

// Box оборачивает дочерний компонент рамкой с заголовком.
type Box struct {
	Title   string
	Child   Component
	Border  string
	Padding int
}

// Render рисует рамку и тело дочернего компонента.
func (b *Box) Render(width int) []string {
	border := b.Border
	if border == "" {
		border = "\x1b[90m"
	}
	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}
	var body []string
	if b.Child != nil {
		body = b.Child.Render(innerW)
	}
	top := border + "┌" + strings.Repeat("─", max(1, innerW)) + "┐\x1b[0m"
	if b.Title != "" {
		title := " " + b.Title + " "
		if len(title) < innerW {
			top = border + "┌" + title + strings.Repeat("─", max(0, innerW-len(title))) + "┐\x1b[0m"
		}
	}
	bot := border + "└" + strings.Repeat("─", max(1, innerW)) + "┘\x1b[0m"
	out := []string{top}
	for _, line := range body {
		out = append(out, border+"│\x1b[0m"+wrap.PadRight(line, innerW)+border+"│\x1b[0m")
	}
	out = append(out, bot)
	return out
}
