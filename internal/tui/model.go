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
		return dimStyle.Render("No JSON document loaded. Press q to quit.")
	}

	rows := jsondoc.VisibleRows(m.Doc.Root)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		if m.height > 0 && len(lines) >= m.height {
			break
		}
		lines = append(lines, renderRow(row, row.Node.ID == m.focusID, m.width))
	}

	return strings.Join(lines, "\n")
}
