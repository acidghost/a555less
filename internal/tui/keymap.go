package tui

import "charm.land/bubbles/v2/key"

type keyMap struct {
	Quit     key.Binding
	Down     key.Binding
	Up       key.Binding
	Toggle   key.Binding
	Left     key.Binding
	Right    key.Binding
	Top      key.Binding
	Bottom   key.Binding
	PageDown key.Binding
	PageUp   key.Binding
	HalfDown key.Binding
	HalfUp   key.Binding
	Parent   key.Binding
	NextSib  key.Binding
	PrevSib  key.Binding
	Help     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Toggle},
		{k.Top, k.Bottom, k.PageUp, k.PageDown, k.HalfUp, k.HalfDown},
		{k.Parent, k.NextSib, k.PrevSib, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" ", "space", "enter"),
		key.WithHelp("space", "toggle"),
	),
	Left: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h/←", "collapse/parent"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l/→", "expand/child"),
	),
	Top: key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G", "shift+g", "end"),
		key.WithHelp("G", "bottom"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "pagedown", "ctrl+f"),
		key.WithHelp("pgdn", "page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "pageup", "ctrl+b"),
		key.WithHelp("pgup", "page up"),
	),
	HalfDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("C-d", "half down"),
	),
	HalfUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("C-u", "half up"),
	),
	Parent: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "parent"),
	),
	NextSib: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "next sibling"),
	),
	PrevSib: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("K", "prev sibling"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}
