package editor

// Input — однострочный ввод.
type Input struct {
	*Editor
	OnSubmit func(value string)
}

// NewInput создаёт однострочный ввод.
func NewInput() *Input {
	return &Input{Editor: NewEditor()}
}

// HandleInput: Enter — submit, остальные клавиши — в Editor.
func (in *Input) HandleInput(data string) {
	if in == nil || in.Editor == nil {
		return
	}
	switch data {
	case "\r", "\n":
		if in.OnSubmit != nil {
			in.OnSubmit(in.Value())
		}
	default:
		in.Editor.HandleInput(data)
	}
}
