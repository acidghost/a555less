package jsondoc

import (
	"strconv"
	"strings"
)

// Path returns an informational path to n using input as the document root.
func Path(n *Node) string {
	if n == nil {
		return "input"
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

	for i, j := 0, len(components)-1; i < j; i, j = i+1, j-1 {
		components[i], components[j] = components[j], components[i]
	}

	return "input" + strings.Join(components, "")
}
