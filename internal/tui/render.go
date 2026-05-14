package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/acidghost/a555less/internal/jsondoc"
)

const (
	previewMaxItems = 5
	previewMaxDepth = 1
)

func renderStatus(path string, filename string, focusIndex int, totalRows int, width int) string {
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
	var text string
	if width > 0 {
		leftWidth := ansi.StringWidth(path)
		rightWidth := ansi.StringWidth(right)
		if leftWidth+1+rightWidth <= width {
			text = path + strings.Repeat(" ", width-leftWidth-rightWidth) + right
		} else {
			text = ansi.Truncate(path+" "+right, width, "…")
		}
	} else {
		text = path + " " + right
	}
	return statusStyle.Render(text)
}

func fillerRow(width int) string {
	line := dimStyle.Render("~")
	if width > 0 {
		line = ansi.Truncate(line, width, "…")
	}
	return line
}

func renderRow(row jsondoc.Row, focused bool, width int) string {
	n := row.Node
	line := strings.Repeat("  ", row.Depth) + renderIndicator(n, focused) + renderLabel(n) + renderValue(n)
	if width > 0 {
		line = ansi.Truncate(line, width, "…")
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
	return punctStyle.Render(indicator)
}

func renderLabel(n *jsondoc.Node) string {
	if n == nil || n.Parent == nil {
		return ""
	}
	if n.HasKey {
		return keyStyle.Render(jsondoc.FormatKey(n.Key)) + punctStyle.Render(": ")
	}
	return indexStyle.Render(fmt.Sprintf("[%d]", n.Index)) + punctStyle.Render(": ")
}

func renderValue(n *jsondoc.Node) string {
	if n == nil {
		return nullStyle.Render("null")
	}
	if n.IsContainer() {
		return countStyle.Render(fmt.Sprintf("(%d) ", len(n.Children))) + renderPreview(n, previewMaxItems, previewMaxDepth)
	}
	return renderPrimitive(n)
}

func renderPrimitive(n *jsondoc.Node) string {
	text := jsondoc.FormatPrimitive(n)
	switch n.Kind {
	case jsondoc.KindNull:
		return nullStyle.Render(text)
	case jsondoc.KindBool:
		return boolStyle.Render(text)
	case jsondoc.KindNumber:
		return numberStyle.Render(text)
	case jsondoc.KindString:
		return stringStyle.Render(text)
	default:
		return text
	}
}

func renderPreview(n *jsondoc.Node, maxItems int, maxDepth int) string {
	if n == nil {
		return nullStyle.Render("null")
	}
	if !n.IsContainer() {
		return renderPrimitive(n)
	}
	if maxDepth <= 0 {
		switch n.Kind {
		case jsondoc.KindObject:
			return punctStyle.Render("{…}")
		case jsondoc.KindArray:
			return punctStyle.Render("[…]")
		}
	}

	switch n.Kind {
	case jsondoc.KindObject:
		return renderContainerPreview(n, maxItems, maxDepth, "{", "}")
	case jsondoc.KindArray:
		return renderContainerPreview(n, maxItems, maxDepth, "[", "]")
	default:
		return renderPrimitive(n)
	}
}

func renderContainerPreview(n *jsondoc.Node, maxItems int, maxDepth int, open string, closeDelim string) string {
	if len(n.Children) == 0 {
		return punctStyle.Render(open + closeDelim)
	}
	if maxItems < 0 {
		maxItems = 0
	}

	parts := make([]string, 0, min(len(n.Children), maxItems)+1)
	limit := min(len(n.Children), maxItems)
	for i := range limit {
		child := n.Children[i]
		part := ""
		if n.Kind == jsondoc.KindObject {
			part += keyStyle.Render(jsondoc.FormatKey(child.Key)) + punctStyle.Render(": ")
		}
		part += renderPreview(child, maxItems, maxDepth-1)
		parts = append(parts, part)
	}
	if limit < len(n.Children) {
		parts = append(parts, punctStyle.Render("…"))
	}

	return punctStyle.Render(open) + strings.Join(parts, punctStyle.Render(", ")) + punctStyle.Render(closeDelim)
}
