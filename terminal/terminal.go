package terminal

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/stelmakhdigital/stell-tui/keys"
)

const (
	hideCursor = "\x1b[?25l"
	showCursor = "\x1b[?25h"
	clearLine  = "\x1b[2K"
	cursorHome = "\x1b[H"
)

// ImageProtocol — протокол инлайн-изображений терминала (Kitty / iTerm2).
type ImageProtocol string

const (
	ImageNone  ImageProtocol = ""
	ImageKitty ImageProtocol = "kitty"
	ImageITerm ImageProtocol = "iterm2"
)

// TerminalCapabilities описывает опциональные возможности терминала.
type TerminalCapabilities struct {
	Images     ImageProtocol
	TrueColor  bool
	Hyperlinks bool
}

// DetectCapabilities определяет поддержку image/truecolor/hyperlink по переменным окружения.
func DetectCapabilities() TerminalCapabilities {
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	term := strings.ToLower(os.Getenv("TERM"))
	colorTerm := strings.ToLower(os.Getenv("COLORTERM"))
	trueColor := colorTerm == "truecolor" || colorTerm == "24bit"

	if os.Getenv("TMUX") != "" || strings.HasPrefix(term, "tmux") || strings.HasPrefix(term, "screen") {
		return TerminalCapabilities{Images: ImageNone, TrueColor: trueColor, Hyperlinks: false}
	}
	cap := TerminalCapabilities{TrueColor: trueColor, Hyperlinks: true}
	switch {
	case termProgram == "kitty" || strings.Contains(term, "kitty") || os.Getenv("KITTY_WINDOW_ID") != "":
		cap.Images = ImageKitty
	case termProgram == "iterm.app" || termProgram == "wezterm" || termProgram == "warpterminal":
		cap.Images = ImageITerm
	case os.Getenv("TERM_PROGRAM") == "ghostty" || strings.Contains(term, "ghostty"):
		cap.Images = ImageKitty
	}
	return cap
}

// Terminal — минимальный I/O-контракт хоста TUI.
type Terminal interface {
	Start(onInput func(data string), onResize func())
	Stop()
	Write(data string)
	Columns() int
	Rows() int
	KittyProtocolActive() bool
	MoveBy(lines int)
	HideCursor()
	ShowCursor()
	ClearLine()
	ClearFromCursor()
	ClearScreen()
	SetTitle(title string)
}

// ProcessTerminal оборачивает stdin/stdout: raw mode, фичи терминала и буфер ввода.
type ProcessTerminal struct {
	mu sync.Mutex

	Reader *bufio.Reader
	In     io.Reader
	Out    io.Writer

	cols int
	rows int

	onInput  func(string)
	onResize func()

	restoreRaw  func()
	restoreFeat func()
	stopResize  func()
	stopRead    chan struct{}
	readDone    chan struct{}
	started     bool
	kittyActive bool
}

// NewProcessTerminal оборачивает r/w (по умолчанию os.Stdin/Stdout).
func NewProcessTerminal(r io.Reader, w io.Writer) *ProcessTerminal {
	if r == nil {
		r = os.Stdin
	}
	if w == nil {
		w = os.Stdout
	}
	cols, rows := 80, 24
	if w, h, err := TermSize(); err == nil {
		cols, rows = w, h
	}
	return &ProcessTerminal{
		Reader: bufio.NewReaderSize(r, 64*1024),
		In:     r,
		Out:    w,
		cols:   cols,
		rows:   rows,
	}
}

// WriteCSI пишет CSI/OSC-запрос в терминал.
func (t *ProcessTerminal) WriteCSI(seq string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := io.WriteString(t.Out, seq)
	return err
}

// QueryCellSize посылает CSI 16 t (размер ячейки в пикселях).
func (t *ProcessTerminal) QueryCellSize() error {
	return t.WriteCSI(seqCellSizeQuery)
}

// Start включает raw mode, фичи терминала, чтение ввода и слежение за resize.
func (t *ProcessTerminal) Start(onInput func(data string), onResize func()) {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return
	}
	t.onInput = onInput
	t.onResize = onResize
	t.started = true
	t.stopRead = make(chan struct{})
	t.readDone = make(chan struct{})
	t.mu.Unlock()

	restoreRaw, err := EnableRawMode()
	if err != nil {
		restoreRaw = func() {}
	}
	t.mu.Lock()
	t.restoreRaw = restoreRaw
	t.mu.Unlock()

	restoreFeat := EnableTerminalFeaturesWriter(t.Out)
	t.mu.Lock()
	t.restoreFeat = restoreFeat
	t.kittyActive = keys.KittyProtocolActive()
	t.mu.Unlock()

	QueryCellSizeWriter(t.Out)

	if w, h, err := TermSize(); err == nil {
		t.mu.Lock()
		t.cols, t.rows = w, h
		t.mu.Unlock()
	}

	t.stopResize = WatchResize(func(w, h int) {
		t.mu.Lock()
		t.cols, t.rows = w, h
		cb := t.onResize
		t.mu.Unlock()
		if cb != nil {
			cb()
		}
	})

	go t.readLoop()
}

// Stop восстанавливает терминал и останавливает читатели.
func (t *ProcessTerminal) Stop() {
	t.mu.Lock()
	if !t.started {
		t.mu.Unlock()
		return
	}
	t.started = false
	stopRead := t.stopRead
	readDone := t.readDone
	stopResize := t.stopResize
	restoreFeat := t.restoreFeat
	restoreRaw := t.restoreRaw
	t.stopResize = nil
	t.restoreFeat = nil
	t.restoreRaw = nil
	t.mu.Unlock()

	if stopRead != nil {
		close(stopRead)
	}
	if readDone != nil {
		<-readDone
	}
	if stopResize != nil {
		stopResize()
	}
	if restoreFeat != nil {
		restoreFeat()
	}
	if restoreRaw != nil {
		restoreRaw()
	}
	keys.SetKittyProtocolActive(false)
}

func (t *ProcessTerminal) readLoop() {
	defer close(t.readDone)
	buf := make([]byte, 4096)
	stdinBuf := keys.NewStdinBuffer(keys.StdinBufferOptions{})
	stdinBuf.OnData = func(data string) {
		t.mu.Lock()
		cb := t.onInput
		t.mu.Unlock()
		if cb != nil {
			cb(data)
		}
	}
	stdinBuf.OnPaste = func(content string) {
		t.mu.Lock()
		cb := t.onInput
		t.mu.Unlock()
		if cb != nil {
			cb("\x1b[200~" + content + "\x1b[201~")
		}
	}

	type readResult struct {
		n   int
		err error
	}
	for {
		ch := make(chan readResult, 1)
		go func() {
			n, err := t.In.Read(buf)
			ch <- readResult{n, err}
		}()
		select {
		case <-t.stopRead:
			return
		case res := <-ch:
			if res.n > 0 {
				chunk := string(buf[:res.n])
				if res.n == 1 && buf[0] > 127 {
					chunk = "\x1b" + string(rune(buf[0]-128))
				}
				stdinBuf.Process(chunk)
			}
			if res.err != nil {
				return
			}
		}
	}
}

// Write пишет сырые данные в вывод терминала.
func (t *ProcessTerminal) Write(data string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, _ = io.WriteString(t.Out, data)
}

// Columns возвращает текущую ширину терминала.
func (t *ProcessTerminal) Columns() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cols <= 0 {
		return 80
	}
	return t.cols
}

// Rows возвращает текущую высоту терминала.
func (t *ProcessTerminal) Rows() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.rows <= 0 {
		return 24
	}
	return t.rows
}

// KittyProtocolActive сообщает, считается ли активным протокол клавиатуры Kitty.
func (t *ProcessTerminal) KittyProtocolActive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.kittyActive || keys.KittyProtocolActive()
}

// MoveBy сдвигает курсор вверх (отрицательное) или вниз (положительное) на n строк.
func (t *ProcessTerminal) MoveBy(lines int) {
	if lines > 0 {
		t.Write(strings.Repeat("\n", lines))
	} else if lines < 0 {
		t.Write("\x1b[" + strconv.Itoa(-lines) + "A")
	}
}

// HideCursor скрывает аппаратный курсор.
func (t *ProcessTerminal) HideCursor() { t.Write(hideCursor) }

// ShowCursor показывает аппаратный курсор.
func (t *ProcessTerminal) ShowCursor() { t.Write(showCursor) }

// ClearLine очищает текущую строку.
func (t *ProcessTerminal) ClearLine() { t.Write(clearLine) }

// ClearFromCursor очищает от курсора до конца экрана.
func (t *ProcessTerminal) ClearFromCursor() { t.Write("\x1b[J") }

// ClearScreen очищает экран и переводит курсор в начало.
func (t *ProcessTerminal) ClearScreen() { t.Write("\x1b[2J" + cursorHome) }

// SetTitle задаёт заголовок окна терминала через OSC 0.
func (t *ProcessTerminal) SetTitle(title string) {
	t.Write("\x1b]0;" + title + "\x07")
}
