// Package overlay — раскладка и композитинг оверлеев TUI.
package overlay

import (
	"strconv"
	"strings"

	"github.com/stelmakhdigital/stell-tui/wrap"
)

// OverlayAnchor задаёт размещение оверлея в терминале.
type OverlayAnchor int

const (
	OverlayAnchorTop OverlayAnchor = iota
	OverlayAnchorCenter
	OverlayAnchorBottom
	OverlayAnchorTopLeft
	OverlayAnchorTopRight
	OverlayAnchorBottomLeft
	OverlayAnchorBottomRight
	OverlayAnchorTopCenter
	OverlayAnchorBottomCenter
	OverlayAnchorLeftCenter
	OverlayAnchorRightCenter
)

// OverlayMargin — отступы от краёв терминала при позиционировании оверлея.
type OverlayMargin struct {
	Top, Right, Bottom, Left int
}

// OverlayOptions — конфигурация раскладки оверлея.
// Поля Anchor и MaxHeightPct сохраняют прежнее поведение CompositeOverlayLines.
type OverlayOptions struct {
	Anchor       OverlayAnchor
	MaxHeightPct int // 1–100; 0 = без процентного лимита

	Width    any
	MinWidth int
	MaxHeight any

	OffsetX int
	OffsetY int

	Row any
	Col any

	Margin OverlayMargin

	Visible func(termWidth, termHeight int) bool

	NonCapturing bool
}

// ClampOverlayLines ограничивает число строк оверлея процентом высоты терминала.
func ClampOverlayLines(lines []string, maxPct, termHeight int) []string {
	if maxPct <= 0 || maxPct > 100 || termHeight <= 0 || len(lines) == 0 {
		return lines
	}
	maxRows := termHeight * maxPct / 100
	if maxRows < 1 {
		maxRows = 1
	}
	if len(lines) <= maxRows {
		return lines
	}
	return lines[:maxRows]
}

// CompositeOverlayLines размещает строки оверлея относительно чата (anchor + max height).
func CompositeOverlayLines(chatLines, overlayLines []string, opts OverlayOptions, termHeight int) []string {
	overlayLines = ClampOverlayLines(overlayLines, opts.MaxHeightPct, termHeight)
	switch opts.Anchor {
	case OverlayAnchorTop, OverlayAnchorTopLeft, OverlayAnchorTopRight, OverlayAnchorTopCenter:
		out := append([]string(nil), overlayLines...)
		return append(out, chatLines...)
	case OverlayAnchorBottom, OverlayAnchorBottomLeft, OverlayAnchorBottomRight, OverlayAnchorBottomCenter:
		out := append([]string(nil), chatLines...)
		return append(out, overlayLines...)
	case OverlayAnchorCenter, OverlayAnchorLeftCenter, OverlayAnchorRightCenter:
		total := termHeight
		if total <= 0 {
			total = len(chatLines) + len(overlayLines)
		}
		topPad := (total - len(overlayLines)) / 2
		if topPad < 0 {
			topPad = 0
		}
		out := make([]string, 0, total)
		for i := 0; i < topPad; i++ {
			out = append(out, "")
		}
		out = append(out, overlayLines...)
		for len(out) < total {
			out = append(out, "")
		}
		if len(out) > total {
			out = out[:total]
		}
		return out
	default:
		return overlayLines
	}
}

func resolveDim(v any, total int) (int, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case int:
		if x <= 0 {
			return 0, false
		}
		return x, true
	case string:
		s := strings.TrimSpace(x)
		if strings.HasSuffix(s, "%") {
			n, err := strconv.Atoi(strings.TrimSuffix(s, "%"))
			if err != nil || n <= 0 {
				return 0, false
			}
			if total <= 0 {
				return 0, false
			}
			return total * n / 100, true
		}
		n, err := strconv.Atoi(s)
		if err != nil || n <= 0 {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

// RenderWidth возвращает ширину рендера оверлея для данного терминала.
func RenderWidth(termWidth int, opts OverlayOptions) int {
	if w, ok := resolveDim(opts.Width, termWidth); ok {
		if opts.MinWidth > 0 && w < opts.MinWidth {
			w = opts.MinWidth
		}
		if w > termWidth {
			w = termWidth
		}
		if w < 1 {
			w = 1
		}
		return w
	}
	if opts.MinWidth > 0 && opts.MinWidth < termWidth {
		return opts.MinWidth
	}
	return termWidth
}

func clampOverlayByOpts(lines []string, opts OverlayOptions, termHeight int) []string {
	lines = ClampOverlayLines(lines, opts.MaxHeightPct, termHeight)
	if maxH, ok := resolveDim(opts.MaxHeight, termHeight); ok && maxH > 0 && len(lines) > maxH {
		return lines[:maxH]
	}
	return lines
}

// CompositeFrame вкладывает оверлей в базовый кадр по OverlayOptions.
func CompositeFrame(base, overlayLines []string, opts OverlayOptions, termWidth, termHeight int) []string {
	if opts.Visible != nil && !opts.Visible(termWidth, termHeight) {
		return base
	}
	overlayLines = clampOverlayByOpts(overlayLines, opts, termHeight)
	if len(overlayLines) == 0 {
		return base
	}

	if termHeight <= 0 {
		termHeight = len(base)
		if termHeight < len(overlayLines) {
			termHeight = len(overlayLines)
		}
	}

	out := make([]string, termHeight)
	for i := range out {
		if i < len(base) {
			out[i] = base[i]
		} else {
			out[i] = ""
		}
	}

	ow := 0
	for _, line := range overlayLines {
		if n := wrap.VisibleLen(line); n > ow {
			ow = n
		}
	}
	if w, ok := resolveDim(opts.Width, termWidth); ok {
		ow = w
	}
	if opts.MinWidth > 0 && ow < opts.MinWidth {
		ow = opts.MinWidth
	}
	oh := len(overlayLines)

	row, col := resolveOverlayPosition(opts, termWidth, termHeight, ow, oh)

	for i, line := range overlayLines {
		r := row + i
		if r < 0 || r >= termHeight {
			continue
		}
		out[r] = overlayLineAt(out[r], line, col, termWidth)
	}
	return out
}

func resolveOverlayPosition(opts OverlayOptions, termW, termH, ow, oh int) (row, col int) {
	if r, ok := resolveDim(opts.Row, termH); ok {
		row = r
	} else {
		switch opts.Anchor {
		case OverlayAnchorTop, OverlayAnchorTopLeft, OverlayAnchorTopRight, OverlayAnchorTopCenter:
			row = 0
		case OverlayAnchorBottom, OverlayAnchorBottomLeft, OverlayAnchorBottomRight, OverlayAnchorBottomCenter:
			row = termH - oh
		default:
			row = (termH - oh) / 2
		}
	}
	if c, ok := resolveDim(opts.Col, termW); ok {
		col = c
	} else {
		switch opts.Anchor {
		case OverlayAnchorTopLeft, OverlayAnchorBottomLeft, OverlayAnchorLeftCenter:
			col = 0
		case OverlayAnchorTopRight, OverlayAnchorBottomRight, OverlayAnchorRightCenter:
			col = termW - ow
		default:
			col = (termW - ow) / 2
		}
	}
	row += opts.OffsetY
	col += opts.OffsetX

	m := opts.Margin
	if row < m.Top {
		row = m.Top
	}
	if col < m.Left {
		col = m.Left
	}
	if row+oh > termH-m.Bottom && termH-m.Bottom-oh >= m.Top {
		row = termH - m.Bottom - oh
	}
	if col+ow > termW-m.Right && termW-m.Right-ow >= m.Left {
		col = termW - m.Right - ow
	}
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	return row, col
}

func overlayLineAt(base, overlayStr string, col, termWidth int) string {
	if col <= 0 {
		line := overlayStr
		if wrap.VisibleLen(line) > termWidth {
			line = wrap.Truncate(overlayStr, termWidth)
		}
		return wrap.PadRight(line, termWidth)
	}
	prefix := wrap.PadRight(wrap.Truncate(base, col), col)
	restW := termWidth - col
	if restW < 1 {
		return wrap.Truncate(prefix, termWidth)
	}
	body := wrap.Truncate(overlayStr, restW)
	return wrap.PadRight(prefix+body, termWidth)
}
