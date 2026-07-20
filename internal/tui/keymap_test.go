package tui

import (
	"strings"
	"testing"
)

func TestKeyBindingsListsAllBindings(t *testing.T) {
	got := KeyBindings()

	wantRows := []string{
		"q, ctrl+c",
		"j, down",
		"k, up",
		"space, enter",
		"/",
		"search",
		"n",
		"next match",
		"N",
		"previous match",
		"pp",
		"print pretty value",
		"ps",
		"print string contents",
		"pq",
		"print jq query",
		"?",
		"toggle help",
	}
	for _, want := range wantRows {
		if !strings.Contains(got, want) {
			t.Fatalf("KeyBindings() missing %q in:\n%s", want, got)
		}
	}

	if strings.Contains(got, "TODO") {
		t.Fatalf("KeyBindings() still contains TODO: %q", got)
	}
	if strings.Contains(got, "|") || strings.Contains(got, "`") {
		t.Fatalf("KeyBindings() should be plain padded text, got:\n%s", got)
	}
}
