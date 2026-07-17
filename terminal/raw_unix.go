//go:build unix

package terminal

import (
	"os"
	"syscall"
	"unsafe"
)

// EnableRawMode переводит stdin в raw mode; restore возвращает прежнее состояние.
func EnableRawMode() (restore func(), err error) {
	fd := int(os.Stdin.Fd())
	var old syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlGetTermios, uintptr(unsafe.Pointer(&old)), 0, 0, 0)
	if errno != 0 {
		return func() {}, errno
	}
	raw := old
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	_, _, errno = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlSetTermios, uintptr(unsafe.Pointer(&raw)), 0, 0, 0)
	if errno != 0 {
		return func() {}, errno
	}
	return func() {
		_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlSetTermios, uintptr(unsafe.Pointer(&old)), 0, 0, 0)
	}, nil
}

// TermSize возвращает текущий размер терминала.
func TermSize() (w, h int, err error) {
	fd := int(os.Stdout.Fd())
	var ws struct {
		Row, Col       uint16
		Xpixel, Ypixel uint16
	}
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)), 0, 0, 0)
	if errno != 0 {
		return 80, 24, errno
	}
	if ws.Col == 0 {
		ws.Col = 80
	}
	if ws.Row == 0 {
		ws.Row = 24
	}
	return int(ws.Col), int(ws.Row), nil
}
