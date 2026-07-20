package jsondoc

import (
	"strconv"
	"strings"
)

// Path returns an informational path to n using input as the document root.
func Path(n *Node) string {
	return "input" + exactPath(n)
}

func exactPath(n *Node) string {
	if n == nil {
		return ""
	}

	components := []string{}
	for cur := n; cur != nil && cur.Parent != nil; cur = cur.Parent {
		if cur.HasKey {
			if IsIdentifier(cur.Key) {
				components = append(components, "."+cur.Key)
			} else {
				components = append(components, "["+strconv.Quote(cur.Key)+"]")
			}
			continue
		}
		components = append(components, "["+strconv.Itoa(cur.Index)+"]")
	}

	reverseStrings(components)
	return strings.Join(components, "")
}

// QueryPath returns a jq filter for the route to n. Array indexes are rendered
// as [] so the filter visits the corresponding value in every array element.
// The synthetic array used to represent JSONL records is omitted because jq
// already evaluates the filter once for each top-level input.
func QueryPath(doc *Document, n *Node) string {
	if n == nil || n.Parent == nil {
		return "."
	}

	components := []string{}
	for cur := n; cur != nil && cur.Parent != nil; cur = cur.Parent {
		if doc != nil && doc.JSONL && cur.Parent == doc.Root {
			continue
		}
		switch {
		case cur.HasKey && isJQIdentifier(cur.Key):
			components = append(components, "."+cur.Key)
		case cur.HasKey:
			components = append(components, "["+jsonString(cur.Key)+"]")
		default:
			components = append(components, "[]")
		}
	}

	reverseStrings(components)
	path := strings.Join(components, "")
	if !strings.HasPrefix(path, ".") {
		path = "." + path
	}
	return path
}

func isJQIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if r != '_' && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
				return false
			}
			continue
		}
		if r != '_' && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func reverseStrings(values []string) {
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
}
