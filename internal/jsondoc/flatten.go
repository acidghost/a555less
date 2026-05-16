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
	appendVisibleRows(&rows, root, 0)
	return rows
}

// VisibleRowsForDocument flattens doc into visible rows. JSONL documents hide
// their synthetic root array and show each item as a top-level row.
func VisibleRowsForDocument(doc *Document) []Row {
	if doc == nil || doc.Root == nil {
		return nil
	}
	if !doc.JSONL {
		return VisibleRows(doc.Root)
	}

	rows := []Row{}
	for _, child := range doc.Root.Children {
		appendVisibleRows(&rows, child, 0)
	}
	return rows
}

func appendVisibleRows(rows *[]Row, n *Node, depth int) {
	*rows = append(*rows, Row{Node: n, Depth: depth})
	if n.IsContainer() && !n.Collapsed {
		for _, child := range n.Children {
			appendVisibleRows(rows, child, depth+1)
		}
	}
}
