//go:build unix

package terminal

import (
	"os"
	"os/signal"
	"syscall"
)

// WatchResize вызывает onResize при изменении размера терминала.
// stop снимает обработчик сигнала и завершает watcher.
func WatchResize(onResize func(w, h int)) (stop func()) {
	if onResize == nil {
		return func() {}
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ch:
				w, h, err := TermSize()
				if err != nil {
					continue
				}
				select {
				case <-done:
					return
				default:
					onResize(w, h)
				}
			}
		}
	}()
	return func() {
		signal.Stop(ch)
		close(done)
	}
}
