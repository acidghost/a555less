package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/acidghost/a555less/internal/jsondoc"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	dimStyle   = lipgloss.NewStyle().Faint(true)
)

// Model is the Bubble Tea application model.
type Model struct {
	Doc *jsondoc.Document

	width  int
	height int
}

// New returns a skeleton TUI model for doc.
func New(doc *jsondoc.Document) Model {
	return Model{Doc: doc}
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
	filename := "stdin"
	size := 0
	if m.Doc != nil {
		if m.Doc.Filename != "" {
			filename = m.Doc.Filename
		}
		size = m.Doc.Size()
	}

	lines := []string{
		titleStyle.Render("a555less"),
		"",
		fmt.Sprintf("Loaded %s (%d bytes)", filename, size),
		"",
		dimStyle.Render("Phase 1 skeleton: JSON input parsed successfully."),
		dimStyle.Render("Press q to quit."),
	}

	if m.height > 0 && len(lines) < m.height {
		for len(lines) < m.height {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}
