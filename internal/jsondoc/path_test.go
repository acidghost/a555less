package jsondoc

import "testing"

func TestPath(t *testing.T) {
	doc, err := Parse([]byte(`{"users":[{"name":"alice","weird key":null}]}`), "path.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	users := doc.Root.Children[0]
	firstUser := users.Children[0]
	name := firstUser.Children[0]
	weird := firstUser.Children[1]

	tests := []struct {
		n    *Node
		want string
	}{
		{doc.Root, "input"},
		{users, "input.users"},
		{firstUser, "input.users[0]"},
		{name, "input.users[0].name"},
		{weird, `input.users[0]["weird key"]`},
		{nil, "input"},
	}

	for _, tt := range tests {
		if got := Path(tt.n); got != tt.want {
			t.Errorf("Path() = %q, want %q", got, tt.want)
		}
	}

	if got, want := QueryPath(doc, name), ".users[].name"; got != want {
		t.Errorf("QueryPath() = %q, want %q", got, want)
	}
	if got, want := QueryPath(doc, weird), `.users[]["weird key"]`; got != want {
		t.Errorf("QueryPath() = %q, want %q", got, want)
	}
	if got, want := QueryPath(doc, doc.Root), "."; got != want {
		t.Errorf("QueryPath(root) = %q, want %q", got, want)
	}
}

func TestQueryPathQuotesJQVariables(t *testing.T) {
	doc, err := Parse([]byte(`{"$foo":1,"normal_key":2}`), "path.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got, want := QueryPath(doc, doc.Root.Children[0]), `.["$foo"]`; got != want {
		t.Errorf("QueryPath($foo) = %q, want %q", got, want)
	}
	if got, want := QueryPath(doc, doc.Root.Children[1]), `.normal_key`; got != want {
		t.Errorf("QueryPath(normal_key) = %q, want %q", got, want)
	}
}

func TestQueryPathOmitsSyntheticJSONLRoot(t *testing.T) {
	doc, err := ParseJSONL([]byte("{\"field\":[1]}\n{\"field\":[2]}\n"), "events.jsonl")
	if err != nil {
		t.Fatalf("ParseJSONL() error = %v", err)
	}

	record := doc.Root.Children[0]
	field := record.Children[0]
	item := field.Children[0]
	if got, want := QueryPath(doc, record), "."; got != want {
		t.Errorf("QueryPath(record) = %q, want %q", got, want)
	}
	if got, want := QueryPath(doc, field), ".field"; got != want {
		t.Errorf("QueryPath(field) = %q, want %q", got, want)
	}
	if got, want := QueryPath(doc, item), ".field[]"; got != want {
		t.Errorf("QueryPath(item) = %q, want %q", got, want)
	}
}
