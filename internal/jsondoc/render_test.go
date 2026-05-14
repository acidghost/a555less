package jsondoc

import (
	"strconv"
	"testing"
)

func TestFormatKey(t *testing.T) {
	tests := map[string]string{
		"abc":       "abc",
		"_abc123":   "_abc123",
		"$schema":   "$schema",
		"weird key": strconv.Quote("weird key"),
		"123abc":    strconv.Quote("123abc"),
		"é":         strconv.Quote("é"),
	}

	for key, want := range tests {
		if got := FormatKey(key); got != want {
			t.Errorf("FormatKey(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestFormatPrimitive(t *testing.T) {
	doc, err := Parse([]byte(`{"s":"a\nb","n":1e-9,"t":true,"f":false,"z":null}`), "primitive.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	want := []string{`"a\nb"`, "1e-9", "true", "false", "null"}
	for i, child := range doc.Root.Children {
		if got := FormatPrimitive(child); got != want[i] {
			t.Errorf("FormatPrimitive(child %d) = %q, want %q", i, got, want[i])
		}
	}
}

func TestPreview(t *testing.T) {
	doc, err := Parse([]byte(`{"id":1,"name":"alice","tags":["go","json"],"meta":{"active":true},"extra":null}`), "preview.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got, want := Preview(doc.Root, 5, 1), `{id: 1, name: "alice", tags: […], meta: {…}, extra: null}`; got != want {
		t.Fatalf("Preview() = %q, want %q", got, want)
	}
	if got, want := Preview(doc.Root.Children[2], 5, 1), `["go", "json"]`; got != want {
		t.Fatalf("Preview(tags) = %q, want %q", got, want)
	}
	if got, want := Preview(doc.Root, 2, 1), `{id: 1, name: "alice", …}`; got != want {
		t.Fatalf("Preview(maxItems) = %q, want %q", got, want)
	}
}
