package keys

import "testing"

func TestKeyMapNormalize(t *testing.T) {
	m := NewKeyMap()
	m.Bind("Ctrl+C", "interrupt")
	if a, ok := m.Lookup("ctrl+c"); !ok || a != "interrupt" {
		t.Fatalf("%v %v", a, ok)
	}
}
