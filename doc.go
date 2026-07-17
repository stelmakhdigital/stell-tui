// Package tui — фреймворк терминального UI с дифференциальным рендерингом.
//
// Публичный импорт — `stell/tui`; реализация разложена по подпакетам
// (component, diff, editor, keys, overlay, terminal, wrap), реэкспорт — в export.go.
//
// Два режима хоста:
//
//   - Standalone: NewWithTerminal + Start/Stop (см. examples/chat_simple)
//   - Hosted: New + RenderNow — цикл событий и raw mode у вызывающего кода
//
// Интерактивный UI coding-agent живёт в stell/coding-agent/internal/tui
// и реэкспортирует этот пакет через type aliases.
package tui
