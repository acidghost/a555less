package tui

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/acidghost/a555less/internal/jsondoc"
)

// Model is the Bubble Tea application model.
type Model struct {
	Doc *jsondoc.Document

	width      int
	height     int
	focusID    int
	top        int
	rows       []jsondoc.Row
	help       help.Model
	helpView   string
	helpHeight int
}

// New returns a skeleton TUI model for doc.
func New(doc *jsondoc.Document) Model {
	m := Model{Doc: doc, focusID: -1, help: help.New()}
	m.refreshHelp()
	m.refreshRows()
	if doc != nil && doc.Root != nil {
		m.focusID = doc.Root.ID
	}
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.SetWidth(msg.Width)
		m.refreshHelp()
		m.ensureVisible()
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.refreshHelp()
			m.ensureVisible()
		case key.Matches(msg, keys.Down):
			m.move(1)
		case key.Matches(msg, keys.Up):
			m.move(-1)
		case key.Matches(msg, keys.Toggle):
			m.toggle()
		case key.Matches(msg, keys.Left):
			m.left()
		case key.Matches(msg, keys.Right):
			m.right()
		case key.Matches(msg, keys.Top):
			m.topRow()
		case key.Matches(msg, keys.Bottom):
			m.bottomRow()
		case key.Matches(msg, keys.PageDown):
			m.page(1)
		case key.Matches(msg, keys.PageUp):
			m.page(-1)
		case key.Matches(msg, keys.HalfDown):
			m.halfPage(1)
		case key.Matches(msg, keys.HalfUp):
			m.halfPage(-1)
		case key.Matches(msg, keys.Parent):
			m.focusParent()
		case key.Matches(msg, keys.NextSib):
			m.focusSibling(1)
		case key.Matches(msg, keys.PrevSib):
			m.focusSibling(-1)
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeNone
	return view
}

func (m Model) render() string {
	if m.Doc == nil || m.Doc.Root == nil {
		return m.renderEmpty()
	}

	focusIdx := indexOfNodeID(m.rows, m.focusID)
	if focusIdx < 0 && len(m.rows) > 0 {
		focusIdx = 0
	}

	focusedNode := m.Doc.Root
	if focusIdx >= 0 && focusIdx < len(m.rows) {
		focusedNode = m.rows[focusIdx].Node
	}

	status := m.renderStatus(jsondoc.Path(focusedNode), focusIdx)
	footer := m.renderFooter(status)

	viewerHeight := m.viewerHeight()

	start := clamp(m.top, max(0, len(m.rows)-viewerHeight))
	lines := make([]string, 0, viewerHeight+1)
	for i := range viewerHeight {
		rowIdx := start + i
		if rowIdx < len(m.rows) {
			row := m.rows[rowIdx]
			lines = append(lines, renderRow(row, row.Node.ID == m.focusID, m.width))
		} else {
			lines = append(lines, fillerRow(m.width))
		}
	}

	return strings.Join(lines, "\n") + "\n" + footer
}

func (m Model) renderEmpty() string {
	status := m.renderStatus("No JSON document loaded.", 0)
	footer := m.renderFooter(status)
	h := m.viewerHeight()
	lines := make([]string, 0, h+2)
	for range h {
		lines = append(lines, fillerRow(m.width))
	}
	return strings.Join(lines, "\n") + "\n" + footer
}

func (m *Model) refreshHelp() {
	m.helpView = m.help.View(keys)
	if m.helpView == "" {
		m.helpHeight = 0
		return
	}
	m.helpHeight = lipgloss.Height(m.helpView)
}

func (m *Model) refreshRows() {
	if m.Doc == nil || m.Doc.Root == nil {
		m.rows = nil
		return
	}
	m.rows = jsondoc.VisibleRows(m.Doc.Root)
}

func (m Model) renderFooter(status string) string {
	if m.helpView == "" {
		return status
	}
	return status + "\n" + m.helpView
}

func indexOfNodeID(rows []jsondoc.Row, id int) int {
	for i, row := range rows {
		if row.Node != nil && row.Node.ID == id {
			return i
		}
	}
	return -1
}

func clamp(v int, high int) int {
	if high < 0 {
		return 0
	}
	if v < 0 {
		return 0
	}
	if v > high {
		return high
	}
	return v
}
