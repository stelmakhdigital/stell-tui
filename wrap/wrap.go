// Package wrap — утилиты визуальной ширины строк с учётом ANSI/OSC.
package wrap

import (
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// WrapLines переносит каждую строку под width.
func WrapLines(lines []string, width int) []string {
	if width <= 0 {
		return append([]string{}, lines...)
	}
	var out []string
	for _, line := range lines {
		out = append(out, WrapLine(line, width)...)
	}
	return out
}

// WrapLine переносит одну строку под width.
func WrapLine(line string, width int) []string {
	if width <= 0 || VisibleLen(line) <= width {
		return []string{line}
	}
	var out []string
	for VisibleLen(line) > width {
		cut := CutAt(line, width)
		if cut <= 0 {
			cut = len(line)
		}
		out = append(out, line[:cut])
		line = line[cut:]
	}
	if line != "" || len(out) == 0 {
		out = append(out, line)
	}
	return out
}

// CutAt возвращает байтовый индекс обрезки по визуальной ширине width.
func CutAt(s string, width int) int {
	n := 0
	for i := 0; i < len(s); {
		skip, next := SkipANSI(s, i)
		if skip {
			i = next
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		w := runeDisplayWidth(r)
		if n+w > width {
			if n == 0 {
				return i + size
			}
			return i
		}
		n += w
		i += size
	}
	return len(s)
}

// VisibleLen возвращает печатную ширину s, игнорируя ANSI/OSC/APC.
func VisibleLen(s string) int {
	n := 0
	for i := 0; i < len(s); {
		skip, next := SkipANSI(s, i)
		if skip {
			i = next
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		n += runeDisplayWidth(r)
		i += size
	}
	return n
}

func runeDisplayWidth(r rune) int {
	w := runewidth.RuneWidth(r)
	if w <= 0 {
		return 0
	}
	return w
}

// SkipANSI пропускает CSI/OSC и прочие ESC-последовательности. Возвращает (skipped, newIndex).
func SkipANSI(s string, i int) (bool, int) {
	if i >= len(s) || s[i] != 0x1b {
		return false, i
	}
	j := i + 1
	if j >= len(s) {
		return true, len(s)
	}
	switch s[j] {
	case '[': // CSI
		j++
		for j < len(s) {
			c := s[j]
			j++
			if c >= 0x40 && c <= 0x7e {
				break
			}
		}
		return true, j
	case ']': // OSC … BEL or ST
		j++
		for j < len(s) {
			if s[j] == 0x07 {
				return true, j + 1
			}
			if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
				return true, j + 2
			}
			j++
		}
		return true, len(s)
	case '_': // APC … BEL or ST (CursorMarker, Kitty graphics, …)
		j++
		for j < len(s) {
			if s[j] == 0x07 {
				return true, j + 1
			}
			if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
				return true, j + 2
			}
			j++
		}
		return true, len(s)
	case '(':
		fallthrough
	case ')':
		if j+1 < len(s) {
			return true, j + 2
		}
		return true, len(s)
	default:
		return true, j + 1
	}
}

// PadVisible дополняет или обрезает строку до визуальной ширины width.
func PadVisible(s string, width int) string {
	n := VisibleLen(s)
	if n > width {
		return Truncate(s, width)
	}
	if n == width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

// PadRight — алиас PadVisible.
func PadRight(s string, width int) string {
	return PadVisible(s, width)
}

// Truncate обрезает строку до визуальной ширины, добавляя многоточие при обрезке.
func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if VisibleLen(s) <= width {
		return s
	}
	if width == 1 {
		if strings.Contains(s, "\x1b") {
			return "…" + "\x1b[0m"
		}
		return "…"
	}
	cut := CutAt(s, width-1)
	out := s[:cut] + "…"
	if strings.Contains(s[:cut], "\x1b") {
		out += "\x1b[0m"
	}
	return out
}

// SplitParas делит текст на строки по \n (с нормализацией \r\n).
func SplitParas(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

// CursorMarker — APC zero-width маркер позиции курсора.
// DiffEngine снимает его и позиционирует аппаратный курсор.
const CursorMarker = "\x1b_stell:c\x07"

// StripANSI удаляет ANSI/OSC/APC escape-последовательности из строки.
func StripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		skip, next := SkipANSI(s, i)
		if skip {
			i = next
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
