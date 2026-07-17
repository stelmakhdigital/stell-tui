package keys

import (
	"sort"
	"strings"
)

var keyNameAliases = map[string]string{
	"escape":   "esc",
	"return":   "enter",
	"pageup":   "pgup",
	"pagedown": "pgdown",
}

// NormalizeKeyChord канонизирует идентификатор клавиши для поиска биндинга.
// Приводит алиасы имён и упорядочивает модификаторы (shift, ctrl, alt, super).
func NormalizeKeyChord(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	if key == "" {
		return ""
	}
	parts := strings.Split(key, "+")
	if len(parts) == 1 {
		if alias, ok := keyNameAliases[parts[0]]; ok {
			return alias
		}
		return parts[0]
	}
	keyPart := parts[len(parts)-1]
	if alias, ok := keyNameAliases[keyPart]; ok {
		keyPart = alias
	}
	var mods []string
	for _, p := range parts[:len(parts)-1] {
		p = strings.TrimSpace(p)
		switch p {
		case "shift", "ctrl", "alt", "super":
			mods = append(mods, p)
		}
	}
	sort.Strings(mods)
	// Канонический порядок: shift, ctrl, alt, super
	order := map[string]int{"shift": 0, "ctrl": 1, "alt": 2, "super": 3}
	sort.SliceStable(mods, func(i, j int) bool {
		return order[mods[i]] < order[mods[j]]
	})
	if len(mods) == 0 {
		return keyPart
	}
	return strings.Join(mods, "+") + "+" + keyPart
}
