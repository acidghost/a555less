package tui

import (
	"testing"

	"github.com/acidghost/a555less/internal/jsondoc"
)

func newTestModel(t *testing.T) Model {
	t.Helper()
	doc, err := jsondoc.Parse([]byte(`{"a":{"b":1},"c":[2,3],"d":4,"e":5,"f":6}`), "nav.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m := New(doc)
	m.width = 80
	m.height = 7 // 5 viewer rows + status + help.
	m.ensureVisible()
	return m
}

func focusedPath(m Model) string {
	return jsondoc.Path(m.focusedNode())
}

func TestMoveAndEnsureVisible(t *testing.T) {
	m := newTestModel(t)
	if got, want := focusedPath(m), "input"; got != want {
		t.Fatalf("initial focus = %q, want %q", got, want)
	}

	m.move(2)
	if got, want := focusedPath(m), "input.a.b"; got != want {
		t.Fatalf("focus after move = %q, want %q", got, want)
	}

	m.move(99)
	if got, want := focusedPath(m), "input.f"; got != want {
		t.Fatalf("focus after large move = %q, want %q", got, want)
	}
	if m.top == 0 {
		t.Fatal("top was not scrolled after moving to bottom")
	}
}

func TestToggleCollapseExpand(t *testing.T) {
	m := newTestModel(t)
	m.move(1) // input.a

	m.toggle()
	if !m.Doc.Root.Children[0].Collapsed {
		t.Fatal("input.a was not collapsed")
	}
	rows := jsondoc.VisibleRows(m.Doc.Root)
	if len(rows) != 8 {
		t.Fatalf("collapsed rows len = %d, want 8", len(rows))
	}

	m.toggle()
	if m.Doc.Root.Children[0].Collapsed {
		t.Fatal("input.a was not expanded")
	}
}

func TestLeftRightNavigation(t *testing.T) {
	m := newTestModel(t)
	m.move(1) // input.a

	m.left()
	if !m.Doc.Root.Children[0].Collapsed {
		t.Fatal("left did not collapse expanded container")
	}
	if got, want := focusedPath(m), "input.a"; got != want {
		t.Fatalf("focus after collapse = %q, want %q", got, want)
	}

	m.left()
	if got, want := focusedPath(m), "input"; got != want {
		t.Fatalf("second left focus = %q, want %q", got, want)
	}

	m.right()
	if got, want := focusedPath(m), "input.a"; got != want {
		t.Fatalf("right from root focus = %q, want %q", got, want)
	}

	m.right()
	if m.Doc.Root.Children[0].Collapsed {
		t.Fatal("right did not expand collapsed container")
	}
	m.right()
	if got, want := focusedPath(m), "input.a.b"; got != want {
		t.Fatalf("right into expanded container = %q, want %q", got, want)
	}
}

func TestTopBottomPageAndSiblings(t *testing.T) {
	m := newTestModel(t)

	m.bottomRow()
	if got, want := focusedPath(m), "input.f"; got != want {
		t.Fatalf("bottom focus = %q, want %q", got, want)
	}

	m.topRow()
	if got, want := focusedPath(m), "input"; got != want {
		t.Fatalf("top focus = %q, want %q", got, want)
	}

	m.page(1)
	if got, want := focusedPath(m), "input.c[1]"; got != want {
		t.Fatalf("page down focus = %q, want %q", got, want)
	}

	m.focusParent()
	if got, want := focusedPath(m), "input.c"; got != want {
		t.Fatalf("parent focus = %q, want %q", got, want)
	}

	m.focusSibling(1)
	if got, want := focusedPath(m), "input.d"; got != want {
		t.Fatalf("next sibling focus = %q, want %q", got, want)
	}

	m.focusSibling(-1)
	if got, want := focusedPath(m), "input.c"; got != want {
		t.Fatalf("previous sibling focus = %q, want %q", got, want)
	}
}

func TestCollapseExpandFocusedSiblings(t *testing.T) {
	m := newTestModel(t)
	m.move(1) // input.a

	m.collapseFocusedSiblings(false)
	if !m.Doc.Root.Children[0].Collapsed {
		t.Fatal("input.a was not collapsed")
	}
	if !m.Doc.Root.Children[1].Collapsed {
		t.Fatal("input.c sibling was not collapsed")
	}
	if m.Doc.Root.Collapsed {
		t.Fatal("root should not be collapsed when focused node has siblings")
	}
	if got, want := focusedPath(m), "input.a"; got != want {
		t.Fatalf("focus after sibling collapse = %q, want %q", got, want)
	}

	m.expandFocusedSiblings(false)
	if m.Doc.Root.Children[0].Collapsed || m.Doc.Root.Children[1].Collapsed {
		t.Fatal("container siblings were not expanded")
	}
}

func TestDeepCollapseExpandFocusedSiblings(t *testing.T) {
	doc, err := jsondoc.Parse([]byte(`{"a":{"b":{"c":1}},"d":[{"e":2}]}`), "nav.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	m := New(doc)
	m.width = 80
	m.height = 10
	m.move(1) // input.a

	m.collapseFocusedSiblings(true)
	if !doc.Root.Children[0].Collapsed {
		t.Fatal("input.a was not collapsed")
	}
	if !doc.Root.Children[0].Children[0].Collapsed {
		t.Fatal("nested input.a.b was not collapsed")
	}
	if !doc.Root.Children[1].Collapsed {
		t.Fatal("input.d sibling was not collapsed")
	}
	if !doc.Root.Children[1].Children[0].Collapsed {
		t.Fatal("nested input.d[0] was not collapsed")
	}
	if doc.Root.Collapsed {
		t.Fatal("root should not be collapsed when focused node has siblings")
	}

	m.expandFocusedSiblings(true)
	if doc.Root.Children[0].Collapsed || doc.Root.Children[0].Children[0].Collapsed || doc.Root.Children[1].Collapsed || doc.Root.Children[1].Children[0].Collapsed {
		t.Fatal("containers below root were not deeply expanded")
	}
}
