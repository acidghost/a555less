package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/acidghost/a555less/internal/jsondoc"
)

func newSearchTestModel(t *testing.T) Model {
	t.Helper()
	doc, err := jsondoc.Parse([]byte(`{"needle":"needle needle","nested":{"x":"needle"}}`), "search.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m := New(doc)
	m.width = 80
	m.height = 12
	return m
}

func TestApplySearchFindsKeysAndValuesAndNavigates(t *testing.T) {
	m := newSearchTestModel(t)
	m.searchInput = "needle"
	m.applySearch()

	if got, want := len(m.searchMatches), 4; got != want {
		t.Fatalf("search matches = %d, want %d", got, want)
	}
	if got, want := m.searchIndex, 0; got != want {
		t.Fatalf("current match = %d, want %d", got, want)
	}
	if got, want := focusedPath(m), "input.needle"; got != want {
		t.Fatalf("focused path = %q, want %q", got, want)
	}

	m.moveSearch(1)
	m.moveSearch(1)
	m.moveSearch(1)
	if got, want := focusedPath(m), "input.nested.x"; got != want {
		t.Fatalf("focused path after next matches = %q, want %q", got, want)
	}

	m.moveSearch(1)
	if got, want := m.searchIndex, 0; got != want {
		t.Fatalf("wrapped current match = %d, want %d", got, want)
	}
	m.moveSearch(-1)
	if got, want := m.searchIndex, 3; got != want {
		t.Fatalf("reverse wrapped current match = %d, want %d", got, want)
	}
}

func TestSearchRevealsMatchInsideCollapsedContainer(t *testing.T) {
	m := newSearchTestModel(t)
	nested := m.Doc.Root.Children[1]
	nested.Collapsed = true
	m.refreshRows()

	m.searchInput = "needle"
	m.applySearch()
	for range 3 {
		m.moveSearch(1)
	}

	if nested.Collapsed {
		t.Fatal("search did not reveal the current match")
	}
	if got, want := focusedPath(m), "input.nested.x"; got != want {
		t.Fatalf("focused path = %q, want %q", got, want)
	}
}

func TestManualMovementKeepsHighlightsAndNextUsesCursorPosition(t *testing.T) {
	m := newSearchTestModel(t)
	m.searchInput = "needle"
	m.applySearch()

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
	m = updated.(Model)
	if !m.searchHighlight {
		t.Fatal("ordinary navigation key disabled search highlighting")
	}
	if got, want := focusedPath(m), "input.nested"; got != want {
		t.Fatalf("manually focused path = %q, want %q", got, want)
	}

	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'n', Text: "n"}))
	m = updated.(Model)
	if got, want := m.searchIndex, 3; got != want {
		t.Fatalf("current match = %d, want closest next match %d", got, want)
	}
	if got, want := focusedPath(m), "input.nested.x"; got != want {
		t.Fatalf("focused path = %q, want %q", got, want)
	}
}

func TestSearchJumpRemainsRelativeAfterCursorReturnsToLastMatch(t *testing.T) {
	m := newSearchTestModel(t)
	m.searchInput = "needle"
	m.applySearch()

	for _, r := range "jk" {
		updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)}))
		m = updated.(Model)
	}
	if got, want := focusedPath(m), "input.needle"; got != want {
		t.Fatalf("manually focused path = %q, want %q", got, want)
	}

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'n', Text: "n"}))
	m = updated.(Model)
	if got, want := m.searchIndex, 3; got != want {
		t.Fatalf("current match = %d, want closest next row's match %d", got, want)
	}
}

func TestPreviousMatchUsesManualCursorPosition(t *testing.T) {
	m := newSearchTestModel(t)
	m.searchInput = "needle"
	m.applySearch()

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'p', Text: "p"}))
	m = updated.(Model)

	// The closest previous result is the final occurrence on input.needle,
	// rather than the result before the last search jump (which would wrap).
	if got, want := m.searchIndex, 2; got != want {
		t.Fatalf("current match = %d, want closest previous match %d", got, want)
	}
	if got, want := focusedPath(m), "input.needle"; got != want {
		t.Fatalf("focused path = %q, want %q", got, want)
	}
}

func TestCtrlCQuitsWhileEditingSearch(t *testing.T) {
	m := newSearchTestModel(t)
	m.startSearch()

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	if cmd == nil {
		t.Fatal("ctrl+c did not return a quit command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c command returned %T, want tea.QuitMsg", msg)
	}
}

func TestCancelSearchPreservesPreviousHighlights(t *testing.T) {
	m := newSearchTestModel(t)
	m.searchInput = "needle"
	m.applySearch()
	previousIndex := m.searchIndex

	m.startSearch()
	m.searchInput = "replacement"
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	m = updated.(Model)

	if m.searchEditing {
		t.Fatal("escape did not close search input")
	}
	if !m.searchHighlight {
		t.Fatal("canceling search disabled previous highlights")
	}
	if got, want := m.searchQuery, "needle"; got != want {
		t.Fatalf("search query = %q, want previous query %q", got, want)
	}
	if got := m.searchIndex; got != previousIndex {
		t.Fatalf("search index = %d, want previous index %d", got, previousIndex)
	}
}

func TestPasteIntoSearchIsSingleLineAndControlSafe(t *testing.T) {
	m := newSearchTestModel(t)
	m.startSearch()

	updated, _ := m.Update(tea.PasteMsg{Content: "first\nsecond\t\x1b[31m\x00"})
	m = updated.(Model)
	if got, want := m.searchInput, "first second [31m"; got != want {
		t.Fatalf("search input = %q, want %q", got, want)
	}

	footer := m.renderFooter("status")
	if got := strings.Count(footer, "\n"); got != 1 {
		t.Fatalf("footer newline count = %d, want 1 in %q", got, footer)
	}
	if got, want := ansi.Strip(footer), "status\n/first second [31m█"; got != want {
		t.Fatalf("plain footer = %q, want %q", got, want)
	}
}

func TestSearchInputIsAppliedWithEnter(t *testing.T) {
	m := newSearchTestModel(t)
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: '/', Text: "/"}))
	m = updated.(Model)
	if !m.searchEditing {
		t.Fatal("slash did not start search input")
	}

	for _, r := range "needle" {
		updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)}))
		m = updated.(Model)
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = updated.(Model)

	if m.searchEditing {
		t.Fatal("enter did not close search input")
	}
	if got, want := m.searchQuery, "needle"; got != want {
		t.Fatalf("search query = %q, want %q", got, want)
	}
	if !m.searchHighlight {
		t.Fatal("applied search was not highlighted")
	}
}

func TestSearchIsCaseInsensitiveUnlessSensitiveSuffixIsUsed(t *testing.T) {
	doc, err := jsondoc.Parse([]byte(`{"first":"VALUE","second":"value"}`), "case.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m := New(doc)

	m.searchInput = "VALUE"
	m.applySearch()
	if got, want := len(m.searchMatches), 2; got != want {
		t.Fatalf("case-insensitive matches = %d, want %d", got, want)
	}
	if m.searchCaseSensitive {
		t.Fatal("search without /s was case sensitive")
	}

	m.searchInput = "VALUE/s"
	m.applySearch()
	if got, want := m.searchQuery, "VALUE"; got != want {
		t.Fatalf("case-sensitive query = %q, want %q", got, want)
	}
	if got, want := len(m.searchMatches), 1; got != want {
		t.Fatalf("case-sensitive matches = %d, want %d", got, want)
	}
	if !m.searchCaseSensitive {
		t.Fatal("search ending in /s was not case sensitive")
	}
}

func TestSearchMatchRanges(t *testing.T) {
	got := searchMatchRanges("éÉ", compileSearchPattern("é", false))
	want := [][2]int{{0, 2}, {2, 4}}
	if len(got) != len(want) {
		t.Fatalf("matches = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("match %d = %v, want %v", i, got[i], want[i])
		}
	}
}
