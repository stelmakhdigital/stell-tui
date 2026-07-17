package wrap

import "testing"

func TestFuzzyFilter(t *testing.T) {
	hits := FuzzyFilter("ab", []string{"alpha-bravo", "zzz", "abacus"}, 10)
	if len(hits) < 2 {
		t.Fatalf("%v", hits)
	}
}
