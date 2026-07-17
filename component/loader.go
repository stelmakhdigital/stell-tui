package component

import (
	"stell/tui/wrap"
	"context"
	"sync"
	"time"
)

// Loader — простой спиннер на одном кадре.
type Loader struct {
	Label string
	Tick  int
	Color string
}

// Advance переключает кадр анимации.
func (l *Loader) Advance() { l.Tick++ }

// Render возвращает строку со спиннером и подписью.
func (l *Loader) Render(width int) []string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	f := frames[l.Tick%len(frames)]
	color := l.Color
	if color == "" {
		color = "\x1b[36m"
	}
	line := color + f + "\x1b[0m " + l.Label
	return []string{wrap.Truncate(line, width)}
}

// CancellableLoader отслеживает асинхронную загрузку с возможностью отмены.
type CancellableLoader struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	running bool
	label   string
	started time.Time
}

// Start запускает загрузку; предыдущая, если была, отменяется.
func (l *CancellableLoader) Start(parent context.Context, label string, fn func(ctx context.Context)) {
	l.mu.Lock()
	if l.cancel != nil {
		l.cancel()
	}
	ctx, cancel := context.WithCancel(parent)
	l.cancel = cancel
	l.running = true
	l.label = label
	l.started = time.Now()
	l.mu.Unlock()
	go func() {
		defer func() {
			l.mu.Lock()
			l.running = false
			l.mu.Unlock()
		}()
		fn(ctx)
	}()
}

// Cancel прерывает текущую загрузку.
func (l *CancellableLoader) Cancel() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cancel != nil {
		l.cancel()
	}
	l.running = false
}

// Running сообщает, идёт ли загрузка.
func (l *CancellableLoader) Running() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}

// Label возвращает текущую подпись загрузки.
func (l *CancellableLoader) Label() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.label
}
