package diff

import (
	"fmt"
	"strings"

	"github.com/stelmakhdigital/stell-tui/wrap"
)

// extractCursorMarker находит wrap.CursorMarker в строках кадра и удаляет его.
// Возвращает row/col (0-based ячейки экрана) и очищенные строки.
func extractCursorMarker(lines []string) (row, col int, cleaned []string, found bool) {
	cleaned = make([]string, len(lines))
	marker := wrap.CursorMarker
	for i, line := range lines {
		if idx := strings.Index(line, marker); idx >= 0 {
			cleaned[i] = line[:idx] + line[idx+len(marker):]
			col = wrap.VisibleLen(line[:idx])
			return i, col, cleaned, true
		}
		cleaned[i] = line
	}
	return 0, 0, cleaned, false
}

func positionHardwareCursor(row, col int) string {
	return fmt.Sprintf("\x1b[%d;%dH%s", row+1, col+1, showCursor)
}
