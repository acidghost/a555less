package jsondoc

// Row is a visible data-mode tree row.
type Row struct {
	Node  *Node
	Depth int
}

// VisibleRows flattens root into visible rows, omitting descendants of
// collapsed containers.
func VisibleRows(root *Node) []Row {
	if root == nil {
		return nil
	}

	rows := []Row{}
	var walk func(*Node, int)
	walk = func(n *Node, depth int) {
		rows = append(rows, Row{Node: n, Depth: depth})
		if n.IsContainer() && !n.Collapsed {
			for _, child := range n.Children {
				walk(child, depth+1)
			}
		}
	}
	walk(root, 0)
	return rows
}
