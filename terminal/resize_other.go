//go:build !unix

package terminal

// WatchResize — no-op на non-unix платформах.
func WatchResize(onResize func(w, h int)) (stop func()) {
	return func() {}
}
