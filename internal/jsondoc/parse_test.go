package jsondoc

import (
	"strings"
	"testing"
)

func TestParsePreservesObjectKeyOrder(t *testing.T) {
	doc, err := Parse([]byte(`{"b":1,"a":2,"c":3}`), "order.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if doc.Root.Kind != KindObject {
		t.Fatalf("root kind = %v, want object", doc.Root.Kind)
	}

	want := []string{"b", "a", "c"}
	if len(doc.Root.Children) != len(want) {
		t.Fatalf("children len = %d, want %d", len(doc.Root.Children), len(want))
	}
	for i, key := range want {
		child := doc.Root.Children[i]
		if child.Key != key {
			t.Errorf("child %d key = %q, want %q", i, child.Key, key)
		}
		if child.Index != i {
			t.Errorf("child %d index = %d, want %d", i, child.Index, i)
		}
	}
}

func TestParsePreservesDuplicateObjectKeys(t *testing.T) {
	doc, err := Parse([]byte(`{"a":1,"a":2}`), "dup.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	children := doc.Root.Children
	if len(children) != 2 {
		t.Fatalf("children len = %d, want 2", len(children))
	}
	for i, child := range children {
		if child.Key != "a" {
			t.Errorf("child %d key = %q, want a", i, child.Key)
		}
		if child.Kind != KindNumber {
			t.Errorf("child %d kind = %v, want number", i, child.Kind)
		}
	}
	if children[0].Number.String() != "1" || children[1].Number.String() != "2" {
		t.Fatalf("numbers = %q, %q; want 1, 2", children[0].Number.String(), children[1].Number.String())
	}
}

func TestParseNestedParentAndIndexFields(t *testing.T) {
	doc, err := Parse([]byte(`{"users":[{"name":"alice"},{"name":"bob"}]}`), "nested.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	users := doc.Root.Children[0]
	if users.Parent != doc.Root {
		t.Fatal("users parent is not root")
	}
	if !users.HasKey || users.Key != "users" || users.Index != 0 {
		t.Fatalf("users metadata = hasKey:%v key:%q index:%d, want true/users/0", users.HasKey, users.Key, users.Index)
	}

	secondUser := users.Children[1]
	if secondUser.Parent != users {
		t.Fatal("second user parent is not users array")
	}
	if secondUser.HasKey || secondUser.Index != 1 {
		t.Fatalf("second user metadata = hasKey:%v index:%d, want false/1", secondUser.HasKey, secondUser.Index)
	}

	name := secondUser.Children[0]
	if name.Parent != secondUser {
		t.Fatal("name parent is not second user")
	}
	if !name.HasKey || name.Key != "name" || name.Index != 0 {
		t.Fatalf("name metadata = hasKey:%v key:%q index:%d, want true/name/0", name.HasKey, name.Key, name.Index)
	}
	if name.Kind != KindString || name.String != "bob" {
		t.Fatalf("name value = kind:%v string:%q, want string bob", name.Kind, name.String)
	}
}

func TestParsePreservesNumberSpelling(t *testing.T) {
	doc, err := Parse([]byte(`{"x":1e-9,"y":12345678901234567890}`), "numbers.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	x := doc.Root.Children[0]
	y := doc.Root.Children[1]
	if x.Number.String() != "1e-9" {
		t.Fatalf("x number = %q, want 1e-9", x.Number.String())
	}
	if y.Number.String() != "12345678901234567890" {
		t.Fatalf("y number = %q, want 12345678901234567890", y.Number.String())
	}
}

func TestParseRejectsTrailingValue(t *testing.T) {
	if _, err := Parse([]byte(`{"a":1} {"b":2}`), "trailing.json"); err == nil {
		t.Fatal("Parse() error = nil, want trailing value error")
	}
}

func TestParseJSONLWrapsValuesInRootArray(t *testing.T) {
	doc, err := ParseJSONL([]byte("{\"a\":1}\n[2,3]\ntrue\n"), "events.jsonl")
	if err != nil {
		t.Fatalf("ParseJSONL() error = %v", err)
	}

	if !doc.JSONL {
		t.Fatal("doc.JSONL = false, want true")
	}
	if doc.Root.Kind != KindArray {
		t.Fatalf("root kind = %v, want array", doc.Root.Kind)
	}
	if len(doc.Root.Children) != 3 {
		t.Fatalf("children len = %d, want 3", len(doc.Root.Children))
	}
	if doc.Root.Children[0].Kind != KindObject || doc.Root.Children[1].Kind != KindArray || doc.Root.Children[2].Kind != KindBool {
		t.Fatalf("child kinds = %v, %v, %v; want object, array, bool", doc.Root.Children[0].Kind, doc.Root.Children[1].Kind, doc.Root.Children[2].Kind)
	}
	for i, child := range doc.Root.Children {
		if child.Parent != doc.Root || child.Index != i {
			t.Fatalf("child %d metadata = parent:%v index:%d, want root/%d", i, child.Parent, child.Index, i)
		}
	}
}

func TestVisibleRowsForDocumentHidesJSONLRoot(t *testing.T) {
	doc, err := ParseJSONL([]byte("{\"a\":1}\n[2]\n"), "events.jsonl")
	if err != nil {
		t.Fatalf("ParseJSONL() error = %v", err)
	}

	rows := VisibleRowsForDocument(doc)
	if len(rows) != 4 {
		t.Fatalf("rows len = %d, want 4", len(rows))
	}
	if rows[0].Node != doc.Root.Children[0] || rows[0].Depth != 0 {
		t.Fatalf("first row = node:%v depth:%d, want first child at depth 0", rows[0].Node, rows[0].Depth)
	}
	if rows[1].Node != doc.Root.Children[0].Children[0] || rows[1].Depth != 1 {
		t.Fatalf("second row = node:%v depth:%d, want nested child at depth 1", rows[1].Node, rows[1].Depth)
	}
	if rows[2].Node != doc.Root.Children[1] || rows[2].Depth != 0 {
		t.Fatalf("third row = node:%v depth:%d, want second child at depth 0", rows[2].Node, rows[2].Depth)
	}
}

func TestParseJSONLSkipsBlankLines(t *testing.T) {
	doc, err := ParseJSONL([]byte("\n  \n{\"a\":1}\n"), "events.jsonl")
	if err != nil {
		t.Fatalf("ParseJSONL() error = %v", err)
	}
	if len(doc.Root.Children) != 1 {
		t.Fatalf("children len = %d, want 1", len(doc.Root.Children))
	}
}

func TestParseJSONLErrorIncludesLine(t *testing.T) {
	_, err := ParseJSONL([]byte("{\"ok\":true}\n{bad}\n"), "events.jsonl")
	if err == nil {
		t.Fatal("ParseJSONL() error = nil, want syntax error")
	}
	if !strings.Contains(err.Error(), "line 2:") {
		t.Fatalf("ParseJSONL() error = %q, want line 2", err.Error())
	}
}

func TestParseSyntaxErrorIncludesLineAndColumn(t *testing.T) {
	_, err := Parse([]byte("{\n  \"a\": \n}"), "invalid.json")
	if err == nil {
		t.Fatal("Parse() error = nil, want syntax error")
	}
	if !strings.Contains(err.Error(), "line 2, column 8") {
		t.Fatalf("Parse() error = %q, want line and column", err.Error())
	}
}

func TestVisibleRowsSkipsCollapsedDescendants(t *testing.T) {
	doc, err := Parse([]byte(`{"a":{"b":1},"c":2}`), "visible.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	rows := VisibleRows(doc.Root)
	if len(rows) != 4 {
		t.Fatalf("expanded rows len = %d, want 4", len(rows))
	}
	if rows[0].Depth != 0 || rows[1].Depth != 1 || rows[2].Depth != 2 || rows[3].Depth != 1 {
		t.Fatalf("expanded depths = [%d %d %d %d], want [0 1 2 1]", rows[0].Depth, rows[1].Depth, rows[2].Depth, rows[3].Depth)
	}

	doc.Root.Children[0].Collapsed = true
	rows = VisibleRows(doc.Root)
	if len(rows) != 3 {
		t.Fatalf("collapsed rows len = %d, want 3", len(rows))
	}
	if rows[0].Node != doc.Root || rows[1].Node != doc.Root.Children[0] || rows[2].Node != doc.Root.Children[1] {
		t.Fatal("collapsed rows do not contain root, a, c in order")
	}
}
