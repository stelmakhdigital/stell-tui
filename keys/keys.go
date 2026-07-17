package keys

import (
	"encoding/json"
	"os"
	"strings"
)

// KeyAction сопоставляет аккорд клавиш с id действия.
type KeyAction struct {
	Key    string `json:"key"`
	Action string `json:"action"`
}

// KeyDef описывает биндинги по умолчанию для одного id.
type KeyDef struct {
	DefaultKeys string
	Description string
}

// KeyMap хранит chord → action.
type KeyMap struct {
	byKey map[string]string
}

// NewKeyMap создаёт пустую карту.
func NewKeyMap() *KeyMap {
	return &KeyMap{byKey: map[string]string{}}
}

// Bind регистрирует key → action (нормализованный lowercase).
func (m *KeyMap) Bind(key, action string) {
	if m.byKey == nil {
		m.byKey = map[string]string{}
	}
	m.byKey[NormalizeKeyChord(NormalizeKey(key))] = action
}

// Lookup возвращает действие для аккорда клавиш.
func (m *KeyMap) Lookup(key string) (string, bool) {
	if m == nil {
		return "", false
	}
	a, ok := m.byKey[NormalizeKeyChord(NormalizeKey(key))]
	return a, ok
}

// KeybindingsManager резолвит namespaced id биндингов.
type KeybindingsManager struct {
	definitions map[string]KeyDef
	user        map[string]string
	keysByID    map[string][]string
}

// DefaultTUIKeybindings возвращает биндинги editor/input по умолчанию (id tui.*).
func DefaultTUIKeybindings() map[string]KeyDef {
	return map[string]KeyDef{
		"tui.editor.cursorLeft":           {DefaultKeys: "left", Description: "Move cursor left"},
		"tui.editor.cursorRight":          {DefaultKeys: "right", Description: "Move cursor right"},
		"tui.editor.cursorUp":             {DefaultKeys: "up", Description: "Move cursor up"},
		"tui.editor.cursorDown":           {DefaultKeys: "down", Description: "Move cursor down"},
		"tui.editor.cursorWordLeft":       {DefaultKeys: "alt+left,ctrl+left,alt+b", Description: "Move cursor word left"},
		"tui.editor.cursorWordRight":      {DefaultKeys: "alt+right,ctrl+right,alt+f", Description: "Move cursor word right"},
		"tui.editor.cursorLineStart":      {DefaultKeys: "home,ctrl+a", Description: "Move to line start"},
		"tui.editor.cursorLineEnd":        {DefaultKeys: "end,ctrl+e", Description: "Move to line end"},
		"tui.editor.jumpForward":          {DefaultKeys: "ctrl+]", Description: "Jump forward to character"},
		"tui.editor.jumpBackward":         {DefaultKeys: "ctrl+alt+]", Description: "Jump backward to character"},
		"tui.editor.pageUp":               {DefaultKeys: "pgup", Description: "Page up"},
		"tui.editor.pageDown":             {DefaultKeys: "pgdown", Description: "Page down"},
		"tui.editor.deleteCharBackward":   {DefaultKeys: "backspace", Description: "Delete character backward"},
		"tui.editor.deleteCharForward":    {DefaultKeys: "delete,ctrl+d", Description: "Delete character forward"},
		"tui.editor.deleteWordBackward":   {DefaultKeys: "ctrl+w,alt+backspace", Description: "Delete word backward"},
		"tui.editor.deleteWordForward":    {DefaultKeys: "alt+d,alt+delete", Description: "Delete word forward"},
		"tui.editor.deleteToLineStart":    {DefaultKeys: "ctrl+u", Description: "Delete to line start"},
		"tui.editor.deleteToLineEnd":      {DefaultKeys: "ctrl+k", Description: "Delete to line end"},
		"tui.editor.yank":                 {DefaultKeys: "ctrl+y", Description: "Yank"},
		"tui.editor.yankPop":              {DefaultKeys: "alt+y", Description: "Yank pop"},
		"tui.editor.undo":                 {DefaultKeys: "ctrl+-", Description: "Undo"},
		"tui.input.newLine":               {DefaultKeys: "shift+enter,ctrl+j", Description: "Insert newline"},
		"tui.input.submit":               {DefaultKeys: "enter", Description: "Submit input"},
		"tui.input.tab":                  {DefaultKeys: "tab", Description: "Tab / autocomplete"},
		"tui.select.up":                  {DefaultKeys: "up", Description: "Move selection up"},
		"tui.select.down":                {DefaultKeys: "down", Description: "Move selection down"},
		"tui.select.pageUp":              {DefaultKeys: "pgup", Description: "Selection page up"},
		"tui.select.pageDown":            {DefaultKeys: "pgdown", Description: "Selection page down"},
		"tui.select.confirm":             {DefaultKeys: "enter", Description: "Confirm selection"},
		"tui.select.cancel":              {DefaultKeys: "esc,ctrl+c", Description: "Cancel selection"},
	}
}

// NewKeybindingsManager создаёт менеджер из определений и опциональных override пользователя.
func NewKeybindingsManager(definitions map[string]KeyDef, user map[string]string) *KeybindingsManager {
	m := &KeybindingsManager{
		definitions: definitions,
		user:        user,
		keysByID:    map[string][]string{},
	}
	m.rebuild()
	return m
}

func (m *KeybindingsManager) rebuild() {
	if m == nil {
		return
	}
	m.keysByID = map[string][]string{}
	for id, def := range m.definitions {
		if keys, ok := m.user[id]; ok && keys != "" {
			m.keysByID[id] = splitKeyList(keys)
			continue
		}
		m.keysByID[id] = splitKeyList(def.DefaultKeys)
	}
}

func splitKeyList(binding string) []string {
	parts := strings.Split(binding, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = NormalizeKey(strings.TrimSpace(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Matches сообщает, совпадает ли keyData с id биндинга.
func (m *KeybindingsManager) Matches(keyData, bindingID string) bool {
	if m == nil {
		return false
	}
	keyData = NormalizeKey(keyData)
	for _, key := range m.keysByID[bindingID] {
		if key == keyData {
			return true
		}
	}
	return false
}

// Keys возвращает настроенные аккорды для id.
func (m *KeybindingsManager) Keys(bindingID string) []string {
	if m == nil {
		return nil
	}
	return append([]string(nil), m.keysByID[bindingID]...)
}

// NormalizeKey lowercases and collapses whitespace around '+'.
func NormalizeKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	parts := strings.Split(key, "+")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return strings.Join(parts, "+")
}

// LoadKeyBindingsJSON загружает карты [{key,action}] или {key: action}.
func LoadKeyBindingsJSON(path string) (*KeyMap, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := NewKeyMap()
	var arr []KeyAction
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		for _, a := range arr {
			m.Bind(a.Key, a.Action)
		}
		return m, nil
	}
	var obj map[string]string
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	for k, v := range obj {
		m.Bind(k, v)
	}
	return m, nil
}
