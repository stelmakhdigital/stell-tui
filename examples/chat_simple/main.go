// Простое демо чата на stell/tui (standalone Start/Stop).
//
//	go run ./examples/chat_simple
package main

import (
	"strings"
	"time"

	"stell/tui"
)

func main() {
	term := tui.NewProcessTerminal(nil, nil)
	ui := tui.NewWithTerminal(term, true)

	welcome := &tui.Text{Lines: []string{
		"Welcome to Simple Chat!",
		"",
		"Type a message and press Enter. Ctrl+C to exit.",
	}}
	ui.AddChild(welcome)

	editor := tui.NewEditor()
	editor.SetPlaceholder("message…")
	ui.AddChild(editor)
	ui.SetFocus(editor)

	ui.AddInputListener(func(data string) bool {
		if tui.MatchesKey(data, "ctrl+c") {
			ui.Stop()
			return true
		}
		return false
	})

	var responding bool
	editor.OnSubmit = func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || responding {
			return
		}
		responding = true
		editor.DisableSubmit = true
		editor.SetValue("")

		ui.RemoveChild(editor)
		ui.AddChild(&tui.Text{Lines: []string{"you: " + trimmed}})
		loader := &tui.Loader{Label: "Thinking...", Color: "\x1b[36m"}
		ui.AddChild(loader)
		ui.AddChild(editor)
		ui.SetFocus(editor)
		ui.RequestRender()

		go func() {
			for i := 0; i < 8; i++ {
				loader.Advance()
				ui.RequestRender()
				time.Sleep(80 * time.Millisecond)
			}
			ui.RemoveChild(loader)
			ui.AddChild(&tui.Text{Lines: []string{"bot: echo → " + trimmed}})
			// Keep editor at the bottom.
			ui.RemoveChild(editor)
			ui.AddChild(editor)
			ui.SetFocus(editor)
			responding = false
			editor.DisableSubmit = false
			ui.RequestRender()
		}()
	}

	ui.Start()
}
