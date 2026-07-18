package component

import (
	"github.com/stelmakhdigital/stell-tui/terminal"
	"github.com/stelmakhdigital/stell-tui/wrap"
	"fmt"
	"regexp"
	"strings"
)

// MarkdownTheme — цвета рендера markdown (ANSI-последовательности).
type MarkdownTheme struct {
	Heading, Link, LinkURL, Code, CodeBlock, CodeBlockBorder string
	Quote, QuoteBorder, HR, ListBullet, Strike, Bold, Italic string
	Reset                                                    string
}

// DefaultMarkdownTheme возвращает тему markdown по умолчанию (ANSI).
func DefaultMarkdownTheme() MarkdownTheme {
	return MarkdownTheme{
		Heading:         "\x1b[33m",
		Link:            "\x1b[36m",
		LinkURL:         "\x1b[90m",
		Code:            "\x1b[36m",
		CodeBlock:       "",
		CodeBlockBorder: "\x1b[90m",
		Quote:           "\x1b[90m",
		QuoteBorder:     "\x1b[90m",
		HR:              "\x1b[90m",
		ListBullet:      "\x1b[36m",
		Strike:          "\x1b[9m",
		Bold:            "\x1b[1m",
		Italic:          "\x1b[3m",
		Reset:           "\x1b[0m",
	}
}

// Markdown рендерит подмножество Markdown с цветами темы.
type Markdown struct {
	Source        string
	Theme         MarkdownTheme
	Hyperlinks    bool // OSC-8 при true
	HighlightCode func(line string) string
	// HideFences: true — скрывать строки ограждений ``` (по умолчанию для layout).
	HideFences bool
	cache      []string
	width      int
}

func NewMarkdown(source string, theme MarkdownTheme) *Markdown {
	return &Markdown{
		Source:     source,
		Theme:      theme,
		Hyperlinks: terminal.DetectCapabilities().Hyperlinks,
		HideFences: true,
	}
}

func (m *Markdown) Invalidate() {
	m.cache = nil
}

func (m *Markdown) SetSource(s string) {
	m.Source = s
	m.cache = nil
}

func (m *Markdown) Render(width int) []string {
	if m.cache != nil && m.width == width {
		return m.cache
	}
	m.width = width
	th := m.Theme
	if th.Reset == "" {
		th = DefaultMarkdownTheme()
	}
	tokens := parseMarkdownBlocks(m.Source)
	out, noWrap := renderMarkdownTokens(tokens, th, m, width)
	out = wrapLinesSelective(out, width, noWrap)
	m.cache = out
	return out
}

func wrapLinesSelective(lines []string, width int, noWrap map[int]bool) []string {
	if width <= 0 {
		return append([]string{}, lines...)
	}
	var out []string
	for i, line := range lines {
		if noWrap[i] {
			// Overflow: truncate rather than mid-line wrap (keeps ASCII/tables intact).
			if wrap.VisibleLen(line) > width {
				out = append(out, wrap.Truncate(line, width))
			} else {
				out = append(out, line)
			}
			continue
		}
		out = append(out, wrap.WrapLine(line, width)...)
	}
	return out
}

var (
	reStrike = regexp.MustCompile(`~~([^~]+)~~`)
	reLink   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reBold   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reCode   = regexp.MustCompile("`([^`]+)`")
)

func colorInline(line string, th MarkdownTheme, hyperlinks bool) string {
	if th.Strike == "" {
		th.Strike = "\x1b[9m"
	}
	if th.Bold == "" {
		th.Bold = "\x1b[1m"
	}
	if th.Italic == "" {
		th.Italic = "\x1b[3m"
	}
	var held []string
	hold := func(s string) string {
		id := fmt.Sprintf("\x00MD%d\x00", len(held))
		held = append(held, s)
		return id
	}
	line = reCode.ReplaceAllStringFunc(line, func(m string) string {
		inner := m[1 : len(m)-1]
		return hold(th.Code + inner + th.Reset)
	})
	line = reLink.ReplaceAllStringFunc(line, func(m string) string {
		sub := reLink.FindStringSubmatch(m)
		if len(sub) != 3 {
			return m
		}
		text, url := sub[1], sub[2]
		if hyperlinks {
			return hold("\x1b]8;;" + url + "\x1b\\" + th.Link + text + th.Reset + "\x1b]8;;\x1b\\")
		}
		return hold(th.Link + text + th.Reset + th.LinkURL + " (" + url + ")" + th.Reset)
	})
	line = reBold.ReplaceAllStringFunc(line, func(m string) string {
		sub := reBold.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		return hold(th.Bold + sub[1] + th.Reset)
	})
	line = reStrike.ReplaceAllStringFunc(line, func(m string) string {
		sub := reStrike.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		return hold(th.Strike + sub[1] + th.Reset)
	})
	for {
		i := strings.Index(line, "*")
		if i < 0 {
			break
		}
		j := strings.Index(line[i+1:], "*")
		if j < 0 {
			break
		}
		j += i + 1
		inner := line[i+1 : j]
		if inner == "" || strings.Contains(inner, "\x00") {
			break
		}
		repl := hold(th.Italic + inner + th.Reset)
		line = line[:i] + repl + line[j+1:]
	}
	for i, s := range held {
		line = strings.Replace(line, fmt.Sprintf("\x00MD%d\x00", i), s, 1)
	}
	return line
}

func isTableRow(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, "|") && strings.HasSuffix(t, "|") && strings.Count(t, "|") >= 2
}

// isTableSepChar reports dash-like chars LLMs emit in GFM separator rows.
func isTableSepChar(r rune) bool {
	switch r {
	case '-', ':', ' ', '+',
		'\u2014', // —
		'\u2013', // –
		'\u2212', // −
		'\u2500', // ─
		'\u2011': // ‑
		return true
	default:
		return false
	}
}

func isTableSep(line string) bool {
	t := strings.TrimSpace(line)
	if !isTableRow(t) {
		return false
	}
	hasDash := false
	for _, cell := range strings.Split(strings.Trim(t, "|"), "|") {
		c := strings.TrimSpace(cell)
		if c == "" {
			continue
		}
		for _, r := range c {
			if !isTableSepChar(r) {
				return false
			}
			if r != ':' && r != ' ' {
				hasDash = true
			}
		}
	}
	return hasDash
}

// lookAheadPipeTable returns (endExclusive, hasSep) for a run of ≥2 pipe rows
// starting at i, or (-1, false) if this is not a table block.
func lookAheadPipeTable(lines []string, i int) (end int, hasSep bool) {
	if i >= len(lines) || !isTableRow(lines[i]) {
		return -1, false
	}
	j := i
	for j < len(lines) && isTableRow(lines[j]) {
		if isTableSep(lines[j]) {
			hasSep = true
		}
		j++
	}
	if j-i < 2 {
		return -1, false
	}
	return j, hasSep
}

func splitTableCells(row string) []string {
	cells := strings.Split(strings.Trim(strings.TrimSpace(row), "|"), "|")
	for j := range cells {
		cells[j] = strings.TrimSpace(cells[j])
	}
	return cells
}

func orderedListPrefix(line string) string {
	for i, r := range line {
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '.' && i > 0 && i+1 < len(line) && line[i+1] == ' ' {
			return line[:i+2]
		}
		return ""
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
