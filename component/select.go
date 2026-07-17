package component

import (
	"fmt"
	"stell/tui/wrap"
)

// SelectList — вертикальный список с фильтрацией и выбором.
type SelectList struct {
	Items    []string
	AllItems []string // полный список для live-фильтра (опционально)
	Query    string
	Cursor   int
	Title    string
	focused  bool
	Selected string
	Accent   string
	Muted    string
	onPick   func(string)
}

// NewSelectList создаёт список выбора; onPick вызывается при Enter.
func NewSelectList(items []string, onPick func(string)) *SelectList {
	cp := append([]string(nil), items...)
	return &SelectList{
		Items:    cp,
		AllItems: cp,
		Accent:   "\x1b[36m",
		Muted:    "\x1b[90m",
		onPick:   onPick,
	}
}

func (s *SelectList) SetFocused(f bool) { s.focused = f }
func (s *SelectList) Focused() bool     { return s.focused }

// SetFilter применяет fuzzy-фильтр к AllItems (или Items).
func (s *SelectList) SetFilter(query string) {
	s.Query = query
	src := s.AllItems
	if len(src) == 0 {
		src = s.Items
	}
	if query == "" {
		s.Items = append([]string(nil), src...)
	} else {
		s.Items = wrap.FuzzyFilter(query, src, len(src))
	}
	if s.Cursor >= len(s.Items) {
		s.Cursor = 0
	}
	if s.Cursor < 0 {
		s.Cursor = 0
	}
}

// HandleInput обрабатывает навигацию, фильтр и выбор.
func (s *SelectList) HandleInput(data string) {
	switch data {
	case "\x1b[A", "k":
		if s.Cursor > 0 {
			s.Cursor--
		}
	case "\x1b[B", "j":
		if s.Cursor < len(s.Items)-1 {
			s.Cursor++
		}
	case "\x7f", "\b":
		if s.Query != "" {
			r := []rune(s.Query)
			s.SetFilter(string(r[:len(r)-1]))
		}
	case "\r", "\n":
		if s.Cursor >= 0 && s.Cursor < len(s.Items) {
			s.Selected = s.Items[s.Cursor]
			if s.onPick != nil {
				s.onPick(s.Selected)
			}
		}
	default:
		if len(data) == 1 && data[0] >= 32 && data[0] < 127 {
			s.SetFilter(s.Query + data)
		}
	}
}

// Render рисует заголовок, фильтр и строки списка.
func (s *SelectList) Render(width int) []string {
	var out []string
	total := len(s.Items)
	header := "select (↑/↓, enter, type to filter)"
	if s.Title != "" {
		header = s.Title + " (↑/↓, enter, type to filter)"
	}
	out = append(out, s.Muted+header+"\x1b[0m")
	if s.Query != "" {
		out = append(out, s.Muted+"filter: "+s.Query+"\x1b[0m")
	}
	if total > 0 {
		out = append(out, s.Muted+fmt.Sprintf("(%d/%d)", s.Cursor+1, total)+"\x1b[0m")
	}
	for i, item := range s.Items {
		prefix := "  "
		line := item
		if i == s.Cursor {
			prefix = s.Accent + "→ " + "\x1b[0m"
			line = s.Accent + item + "\x1b[0m"
		} else {
			line = s.Muted + item + "\x1b[0m"
		}
		out = append(out, wrap.Truncate(prefix+line, width))
	}
	if len(out) <= 1 {
		out = append(out, s.Muted+"(empty)"+"\x1b[0m")
	}
	return out
}
