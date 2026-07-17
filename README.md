# stell/tui

Фреймворк терминального UI с дифференциальным рендерингом.

> Дифференциальный рендеринг - при каждом обновлении перерисовывается не весь экран, а только разница с предыдущим кадром — изменённые строки/ячейки.

## Возможности

- Дифференциальный рендер (`DiffFull` / `DiffPatch` / `DiffScroll`) и CSI 2026
- `ProcessTerminal` — raw mode, bracketed paste, подсказки Kitty keyboard, resize
- Standalone `TUI.Start` / `Stop` или hosted `RenderNow` для встраивания
- Компоненты: `Container`, `Text`, `Box`, `Loader`, `SelectList`, `SettingsList`, `Markdown`, `Image`, `Editor`, `Input`
- Стек оверлеев: `OverlayOptions` / `OverlayHandle`

## Карта директорий

| Путь | Назначение |
|------|------------|
| `tui.go`, `overlay_host.go`, `export.go` | хост Start/Stop, фокус, оверлеи, публичные aliases |
| `component/` | UI-компоненты |
| `diff/` | DiffEngine |
| `terminal/` | Terminal / ProcessTerminal |
| `keys/` | клавиши и буфер stdin |
| `overlay/` | раскладка и композитинг оверлеев |
| `editor/` | редактор, input, autocomplete |
| `wrap/` | ширина / ANSI / fuzzy |
| `examples/` | демо |

Внешний код импортирует только `stell/tui` (не подпакеты).

## Standalone

```go
package main

import "stell/tui"

func main() {
	term := tui.NewProcessTerminal(nil, nil)
	ui := tui.NewWithTerminal(term, true)

	ui.AddChild(&tui.Text{Lines: []string{"Hello"}})
	ed := tui.NewEditor()
	ed.OnSubmit = func(v string) { /* ... */ }
	ui.AddChild(ed)
	ui.SetFocus(ed)

	ui.AddInputListener(func(data string) bool {
		if tui.MatchesKey(data, "ctrl+c") {
			ui.Stop()
			return true
		}
		return false
	})

	ui.Start()
}
```

Демо:

```bash
go run ./examples/chat_simple
```

## Hosted (встраивание)

Для приложений со своим циклом (например `stell/coding-agent`):

```go
ui := tui.New(os.Stdout, true)
defer ui.Close()

restore, _ := tui.EnableRawMode()
defer restore()
defer tui.EnableTerminalFeatures()()

w, h, _ := tui.TermSize()
ui.SetSize(w, h)
ui.SetRoot(tui.NewContainer(root))
_ = ui.RenderNow() // после каждого обновления модели
```

## Маркер курсора

`CursorMarker` — APC `"\x1b_stell:c\x07"`.
При `SetShowHardwareCursor(true)` движок ищет полную последовательность, снимает её и позиционирует аппаратный курсор. `VisibleLen` / `StripANSI` считают APC zero-width.
