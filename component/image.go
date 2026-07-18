package component

import (
	"github.com/stelmakhdigital/stell-tui/terminal"
	"strings"
	"github.com/stelmakhdigital/stell-tui/wrap"
)

// Image — плейсхолдер/инлайн-изображение (Kitty/iTerm при поддержке терминала).
type Image struct {
	Path    string
	Alt     string
	RawData []byte
	Rows    int
}

// NewImage создаёт компонент изображения по пути и alt-тексту.
func NewImage(path, alt string) *Image {
	return &Image{Path: path, Alt: alt, Rows: 1}
}

// Render выводит протокол изображения или текстовый stub.
func (img *Image) Render(width int) []string {
	label := img.Alt
	if label == "" {
		label = img.Path
	}
	if label == "" {
		label = "[image]"
	}
	if len(img.RawData) > 0 {
		cap := terminal.DetectCapabilities()
		if seq := terminal.EncodeTerminalImage(cap.Images, "image/png", img.RawData, terminal.ImageRenderOptions{
			MaxWidthCells: width, ImageID: 1,
		}); seq != "" {
			return []string{seq}
		}
		h := img.Rows
		if h < 3 {
			h = 3
		}
		return strings.Split(terminal.ImageStub(width, h, label), "\n")
	}
	line := "[image] " + label
	n := img.Rows
	if n < 1 {
		n = 1
	}
	out := make([]string, n)
	out[0] = wrap.Truncate(line, width)
	for i := 1; i < n; i++ {
		out[i] = ""
	}
	return out
}
