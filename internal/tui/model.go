package tui

import (
	"regexp"

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

	searchEditing       bool
	searchInput         string
	searchQuery         string
	searchCaseSensitive bool
	searchPattern       *regexp.Regexp
	searchMatches       []searchMatch
	searchMatchesByPart map[searchTarget][][2]int
	searchIndex         int
	searchHighlight     bool
	searchCursorMoved   bool
}

// New returns a skeleton TUI model for doc.
func New(doc *jsondoc.Document) Model {
	m := Model{Doc: doc, focusID: -1, searchIndex: -1, help: help.New()}
	m.help.Styles.FullKey = helpFullKeyStyle
	m.refreshHelp()
	m.refreshRows()
	if len(m.rows) > 0 {
		m.focusID = m.rows[0].Node.ID
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
		if m.searchEditing {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.updateSearch(msg)
			return m, nil
		}

		previousFocusID := m.focusID
		searchJump := false
		switch {
		case key.Matches(msg, keys.NextMatch):
			searchJump = true
			m.moveSearch(1)
		case key.Matches(msg, keys.PrevMatch):
			searchJump = true
			m.moveSearch(-1)
		case key.Matches(msg, keys.Search):
			m.startSearch()
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
		case key.Matches(msg, keys.Collapse):
			m.collapseFocusedSiblings(false)
		case key.Matches(msg, keys.CollapseDeep):
			m.collapseFocusedSiblings(true)
		case key.Matches(msg, keys.Expand):
			m.expandFocusedSiblings(false)
		case key.Matches(msg, keys.ExpandDeep):
			m.expandFocusedSiblings(true)
		}
		if !searchJump && m.focusID != previousFocusID {
			m.searchCursorMoved = true
		}
	case tea.PasteMsg:
		if m.searchEditing {
			m.searchInput += sanitizeSearchInput(msg.Content)
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
	m.rows = jsondoc.VisibleRowsForDocument(m.Doc)
}
