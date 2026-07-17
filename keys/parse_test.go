package keys

import "testing"

func TestParseKeyModifyOtherKeysEscape(t *testing.T) {
	SetKittyProtocolActive(false)
	got := NormalizeKeyChord(ParseKey("\x1b[27;1;27~"))
	if got != "esc" {
		t.Fatalf("got %q want esc", got)
	}
	if !MatchesKey("\x1b[27;1;27~", "escape") {
		t.Fatal("MatchesKey escape")
	}
}

func TestParseKeyModifyOtherKeysCtrl(t *testing.T) {
	SetKittyProtocolActive(false)
	for _, tc := range []struct {
		data string
		want string
	}{
		{"\x1b[27;5;99~", "ctrl+c"},
		{"\x1b[27;5;100~", "ctrl+d"},
		{"\x1b[27;5;122~", "ctrl+z"},
		{"\x1b[27;2;13~", "shift+enter"},
		{"\x1b[27;2;9~", "shift+tab"},
		{"\x1b[27;1;127~", "backspace"},
		{"\x1b[27;5;127~", "ctrl+backspace"},
		{"\x1b[27;5;47~", "ctrl+/"},
		{"\x1b[27;5;49~", "ctrl+1"},
		{"\x1b[27;2;49~", "shift+1"},
		{"\x1b[27;2;69~", "shift+e"},
		{"\x1b[27;6;69~", "shift+ctrl+e"},
		{"\x1b[27;7;104~", "ctrl+alt+h"},
	} {
		got := NormalizeKeyChord(ParseKey(tc.data))
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.data, got, tc.want)
		}
	}
}

func TestParseKeyKittyCSIu(t *testing.T) {
	SetKittyProtocolActive(true)
	for _, tc := range []struct {
		data string
		want string
	}{
		{"\x1b[27u", "esc"},
		{"\x1b[99;5u", "ctrl+c"},
		{"\x1b[107;9u", "super+k"},
	} {
		got := NormalizeKeyChord(ParseKey(tc.data))
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.data, got, tc.want)
		}
	}
}

func TestParseKeyLegacy(t *testing.T) {
	SetKittyProtocolActive(false)
	for _, tc := range []struct {
		data string
		want string
	}{
		{"\x1b", "esc"},
		{"\x03", "ctrl+c"},
		{"\x1b[A", "up"},
		{"\x1b[Z", "shift+tab"},
		{"\x1ba", "alt+a"},
		{"\x1bOP", "f1"},
		{"\x1b[24~", "f12"},
	} {
		got := NormalizeKeyChord(ParseKey(tc.data))
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.data, got, tc.want)
		}
	}
}

func TestParseKeyKittyActiveShiftEnter(t *testing.T) {
	SetKittyProtocolActive(true)
	if got := ParseKey("\n"); got != "shift+enter" {
		t.Fatalf("newline: got %q", got)
	}
	if got := ParseKey("\x1b\r"); got != "shift+enter" {
		t.Fatalf("esc+cr: got %q", got)
	}
}

func TestDecodePrintableKey(t *testing.T) {
	if got := DecodePrintableKey("\x1b[109u"); got != "m" {
		t.Fatalf("got %q want m", got)
	}
}

func TestNormalizeKeyChordAliases(t *testing.T) {
	if got := NormalizeKeyChord("escape"); got != "esc" {
		t.Fatalf("escape -> %q", got)
	}
	if got := NormalizeKeyChord("pageUp"); got != "pgup" {
		t.Fatalf("pageUp -> %q", got)
	}
	if got := NormalizeKeyChord("ctrl+shift+m"); got != "shift+ctrl+m" {
		t.Fatalf("modifier order: %q", got)
	}
}
