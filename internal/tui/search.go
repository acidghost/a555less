package tui

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/acidghost/a555less/internal/jsondoc"
)

type searchPart uint8

const (
	searchPartKey searchPart = iota
	searchPartValue
)

type searchTarget struct {
	nodeID int
	part   searchPart
}

type searchMatch struct {
	node  *jsondoc.Node
	part  searchPart
	start int
	end   int
}

func (m *Model) startSearch() {
	m.searchEditing = true
	m.searchInput = ""
	m.ensureVisible()
}

func (m *Model) updateSearch(msg tea.KeyPressMsg) {
	switch msg.Key().Code {
	case tea.KeyEnter, tea.KeyKpEnter:
		m.applySearch()
	case tea.KeyEscape:
		m.cancelSearch()
	case tea.KeyBackspace, tea.KeyDelete:
		m.searchInput = removeLastRune(m.searchInput)
	default:
		m.searchInput += msg.Key().Text
	}
}

func (m *Model) cancelSearch() {
	m.searchEditing = false
	m.searchInput = ""
	m.ensureVisible()
}

func (m *Model) applySearch() {
	m.searchEditing = false
	m.searchQuery = m.searchInput
	m.searchCaseSensitive = strings.HasSuffix(m.searchQuery, "/s")
	if m.searchCaseSensitive {
		m.searchQuery = strings.TrimSuffix(m.searchQuery, "/s")
	}
	m.searchInput = ""
	m.searchPattern = compileSearchPattern(m.searchQuery, m.searchCaseSensitive)
	m.searchMatches = nil
	m.searchMatchesByPart = make(map[searchTarget][][2]int)
	m.searchIndex = -1
	m.searchHighlight = false
	m.searchCursorMoved = false

	if m.searchPattern == nil || m.Doc == nil || m.Doc.Root == nil {
		m.ensureVisible()
		return
	}

	m.collectSearchMatches(m.Doc.Root, m.Doc.JSONL)
	m.searchHighlight = true
	if len(m.searchMatches) == 0 {
		m.ensureVisible()
		return
	}

	// Start at the first occurrence on or after the focused node, wrapping to
	// the beginning when the cursor is below the final match.
	m.searchIndex = 0
	for i, match := range m.searchMatches {
		if match.node.ID >= m.focusID {
			m.searchIndex = i
			break
		}
	}
	m.focusSearchMatch()
}

func (m *Model) collectSearchMatches(n *jsondoc.Node, skip bool) {
	if n == nil {
		return
	}

	if !skip {
		if n.HasKey {
			m.appendSearchMatches(n, searchPartKey, jsondoc.FormatKey(n.Key))
		}
		if !n.IsContainer() {
			m.appendSearchMatches(n, searchPartValue, jsondoc.FormatPrimitive(n))
		}
	}

	for _, child := range n.Children {
		m.collectSearchMatches(child, false)
	}
}

func (m *Model) appendSearchMatches(node *jsondoc.Node, part searchPart, text string) {
	target := searchTarget{nodeID: node.ID, part: part}
	for _, matchRange := range searchMatchRanges(text, m.searchPattern) {
		m.searchMatches = append(m.searchMatches, searchMatch{
			node:  node,
			part:  part,
			start: matchRange[0],
			end:   matchRange[1],
		})
		m.searchMatchesByPart[target] = append(m.searchMatchesByPart[target], matchRange)
	}
}

func (m *Model) moveSearch(delta int) {
	if m.searchQuery == "" {
		return
	}
	m.searchHighlight = true
	if len(m.searchMatches) == 0 {
		return
	}

	if m.searchCursorMoved || !m.currentSearchMatchIsFocused() {
		m.searchIndex = m.closestSearchMatch(delta)
	} else {
		m.searchIndex = (m.searchIndex + delta) % len(m.searchMatches)
		if m.searchIndex < 0 {
			m.searchIndex += len(m.searchMatches)
		}
	}
	m.focusSearchMatch()
}

func (m Model) currentSearchMatchIsFocused() bool {
	return m.searchIndex >= 0 &&
		m.searchIndex < len(m.searchMatches) &&
		m.searchMatches[m.searchIndex].node.ID == m.focusID
}

func (m Model) closestSearchMatch(delta int) int {
	if delta >= 0 {
		for i, match := range m.searchMatches {
			if match.node.ID > m.focusID {
				return i
			}
		}
		return 0
	}

	for i := len(m.searchMatches) - 1; i >= 0; i-- {
		if m.searchMatches[i].node.ID < m.focusID {
			return i
		}
	}
	return len(m.searchMatches) - 1
}

func (m *Model) focusSearchMatch() {
	if m.searchIndex < 0 || m.searchIndex >= len(m.searchMatches) {
		return
	}

	match := m.searchMatches[m.searchIndex]
	for parent := match.node.Parent; parent != nil; parent = parent.Parent {
		parent.Collapsed = false
	}
	m.refreshRows()
	m.focusID = match.node.ID
	m.searchCursorMoved = false
	m.ensureVisible()
}

func compileSearchPattern(query string, caseSensitive bool) *regexp.Regexp {
	if query == "" {
		return nil
	}
	pattern := regexp.QuoteMeta(query)
	if !caseSensitive {
		pattern = "(?i:" + pattern + ")"
	}
	return regexp.MustCompile(pattern)
}

func searchMatchRanges(text string, pattern *regexp.Regexp) [][2]int {
	if pattern == nil {
		return nil
	}
	indices := pattern.FindAllStringIndex(text, -1)
	matches := make([][2]int, len(indices))
	for i, match := range indices {
		matches[i] = [2]int{match[0], match[1]}
	}
	return matches
}

func removeLastRune(s string) string {
	if s == "" {
		return s
	}
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

func sanitizeSearchInput(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t':
			return ' '
		default:
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}
	}, s)
}

func (m Model) renderSearchText(text string, base lipgloss.Style, nodeID int, part searchPart) string {
	if !m.searchHighlight || m.searchQuery == "" {
		return base.Render(text)
	}

	ranges := m.searchMatchesByPart[searchTarget{nodeID: nodeID, part: part}]
	if len(ranges) == 0 {
		return base.Render(text)
	}

	var out strings.Builder
	pos := 0
	for _, matchRange := range ranges {
		out.WriteString(base.Render(text[pos:matchRange[0]]))
		style := base.Background(searchMatchBackground).Foreground(searchMatchForeground)
		if m.isCurrentSearchMatch(nodeID, part, matchRange[0], matchRange[1]) {
			style = style.Background(searchCurrentBackground).Bold(true)
		}
		out.WriteString(style.Render(text[matchRange[0]:matchRange[1]]))
		pos = matchRange[1]
	}
	out.WriteString(base.Render(text[pos:]))
	return out.String()
}

func (m Model) isCurrentSearchMatch(nodeID int, part searchPart, start, end int) bool {
	if m.searchIndex < 0 || m.searchIndex >= len(m.searchMatches) {
		return false
	}
	current := m.searchMatches[m.searchIndex]
	return current.node.ID == nodeID && current.part == part && current.start == start && current.end == end
}
