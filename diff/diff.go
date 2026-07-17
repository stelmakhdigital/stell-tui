package diff

import (
	"fmt"
	"io"
	"strings"
)

const (
	syncOutputOn  = "\x1b[?2026h"
	syncOutputOff = "\x1b[?2026l"
	clearLine     = "\x1b[2K"
	cursorHome    = "\x1b[H"
	hideCursor    = "\x1b[?25l"
	showCursor    = "\x1b[?25h"
	altScreenOn   = "\x1b[?1049h"
	altScreenOff  = "\x1b[?1049l"
)

// DiffEngine обновляет терминал дифференциальным рендерингом.
type DiffEngine struct {
	prev              []string
	width             int
	height            int
	first             bool
	altScreen         bool
	out               io.Writer
	strategy          DiffStrategy
	showHardwareCursor bool
	prevViewportTop   int
}

func NewDiffEngine(w io.Writer, altScreen bool) *DiffEngine {
	return &DiffEngine{out: w, first: true, altScreen: altScreen, strategy: DiffPatch}
}

// SetStrategy выбирает режим DiffFull / DiffPatch / DiffScroll.
func (d *DiffEngine) SetStrategy(s DiffStrategy) {
	if d == nil {
		return
	}
	d.strategy = s
	d.first = true
	d.prev = nil
}

// Strategy возвращает текущую стратегию дифференциального рендера.
func (d *DiffEngine) Strategy() DiffStrategy {
	if d == nil {
		return DiffFull
	}
	return d.strategy
}

// SetShowHardwareCursor включает извлечение CursorMarker и CSI-позиционирование.
func (d *DiffEngine) SetShowHardwareCursor(on bool) {
	if d == nil {
		return
	}
	d.showHardwareCursor = on
}

func (d *DiffEngine) Resize(width, height int) {
	if width != d.width || height != d.height {
		d.width = width
		d.height = height
		d.first = true
		d.prev = nil
	}
}

// Invalidate заставляет следующий Render полностью перерисовать кадр.
func (d *DiffEngine) Invalidate() {
	d.first = true
	d.prev = nil
	d.prevViewportTop = 0
}

// Render записывает строки полного кадра.
func (d *DiffEngine) Render(lines []string) error {
	viewportTop := 0
	if d.height > 0 && len(lines) > d.height {
		viewportTop = len(lines) - d.height
		lines = lines[viewportTop:]
	}
	hwRow, hwCol := -1, -1
	if d.showHardwareCursor {
		r, c, cleaned, found := extractCursorMarker(lines)
		lines = cleaned
		if found {
			hwRow, hwCol = r, c
		}
	}
	var b strings.Builder
	b.WriteString(syncOutputOn)
	strategy := d.strategy
	// DiffScroll: при выходе контента за viewport прокручиваем через \r\n.
	if strategy == DiffScroll && !d.first && d.prev != nil && viewportTop > d.prevViewportTop {
		scrollBy := viewportTop - d.prevViewportTop
		if scrollBy > 0 && scrollBy < d.height {
			b.WriteString(hideCursor)
			b.WriteString(strings.Repeat("\r\n", scrollBy))
			// Перезаписываем нижние scrollBy строк после прокрутки.
			start := len(lines) - scrollBy
			if start < 0 {
				start = 0
			}
			b.WriteString(fmt.Sprintf("\x1b[%d;1H", start+1))
			for i := start; i < len(lines); i++ {
				b.WriteString(clearLine)
				b.WriteString(lines[i])
				if i < len(lines)-1 {
					b.WriteByte('\n')
				}
			}
			b.WriteString(syncOutputOff)
			if hwRow >= 0 {
				b.WriteString(positionHardwareCursor(hwRow, hwCol))
			}
			d.prev = append([]string{}, lines...)
			d.prevViewportTop = viewportTop
			_, err := io.WriteString(d.out, b.String())
			return err
		}
	}
	if d.first || d.prev == nil || strategy == DiffFull {
		if d.altScreen {
			b.WriteString(altScreenOn)
		}
		b.WriteString(hideCursor)
		b.WriteString(cursorHome)
		for i, line := range lines {
			b.WriteString(clearLine)
			b.WriteString(line)
			if i < len(lines)-1 {
				b.WriteByte('\n')
			}
		}
		// Очищаем остаток предыдущего более высокого кадра.
		for i := len(lines); i < len(d.prev); i++ {
			b.WriteByte('\n')
			b.WriteString(clearLine)
		}
		d.first = false
	} else if strategy == DiffScroll {
		// Перезапись стабильной середины (prefix/suffix), если viewport не прокручивался.
		start, endNew, endPrev := 0, len(lines), len(d.prev)
		minLen := len(lines)
		if len(d.prev) < minLen {
			minLen = len(d.prev)
		}
		for start < minLen && lines[start] == d.prev[start] {
			start++
		}
		for endNew > start && endPrev > start && lines[endNew-1] == d.prev[endPrev-1] {
			endNew--
			endPrev--
		}
		if start == len(lines) && start == len(d.prev) {
			b.WriteString(syncOutputOff)
			if hwRow >= 0 {
				b.WriteString(positionHardwareCursor(hwRow, hwCol))
			}
			d.prevViewportTop = viewportTop
			_, err := io.WriteString(d.out, b.String())
			return err
		}
		b.WriteString(hideCursor)
		if start > 0 {
			b.WriteString(fmt.Sprintf("\x1b[%d;1H", start+1))
		} else {
			b.WriteString(cursorHome)
		}
		for i := start; i < endNew; i++ {
			b.WriteString(clearLine)
			b.WriteString(lines[i])
			if i < endNew-1 {
				b.WriteByte('\n')
			}
		}
		for i := endNew; i < endPrev; i++ {
			b.WriteByte('\n')
			b.WriteString(clearLine)
		}
	} else {
		// DiffPatch: перезапись с первой изменённой строки до конца.
		start := 0
		minLen := len(lines)
		if len(d.prev) < minLen {
			minLen = len(d.prev)
		}
		for start < minLen && lines[start] == d.prev[start] {
			start++
		}
		if start == len(lines) && start == len(d.prev) {
			b.WriteString(syncOutputOff)
			_, err := io.WriteString(d.out, b.String())
			return err
		}
		if start > 0 {
			b.WriteString(fmt.Sprintf("\x1b[%d;1H", start+1))
		} else {
			b.WriteString(cursorHome)
		}
		for i := start; i < len(lines); i++ {
			b.WriteString(clearLine)
			b.WriteString(lines[i])
			if i < len(lines)-1 {
				b.WriteByte('\n')
			}
		}
		for i := len(lines); i < len(d.prev); i++ {
			b.WriteByte('\n')
			b.WriteString(clearLine)
		}
	}
	b.WriteString(syncOutputOff)
	if hwRow >= 0 {
		b.WriteString(positionHardwareCursor(hwRow, hwCol))
	}
	d.prev = append([]string{}, lines...)
	d.prevViewportTop = viewportTop
	_, err := io.WriteString(d.out, b.String())
	return err
}

// Close восстанавливает курсор и при необходимости выходит из alt screen.
func (d *DiffEngine) Close() error {
	var b strings.Builder
	b.WriteString(showCursor)
	if d.altScreen {
		b.WriteString(altScreenOff)
	}
	_, err := io.WriteString(d.out, b.String())
	return err
}
