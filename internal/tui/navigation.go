package tui

import "github.com/acidghost/a555less/internal/jsondoc"

const defaultScrolloff = 2

func (m *Model) move(delta int) {
	if len(m.rows) == 0 {
		return
	}
	i := m.focusedIndex()
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
	if n == nil || n.Parent == nil || m.isHiddenRoot(n.Parent) {
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

func (m *Model) collapseFocusedSiblings(deep bool) {
	m.setFocusedSiblingsCollapseState(true, deep)
}

func (m *Model) expandFocusedSiblings(deep bool) {
	m.setFocusedSiblingsCollapseState(false, deep)
}

func (m *Model) setFocusedSiblingsCollapseState(collapsed, deep bool) {
	n := m.focusedNode()
	if n == nil {
		return
	}

	siblings := []*jsondoc.Node{n}
	if n.Parent != nil {
		siblings = n.Parent.Children
	}

	for _, sibling := range siblings {
		setCollapseState(sibling, collapsed, deep)
	}
	m.refreshRows()
	m.ensureVisible()
}

func (m *Model) isHiddenRoot(n *jsondoc.Node) bool {
	return m.Doc != nil && m.Doc.JSONL && n == m.Doc.Root
}

func setCollapseState(n *jsondoc.Node, collapsed, deep bool) {
	if n == nil {
		return
	}
	if n.IsContainer() {
		n.Collapsed = collapsed
	}
	if !deep {
		return
	}
	for _, child := range n.Children {
		setCollapseState(child, collapsed, true)
	}
}

// focusedNode returns the node currently identified by m.focusID.
// If the cached rows are empty, it returns nil. If m.focusID is stale,
// focusIndex repairs it to the first visible row before returning that node.
func (m *Model) focusedNode() *jsondoc.Node {
	if len(m.rows) == 0 {
		return nil
	}
	i := m.focusedIndex()
	return m.rows[i].Node
}

// focusedIndex returns the index of m.focusID within rows.
// If rows is empty, it returns -1. If m.focusID is not present, it resets
// focus to the first row and returns 0.
func (m *Model) focusedIndex() int {
	if len(m.rows) == 0 {
		return -1
	}
	i := indexOfNodeID(m.rows, m.focusID)
	if i < 0 {
		m.focusID = m.rows[0].Node.ID
		return 0
	}
	return i
}

// ensureVisible adjusts m.top so the focused row is visible in the viewer.
// It keeps up to defaultScrolloff rows of context around the focus when
// possible, clamps m.top to the valid scroll range, and resets focus/scroll
// state when there are no visible rows.
func (m *Model) ensureVisible() {
	if len(m.rows) == 0 {
		m.focusID = -1
		m.top = 0
		return
	}

	viewerHeight := m.viewerHeight()
	maxTop := max(0, len(m.rows)-viewerHeight)
	focusIdx := m.focusedIndex()

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
	if m.search.editing() || m.printPending || m.message != "" {
		footerHeight = 2 // status line + one-line prompt or message.
	}
	return max(1, m.height-footerHeight)
}
