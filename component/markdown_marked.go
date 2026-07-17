package component

import (
	"stell/tui/wrap"
	"strconv"
	"strings"
)

// Виды block-токенов markdown (подмножество marked-стиля).
type mdBlockKind int

const (
	mdHeading mdBlockKind = iota
	mdParagraph
	mdCode
	mdList
	mdTable
	mdQuote
	mdHR
	mdSpace
)

type mdBlock struct {
	kind     mdBlockKind
	level    int      // heading level
	lang     string   // code fence language
	text     string   // paragraph / heading / quote
	lines    []string // code lines (without fences)
	items    []string // list item bodies
	ordered  bool
	headers  []string
	rows     [][]string
	rawTable string // fallback when too narrow
}

// parseMarkdownBlocks токенизирует markdown в блоки (подмножество marked).
func parseMarkdownBlocks(src string) []mdBlock {
	lines := wrap.SplitParas(src)
	var tokens []mdBlock
	i := 0
	for i < len(lines) {
		line := lines[i]
		trim := strings.TrimSpace(line)

		if trim == "" {
			tokens = append(tokens, mdBlock{kind: mdSpace})
			i++
			continue
		}

		if fence, ok := fenceMarker(trim); ok {
			lang := strings.TrimSpace(trim[len(fence):])
			i++
			var body []string
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				if strings.HasPrefix(t, fence) && (len(t) == len(fence) || !isFenceChar(t[len(fence)])) {
					i++
					break
				}
				// Частично закрытый fence при стриминге.
				if isPartialClosingFence(t, fence) {
					i++
					break
				}
				body = append(body, lines[i])
				i++
			}
			tokens = append(tokens, mdBlock{kind: mdCode, lang: lang, lines: body})
			continue
		}

		if end, _ := lookAheadPipeTable(lines, i); end > 0 {
			start := i
			rawRows := lines[i:end]
			i = end
			tok := parseTableToken(rawRows)
			tok.rawTable = strings.Join(lines[start:end], "\n")
			tokens = append(tokens, tok)
			continue
		}

		if strings.HasPrefix(trim, "#") {
			level := 0
			for level < len(trim) && trim[level] == '#' {
				level++
			}
			text := strings.TrimSpace(trim[level:])
			tokens = append(tokens, mdBlock{kind: mdHeading, level: level, text: text})
			i++
			continue
		}

		if strings.HasPrefix(trim, ">") {
			var qlines []string
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(t, ">") {
					break
				}
				qlines = append(qlines, strings.TrimSpace(strings.TrimPrefix(t, ">")))
				i++
			}
			tokens = append(tokens, mdBlock{kind: mdQuote, text: strings.Join(qlines, "\n")})
			continue
		}

		if trim == "---" || trim == "***" || trim == "___" {
			tokens = append(tokens, mdBlock{kind: mdHR})
			i++
			continue
		}

		if strings.HasPrefix(trim, "- ") || strings.HasPrefix(trim, "* ") || orderedListPrefix(line) != "" {
			ordered := orderedListPrefix(line) != ""
			var items []string
			for i < len(lines) {
				l := lines[i]
				t := strings.TrimSpace(l)
				if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") {
					items = append(items, t[2:])
					i++
					continue
				}
				if pref := orderedListPrefix(l); pref != "" {
					items = append(items, strings.TrimSpace(strings.TrimPrefix(l, pref)))
					i++
					continue
				}
				break
			}
			tokens = append(tokens, mdBlock{kind: mdList, items: items, ordered: ordered})
			continue
		}

		var paras []string
		for i < len(lines) {
			l := lines[i]
			t := strings.TrimSpace(l)
			if t == "" {
				break
			}
			if _, ok := fenceMarker(t); ok {
				break
			}
			if end, _ := lookAheadPipeTable(lines, i); end > 0 {
				break
			}
			if strings.HasPrefix(t, "#") || strings.HasPrefix(t, ">") ||
				t == "---" || t == "***" || t == "___" ||
				strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") ||
				orderedListPrefix(l) != "" {
				break
			}
			paras = append(paras, l)
			i++
		}
		tokens = append(tokens, mdBlock{kind: mdParagraph, text: strings.Join(paras, "\n")})
	}
	return tokens
}

func fenceMarker(trim string) (string, bool) {
	for _, ch := range []byte{'`', '~'} {
		n := 0
		for n < len(trim) && trim[n] == ch {
			n++
		}
		if n >= 3 {
			return trim[:n], true
		}
	}
	return "", false
}

func isFenceChar(b byte) bool {
	return b == '`' || b == '~'
}

func isPartialClosingFence(trim, marker string) bool {
	if trim == "" || len(trim) >= len(marker) {
		return false
	}
	ch := marker[0]
	for i := 0; i < len(trim); i++ {
		if trim[i] != ch {
			return false
		}
	}
	return true
}

func parseTableToken(rawRows []string) mdBlock {
	var headers []string
	var rows [][]string
	sepSeen := false
	for _, row := range rawRows {
		if isTableSep(row) {
			sepSeen = true
			continue
		}
		cells := splitTableCells(row)
		if !sepSeen && headers == nil {
			headers = cells
			continue
		}
		rows = append(rows, cells)
	}
	return mdBlock{kind: mdTable, headers: headers, rows: rows}
}

func renderCodeLine(line string, th MarkdownTheme, m *Markdown) string {
	if m.HighlightCode != nil {
		return m.HighlightCode(line)
	}
	if th.CodeBlock != "" {
		return th.CodeBlock + line + th.Reset
	}
	return line
}

func isMarkdownFenceLang(lang string) bool {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "", "markdown", "md", "gfm":
		return true
	default:
		return false
	}
}

// splitFencePipeTable finds a GFM/pipe table inside a markdown-ish fence.
func splitFencePipeTable(lines []string) (before []string, table mdBlock, after []string, ok bool) {
	for i := 0; i < len(lines); i++ {
		end, _ := lookAheadPipeTable(lines, i)
		if end < 0 {
			continue
		}
		before = lines[:i]
		table = parseTableToken(lines[i:end])
		table.rawTable = strings.Join(lines[i:end], "\n")
		after = lines[end:]
		// Require table to be the main content (not a tiny fragment in a large code sample).
		tableLines := end - i
		nonEmpty := 0
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				nonEmpty++
			}
		}
		if nonEmpty > 0 && tableLines*2 >= nonEmpty {
			return before, table, after, true
		}
	}
	return nil, mdBlock{}, nil, false
}

func renderMarkdownTokens(tokens []mdBlock, th MarkdownTheme, m *Markdown, width int) (out []string, noWrap map[int]bool) {
	noWrap = map[int]bool{}
	markNoWrap := func() { noWrap[len(out)] = true }

	for ti, tok := range tokens {
		nextKind := mdSpace
		if ti+1 < len(tokens) {
			nextKind = tokens[ti+1].kind
		}
		switch tok.kind {
		case mdSpace:
			if len(out) > 0 && out[len(out)-1] != "" {
				out = append(out, "")
			}
		case mdHeading:
			prefix := strings.Repeat("#", max(1, tok.level)) + " "
			out = append(out, th.Heading+prefix+colorInline(tok.text, th, m.Hyperlinks)+th.Reset)
		case mdHR:
			out = append(out, th.HR+strings.Repeat("─", max(1, width))+th.Reset)
		case mdQuote:
			for _, ql := range strings.Split(tok.text, "\n") {
				out = append(out, th.QuoteBorder+"│ "+th.Quote+colorInline(ql, th, m.Hyperlinks)+th.Reset)
			}
		case mdCode:
			if isMarkdownFenceLang(tok.lang) {
				if before, table, after, ok := splitFencePipeTable(tok.lines); ok {
					for _, line := range before {
						out = append(out, renderCodeLine(line, th, m))
						markNoWrap()
					}
					tableLines := renderMarkedTable(table, th, m, width)
					for _, row := range tableLines {
						out = append(out, row)
						markNoWrap()
					}
					for _, line := range after {
						out = append(out, renderCodeLine(line, th, m))
						markNoWrap()
					}
					if nextKind != mdSpace && nextKind != mdCode {
						out = append(out, "")
					}
					break
				}
			}
			for _, line := range tok.lines {
				out = append(out, renderCodeLine(line, th, m))
				markNoWrap()
			}
			if nextKind != mdSpace && nextKind != mdCode {
				out = append(out, "")
			}
		case mdList:
			for i, item := range tok.items {
				bullet := "• "
				if tok.ordered {
					bullet = strconv.Itoa(i+1) + ". "
				}
				out = append(out, th.ListBullet+bullet+th.Reset+colorInline(item, th, m.Hyperlinks))
			}
		case mdParagraph:
			for _, pl := range strings.Split(tok.text, "\n") {
				out = append(out, colorInline(pl, th, m.Hyperlinks))
			}
		case mdTable:
			tableLines := renderMarkedTable(tok, th, m, width)
			for _, row := range tableLines {
				out = append(out, row)
				markNoWrap()
			}
			if nextKind != mdSpace && len(tableLines) > 0 {
				out = append(out, "")
			}
		}
	}
	return out, noWrap
}

// renderMarkedTable рисует GFM-таблицы с box-рамкой и переносом ячеек.
func renderMarkedTable(tok mdBlock, th MarkdownTheme, m *Markdown, availableWidth int) []string {
	numCols := len(tok.headers)
	for _, row := range tok.rows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}
	if numCols == 0 {
		return nil
	}

	styleCell := func(s string) string {
		return colorInline(s, th, m.Hyperlinks)
	}

	headerTexts := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		if i < len(tok.headers) {
			headerTexts[i] = styleCell(tok.headers[i])
		}
	}
	rowTexts := make([][]string, len(tok.rows))
	for r, row := range tok.rows {
		rowTexts[r] = make([]string, numCols)
		for i := 0; i < numCols; i++ {
			if i < len(row) {
				rowTexts[r][i] = styleCell(row[i])
			}
		}
	}

	borderOverhead := 3*numCols + 1
	availableForCells := availableWidth - borderOverhead
	if availableWidth > 0 && availableForCells < numCols {
		if tok.rawTable != "" {
			return wrap.WrapLine(tok.rawTable, availableWidth)
		}
		return nil
	}

	natural := make([]int, numCols)
	for i, t := range headerTexts {
		natural[i] = wrap.VisibleLen(t)
	}
	for _, row := range rowTexts {
		for i, t := range row {
			if w := wrap.VisibleLen(t); w > natural[i] {
				natural[i] = w
			}
		}
	}

	columnWidths := append([]int(nil), natural...)
	if availableWidth > 0 && availableForCells > 0 {
		total := 0
		for _, w := range columnWidths {
			total += w
		}
		for total > availableForCells {
			maxi := 0
			for j := 1; j < numCols; j++ {
				if columnWidths[j] > columnWidths[maxi] {
					maxi = j
				}
			}
			if columnWidths[maxi] <= 1 {
				break
			}
			columnWidths[maxi]--
			total--
		}
	}

	wrapCells := func(cells []string) [][]string {
		out := make([][]string, numCols)
		for i := 0; i < numCols; i++ {
			text := ""
			if i < len(cells) {
				text = cells[i]
			}
			out[i] = wrap.WrapLine(text, max(1, columnWidths[i]))
		}
		return out
	}

	var lines []string
	topCells := make([]string, numCols)
	for i, w := range columnWidths {
		topCells[i] = strings.Repeat("─", w)
	}
	lines = append(lines, "┌─"+strings.Join(topCells, "─┬─")+"─┐")

	headerWrapped := wrapCells(headerTexts)
	headerH := 1
	for _, cellLines := range headerWrapped {
		if len(cellLines) > headerH {
			headerH = len(cellLines)
		}
	}
	for lineIdx := 0; lineIdx < headerH; lineIdx++ {
		parts := make([]string, numCols)
		for col := 0; col < numCols; col++ {
			text := ""
			if lineIdx < len(headerWrapped[col]) {
				text = headerWrapped[col][lineIdx]
			}
			parts[col] = th.Bold + wrap.PadVisible(text, columnWidths[col]) + th.Reset
		}
		lines = append(lines, "│ "+strings.Join(parts, " │ ")+" │")
	}

	sepCells := make([]string, numCols)
	for i, w := range columnWidths {
		sepCells[i] = strings.Repeat("─", w)
	}
	separator := "├─" + strings.Join(sepCells, "─┼─") + "─┤"
	lines = append(lines, th.HR+separator+th.Reset)

	for ri, row := range rowTexts {
		wrapped := wrapCells(row)
		rowH := 1
		for _, cellLines := range wrapped {
			if len(cellLines) > rowH {
				rowH = len(cellLines)
			}
		}
		for lineIdx := 0; lineIdx < rowH; lineIdx++ {
			parts := make([]string, numCols)
			for col := 0; col < numCols; col++ {
				text := ""
				if lineIdx < len(wrapped[col]) {
					text = wrapped[col][lineIdx]
				}
				parts[col] = wrap.PadVisible(text, columnWidths[col])
			}
			lines = append(lines, "│ "+strings.Join(parts, " │ ")+" │")
		}
		if ri < len(rowTexts)-1 {
			lines = append(lines, th.HR+separator+th.Reset)
		}
	}

	botCells := make([]string, numCols)
	for i, w := range columnWidths {
		botCells[i] = strings.Repeat("─", w)
	}
	lines = append(lines, "└─"+strings.Join(botCells, "─┴─")+"─┘")
	return lines
}
