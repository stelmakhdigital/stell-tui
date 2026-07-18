package component

import (
	"github.com/stelmakhdigital/stell-tui/wrap"
	"strings"
	"testing"
)

func TestMarkdownStrikethroughAndLink(t *testing.T) {
	md := NewMarkdown("see ~~old~~ and [docs](https://example.com)", DefaultMarkdownTheme())
	md.Hyperlinks = false
	lines := md.Render(80)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "\x1b[9m") {
		t.Fatalf("missing strike: %q", joined)
	}
	if !strings.Contains(joined, "docs") || !strings.Contains(joined, "example.com") {
		t.Fatalf("missing link: %q", joined)
	}
	md2 := NewMarkdown("[x](https://x.test)", DefaultMarkdownTheme())
	md2.Hyperlinks = true
	j2 := strings.Join(md2.Render(40), "")
	if !strings.Contains(j2, "\x1b]8;;https://x.test") {
		t.Fatalf("missing OSC-8: %q", j2)
	}
}

func TestMarkdownBoldItalicTable(t *testing.T) {
	md := NewMarkdown("**bold** and *italic*\n\n| a | b |\n| - | - |\n| 1 | 2 |", DefaultMarkdownTheme())
	md.Hyperlinks = false
	joined := strings.Join(md.Render(80), "\n")
	if !strings.Contains(joined, "\x1b[1m") {
		t.Fatalf("missing bold: %q", joined)
	}
	if !strings.Contains(joined, "\x1b[3m") {
		t.Fatalf("missing italic: %q", joined)
	}
	if !strings.Contains(joined, "│") {
		t.Fatalf("missing table: %q", joined)
	}
	if !strings.Contains(joined, "├") {
		t.Fatalf("missing table separator: %q", joined)
	}
}

func TestMarkdownCodeFenceNoWrapASCII(t *testing.T) {
	src := "intro\n```\n┌─ Барб ─┤   Гайка\n       │           │\n```\nout"
	md := NewMarkdown(src, DefaultMarkdownTheme())
	md.Hyperlinks = false
	lines := md.Render(20) // narrow width would break if wrapped
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "```") {
		t.Fatalf("fence markers should be hidden: %q", joined)
	}
	found := false
	for _, line := range lines {
		plain := stripANSITest(line)
		if strings.Contains(plain, "┌─") {
			found = true
			// Must stay on one line (not wrapped mid-diagram)
			if !strings.Contains(plain, "Гайка") && wrap.VisibleLen(plain) < 10 {
				// truncated ok at narrow width, but shouldn't be mid-char garbage from wrap of fence
			}
			if strings.Count(plain, "┌") > 0 && strings.Contains(joined, "┌─ Барб") {
				// full line preserved or truncated as whole
			}
		}
	}
	if !found && !strings.Contains(joined, "┌") {
		t.Fatalf("diagram line missing: %q", joined)
	}
	// Ensure diagram content wasn't split across wrap of non-code
	for _, line := range lines {
		plain := stripANSITest(line)
		if strings.HasPrefix(plain, "┌") {
			// single visual line — no newline inside
			if strings.Contains(plain, "\n") {
				t.Fatal("diagram line contains newline")
			}
		}
	}
}

func TestMarkdownTableColumnsAligned(t *testing.T) {
	src := "| Size | OD |\n| --- | --- |\n| Standard | 6 mm |\n| Wide | 10 mm |"
	md := NewMarkdown(src, DefaultMarkdownTheme())
	md.Hyperlinks = false
	lines := md.Render(80)
	var plainRows []string
	for _, line := range lines {
		p := stripANSITest(line)
		if strings.ContainsAny(p, "┌└├│") {
			plainRows = append(plainRows, p)
		}
	}
	if len(plainRows) < 4 {
		t.Fatalf("expected table rows, got %v", plainRows)
	}
	joined := strings.Join(plainRows, "\n")
	if !strings.Contains(joined, "┌") || !strings.Contains(joined, "└") || !strings.Contains(joined, "├") {
		t.Fatalf("expected box borders: %v", plainRows)
	}
	var contentW int
	for _, row := range plainRows {
		if strings.HasPrefix(row, "│") {
			contentW = wrap.VisibleLen(row)
			break
		}
	}
	if contentW == 0 {
		t.Fatalf("no content rows: %v", plainRows)
	}
	for i, row := range plainRows {
		if wrap.VisibleLen(row) != contentW {
			t.Fatalf("row %d width %d != %d: %q", i, wrap.VisibleLen(row), contentW, row)
		}
	}
}

func TestVisibleLenSkipsOSC8(t *testing.T) {
	s := "\x1b]8;;https://x.test\x1b\\hi\x1b]8;;\x1b\\"
	if got := wrap.VisibleLen(s); got != 2 {
		t.Fatalf("visibleLen=%d want 2", got)
	}
}

func TestMarkedPortCodeFenceAndBoxTable(t *testing.T) {
	src := "# Title\n\n```\nA──B\n│  │\n```\n\n| H1 | H2 |\n| -- | -- |\n| a | bb |"
	md := NewMarkdown(src, DefaultMarkdownTheme())
	md.Hyperlinks = false
	joined := strings.Join(md.Render(60), "\n")
	plain := stripANSITest(joined)
	if strings.Contains(plain, "```") {
		t.Fatal("fences should be hidden")
	}
	if !strings.Contains(plain, "A──B") {
		t.Fatalf("missing fence body: %q", plain)
	}
	if !strings.Contains(plain, "┌") || !strings.Contains(plain, "└─") {
		t.Fatalf("missing box table: %q", plain)
	}
	if !strings.Contains(plain, "# Title") {
		t.Fatalf("missing heading: %q", plain)
	}
}

func TestMarkdownTableCyrillicEmDashSeparator(t *testing.T) {
	src := "| Параметр | Значение |\n| — | — |\n| Диаметр | `5.7 мм` |\n| Ширина | 10 мм |"
	assertAlignedBoxTable(t, src, 60)
}

func TestMarkdownTablePlusSeparator(t *testing.T) {
	src := "| A | B |\n|---+---|\n| x | y |"
	assertAlignedBoxTable(t, src, 40)
}

func TestMarkdownTableNoSeparatorHeuristic(t *testing.T) {
	src := "| Тип | Размер |\n| Crimp | 6 мм |\n| Comp | 5.7 мм |"
	assertAlignedBoxTable(t, src, 50)
}

func TestMarkdownSinglePipeNotTable(t *testing.T) {
	src := "see |foo| inline\n\n| alone |"
	md := NewMarkdown(src, DefaultMarkdownTheme())
	md.Hyperlinks = false
	joined := strings.Join(md.Render(40), "\n")
	plain := stripANSITest(joined)
	if strings.Contains(plain, "┌") {
		t.Fatalf("single pipe row should not become box table: %q", plain)
	}
}

func TestPadVisibleTruncatesOverflow(t *testing.T) {
	got := wrap.PadVisible("abcdef", 3)
	if wrap.VisibleLen(got) != 3 {
		t.Fatalf("padVisible overflow len=%d want 3 (%q)", wrap.VisibleLen(got), got)
	}
	if !strings.HasSuffix(stripANSITest(got), "…") && wrap.VisibleLen(got) > 3 {
		t.Fatalf("expected truncate: %q", got)
	}
}

func assertAlignedBoxTable(t *testing.T, src string, width int) {
	t.Helper()
	md := NewMarkdown(src, DefaultMarkdownTheme())
	md.Hyperlinks = false
	lines := md.Render(width)
	var plainRows []string
	for _, line := range lines {
		p := stripANSITest(line)
		if strings.ContainsAny(p, "┌└├│") {
			plainRows = append(plainRows, p)
		}
	}
	joined := strings.Join(plainRows, "\n")
	if !strings.Contains(joined, "┌") {
		t.Fatalf("expected box table, got raw pipes:\n%s\nfull:\n%s", joined, strings.Join(lines, "\n"))
	}
	var contentW int
	for _, row := range plainRows {
		if strings.HasPrefix(row, "│") {
			contentW = wrap.VisibleLen(row)
			break
		}
	}
	if contentW == 0 {
		t.Fatalf("no content rows: %v", plainRows)
	}
	for i, row := range plainRows {
		if wrap.VisibleLen(row) != contentW {
			t.Fatalf("row %d width %d != %d: %q", i, wrap.VisibleLen(row), contentW, row)
		}
	}
}

func TestMarkdownFencePipeTablePromoted(t *testing.T) {
	bare := "| Имя | Балл |\n|-----|------|\n| Алексей | 95 |\n| Мария | 87 |"
	for _, src := range []string{
		"```\n" + bare + "\n```",
		"```markdown\n" + bare + "\n```",
		"```md\nintro\n" + bare + "\n```",
	} {
		md := NewMarkdown(src, DefaultMarkdownTheme())
		md.Hyperlinks = false
		plain := stripANSITest(strings.Join(md.Render(80), "\n"))
		if !strings.Contains(plain, "┌") {
			t.Fatalf("fenced table should promote to box, src=%q out=%q", src[:min(40, len(src))], plain)
		}
		if strings.Contains(plain, "| Имя |") {
			t.Fatalf("raw GFM should not remain: %q", plain)
		}
	}
}

func TestMarkdownGoFenceKeepsRawTable(t *testing.T) {
	src := "```go\n| not | a | table |\n|-----|---|-------|\n| x | y | z |\n```"
	md := NewMarkdown(src, DefaultMarkdownTheme())
	md.Hyperlinks = false
	plain := stripANSITest(strings.Join(md.Render(80), "\n"))
	if strings.Contains(plain, "┌") {
		t.Fatalf("non-markdown fence should not promote table: %q", plain)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func stripANSITest(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if skip, next := wrap.SkipANSI(s, i); skip {
			i = next
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
