package tui

import "github.com/acidghost/a555less/internal/jsondoc"

const defaultScrolloff = 2

func (m *Model) move(delta int) {
	rows := m.visibleRows()
	if len(rows) == 0 {
		return
	}
	i := m.focusIndex(rows)
	i = clamp(i+delta, len(rows)-1)
	m.focusID = rows[i].Node.ID
	m.ensureVisibleRows(rows)
}

func (m *Model) page(delta int) {
	step := m.viewerHeight()
	if step < 1 {
		step = 1
	}
	m.move(delta * step)
}

func (m *Model) halfPage(delta int) {
	step := m.viewerHeight() / 2
	if step < 1 {
		step = 1
	}
	m.move(delta * step)
}

func (m *Model) toggle() {
	n := m.focusedNode()
	if n == nil || !n.IsContainer() {
		return
	}
	n.Collapsed = !n.Collapsed
	m.ensureVisible()
}

func (m *Model) left() {
	n := m.focusedNode()
	if n == nil {
		return
	}
	if n.IsContainer() && !n.Collapsed {
		n.Collapsed = true
		m.ensureVisible()
		return
	}
	m.focusParent()
}

func (m *Model) right() {
	n := m.focusedNode()
	if n == nil || !n.IsContainer() {
		return
	}
	if n.Collapsed {
		n.Collapsed = false
		m.ensureVisible()
		return
	}
	if len(n.Children) > 0 {
		m.focusID = n.Children[0].ID
		m.ensureVisible()
	}
}

func (m *Model) topRow() {
	rows := m.visibleRows()
	if len(rows) == 0 {
		return
	}
	m.focusID = rows[0].Node.ID
	m.ensureVisibleRows(rows)
}

func (m *Model) bottomRow() {
	rows := m.visibleRows()
	if len(rows) == 0 {
		return
	}
	m.focusID = rows[len(rows)-1].Node.ID
	m.ensureVisibleRows(rows)
}

func (m *Model) focusParent() {
	n := m.focusedNode()
	if n == nil || n.Parent == nil {
		return
	}
	m.focusID = n.Parent.ID
	m.ensureVisible()
}

func (m *Model) focusSibling(delta int) {
	n := m.focusedNode()
	if n == nil || n.Parent == nil {
		return
	}
	siblings := n.Parent.Children
	idx := -1
	for i, sibling := range siblings {
		if sibling == n {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	idx = clamp(idx+delta, len(siblings)-1)
	m.focusID = siblings[idx].ID
	m.ensureVisible()
}

func (m *Model) visibleRows() []jsondoc.Row {
	if m.Doc == nil || m.Doc.Root == nil {
		return nil
	}
	return jsondoc.VisibleRows(m.Doc.Root)
}

func (m *Model) focusedNode() *jsondoc.Node {
	rows := m.visibleRows()
	if len(rows) == 0 {
		return nil
	}
	i := m.focusIndex(rows)
	return rows[i].Node
}

func (m *Model) focusIndex(rows []jsondoc.Row) int {
	if len(rows) == 0 {
		return -1
	}
	i := indexOfNodeID(rows, m.focusID)
	if i < 0 {
		m.focusID = rows[0].Node.ID
		return 0
	}
	return i
}

func (m *Model) ensureVisible() {
	m.ensureVisibleRows(m.visibleRows())
}

func (m *Model) ensureVisibleRows(rows []jsondoc.Row) {
	if len(rows) == 0 {
		m.focusID = -1
		m.top = 0
		return
	}

	viewerHeight := m.viewerHeight()
	maxTop := max(0, len(rows)-viewerHeight)
	focusIdx := m.focusIndex(rows)

	m.top = clamp(m.top, maxTop)
	so := min(defaultScrolloff, max(0, (viewerHeight-1)/2))

	if focusIdx < m.top+so {
		m.top = focusIdx - so
	}
	if focusIdx >= m.top+viewerHeight-so {
		m.top = focusIdx - viewerHeight + so + 1
	}
	m.top = clamp(m.top, maxTop)
}

func (m Model) viewerHeight() int {
	if m.height <= 0 {
		return max(1, len(m.visibleRows()))
	}
	return max(1, m.height-1)
}
