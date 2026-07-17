package component

import (
	"strings"
	"stell/tui/wrap"
)

// SettingsItem — одна строка в SettingsList.
type SettingsItem struct {
	ID           string
	Label        string
	Description  string
	CurrentValue string
	Values       []string // если задано, Enter/Space циклически переключает
}

// SettingsList — список настроек с опциональным перебором значений.
type SettingsList struct {
	Items    []SettingsItem
	Cursor   int
	focused  bool
	Accent   string
	Muted    string
	OnChange func(id, value string)
	OnCancel func()
}

// NewSettingsList создаёт панель настроек.
func NewSettingsList(items []SettingsItem, onChange func(string, string), onCancel func()) *SettingsList {
	return &SettingsList{
		Items:    items,
		Accent:   "\x1b[36m",
		Muted:    "\x1b[90m",
		OnChange: onChange,
		OnCancel: onCancel,
	}
}

func (s *SettingsList) SetFocused(f bool) { s.focused = f }
func (s *SettingsList) Focused() bool     { return s.focused }

// HandleInput обрабатывает навигацию, цикл значений и отмену.
func (s *SettingsList) HandleInput(data string) {
	switch data {
	case "\x1b[A", "k":
		if s.Cursor > 0 {
			s.Cursor--
		}
	case "\x1b[B", "j":
		if s.Cursor < len(s.Items)-1 {
			s.Cursor++
		}
	case "\x1b":
		if s.OnCancel != nil {
			s.OnCancel()
		}
	case "\r", "\n", " ":
		if s.Cursor < 0 || s.Cursor >= len(s.Items) {
			return
		}
		it := &s.Items[s.Cursor]
		if len(it.Values) == 0 {
			return
		}
		idx := 0
		for i, v := range it.Values {
			if v == it.CurrentValue {
				idx = (i + 1) % len(it.Values)
				break
			}
		}
		it.CurrentValue = it.Values[idx]
		if s.OnChange != nil {
			s.OnChange(it.ID, it.CurrentValue)
		}
	}
}

// Render рисует строки настроек.
func (s *SettingsList) Render(width int) []string {
	var out []string
	out = append(out, s.Muted+"settings (↑/↓, enter/space cycle, esc)"+"\x1b[0m")
	for i, it := range s.Items {
		prefix := "  "
		style := s.Muted
		if i == s.Cursor {
			prefix = "> "
			style = s.Accent
		}
		line := prefix + it.Label
		if it.CurrentValue != "" {
			pad := width - wrap.VisibleLen(line) - wrap.VisibleLen(it.CurrentValue) - 1
			if pad < 1 {
				pad = 1
			}
			line += strings.Repeat(" ", pad) + it.CurrentValue
		}
		out = append(out, wrap.Truncate(style+line+"\x1b[0m", width))
		if i == s.Cursor && it.Description != "" {
			out = append(out, wrap.Truncate(s.Muted+"  "+it.Description+"\x1b[0m", width))
		}
	}
	return out
}
