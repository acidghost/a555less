package jsondoc

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PrettyValue returns n as indented JSON while preserving object key order and
// the original representation of numbers.
func PrettyValue(n *Node) string {
	var out strings.Builder
	writePrettyValue(&out, n, 0)
	return out.String()
}

func writePrettyValue(out *strings.Builder, n *Node, depth int) {
	if n == nil {
		out.WriteString("null")
		return
	}

	switch n.Kind {
	case KindNull:
		out.WriteString("null")
	case KindBool:
		if n.Bool {
			out.WriteString("true")
		} else {
			out.WriteString("false")
		}
	case KindNumber:
		out.WriteString(n.Number.String())
	case KindString:
		out.WriteString(jsonString(n.String))
	case KindObject:
		writePrettyObject(out, n, depth)
	case KindArray:
		writePrettyArray(out, n, depth)
	}
}

func writePrettyObject(out *strings.Builder, n *Node, depth int) {
	if len(n.Children) == 0 {
		out.WriteString("{}")
		return
	}

	out.WriteString("{\n")
	for i, child := range n.Children {
		writeIndent(out, depth+1)
		out.WriteString(jsonString(child.Key))
		out.WriteString(": ")
		writePrettyValue(out, child, depth+1)
		if i < len(n.Children)-1 {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	writeIndent(out, depth)
	out.WriteByte('}')
}

func writePrettyArray(out *strings.Builder, n *Node, depth int) {
	if len(n.Children) == 0 {
		out.WriteString("[]")
		return
	}

	out.WriteString("[\n")
	for i, child := range n.Children {
		writeIndent(out, depth+1)
		writePrettyValue(out, child, depth+1)
		if i < len(n.Children)-1 {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	writeIndent(out, depth)
	out.WriteByte(']')
}

func writeIndent(out *strings.Builder, depth int) {
	out.WriteString(strings.Repeat("  ", depth))
}

func jsonString(s string) string {
	encoded, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(encoded)
}

// PrintableStringContents returns decoded string contents without surrounding
// quotes. Control characters stay escaped so printing cannot affect terminal
// state or layout.
func PrintableStringContents(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch r {
		case '\b':
			out.WriteString(`\b`)
		case '\f':
			out.WriteString(`\f`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
				_, _ = fmt.Fprintf(&out, `\u%04X`, r)
			} else {
				out.WriteRune(r)
			}
		}
	}
	return out.String()
}
