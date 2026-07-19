package jsondoc

import (
	"strconv"
	"strings"
)

// IsIdentifier reports whether s can be rendered as an unquoted JavaScript-like
// identifier in object labels and paths.
func IsIdentifier(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		if i == 0 {
			if !isIdentifierStart(r) {
				return false
			}
			continue
		}
		if !isIdentifierPart(r) {
			return false
		}
	}
	return true
}

func isIdentifierStart(r rune) bool {
	return r == '_' || r == '$' || ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isIdentifierPart(r rune) bool {
	return isIdentifierStart(r) || ('0' <= r && r <= '9')
}

// FormatKey formats an object key for display.
func FormatKey(key string) string {
	if IsIdentifier(key) {
		return key
	}
	return strconv.Quote(key)
}

// FormatPrimitive formats a primitive JSON node for display.
func FormatPrimitive(n *Node) string {
	if n == nil {
		return "null"
	}

	switch n.Kind {
	case KindNull:
		return "null"
	case KindBool:
		if n.Bool {
			return "true"
		}
		return "false"
	case KindNumber:
		return n.Number.String()
	case KindString:
		return strconv.Quote(n.String)
	default:
		return ""
	}
}

// Preview returns a compact, unstyled preview for n.
func Preview(n *Node, maxItems int, maxDepth int) string {
	if n == nil {
		return "null"
	}
	if !n.IsContainer() {
		return FormatPrimitive(n)
	}
	if maxDepth <= 0 {
		switch n.Kind {
		case KindObject:
			return "{…}"
		case KindArray:
			return "[…]"
		}
	}

	switch n.Kind {
	case KindObject:
		return previewContainer(n, maxItems, maxDepth, "{", "}")
	case KindArray:
		return previewContainer(n, maxItems, maxDepth, "[", "]")
	default:
		return FormatPrimitive(n)
	}
}

func previewContainer(n *Node, maxItems int, maxDepth int, open string, closeDelim string) string {
	if len(n.Children) == 0 {
		return open + closeDelim
	}
	if maxItems < 0 {
		maxItems = 0
	}

	var out strings.Builder
	out.WriteString(open)
	limit := min(len(n.Children), maxItems)
	for i := range limit {
		if i > 0 {
			out.WriteString(", ")
		}
		child := n.Children[i]
		if n.Kind == KindObject {
			out.WriteString(FormatKey(child.Key))
			out.WriteString(": ")
		}
		out.WriteString(Preview(child, maxItems, maxDepth-1))
	}
	if limit < len(n.Children) {
		if limit > 0 {
			out.WriteString(", ")
		}
		out.WriteString("…")
	}
	out.WriteString(closeDelim)
	return out.String()
}
