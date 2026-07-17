package editor

import (
	"stell/tui/wrap"
	"strings"
	"unicode/utf8"
)

// KillRing — буфер kill/yank в стиле Emacs.
type KillRing struct {
	entries []string
	max     int
}

func NewKillRing(max int) *KillRing {
	if max < 1 {
		max = 32
	}
	return &KillRing{max: max}
}

func (k *KillRing) Push(text string, prepend, accumulate bool) {
	if text == "" || k == nil {
		return
	}
	if accumulate && len(k.entries) > 0 {
		last := k.entries[len(k.entries)-1]
		if prepend {
			k.entries[len(k.entries)-1] = text + last
		} else {
			k.entries[len(k.entries)-1] = last + text
		}
		return
	}
	k.entries = append(k.entries, text)
	if len(k.entries) > k.max {
		k.entries = k.entries[len(k.entries)-k.max:]
	}
}

func (k *KillRing) Peek() string {
	if k == nil || len(k.entries) == 0 {
		return ""
	}
	return k.entries[len(k.entries)-1]
}

func (k *KillRing) Rotate() {
	if k == nil || len(k.entries) < 2 {
		return
	}
	last := k.entries[len(k.entries)-1]
	k.entries = append([]string{last}, k.entries[:len(k.entries)-1]...)
}

type undoEntry struct {
	value  string
	cursor int
}

// Editor — многострочный редактор с переносом, undo и kill-ring.
type Editor struct {
	value              string
	cursor             int
	focused            bool
	placeholder        string
	borderColor        string
	boxBorder          bool
	pasteThresh        int
	pasteN             int
	undo               []undoEntry
	undoMax            int
	kill               *KillRing
	accumKill          bool
	showHardwareCursor bool
	paddingX           int
	jumpMode           string // "forward" | "backward" | ""
	OnSubmit           func(value string)
	DisableSubmit      bool
}

// SetShowHardwareCursor при фокусе использует wrap.CursorMarker вместо блочного глифа.
func (e *Editor) SetShowHardwareCursor(on bool) {
	if e != nil {
		e.showHardwareCursor = on
	}
}

// SetPaddingX задаёт левый отступ строк тела редактора.
func (e *Editor) SetPaddingX(n int) {
	if e == nil {
		return
	}
	if n < 0 {
		n = 0
	}
	if n > 8 {
		n = 8
	}
	e.paddingX = n
}

func NewEditor() *Editor {
	return &Editor{
		pasteThresh: 10,
		undoMax:     64,
		kill:        NewKillRing(32),
	}
}

func (e *Editor) Value() string { return e.value }

func (e *Editor) SetValue(v string) {
	e.pushUndo()
	e.value = v
	e.cursor = len(v)
	e.accumKill = false
}

func (e *Editor) SetPlaceholder(s string) { e.placeholder = s }

func (e *Editor) SetBorderColor(ansi string) { e.borderColor = ansi }

// SetBoxBorder включает полную рамку (боковые │), а не только верх/низ.
func (e *Editor) SetBoxBorder(on bool) {
	if e != nil {
		e.boxBorder = on
	}
}

func (e *Editor) Focused() bool     { return e.focused }
func (e *Editor) SetFocused(f bool) { e.focused = f }

func (e *Editor) pushUndo() {
	if e.undoMax <= 0 {
		return
	}
	e.undo = append(e.undo, undoEntry{value: e.value, cursor: e.cursor})
	if len(e.undo) > e.undoMax {
		e.undo = e.undo[len(e.undo)-e.undoMax:]
	}
}

func (e *Editor) Undo() {
	if len(e.undo) == 0 {
		return
	}
	u := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.value = u.value
	e.cursor = u.cursor
	e.accumKill = false
}

func (e *Editor) HandleInput(data string) {
	// Payload bracketed paste от хоста.
	if strings.HasPrefix(data, "\x1b[200~") && strings.HasSuffix(data, "\x1b[201~") {
		inner := data[6 : len(data)-6]
		e.pushUndo()
		e.insert(inner)
		e.accumKill = false
		return
	}
	if e.jumpMode != "" {
		switch data {
		case "\x1d", "\x1b\x1d":
			e.jumpMode = ""
			return
		}
		if strings.HasPrefix(data, "\x1b") && len(data) > 1 {
			e.jumpMode = ""
		} else if len(data) == 1 && data[0] >= 32 {
			mode := e.jumpMode
			e.jumpMode = ""
			e.jumpToChar(data, mode)
			return
		} else if len(data) > 0 && data[0] >= 32 {
			mode := e.jumpMode
			e.jumpMode = ""
			e.jumpToChar(data, mode)
			return
		} else {
			e.jumpMode = ""
		}
	}
	switch data {
	case "\x7f", "\b":
		e.backspace()
	case "\x1b[3~":
		e.deleteForward()
	case "\x1b[D":
		e.move(-1)
		e.accumKill = false
	case "\x1b[C":
		e.move(1)
		e.accumKill = false
	case "\x1b[1;5D", "\x1b[5D", "\x1bb": // Ctrl/Alt-Left, Alt-b
		e.moveWord(-1)
		e.accumKill = false
	case "\x1b[1;5C", "\x1b[5C", "\x1bf": // Ctrl/Alt-Right, Alt-f
		e.moveWord(1)
		e.accumKill = false
	case "\x1bd", "\x1bD": // Alt-d — удалить слово вперёд
		e.killWord(1)
	case "\x17": // Ctrl-W — удалить слово назад
		e.killWord(-1)
	case "\x01": // Ctrl-A — начало строки
		e.moveToLineStart()
	case "\x05": // Ctrl-E — конец строки
		e.moveToLineEnd()
	case "\x1b[H", "\x1b[1~", "\x1b[7~":
		e.moveToLineStart()
	case "\x1b[F", "\x1b[4~", "\x1b[8~":
		e.moveToLineEnd()
	case "\x1f":
		e.Undo()
	case "\x15":
		e.killLine(true)
	case "\x0b":
		e.killLine(false)
	case "\x19":
		e.Yank()
	case "\x03":
		return
	case "\n", "\r":
		if e.OnSubmit != nil && !e.DisableSubmit {
			e.OnSubmit(e.Value())
			return
		}
		e.pushUndo()
		e.insert("\n")
		e.accumKill = false
	case "\x1b[A":
		e.moveVertical(-1)
		e.accumKill = false
	case "\x1b[B":
		e.moveVertical(1)
		e.accumKill = false
	case "\x1b[5~":
		e.PageUp()
	case "\x1b[6~":
		e.PageDown()
	case "\x1d":
		e.jumpMode = "forward"
	case "\x1b\x1d":
		e.jumpMode = "backward"
	case "\x1by", "\x1bY":
		e.YankPop()
	default:
		if strings.HasPrefix(data, "\x1b") {
			return
		}
		if strings.Count(data, "\n") >= e.pasteThresh {
			e.pasteN++
			marker := "[paste #" + itoa(e.pasteN) + " +" + itoa(strings.Count(data, "\n")) + " lines]"
			e.pushUndo()
			e.insert(marker)
			e.accumKill = false
			return
		}
		e.pushUndo()
		e.insert(data)
		e.accumKill = false
	}
}

func (e *Editor) Yank() {
	if t := e.kill.Peek(); t != "" {
		e.pushUndo()
		e.insert(t)
	}
}

func (e *Editor) YankPop() {
	if e.kill == nil || len(e.kill.entries) < 2 {
		return
	}
	prev := e.kill.Peek()
	e.kill.Rotate()
	next := e.kill.Peek()
	if prev == "" || next == "" {
		return
	}
	if e.cursor >= len(prev) && strings.HasSuffix(e.value[:e.cursor], prev) {
		e.pushUndo()
		start := e.cursor - len(prev)
		e.value = e.value[:start] + next + e.value[e.cursor:]
		e.cursor = start + len(next)
	}
}

func (e *Editor) PageUp() {
	for i := 0; i < 10; i++ {
		e.moveVertical(-1)
	}
	e.accumKill = false
}

func (e *Editor) PageDown() {
	for i := 0; i < 10; i++ {
		e.moveVertical(1)
	}
	e.accumKill = false
}

func (e *Editor) moveToLineStart() {
	start := 0
	if e.cursor > 0 {
		if i := strings.LastIndexByte(e.value[:e.cursor], '\n'); i >= 0 {
			start = i + 1
		}
	}
	e.cursor = start
	e.accumKill = false
}

func (e *Editor) moveToLineEnd() {
	end := len(e.value)
	if i := strings.IndexByte(e.value[e.cursor:], '\n'); i >= 0 {
		end = e.cursor + i
	}
	e.cursor = end
	e.accumKill = false
}

func (e *Editor) moveVertical(dir int) {
	if e.value == "" {
		return
	}
	lines := strings.Split(e.value, "\n")
	// Находим текущую строку и колонку по байтовому смещению.
	off := 0
	lineIdx := 0
	col := 0
	for i, line := range lines {
		end := off + len(line)
		if e.cursor <= end || i == len(lines)-1 {
			lineIdx = i
			col = e.cursor - off
			if col < 0 {
				col = 0
			}
			if col > len(line) {
				col = len(line)
			}
			break
		}
		off = end + 1 // +1 за перевод строки
	}
	target := lineIdx + dir
	if target < 0 || target >= len(lines) {
		if dir < 0 {
			e.cursor = 0
		} else {
			e.cursor = len(e.value)
		}
		return
	}
	off = 0
	for i := 0; i < target; i++ {
		off += len(lines[i]) + 1
	}
	if col > len(lines[target]) {
		col = len(lines[target])
	}
	e.cursor = off + col
}

func (e *Editor) insert(s string) {
	if e.cursor >= len(e.value) {
		e.value += s
		e.cursor = len(e.value)
		return
	}
	e.value = e.value[:e.cursor] + s + e.value[e.cursor:]
	e.cursor += len(s)
}

func (e *Editor) backspace() {
	if e.cursor == 0 {
		return
	}
	_, size := utf8.DecodeLastRuneInString(e.value[:e.cursor])
	killed := e.value[e.cursor-size : e.cursor]
	e.pushUndo()
	e.value = e.value[:e.cursor-size] + e.value[e.cursor:]
	e.cursor -= size
	e.kill.Push(killed, true, e.accumKill)
	e.accumKill = true
}

// DeleteForward удаляет один символ под курсором (ctrl+d / Delete).
func (e *Editor) DeleteForward() { e.deleteForward() }

func (e *Editor) deleteForward() {
	if e.cursor >= len(e.value) {
		return
	}
	_, size := utf8.DecodeRuneInString(e.value[e.cursor:])
	killed := e.value[e.cursor : e.cursor+size]
	e.pushUndo()
	e.value = e.value[:e.cursor] + e.value[e.cursor+size:]
	e.kill.Push(killed, false, e.accumKill)
	e.accumKill = true
}

func (e *Editor) killLine(toStart bool) {
	if toStart {
		if e.cursor == 0 {
			return
		}
		killed := e.value[:e.cursor]
		e.pushUndo()
		e.value = e.value[e.cursor:]
		e.cursor = 0
		e.kill.Push(killed, true, false)
		e.accumKill = false
		return
	}
	if e.cursor >= len(e.value) {
		return
	}
	killed := e.value[e.cursor:]
	e.pushUndo()
	e.value = e.value[:e.cursor]
	e.kill.Push(killed, false, false)
	e.accumKill = false
}

func (e *Editor) move(delta int) {
	if delta < 0 {
		if e.cursor == 0 {
			return
		}
		_, size := utf8.DecodeLastRuneInString(e.value[:e.cursor])
		e.cursor -= size
		return
	}
	if e.cursor >= len(e.value) {
		return
	}
	_, size := utf8.DecodeRuneInString(e.value[e.cursor:])
	e.cursor += size
}

func (e *Editor) moveWord(dir int) {
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }
	if dir < 0 {
		for e.cursor > 0 {
			r, size := utf8.DecodeLastRuneInString(e.value[:e.cursor])
			if !isSpace(r) {
				break
			}
			e.cursor -= size
		}
		for e.cursor > 0 {
			r, size := utf8.DecodeLastRuneInString(e.value[:e.cursor])
			if isSpace(r) {
				break
			}
			e.cursor -= size
		}
		return
	}
	for e.cursor < len(e.value) {
		r, size := utf8.DecodeRuneInString(e.value[e.cursor:])
		if !isSpace(r) {
			break
		}
		e.cursor += size
	}
	for e.cursor < len(e.value) {
		r, size := utf8.DecodeRuneInString(e.value[e.cursor:])
		if isSpace(r) {
			break
		}
		e.cursor += size
	}
}

func (e *Editor) killWord(dir int) {
	start := e.cursor
	e.moveWord(dir)
	end := e.cursor
	if start == end {
		return
	}
	if start > end {
		start, end = end, start
	}
	killed := e.value[start:end]
	e.pushUndo()
	e.value = e.value[:start] + e.value[end:]
	e.cursor = start
	e.kill.Push(killed, dir < 0, e.accumKill)
	e.accumKill = true
}

func (e *Editor) jumpToChar(char, mode string) {
	if char == "" {
		return
	}
	forward := mode != "backward"
	lines := strings.Split(e.value, "\n")
	off := 0
	lineIdx := 0
	col := 0
	for i, line := range lines {
		end := off + len(line)
		if e.cursor <= end || i == len(lines)-1 {
			lineIdx = i
			col = e.cursor - off
			break
		}
		off = end + 1
	}
	if forward {
		for li := lineIdx; li < len(lines); li++ {
			line := lines[li]
			start := 0
			if li == lineIdx {
				start = col + 1
			}
			if idx := strings.Index(line[start:], char); idx >= 0 {
				e.cursor = offForLine(lines, li) + start + idx
				e.accumKill = false
				return
			}
		}
		return
	}
	for li := lineIdx; li >= 0; li-- {
		line := lines[li]
		end := len(line)
		if li == lineIdx {
			end = col - 1
		}
		if end < 0 {
			continue
		}
		if idx := strings.LastIndex(line[:end+1], char); idx >= 0 {
			e.cursor = offForLine(lines, li) + idx
			e.accumKill = false
			return
		}
	}
}

func offForLine(lines []string, lineIdx int) int {
	off := 0
	for i := 0; i < lineIdx; i++ {
		off += len(lines[i]) + 1
	}
	return off
}

func (e *Editor) Render(width int) []string {
	innerW := width - 2
	if !e.boxBorder {
		innerW = width
	}
	if innerW < 8 {
		innerW = 8
	}
	body := e.value
	if body == "" && e.placeholder != "" && !e.focused {
		body = e.placeholder
	}
	display := body
	if e.focused {
		marker := "█"
		if e.showHardwareCursor {
			marker = wrap.CursorMarker
		}
		if e.cursor >= len(e.value) {
			display = e.value + marker
		} else {
			display = e.value[:e.cursor] + marker + e.value[e.cursor:]
		}
	}
	padInner := innerW - e.paddingX
	if padInner < 8 {
		padInner = 8
	}
	lines := wrap.WrapLines(wrap.SplitParas(display), padInner)
	if len(lines) == 0 {
		lines = []string{""}
	}
	if e.paddingX > 0 {
		pad := strings.Repeat(" ", e.paddingX)
		for i := range lines {
			lines[i] = pad + lines[i]
		}
	}
	if !e.boxBorder {
		border := "─"
		if width > 2 {
			border = strings.Repeat("─", width)
		}
		top, bot := border, border
		if e.borderColor != "" {
			top = e.borderColor + border + "\x1b[0m"
			bot = top
		}
		out := []string{top}
		out = append(out, lines...)
		out = append(out, bot)
		return out
	}
	top := "┌" + strings.Repeat("─", innerW) + "┐"
	bot := "└" + strings.Repeat("─", innerW) + "┘"
	if e.borderColor != "" {
		top = e.borderColor + top + "\x1b[0m"
		bot = e.borderColor + bot + "\x1b[0m"
	}
	out := []string{top}
	for _, line := range lines {
		vis := runeLen(wrap.StripANSI(line))
		pad := innerW - vis
		if pad < 0 {
			pad = 0
		}
		mid := "│" + line + strings.Repeat(" ", pad) + "│"
		if e.borderColor != "" {
			mid = e.borderColor + "│\x1b[0m" + line + strings.Repeat(" ", pad) + e.borderColor + "│\x1b[0m"
		}
		out = append(out, mid)
	}
	out = append(out, bot)
	return out
}

func runeLen(s string) int { return len([]rune(s)) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d [20]byte
	i := len(d)
	for n > 0 {
		i--
		d[i] = byte('0' + n%10)
		n /= 10
	}
	return string(d[i:])
}
