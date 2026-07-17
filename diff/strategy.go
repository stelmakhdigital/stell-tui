package diff

// DiffStrategy — стратегии дифференциального рендера (синхронизированный вывод CSI 2026).
type DiffStrategy int

const (
	DiffFull DiffStrategy = iota
	DiffPatch
	DiffScroll
)
