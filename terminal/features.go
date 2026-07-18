package terminal

import (
	"io"
	"os"

	"github.com/stelmakhdigital/stell-tui/keys"
)

const (
	seqBracketedPasteOn  = "\x1b[?2004h"
	seqBracketedPasteOff = "\x1b[?2004l"
	seqKittyKeyboardOn   = "\x1b[>1u" // progressive enhancement request
	seqKittyKeyboardOff  = "\x1b[<u"
	seqModifyOtherKeys   = "\x1b[>4;1m"
	seqCellSizeQuery     = "\x1b[16t"
)

// EnableTerminalFeatures включает bracketed paste и расширенную отчётность клавиш на stdout.
// Безопасно после EnableRawMode; closer восстанавливает режимы.
func EnableTerminalFeatures() (restore func()) {
	return EnableTerminalFeaturesWriter(os.Stdout)
}

// EnableTerminalFeaturesWriter включает фичи на указанном writer.
func EnableTerminalFeaturesWriter(w io.Writer) (restore func()) {
	if w == nil {
		w = os.Stdout
	}
	_, _ = io.WriteString(w, seqBracketedPasteOn+seqModifyOtherKeys+seqKittyKeyboardOn)
	// Do not block startup reading stdin; keyboard negotiation is best-effort.
	keys.SetKittyProtocolActive(false)
	return func() {
		_, _ = io.WriteString(w, seqBracketedPasteOff+seqKittyKeyboardOff)
		keys.SetKittyProtocolActive(false)
	}
}

// QueryCellSize запрашивает размер ячейки (CSI 16 t) на stdout.
func QueryCellSize() {
	QueryCellSizeWriter(os.Stdout)
}

// QueryCellSizeWriter запрашивает размер ячейки на указанном writer.
func QueryCellSizeWriter(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	_, _ = io.WriteString(w, seqCellSizeQuery)
}
