package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/acidghost/a555less/internal/jsondoc"
)

// Model is the Bubble Tea application model.
type Model struct {
	Doc *jsondoc.Document

	width   int
	height  int
	focusID int
	top     int
}

// New returns a skeleton TUI model for doc.
func New(doc *jsondoc.Document) Model {
	m := Model{Doc: doc, focusID: -1}
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
	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
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
		return renderEmpty(m.width)
	}

	rows := jsondoc.VisibleRows(m.Doc.Root)
	focusIdx := indexOfNodeID(rows, m.focusID)
	if focusIdx < 0 && len(rows) > 0 {
		focusIdx = 0
	}

	viewerHeight := len(rows)
	if m.height > 0 {
		viewerHeight = max(0, m.height-1)
	}

	start := clamp(m.top, 0, max(0, len(rows)-viewerHeight))
	lines := make([]string, 0, viewerHeight+1)
	for i := 0; i < viewerHeight; i++ {
		rowIdx := start + i
		if rowIdx < len(rows) {
			row := rows[rowIdx]
			lines = append(lines, renderRow(row, row.Node.ID == m.focusID, m.width))
		} else {
			lines = append(lines, fillerRow(m.width))
		}
	}

	focusedNode := m.Doc.Root
	if focusIdx >= 0 && focusIdx < len(rows) {
		focusedNode = rows[focusIdx].Node
	}
	lines = append(lines, renderStatus(jsondoc.Path(focusedNode), m.Doc.Filename, focusIdx, len(rows), m.width))

	return strings.Join(lines, "\n")
}

func renderEmpty(width int) string {
	return renderStatus("No JSON document loaded. Press q to quit.", "", 0, 0, width)
}

func indexOfNodeID(rows []jsondoc.Row, id int) int {
	for i, row := range rows {
		if row.Node != nil && row.Node.ID == id {
			return i
		}
	}
	return -1
}

func clamp(v int, low int, high int) int {
	if high < low {
		return low
	}
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
