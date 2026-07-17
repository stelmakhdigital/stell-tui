//go:build !unix

package terminal

import "fmt"

// EnableRawMode не поддерживается на non-unix платформах.
func EnableRawMode() (func(), error) {
	return func() {}, fmt.Errorf("raw mode unsupported on this platform")
}

// TermSize возвращает размер по умолчанию на non-unix платформах.
func TermSize() (int, int, error) {
	return 80, 24, nil
}
