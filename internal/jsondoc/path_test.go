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
}
