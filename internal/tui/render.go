package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/acidghost/a555less/internal/jsondoc"
)

const (
	previewMaxItems = 5
	previewMaxDepth = 1
)

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
			lines = append(lines, m.renderRow(row, row.Node.ID == m.focusID))
		} else {
			lines = append(lines, m.fillerRow())
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
		lines = append(lines, m.fillerRow())
	}
	return strings.Join(lines, "\n") + "\n" + footer
}

func (m Model) renderStatus(path string, focusIndex int) string {
	filename := m.Doc.Filename
	totalRows := len(m.rows)
	if filename == "" {
		filename = "stdin"
	}
	if focusIndex < 0 {
		focusIndex = 0
	}
	if totalRows < 0 {
		totalRows = 0
	}

	right := fmt.Sprintf("%s  %d/%d", filename, focusIndex+1, totalRows)
	if matchIndex, matchTotal, ok := m.search.matchPosition(); ok {
		right = fmt.Sprintf("[%d/%d]  %s", matchIndex, matchTotal, right)
	}
	var text string
	if m.width > 0 {
		leftWidth := lipgloss.Width(path)
		rightWidth := lipgloss.Width(right)
		switch {
		case leftWidth+rightWidth+1 <= m.width:
			text = path + strings.Repeat(" ", m.width-leftWidth-rightWidth) + right
		default:
			text = ansi.Truncate(path+" "+right, m.width, "…")
		}
	} else {
		text = path + " " + right
	}
	return statusStyle.Render(text)
}

func (m Model) renderFooter(status string) string {
	if m.search.editing() {
		prompt := searchPromptStyle.Render("/") + m.search.input() + searchPromptStyle.Render("█")
		if m.width > 0 {
			prompt = ansi.Truncate(prompt, m.width, "…")
		}
		return status + "\n" + prompt
	}
	if m.printPending {
		return status + "\n" + printPromptStyle.Render("p█")
	}
	if m.message != "" {
		message := warningStyle.Render(m.message)
		if m.width > 0 {
			message = ansi.Truncate(message, m.width, "…")
		}
		return status + "\n" + message
	}
	if m.helpView == "" {
		return status
	}
	return status + "\n" + m.helpView
}

func (m Model) fillerRow() string {
	line := dimStyle.Render("~")
	if m.width > 0 {
		line = ansi.Truncate(line, m.width, "…")
	}
	return line
}

func (m Model) renderRow(row jsondoc.Row, focused bool) string {
	n := row.Node
	line := strings.Repeat("  ", row.Depth) +
		renderIndicator(n, focused) +
		m.renderLabel(n) +
		m.renderValue(n)
	if m.width > 0 {
		line = ansi.Truncate(line, m.width, "…")
	}
	return line
}

func renderIndicator(n *jsondoc.Node, focused bool) string {
	indicator := "  "
	if n.IsContainer() {
		switch {
		case n.Collapsed && focused:
			indicator = "▶ "
		case n.Collapsed:
			indicator = "▷ "
		case focused:
			indicator = "▼ "
		default:
			indicator = "▽ "
		}
	} else if focused {
		indicator = "▶ "
	}

	if focused {
		return indicatorStyle.Render(indicator)
	}
	return dimStyle.Render(indicator)
}

func (m Model) renderLabel(n *jsondoc.Node) string {
	if n == nil || n.Parent == nil {
		return ""
	}
	var label string
	if n.HasKey {
		label = m.renderSearchText(jsondoc.FormatKey(n.Key), keyStyle, n, searchPartKey)
	} else {
		label = indexStyle.Render(fmt.Sprintf("[%d]", n.Index))
	}
	return label + dimStyle.Render(": ")
}

func (m Model) renderValue(n *jsondoc.Node) string {
	if n == nil {
		return nullStyle.Render("null")
	}
	if n.IsContainer() {
		return dimStyle.Render(
			fmt.Sprintf("(%d) %s",
				len(n.Children),
				jsondoc.Preview(n, previewMaxItems, previewMaxDepth),
			),
		)
	}
	return m.renderPrimitive(n)
}

func (m Model) renderPrimitive(n *jsondoc.Node) string {
	text := jsondoc.FormatPrimitive(n)
	style := lipgloss.NewStyle()
	switch n.Kind {
	case jsondoc.KindNull:
		style = nullStyle
	case jsondoc.KindBool:
		style = boolStyle
	case jsondoc.KindNumber:
		style = numberStyle
	case jsondoc.KindString:
		style = stringStyle
	}
	return m.renderSearchText(text, style, n, searchPartValue)
}
