package tui

import "charm.land/lipgloss/v2"

var (
	ColorLavander   = lipgloss.Color("105")
	ColorBlue       = lipgloss.Color("117")
	ColorVioletDark = lipgloss.Color("129")
	ColorViolet     = lipgloss.Color("135")
	ColorPurple     = lipgloss.Color("141")
	ColorPink       = lipgloss.Color("219")

	indicatorStyle   = lipgloss.NewStyle().Bold(true).Foreground(ColorPink)
	keyStyle         = lipgloss.NewStyle().Foreground(ColorLavander)
	indexStyle       = lipgloss.NewStyle().Faint(true)
	stringStyle      = lipgloss.NewStyle().Foreground(ColorBlue)
	numberStyle      = lipgloss.NewStyle().Foreground(ColorViolet)
	boolStyle        = lipgloss.NewStyle().Foreground(ColorVioletDark)
	nullStyle        = lipgloss.NewStyle().Faint(true)
	dimStyle         = lipgloss.NewStyle().Faint(true)
	statusStyle      = lipgloss.NewStyle().Foreground(ColorLavander)
	helpFullKeyStyle = lipgloss.NewStyle().Foreground(ColorBlue)
)
