package editor

import (
	"github.com/stelmakhdigital/stell-tui/wrap"
)

// Completer возвращает варианты автодополнения для запроса.
type Completer interface {
	Complete(query string) []CompleteItem
}

// CompleteItem — одна строка подсказки.
type CompleteItem struct {
	Label       string
	Value       string
	Description string
}

// Autocomplete хранит состояние попапа автодополнения.
type Autocomplete struct {
	Completer Completer
	Query     string
	Index     int
	Items     []CompleteItem
	Open      bool
}

// Update обновляет элементы по запросу (индекс сохраняется, если запрос тот же).
func (a *Autocomplete) Update(query string) {
	if a == nil || a.Completer == nil {
		return
	}
	if a.Open && a.Query == query && len(a.Items) > 0 {
		return
	}
	a.Query = query
	a.Items = a.Completer.Complete(query)
	if a.Index >= len(a.Items) {
		a.Index = 0
	}
	a.Open = len(a.Items) > 0
}

// Move сдвигает курсор выбора.
func (a *Autocomplete) Move(delta int) {
	if a == nil || len(a.Items) == 0 {
		return
	}
	a.Index += delta
	if a.Index < 0 {
		a.Index = 0
	}
	if a.Index >= len(a.Items) {
		a.Index = len(a.Items) - 1
	}
}

// Selected возвращает текущий выбранный элемент.
func (a *Autocomplete) Selected() (CompleteItem, bool) {
	if a == nil || !a.Open || a.Index < 0 || a.Index >= len(a.Items) {
		return CompleteItem{}, false
	}
	return a.Items[a.Index], true
}

// Close закрывает попап.
func (a *Autocomplete) Close() {
	if a == nil {
		return
	}
	a.Open = false
	a.Items = nil
	a.Index = 0
}

// StaticCompleter фильтрует фиксированный список через FuzzyFilter.
type StaticCompleter struct {
	Candidates []string
	Limit      int
}

func (s StaticCompleter) Complete(query string) []CompleteItem {
	limit := s.Limit
	if limit <= 0 {
		limit = 20
	}
	hits := wrap.FuzzyFilter(query, s.Candidates, limit)
	out := make([]CompleteItem, len(hits))
	for i, h := range hits {
		out[i] = CompleteItem{Label: h, Value: h}
	}
	return out
}
