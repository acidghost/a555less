package tui

import "charm.land/lipgloss/v2"

var (
	indicatorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	keyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	indexStyle     = lipgloss.NewStyle().Faint(true)
	stringStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	numberStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	boolStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	nullStyle      = lipgloss.NewStyle().Faint(true)
	punctStyle     = lipgloss.NewStyle().Faint(true)
	countStyle     = lipgloss.NewStyle().Faint(true)
	dimStyle       = lipgloss.NewStyle().Faint(true)
)
