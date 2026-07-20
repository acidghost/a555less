package tui

import (
	"bytes"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/acidghost/a555less/internal/jsondoc"
)

func TestPrintContentFormats(t *testing.T) {
	doc, err := jsondoc.Parse([]byte(`{"users":[{"name":"Ada\nLovelace","weird key":true}]}`), "print.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	users := doc.Root.Children[0]
	name := users.Children[0].Children[0]

	tests := []struct {
		name   string
		node   *jsondoc.Node
		target printTarget
		want   string
	}{
		{"pretty", users.Children[0], printPrettyValue, "{\n  \"name\": \"Ada\\nLovelace\",\n  \"weird key\": true\n}"},
		{"string", name, printString, `Ada\nLovelace`},
		{"query", name, printQueryPath, `.users[].name`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := printContent(doc, tt.node, tt.target)
			if err != nil {
				t.Fatalf("printContent() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("printContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintTargetForKey(t *testing.T) {
	tests := map[string]printTarget{
		"p": printPrettyValue,
		"s": printString,
		"q": printQueryPath,
	}
	for key, want := range tests {
		if got, ok := printTargetForKey(key); !ok || got != want {
			t.Errorf("printTargetForKey(%q) = %v, %v, want %v, true", key, got, ok, want)
		}
	}
	if _, ok := printTargetForKey("x"); ok {
		t.Error("printTargetForKey(x) should reject an unknown command")
	}
}

func TestPrintContentRejectsInvalidTargets(t *testing.T) {
	doc, err := jsondoc.Parse([]byte(`{"value":1}`), "print.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if _, err := printContent(doc, doc.Root, printString); err != errPrintNotString {
		t.Errorf("print string error = %v, want %v", err, errPrintNotString)
	}
}

func TestTerminalPrintCommandWritesUntruncatedContent(t *testing.T) {
	content := strings.Repeat("0123456789", 1000)
	stdin := bytes.NewBufferString("x")
	var stdout bytes.Buffer
	cmd := terminalPrintCommand{content: content, stdin: stdin, stdout: &stdout}

	if err := cmd.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := "\x1b[2J\x1b[H" + content + "\n\nPress any key to continue."
	if got := stdout.String(); got != want {
		t.Fatalf("Run() wrote %d bytes, want %d", len(got), len(want))
	}
}

func TestPrintCommandSuspendsTerminal(t *testing.T) {
	doc, err := jsondoc.Parse([]byte(`{"value":1}`), "print.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m := New(doc)

	updated, _ := m.Update(printKeyMsg('p'))
	m = updated.(Model)
	if !m.printPending {
		t.Fatal("p should start a pending print command")
	}

	updated, cmd := m.Update(printKeyMsg('p'))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("pp should return a terminal exec command")
	}
	if m.printPending {
		t.Fatal("pp should leave pending print mode")
	}
	if !m.View().AltScreen {
		t.Fatal("the regular TUI view should remain in the alternate screen")
	}
}

func TestCtrlCQuitsPendingPrintCommand(t *testing.T) {
	doc, err := jsondoc.Parse([]byte(`{"value":1}`), "print.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m := New(doc)
	updated, _ := m.Update(printKeyMsg('p'))
	m = updated.(Model)

	ctrlC := tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl})
	_, cmd := m.Update(ctrlC)
	if cmd == nil {
		t.Fatal("ctrl+c in pending print mode should quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c command returned %T, want tea.QuitMsg", cmd())
	}
}

func printKeyMsg(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)})
}
