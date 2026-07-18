package tui_test

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stelmakhdigital/stell-tui"
)

// fakeTerminal — in-memory Terminal для интеграционного теста хоста.
type fakeTerminal struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	cols     int
	rows     int
	onInput  func(string)
	onResize func()
	started  bool
	stopped  bool
}

func newFakeTerminal(cols, rows int) *fakeTerminal {
	return &fakeTerminal{cols: cols, rows: rows}
}

func (f *fakeTerminal) Start(onInput func(string), onResize func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onInput = onInput
	f.onResize = onResize
	f.started = true
}

func (f *fakeTerminal) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopped = true
	f.started = false
}

func (f *fakeTerminal) Write(data string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, _ = f.buf.WriteString(data)
}

func (f *fakeTerminal) Columns() int              { return f.cols }
func (f *fakeTerminal) Rows() int                 { return f.rows }
func (f *fakeTerminal) KittyProtocolActive() bool { return false }
func (f *fakeTerminal) MoveBy(lines int) {
	if lines < 0 {
		f.Write("\x1b[" + strconv.Itoa(-lines) + "A")
	}
}
func (f *fakeTerminal) HideCursor()      {}
func (f *fakeTerminal) ShowCursor()      {}
func (f *fakeTerminal) ClearLine()       {}
func (f *fakeTerminal) ClearFromCursor() {}
func (f *fakeTerminal) ClearScreen()     {}
func (f *fakeTerminal) SetTitle(string)  {}

func (f *fakeTerminal) inject(data string) {
	f.mu.Lock()
	cb := f.onInput
	f.mu.Unlock()
	if cb != nil {
		cb(data)
	}
}

func TestTUIStartStopWithFakeTerminal(t *testing.T) {
	var out bytes.Buffer
	term := newFakeTerminal(40, 12)
	ui := tui.New(&out, false)
	ui.SetTerminal(term)
	ui.AddChild(&tui.Text{Lines: []string{"hello"}})

	done := make(chan struct{})
	go func() {
		ui.Start()
		close(done)
	}()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		term.mu.Lock()
		ok := term.started
		term.mu.Unlock()
		if ok {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	term.inject("x")
	ui.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not exit after Stop")
	}

	term.mu.Lock()
	stopped := term.stopped
	term.mu.Unlock()
	if !stopped {
		t.Fatal("expected terminal Stop")
	}
	if out.Len() == 0 {
		t.Fatal("expected DiffEngine output")
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("expected hello in output: %q", out.String())
	}
}
