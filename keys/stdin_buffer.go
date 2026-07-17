package keys

import (
	"regexp"
	"strings"
	"time"
)

const (
	escByte             = "\x1b"
	bracketedPasteStart = "\x1b[200~"
	bracketedPasteEnd   = "\x1b[201~"
)

var reKittyPlainPrintable = regexp.MustCompile(`^\x1b\[(\d+)(?::\d*)?(?::\d+)?u$`)

// StdinBufferOptions настраивает буферизацию последовательностей stdin.
type StdinBufferOptions struct {
	Timeout time.Duration
}

// StdinBuffer накапливает куски stdin и отдаёт полные escape-последовательности.
type StdinBuffer struct {
	buffer                         string
	timeout                        time.Duration
	timer                          *time.Timer
	pasteMode                      bool
	pasteBuffer                    string
	pendingKittyPrintableCodepoint int
	OnData                         func(string)
	OnPaste                        func(string)
}

// NewStdinBuffer создаёт буфер stdin с опциональным timeout (по умолчанию 10ms).
func NewStdinBuffer(opts StdinBufferOptions) *StdinBuffer {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Millisecond
	}
	return &StdinBuffer{timeout: timeout}
}

// Process подаёт данные ввода в буфер.
func (b *StdinBuffer) Process(data string) {
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	if data == "" && b.buffer == "" {
		b.emitDataSequence("")
		return
	}
	b.buffer += data

	if b.pasteMode {
		b.pasteBuffer += b.buffer
		b.buffer = ""
		if idx := strings.Index(b.pasteBuffer, bracketedPasteEnd); idx >= 0 {
			content := b.pasteBuffer[:idx]
			remaining := b.pasteBuffer[idx+len(bracketedPasteEnd):]
			b.pasteMode = false
			b.pasteBuffer = ""
			b.pendingKittyPrintableCodepoint = 0
			if b.OnPaste != nil {
				b.OnPaste(content)
			}
			if remaining != "" {
				b.Process(remaining)
			}
		}
		return
	}

	if idx := strings.Index(b.buffer, bracketedPasteStart); idx >= 0 {
		if idx > 0 {
			seqs, rem := extractCompleteSequences(b.buffer[:idx])
			for _, s := range seqs {
				b.emitDataSequence(s)
			}
			if rem != "" {
				b.emitDataSequence(rem)
			}
		}
		b.pendingKittyPrintableCodepoint = 0
		b.buffer = b.buffer[idx+len(bracketedPasteStart):]
		b.pasteMode = true
		b.pasteBuffer = b.buffer
		b.buffer = ""
		if end := strings.Index(b.pasteBuffer, bracketedPasteEnd); end >= 0 {
			content := b.pasteBuffer[:end]
			remaining := b.pasteBuffer[end+len(bracketedPasteEnd):]
			b.pasteMode = false
			b.pasteBuffer = ""
			b.pendingKittyPrintableCodepoint = 0
			if b.OnPaste != nil {
				b.OnPaste(content)
			}
			if remaining != "" {
				b.Process(remaining)
			}
		}
		return
	}

	seqs, rem := extractCompleteSequences(b.buffer)
	b.buffer = rem
	for _, s := range seqs {
		b.emitDataSequence(s)
	}
	if b.buffer != "" {
		buf := b
		b.timer = time.AfterFunc(b.timeout, func() {
			for _, s := range buf.Flush() {
				buf.emitDataSequence(s)
			}
		})
	}
}

func (b *StdinBuffer) emitDataSequence(sequence string) {
	if sequence == "" {
		if b.OnData != nil {
			b.OnData(sequence)
		}
		return
	}
	if len(sequence) == 1 {
		if int(sequence[0]) == b.pendingKittyPrintableCodepoint {
			b.pendingKittyPrintableCodepoint = 0
			return
		}
	}
	b.pendingKittyPrintableCodepoint = parseUnmodifiedKittyPrintableCodepoint(sequence)
	if b.OnData != nil {
		b.OnData(sequence)
	}
}

// Flush возвращает и очищает незавершённую буферизованную последовательность.
func (b *StdinBuffer) Flush() []string {
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	if b.buffer == "" {
		return nil
	}
	out := []string{b.buffer}
	b.buffer = ""
	b.pendingKittyPrintableCodepoint = 0
	return out
}

// Clear сбрасывает состояние буфера.
func (b *StdinBuffer) Clear() {
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	b.buffer = ""
	b.pasteMode = false
	b.pasteBuffer = ""
	b.pendingKittyPrintableCodepoint = 0
}

func parseUnmodifiedKittyPrintableCodepoint(sequence string) int {
	m := reKittyPlainPrintable.FindStringSubmatch(sequence)
	if m == nil {
		return 0
	}
	cp := atoi(m[1])
	if cp >= 32 {
		return cp
	}
	return 0
}

type seqStatus int

const (
	seqNotEscape seqStatus = iota
	seqComplete
	seqIncomplete
)

func sequenceStatus(data string) seqStatus {
	if !strings.HasPrefix(data, escByte) {
		return seqNotEscape
	}
	if len(data) == 1 {
		return seqIncomplete
	}
	after := data[1:]
	switch {
	case strings.HasPrefix(after, "["):
		if strings.HasPrefix(after, "[M") {
			if len(data) >= 6 {
				return seqComplete
			}
			return seqIncomplete
		}
		return csiStatus(data)
	case strings.HasPrefix(after, "]"):
		if strings.HasSuffix(data, escByte+"\\") || strings.HasSuffix(data, "\x07") {
			return seqComplete
		}
		return seqIncomplete
	case strings.HasPrefix(after, "P"), strings.HasPrefix(after, "_"):
		if strings.HasSuffix(data, escByte+"\\") {
			return seqComplete
		}
		return seqIncomplete
	case strings.HasPrefix(after, "O"):
		if len(after) >= 2 {
			return seqComplete
		}
		return seqIncomplete
	case len(after) == 1:
		return seqComplete
	default:
		return seqComplete
	}
}

func csiStatus(data string) seqStatus {
	if len(data) < 3 {
		return seqIncomplete
	}
	payload := data[2:]
	last := payload[len(payload)-1]
	if last >= 0x40 && last <= 0x7e {
		if strings.HasPrefix(payload, "<") && (last == 'M' || last == 'm') {
			parts := strings.Split(payload[1:len(payload)-1], ";")
			if len(parts) == 3 {
				return seqComplete
			}
			return seqIncomplete
		}
		return seqComplete
	}
	return seqIncomplete
}

func extractCompleteSequences(buffer string) (sequences []string, remainder string) {
	pos := 0
	for pos < len(buffer) {
		remaining := buffer[pos:]
		if strings.HasPrefix(remaining, escByte) {
			seqEnd := 1
			handled := false
			for seqEnd <= len(remaining) {
				candidate := remaining[:seqEnd]
				switch sequenceStatus(candidate) {
				case seqNotEscape:
					sequences = append(sequences, candidate)
					pos += seqEnd
					handled = true
				case seqComplete:
					if candidate == "\x1b\x1b" && seqEnd < len(remaining) {
						next := remaining[seqEnd]
						if next == '[' || next == ']' || next == 'O' || next == 'P' || next == '_' {
							sequences = append(sequences, escByte)
							pos++
							handled = true
							break
						}
					}
					sequences = append(sequences, candidate)
					pos += seqEnd
					handled = true
				case seqIncomplete:
					seqEnd++
					continue
				}
				break
			}
			if handled {
				continue
			}
			return sequences, remaining
		}
		sequences = append(sequences, remaining[:1])
		pos++
	}
	return sequences, ""
}
