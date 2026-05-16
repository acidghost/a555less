package tui

import "github.com/acidghost/a555less/internal/jsondoc"

const defaultScrolloff = 2

func (m *Model) move(delta int) {
	if len(m.rows) == 0 {
		return
	}
	i := m.focusIndex(m.rows)
	i = clamp(i+delta, len(m.rows)-1)
	m.focusID = m.rows[i].Node.ID
	m.ensureVisible()
}

func (m *Model) page(delta int) {
	step := max(m.viewerHeight(), 1)
	m.move(delta * step)
}

func (m *Model) halfPage(delta int) {
	step := max(m.viewerHeight()/2, 1)
	m.move(delta * step)
}

func (m *Model) toggle() {
	n := m.focusedNode()
	if n == nil || !n.IsContainer() {
		return
	}
	n.Collapsed = !n.Collapsed
	m.refreshRows()
	m.ensureVisible()
}

func (m *Model) left() {
	n := m.focusedNode()
	if n == nil {
		return
	}
	if n.IsContainer() && !n.Collapsed {
		n.Collapsed = true
		m.refreshRows()
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
		m.refreshRows()
		m.ensureVisible()
		return
	}
	if len(n.Children) > 0 {
		m.focusID = n.Children[0].ID
		m.ensureVisible()
	}
}

func (m *Model) topRow() {
	if len(m.rows) == 0 {
		return
	}
	m.focusID = m.rows[0].Node.ID
	m.ensureVisible()
}

func (m *Model) bottomRow() {
	if len(m.rows) == 0 {
		return
	}
	m.focusID = m.rows[len(m.rows)-1].Node.ID
	m.ensureVisible()
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

// focusedNode  TODO docs
func (m *Model) focusedNode() *jsondoc.Node {
	if len(m.rows) == 0 {
		return nil
	}
	i := m.focusIndex(m.rows)
	return m.rows[i].Node
}

// focusIndex TODO docs
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

// ensureVisible TODO docs
func (m *Model) ensureVisible() {
	if len(m.rows) == 0 {
		m.focusID = -1
		m.top = 0
		return
	}

	viewerHeight := m.viewerHeight()
	maxTop := max(0, len(m.rows)-viewerHeight)
	focusIdx := m.focusIndex(m.rows)

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
		return max(1, len(m.rows))
	}
	footerHeight := 1 + m.helpHeight // status line + cached help view height.
	return max(1, m.height-footerHeight)
}
