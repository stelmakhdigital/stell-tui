package keys

import (
	"os"
	"regexp"
	"strings"
)

const (
	modShift = 1
	modAlt   = 2
	modCtrl  = 4
	modSuper = 8
	lockMask = 64 + 128
)

const (
	cpEscape    = 27
	cpTab       = 9
	cpEnter     = 13
	cpSpace     = 32
	cpBackspace = 127
	cpKpEnter   = 57414
)

const (
	cpArrowUp    = -1
	cpArrowDown  = -2
	cpArrowRight = -3
	cpArrowLeft  = -4
)

const (
	cpDelete   = -10
	cpInsert   = -11
	cpPageUp   = -12
	cpPageDown = -13
	cpHome     = -14
	cpEnd      = -15
)

var symbolKeys = map[string]struct{}{
	"`": {}, "-": {}, "=": {}, "[": {}, "]": {}, "\\": {}, ";": {}, "'": {}, ",": {}, ".": {}, "/": {},
	"!": {}, "@": {}, "#": {}, "$": {}, "%": {}, "^": {}, "&": {}, "*": {}, "(": {}, ")": {},
	"_": {}, "+": {}, "|": {}, "~": {}, "{": {}, "}": {}, ":": {}, "<": {}, ">": {}, "?": {},
}

var kittyFunctionalEquivalents = map[int]int{
	57399: 48, 57400: 49, 57401: 50, 57402: 51, 57403: 52,
	57404: 53, 57405: 54, 57406: 55, 57407: 56, 57408: 57,
	57409: 46, 57410: 47, 57411: 42, 57412: 45, 57413: 43,
	57415: 61, 57416: 44,
	57417: cpArrowLeft, 57418: cpArrowRight, 57419: cpArrowUp, 57420: cpArrowDown,
	57421: cpPageUp, 57422: cpPageDown, 57423: cpHome, 57424: cpEnd,
	57425: cpInsert, 57426: cpDelete,
}

var legacySequenceKeyIDs = map[string]string{
	"\x1bOA": "up", "\x1bOB": "down", "\x1bOC": "right", "\x1bOD": "left",
	"\x1bOH": "home", "\x1bOF": "end",
	"\x1b[E": "clear", "\x1bOE": "clear", "\x1bOe": "ctrl+clear", "\x1b[e": "shift+clear",
	"\x1b[2~": "insert", "\x1b[2$": "shift+insert", "\x1b[2^": "ctrl+insert",
	"\x1b[3$": "shift+delete", "\x1b[3^": "ctrl+delete",
	"\x1b[[5~": "pageUp", "\x1b[[6~": "pageDown",
	"\x1b[a": "shift+up", "\x1b[b": "shift+down", "\x1b[c": "shift+right", "\x1b[d": "shift+left",
	"\x1bOa": "ctrl+up", "\x1bOb": "ctrl+down", "\x1bOc": "ctrl+right", "\x1bOd": "ctrl+left",
	"\x1b[5$": "shift+pageUp", "\x1b[6$": "shift+pageDown",
	"\x1b[7$": "shift+home", "\x1b[8$": "shift+end",
	"\x1b[5^": "ctrl+pageUp", "\x1b[6^": "ctrl+pageDown",
	"\x1b[7^": "ctrl+home", "\x1b[8^": "ctrl+end",
	"\x1bOP": "f1", "\x1bOQ": "f2", "\x1bOR": "f3", "\x1bOS": "f4",
	"\x1b[11~": "f1", "\x1b[12~": "f2", "\x1b[13~": "f3", "\x1b[14~": "f4",
	"\x1b[[A": "f1", "\x1b[[B": "f2", "\x1b[[C": "f3", "\x1b[[D": "f4", "\x1b[[E": "f5",
	"\x1b[15~": "f5", "\x1b[17~": "f6", "\x1b[18~": "f7", "\x1b[19~": "f8",
	"\x1b[20~": "f9", "\x1b[21~": "f10", "\x1b[23~": "f11", "\x1b[24~": "f12",
	"\x1bb": "alt+left", "\x1bf": "alt+right", "\x1bp": "alt+up", "\x1bn": "alt+down",
}

var (
	reKittyCSIu      = regexp.MustCompile(`^\x1b\[(\d+)(?::(\d*))?(?::(\d+))?(?:;(\d+))?(?::(\d+))?u$`)
	reKittyArrow     = regexp.MustCompile(`^\x1b\[1;(\d+)(?::(\d+))?([ABCD])$`)
	reKittyFunc      = regexp.MustCompile(`^\x1b\[(\d+)(?:;(\d+))?(?::(\d+))?~$`)
	reKittyHomeEnd   = regexp.MustCompile(`^\x1b\[1;(\d+)(?::(\d+))?([HF])$`)
	reModifyOther    = regexp.MustCompile(`^\x1b\[27;(\d+);(\d+)~$`)
)

type parsedKitty struct {
	codepoint     int
	shiftedKey    int
	baseLayoutKey int
	modifier      int
	hasBase       bool
}

type parsedModifyOtherKeys struct {
	codepoint int
	modifier  int
}

func normalizeKittyFunctional(codepoint int) int {
	if v, ok := kittyFunctionalEquivalents[codepoint]; ok {
		return v
	}
	return codepoint
}

func normalizeShiftedLetterIdentity(codepoint, modifier int) int {
	effective := modifier & ^lockMask
	if effective&modShift != 0 && codepoint >= 65 && codepoint <= 90 {
		return codepoint + 32
	}
	return codepoint
}

func isSymbolKey(ch string) bool {
	_, ok := symbolKeys[ch]
	return ok
}

func isWindowsTerminalSession() bool {
	if os.Getenv("WT_SESSION") == "" {
		return false
	}
	return os.Getenv("SSH_CONNECTION") == "" && os.Getenv("SSH_CLIENT") == "" && os.Getenv("SSH_TTY") == ""
}

func parseKittySequence(data string) *parsedKitty {
	if m := reKittyCSIu.FindStringSubmatch(data); m != nil {
		cp := atoi(m[1])
		var shifted, base int
		hasBase := false
		if m[2] != "" {
			shifted = atoi(m[2])
		}
		if m[3] != "" {
			base = atoi(m[3])
			hasBase = true
		}
		mod := 0
		if m[4] != "" {
			mod = atoi(m[4]) - 1
		}
		_ = m[5] // тип события
		p := &parsedKitty{codepoint: cp, modifier: mod, hasBase: hasBase, shiftedKey: shifted, baseLayoutKey: base}
		return p
	}
	if m := reKittyArrow.FindStringSubmatch(data); m != nil {
		arrow := map[string]int{"A": cpArrowUp, "B": cpArrowDown, "C": cpArrowRight, "D": cpArrowLeft}
		return &parsedKitty{codepoint: arrow[m[3]], modifier: atoi(m[1]) - 1}
	}
	if m := reKittyFunc.FindStringSubmatch(data); m != nil {
		funcCodes := map[int]int{2: cpInsert, 3: cpDelete, 5: cpPageUp, 6: cpPageDown, 7: cpHome, 8: cpEnd}
		keyNum := atoi(m[1])
		cp, ok := funcCodes[keyNum]
		if !ok {
			return nil
		}
		mod := 0
		if m[2] != "" {
			mod = atoi(m[2]) - 1
		}
		return &parsedKitty{codepoint: cp, modifier: mod}
	}
	if m := reKittyHomeEnd.FindStringSubmatch(data); m != nil {
		cp := cpEnd
		if m[3] == "H" {
			cp = cpHome
		}
		return &parsedKitty{codepoint: cp, modifier: atoi(m[1]) - 1}
	}
	return nil
}

func parseModifyOtherKeysSequence(data string) *parsedModifyOtherKeys {
	m := reModifyOther.FindStringSubmatch(data)
	if m == nil {
		return nil
	}
	return &parsedModifyOtherKeys{
		codepoint: atoi(m[2]),
		modifier:  atoi(m[1]) - 1,
	}
}

func formatKeyNameWithModifiers(keyName string, modifier int) string {
	effective := modifier & ^lockMask
	supported := modShift | modCtrl | modAlt | modSuper
	if effective&^supported != 0 {
		return ""
	}
	var mods []string
	if effective&modShift != 0 {
		mods = append(mods, "shift")
	}
	if effective&modCtrl != 0 {
		mods = append(mods, "ctrl")
	}
	if effective&modAlt != 0 {
		mods = append(mods, "alt")
	}
	if effective&modSuper != 0 {
		mods = append(mods, "super")
	}
	if len(mods) == 0 {
		return keyName
	}
	return strings.Join(mods, "+") + "+" + keyName
}

func formatParsedKey(codepoint, modifier int, baseLayoutKey int) string {
	normalized := normalizeKittyFunctional(codepoint)
	identity := normalizeShiftedLetterIdentity(normalized, modifier)

	isLatin := identity >= 97 && identity <= 122
	isDigit := identity >= 48 && identity <= 57
	isKnownSymbol := isSymbolKey(string(rune(identity)))
	effective := identity
	if !isLatin && !isDigit && !isKnownSymbol && baseLayoutKey != 0 {
		effective = baseLayoutKey
	}

	var keyName string
	switch effective {
	case cpEscape:
		keyName = "escape"
	case cpTab:
		keyName = "tab"
	case cpEnter, cpKpEnter:
		keyName = "enter"
	case cpSpace:
		keyName = "space"
	case cpBackspace:
		keyName = "backspace"
	case cpDelete:
		keyName = "delete"
	case cpInsert:
		keyName = "insert"
	case cpHome:
		keyName = "home"
	case cpEnd:
		keyName = "end"
	case cpPageUp:
		keyName = "pageUp"
	case cpPageDown:
		keyName = "pageDown"
	case cpArrowUp:
		keyName = "up"
	case cpArrowDown:
		keyName = "down"
	case cpArrowLeft:
		keyName = "left"
	case cpArrowRight:
		keyName = "right"
	default:
		if effective >= 48 && effective <= 57 {
			keyName = string(rune(effective))
		} else if effective >= 97 && effective <= 122 {
			keyName = string(rune(effective))
		} else if isSymbolKey(string(rune(effective))) {
			keyName = string(rune(effective))
		}
	}
	if keyName == "" {
		return ""
	}
	return formatKeyNameWithModifiers(keyName, modifier)
}

// ParseKey разбирает сырой ввод терминала в идентификатор клавиши.
// Пустая строка — нераспознанный печатный ввод.
func ParseKey(data string) string {
	if kitty := parseKittySequence(data); kitty != nil {
		base := 0
		if kitty.hasBase {
			base = kitty.baseLayoutKey
		}
		return formatParsedKey(kitty.codepoint, kitty.modifier, base)
	}
	if mok := parseModifyOtherKeysSequence(data); mok != nil {
		return formatParsedKey(mok.codepoint, mok.modifier, 0)
	}
	if kittyProtocolActive {
		if data == "\x1b\r" || data == "\n" {
			return "shift+enter"
		}
	}
	if id, ok := legacySequenceKeyIDs[data]; ok {
		return id
	}
	switch data {
	case "\x1b":
		return "escape"
	case "\x1c":
		return "ctrl+\\"
	case "\x1d":
		return "ctrl+]"
	case "\x1f":
		return "ctrl+-"
	case "\x1b\x1b":
		return "ctrl+alt+["
	case "\x1b\x1c":
		return "ctrl+alt+\\"
	case "\x1b\x1d":
		return "ctrl+alt+]"
	case "\x1b\x1f":
		return "ctrl+alt+-"
	case "\t":
		return "tab"
	case "\r":
		return "enter"
	case "\n":
		if !kittyProtocolActive {
			return "enter"
		}
	case "\x1bOM":
		return "enter"
	case "\x00":
		return "ctrl+space"
	case " ":
		return "space"
	case "\x7f":
		return "backspace"
	case "\x08":
		if isWindowsTerminalSession() {
			return "ctrl+backspace"
		}
		return "backspace"
	case "\x1b[Z":
		return "shift+tab"
	case "\x1b[13;2~", "\x1b[27;2;13~":
		return "shift+enter"
	case "\x1b[27;2;9~":
		return "shift+tab"
	case "\x1b[27;6;80~", "\x1b[80;6u":
		return "shift+ctrl+p"
	case "\x16", "\x1b[118;5u":
		return "ctrl+v"
	case "\x1b\r":
		if !kittyProtocolActive {
			return "alt+enter"
		}
	case "\x1b ":
		if !kittyProtocolActive {
			return "alt+space"
		}
	case "\x1b\x7f", "\x1b\b":
		return "alt+backspace"
	case "\x1bB":
		if !kittyProtocolActive {
			return "alt+left"
		}
	case "\x1bF":
		if !kittyProtocolActive {
			return "alt+right"
		}
	case "\x1b[A":
		return "up"
	case "\x1b[B":
		return "down"
	case "\x1b[C":
		return "right"
	case "\x1b[D":
		return "left"
	case "\x1b[H", "\x1bOH":
		return "home"
	case "\x1b[F", "\x1bOF":
		return "end"
	case "\x1b[3~":
		return "delete"
	case "\x1b[5~":
		return "pageUp"
	case "\x1b[6~":
		return "pageDown"
	case "\x1b[5;3~":
		return "alt+pgup"
	case "\x1b[6;3~":
		return "alt+pgdown"
	case "\x1bv", "\x1bV":
		return "alt+v"
	}
	if !kittyProtocolActive && len(data) == 2 && data[0] == '\x1b' {
		code := data[1]
		if code >= 1 && code <= 26 {
			return "ctrl+alt+" + string(rune(code+96))
		}
		if (code >= 'a' && code <= 'z') || (code >= '0' && code <= '9') || isSymbolKey(string(rune(code))) {
			return "alt+" + strings.ToLower(string(rune(code)))
		}
	}
	if len(data) == 1 {
		code := data[0]
		if code >= 1 && code <= 26 {
			return "ctrl+" + string(rune(code+96))
		}
		if code >= 32 && code <= 126 {
			return data
		}
	}
	return ""
}

// MatchesKey сообщает, совпадает ли сырой ввод терминала с идентификатором клавиши.
func MatchesKey(data, keyID string) bool {
	parsed := ParseKey(data)
	if parsed == "" {
		return false
	}
	return NormalizeKeyChord(parsed) == NormalizeKeyChord(keyID)
}

// DecodePrintableKey извлекает печатный символ из CSI Kitty/modifyOtherKeys.
func DecodePrintableKey(data string) string {
	if ch := decodeKittyPrintable(data); ch != "" {
		return ch
	}
	return decodeModifyOtherKeysPrintable(data)
}

func decodeKittyPrintable(data string) string {
	full := reKittyCSIu.FindStringSubmatch(data)
	if full == nil {
		return ""
	}
	cp := atoi(full[1])
	modValue := 1
	if full[4] != "" {
		modValue = atoi(full[4])
	}
	modifier := modValue - 1
	allowed := modShift | lockMask
	if modifier&^allowed != 0 {
		return ""
	}
	if modifier&(modAlt|modCtrl) != 0 {
		return ""
	}
	effective := cp
	if modifier&modShift != 0 && full[2] != "" {
		effective = atoi(full[2])
	}
	effective = normalizeKittyFunctional(effective)
	if effective < 32 {
		return ""
	}
	return string(rune(effective))
}

func decodeModifyOtherKeysPrintable(data string) string {
	parsed := parseModifyOtherKeysSequence(data)
	if parsed == nil {
		return ""
	}
	modifier := parsed.modifier & ^lockMask
	if modifier&^modShift != 0 {
		return ""
	}
	if parsed.codepoint < 32 {
		return ""
	}
	return string(rune(parsed.codepoint))
}

func atoi(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// IsKeyRelease сообщает, похожи ли данные на событие отпускания клавиши Kitty.
func IsKeyRelease(data string) bool {
	if strings.Contains(data, "\x1b[200~") {
		return false
	}
	for _, frag := range []string{":3u", ":3~", ":3A", ":3B", ":3C", ":3D", ":3H", ":3F"} {
		if strings.Contains(data, frag) {
			return true
		}
	}
	return false
}

// ParseKeyboardProtocolResponse разбирает ответы согласования протокола Kitty.
func ParseKeyboardProtocolResponse(sequence string) (flags int, ok bool) {
	if len(sequence) >= 4 && sequence[:2] == "\x1b[" && sequence[len(sequence)-1] == 'u' && sequence[2] == '?' {
		flags = atoi(sequence[3 : len(sequence)-1])
		return flags, true
	}
	return 0, false
}
