package component

import (
	"strings"
	"testing"
)

func TestSettingsListCycle(t *testing.T) {
	var gotID, gotVal string
	s := NewSettingsList([]SettingsItem{{
		ID: "mode", Label: "Mode", CurrentValue: "a", Values: []string{"a", "b"},
	}}, func(id, v string) { gotID, gotVal = id, v }, nil)
	s.HandleInput("\r")
	if gotID != "mode" || gotVal != "b" {
		t.Fatalf("got %s=%s", gotID, gotVal)
	}
	lines := s.Render(40)
	if len(lines) == 0 || !strings.Contains(lines[0], "settings") {
		t.Fatalf("render: %v", lines)
	}
}
