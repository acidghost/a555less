package jsondoc

import "testing"

func TestPrettyValue(t *testing.T) {
	doc, err := Parse([]byte(`{"z":[1,{"name":"Ada"}],"empty":{},"enabled":true,"nothing":null}`), "value.json")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	want := `{
  "z": [
    1,
    {
      "name": "Ada"
    }
  ],
  "empty": {},
  "enabled": true,
  "nothing": null
}`
	if got := PrettyValue(doc.Root); got != want {
		t.Fatalf("PrettyValue() =\n%s\nwant:\n%s", got, want)
	}

	if got, want := PrettyValue(doc.Root.Children[0]), "[\n  1,\n  {\n    \"name\": \"Ada\"\n  }\n]"; got != want {
		t.Fatalf("PrettyValue(child) = %q, want %q", got, want)
	}
	if got, want := PrettyValue(doc.Root.Children[2]), "true"; got != want {
		t.Fatalf("PrettyValue(primitive) = %q, want %q", got, want)
	}
}

func TestPrintableStringContents(t *testing.T) {
	input := "quote: \" slash: \\ euro: € controls: \n\t\x1b\x7f\u0080"
	want := `quote: " slash: \ euro: € controls: \n\t\u001B\u007F\u0080`
	if got := PrintableStringContents(input); got != want {
		t.Fatalf("PrintableStringContents() = %q, want %q", got, want)
	}
}
