package tui

import "github.com/acidghost/a555less/internal/jsondoc"

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
