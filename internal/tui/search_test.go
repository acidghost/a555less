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

func applySearchForTest(m *Model, query string) {
	m.search.begin()
	m.search.appendInput(query)
	m.applySearch()
}

func TestApplySearchFindsKeysAndValuesAndNavigates(t *testing.T) {
	m := newSearchTestModel(t)
	applySearchForTest(&m, "needle")

	if got, want := len(m.search.result.matches), 4; got != want {
		t.Fatalf("search matches = %d, want %d", got, want)
	}
	if got, want := m.search.result.current, 0; got != want {
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
	if got, want := m.search.result.current, 0; got != want {
		t.Fatalf("wrapped current match = %d, want %d", got, want)
	}
	m.moveSearch(-1)
	if got, want := m.search.result.current, 3; got != want {
		t.Fatalf("reverse wrapped current match = %d, want %d", got, want)
	}
}

func TestSearchRevealsMatchInsideCollapsedContainer(t *testing.T) {
	m := newSearchTestModel(t)
	nested := m.Doc.Root.Children[1]
	nested.Collapsed = true
	m.refreshRows()

	applySearchForTest(&m, "needle")
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
	applySearchForTest(&m, "needle")

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
	m = updated.(Model)
	if !m.search.highlighting() {
		t.Fatal("ordinary navigation key disabled search highlighting")
	}
	if got, want := focusedPath(m), "input.nested"; got != want {
		t.Fatalf("manually focused path = %q, want %q", got, want)
	}

	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'n', Text: "n"}))
	m = updated.(Model)
	if got, want := m.search.result.current, 3; got != want {
		t.Fatalf("current match = %d, want closest next match %d", got, want)
	}
	if got, want := focusedPath(m), "input.nested.x"; got != want {
		t.Fatalf("focused path = %q, want %q", got, want)
	}
}

func TestSearchJumpRemainsRelativeAfterCursorReturnsToLastMatch(t *testing.T) {
	m := newSearchTestModel(t)
	applySearchForTest(&m, "needle")

	for _, r := range "jk" {
		updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)}))
		m = updated.(Model)
	}
	if got, want := focusedPath(m), "input.needle"; got != want {
		t.Fatalf("manually focused path = %q, want %q", got, want)
	}

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'n', Text: "n"}))
	m = updated.(Model)
	if got, want := m.search.result.current, 3; got != want {
		t.Fatalf("current match = %d, want closest next row's match %d", got, want)
	}
}

func TestPreviousMatchUsesManualCursorPosition(t *testing.T) {
	m := newSearchTestModel(t)
	applySearchForTest(&m, "needle")

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'N', Text: "N"}))
	m = updated.(Model)

	// The closest previous result is the final occurrence on input.needle,
	// rather than the result before the last search jump (which would wrap).
	if got, want := m.search.result.current, 2; got != want {
		t.Fatalf("current match = %d, want closest previous match %d", got, want)
	}
	if got, want := focusedPath(m), "input.needle"; got != want {
		t.Fatalf("focused path = %q, want %q", got, want)
	}
}

func TestCursorRelativeSearchUsesDocumentOrderInsteadOfNodeIDs(t *testing.T) {
	doc, err := jsondoc.Parse([]byte(`{"first":"hit","middle":0,"last":"hit"}`), "order.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	first := doc.Root.Children[0]
	middle := doc.Root.Children[1]
	last := doc.Root.Children[2]

	// Keep IDs unique for model focus lookup while deliberately making their
	// numeric order different from document order.
	first.ID = 100
	middle.ID = 1
	last.ID = 50

	m := New(doc)
	applySearchForTest(&m, "hit")
	m.focusID = middle.ID
	m.search.markCursorMoved()
	m.moveSearch(1)
	if got, want := m.focusedNode(), last; got != want {
		t.Fatalf("next match node = %p, want document-order node %p", got, want)
	}

	m.focusID = middle.ID
	m.search.markCursorMoved()
	m.moveSearch(-1)
	if got, want := m.focusedNode(), first; got != want {
		t.Fatalf("previous match node = %p, want document-order node %p", got, want)
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
	applySearchForTest(&m, "needle")
	previousIndex := m.search.result.current

	m.startSearch()
	m.search.appendInput("replacement")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	m = updated.(Model)

	if m.search.editing() {
		t.Fatal("escape did not close search input")
	}
	if !m.search.highlighting() {
		t.Fatal("canceling search disabled previous highlights")
	}
	if got, want := m.search.query(), "needle"; got != want {
		t.Fatalf("search query = %q, want previous query %q", got, want)
	}
	if got := m.search.result.current; got != previousIndex {
		t.Fatalf("search index = %d, want previous index %d", got, previousIndex)
	}
}

func TestPasteIntoSearchIsSingleLineAndControlSafe(t *testing.T) {
	m := newSearchTestModel(t)
	m.startSearch()

	updated, _ := m.Update(tea.PasteMsg{Content: "first\nsecond\t\x1b[31m\x00"})
	m = updated.(Model)
	if got, want := m.search.input(), "first second [31m"; got != want {
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
	if !m.search.editing() {
		t.Fatal("slash did not start search input")
	}

	for _, r := range "needle" {
		updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)}))
		m = updated.(Model)
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = updated.(Model)

	if m.search.editing() {
		t.Fatal("enter did not close search input")
	}
	if got, want := m.search.query(), "needle"; got != want {
		t.Fatalf("search query = %q, want %q", got, want)
	}
	if !m.search.highlighting() {
		t.Fatal("applied search was not highlighted")
	}
}

func TestSearchIsCaseInsensitiveUnlessSensitiveSuffixIsUsed(t *testing.T) {
	doc, err := jsondoc.Parse([]byte(`{"first":"VALUE","second":"value"}`), "case.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m := New(doc)

	applySearchForTest(&m, "VALUE")
	if got, want := len(m.search.result.matches), 2; got != want {
		t.Fatalf("case-insensitive matches = %d, want %d", got, want)
	}
	if m.search.result.caseSensitive {
		t.Fatal("search without /s was case sensitive")
	}

	applySearchForTest(&m, "VALUE/s")
	if got, want := m.search.query(), "VALUE"; got != want {
		t.Fatalf("case-sensitive query = %q, want %q", got, want)
	}
	if got, want := len(m.search.result.matches), 1; got != want {
		t.Fatalf("case-sensitive matches = %d, want %d", got, want)
	}
	if !m.search.result.caseSensitive {
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
