package wrap

import (
	"strings"
	"unicode"
)

// FuzzyScore оценивает совпадение query с candidate (больше — лучше; -1 — нет совпадения).
func FuzzyScore(query, candidate string) int {
	q := []rune(strings.ToLower(query))
	c := []rune(strings.ToLower(candidate))
	if len(q) == 0 {
		return 0
	}
	if len(c) == 0 {
		return -1
	}
	qi := 0
	score := 0
	prevMatch := -2
	for ci, r := range c {
		if qi < len(q) && r == q[qi] {
			score += 10
			if ci == prevMatch+1 {
				score += 5
			}
			if ci == 0 || !unicode.IsLetter(c[ci-1]) {
				score += 3
			}
			prevMatch = ci
			qi++
		}
	}
	if qi < len(q) {
		return -1
	}
	score -= len(c) - len(q)
	return score
}

// FuzzyFilter возвращает совпадения с query, лучшие первыми.
func FuzzyFilter(query string, candidates []string, limit int) []string {
	type scored struct {
		s string
		n int
	}
	var hits []scored
	for _, c := range candidates {
		n := FuzzyScore(query, c)
		if n < 0 {
			continue
		}
		hits = append(hits, scored{c, n})
	}
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if hits[j].n > hits[i].n {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	out := make([]string, len(hits))
	for i, h := range hits {
		out[i] = h.s
	}
	return out
}
