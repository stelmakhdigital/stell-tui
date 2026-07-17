package terminal

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// ImageRenderOptions управляет размещением инлайн-изображения.
type ImageRenderOptions struct {
	MaxWidthCells  int
	MaxHeightCells int
	ImageID        int
	MoveCursor     bool
}

// EncodeTerminalImage возвращает escape-последовательности для инлайн-показа изображения.
// Пустая строка — протокол не поддерживается (caller может показать stub).
func EncodeTerminalImage(protocol ImageProtocol, mime string, data []byte, opts ImageRenderOptions) string {
	if len(data) == 0 || protocol == ImageNone {
		return ""
	}
	if mime == "" {
		mime = "image/png"
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	switch protocol {
	case ImageKitty:
		return encodeKittyImage(b64, opts)
	case ImageITerm:
		return encodeITermImage(b64, mime, opts)
	default:
		return ""
	}
}

func encodeKittyImage(b64 string, opts ImageRenderOptions) string {
	// Протокол графики Kitty: APC G … ST, payload по частям.
	const chunk = 4096
	var b strings.Builder
	id := opts.ImageID
	if id <= 0 {
		id = 1
	}
	first := true
	for len(b64) > 0 {
		part := b64
		more := 0
		if len(part) > chunk {
			part = b64[:chunk]
			b64 = b64[chunk:]
			more = 1
		} else {
			b64 = ""
		}
		if first {
			fmt.Fprintf(&b, "\x1b_Ga=T,f=100,m=%d,i=%d", more, id)
			if opts.MaxWidthCells > 0 {
				fmt.Fprintf(&b, ",c=%d", opts.MaxWidthCells)
			}
			if opts.MaxHeightCells > 0 {
				fmt.Fprintf(&b, ",r=%d", opts.MaxHeightCells)
			}
			b.WriteByte(';')
			b.WriteString(part)
			b.WriteString("\x1b\\")
			first = false
		} else {
			fmt.Fprintf(&b, "\x1b_Gm=%d;%s\x1b\\", more, part)
		}
	}
	return b.String()
}

func encodeITermImage(b64, mime string, opts ImageRenderOptions) string {
	var b strings.Builder
	b.WriteString("\x1b]1337;File=inline=1")
	if opts.MaxWidthCells > 0 {
		fmt.Fprintf(&b, ";width=%d", opts.MaxWidthCells)
	}
	if opts.MaxHeightCells > 0 {
		fmt.Fprintf(&b, ";height=%d", opts.MaxHeightCells)
	}
	fmt.Fprintf(&b, ";name=%s", base64.StdEncoding.EncodeToString([]byte(mime)))
	b.WriteByte(':')
	b.WriteString(b64)
	b.WriteByte('\a')
	return b.String()
}

// ImageStub — текстовый fallback, если графические протоколы недоступны.
func ImageStub(width, height int, label string) string {
	if width < 8 {
		width = 8
	}
	if height < 1 {
		height = 1
	}
	if label == "" {
		label = "image"
	}
	border := strings.Repeat("─", width-2)
	lines := make([]string, 0, height)
	lines = append(lines, "┌"+border+"┐")
	pad := (width - 2 - len(label)) / 2
	if pad < 0 {
		pad = 0
	}
	mid := strings.Repeat(" ", pad) + label
	if len(mid) > width-2 {
		mid = mid[:width-2]
	}
	mid = mid + strings.Repeat(" ", width-2-len(mid))
	for i := 1; i < height-1; i++ {
		if i == height/2 {
			lines = append(lines, "│"+mid+"│")
		} else {
			lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
		}
	}
	if height > 1 {
		lines = append(lines, "└"+border+"┘")
	}
	return strings.Join(lines, "\n")
}
