package tui

import (
	"regexp"
	"sort"
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
	node *jsondoc.Node
	part searchPart
}

type searchMatch struct {
	node  *jsondoc.Node
	order int
	part  searchPart
	start int
	end   int
}

type searchDraft struct {
	text string
}

type searchResult struct {
	query         string
	caseSensitive bool
	matches       []searchMatch
	ranges        map[searchTarget][][2]int
	nodeOrder     map[*jsondoc.Node]int
	current       int
	cursorMoved   bool
}

type searchState struct {
	draft  *searchDraft
	result *searchResult
}

func (s *searchState) begin() {
	s.draft = &searchDraft{}
}

func (s *searchState) cancel() {
	s.draft = nil
}

func (s searchState) editing() bool {
	return s.draft != nil
}

func (s searchState) input() string {
	if s.draft == nil {
		return ""
	}
	return s.draft.text
}

func (s *searchState) appendInput(text string) {
	if s.draft == nil {
		return
	}
	s.draft.text += sanitizeSearchInput(text)
}

func (s *searchState) backspace() {
	if s.draft == nil {
		return
	}
	s.draft.text = removeLastRune(s.draft.text)
}

func (s *searchState) commit(doc *jsondoc.Document, focused *jsondoc.Node) *jsondoc.Node {
	if s.draft == nil {
		return nil
	}

	query := s.draft.text
	s.draft = nil
	caseSensitive := strings.HasSuffix(query, "/s")
	if caseSensitive {
		query = strings.TrimSuffix(query, "/s")
	}

	pattern := compileSearchPattern(query, caseSensitive)
	if pattern == nil || doc == nil || doc.Root == nil {
		s.result = nil
		return nil
	}

	result := &searchResult{
		query:         query,
		caseSensitive: caseSensitive,
		ranges:        make(map[searchTarget][][2]int),
		nodeOrder:     make(map[*jsondoc.Node]int),
		current:       -1,
	}
	nextOrder := 0
	result.collect(doc.Root, doc.JSONL, pattern, &nextOrder)
	s.result = result

	if len(result.matches) == 0 {
		return nil
	}

	focusOrder, ok := result.nodeOrder[focused]
	if !ok {
		focusOrder = 0
	}
	result.current = sort.Search(len(result.matches), func(i int) bool {
		return result.matches[i].order >= focusOrder
	})
	if result.current == len(result.matches) {
		result.current = 0
	}
	return result.matches[result.current].node
}

func (r *searchResult) collect(n *jsondoc.Node, skip bool, pattern *regexp.Regexp, nextOrder *int) {
	if n == nil {
		return
	}

	order := *nextOrder
	*nextOrder = order + 1
	r.nodeOrder[n] = order

	if !skip {
		if n.HasKey {
			r.appendMatches(n, order, searchPartKey, jsondoc.FormatKey(n.Key), pattern)
		}
		if !n.IsContainer() {
			r.appendMatches(n, order, searchPartValue, jsondoc.FormatPrimitive(n), pattern)
		}
	}

	for _, child := range n.Children {
		r.collect(child, false, pattern, nextOrder)
	}
}

func (r *searchResult) appendMatches(node *jsondoc.Node, order int, part searchPart, text string, pattern *regexp.Regexp) {
	target := searchTarget{node: node, part: part}
	for _, matchRange := range searchMatchRanges(text, pattern) {
		r.matches = append(r.matches, searchMatch{
			node:  node,
			order: order,
			part:  part,
			start: matchRange[0],
			end:   matchRange[1],
		})
		r.ranges[target] = append(r.ranges[target], matchRange)
	}
}

func (s *searchState) move(delta int, focused *jsondoc.Node) *jsondoc.Node {
	result := s.result
	if result == nil || len(result.matches) == 0 {
		return nil
	}

	if result.cursorMoved || !result.currentMatchIsFocused(focused) {
		result.current = result.closestMatch(delta, focused)
	} else {
		result.current = (result.current + delta) % len(result.matches)
		if result.current < 0 {
			result.current += len(result.matches)
		}
	}
	result.cursorMoved = false
	return result.matches[result.current].node
}

func (r searchResult) currentMatchIsFocused(focused *jsondoc.Node) bool {
	return r.current >= 0 &&
		r.current < len(r.matches) &&
		r.matches[r.current].node == focused
}

func (r searchResult) closestMatch(delta int, focused *jsondoc.Node) int {
	focusOrder, ok := r.nodeOrder[focused]
	if !ok {
		if delta >= 0 {
			return 0
		}
		return len(r.matches) - 1
	}

	if delta >= 0 {
		i := sort.Search(len(r.matches), func(i int) bool {
			return r.matches[i].order > focusOrder
		})
		if i == len(r.matches) {
			return 0
		}
		return i
	}

	i := sort.Search(len(r.matches), func(i int) bool {
		return r.matches[i].order >= focusOrder
	}) - 1
	if i < 0 {
		return len(r.matches) - 1
	}
	return i
}

func (s *searchState) markCursorMoved() {
	if s.result != nil {
		s.result.cursorMoved = true
	}
}

func (s searchState) highlighting() bool {
	return s.result != nil
}

func (s searchState) query() string {
	if s.result == nil {
		return ""
	}
	return s.result.query
}

func (s searchState) matchPosition() (current, total int, ok bool) {
	if s.result == nil {
		return 0, 0, false
	}
	if s.result.current >= 0 {
		current = s.result.current + 1
	}
	return current, len(s.result.matches), true
}

func (s searchState) ranges(node *jsondoc.Node, part searchPart) [][2]int {
	if s.result == nil {
		return nil
	}
	return s.result.ranges[searchTarget{node: node, part: part}]
}

func (s searchState) isCurrent(node *jsondoc.Node, part searchPart, start, end int) bool {
	if s.result == nil || s.result.current < 0 || s.result.current >= len(s.result.matches) {
		return false
	}
	current := s.result.matches[s.result.current]
	return current.node == node && current.part == part && current.start == start && current.end == end
}

func (m *Model) startSearch() {
	m.search.begin()
	m.ensureVisible()
}

func (m *Model) updateSearch(msg tea.KeyPressMsg) {
	switch msg.Key().Code {
	case tea.KeyEnter, tea.KeyKpEnter:
		m.applySearch()
	case tea.KeyEscape:
		m.cancelSearch()
	case tea.KeyBackspace, tea.KeyDelete:
		m.search.backspace()
	default:
		m.search.appendInput(msg.Key().Text)
	}
}

func (m *Model) cancelSearch() {
	m.search.cancel()
	m.ensureVisible()
}

func (m *Model) applySearch() {
	target := m.search.commit(m.Doc, m.focusedNode())
	if target != nil {
		m.focusSearchNode(target)
		return
	}
	m.ensureVisible()
}

func (m *Model) moveSearch(delta int) {
	target := m.search.move(delta, m.focusedNode())
	if target != nil {
		m.focusSearchNode(target)
	}
}

func (m *Model) focusSearchNode(n *jsondoc.Node) {
	for parent := n.Parent; parent != nil; parent = parent.Parent {
		parent.Collapsed = false
	}
	m.refreshRows()
	m.focusID = n.ID
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

func (m Model) renderSearchText(text string, base lipgloss.Style, node *jsondoc.Node, part searchPart) string {
	if !m.search.highlighting() {
		return base.Render(text)
	}

	ranges := m.search.ranges(node, part)
	if len(ranges) == 0 {
		return base.Render(text)
	}

	var out strings.Builder
	pos := 0
	for _, matchRange := range ranges {
		out.WriteString(base.Render(text[pos:matchRange[0]]))
		style := base.Background(searchMatchBackground).Foreground(searchMatchForeground)
		if m.search.isCurrent(node, part, matchRange[0], matchRange[1]) {
			style = style.Background(searchCurrentBackground).Bold(true)
		}
		out.WriteString(style.Render(text[matchRange[0]:matchRange[1]]))
		pos = matchRange[1]
	}
	out.WriteString(base.Render(text[pos:]))
	return out.String()
}
