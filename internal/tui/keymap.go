package tui

import (
	"fmt"
	"reflect"
	"strings"

	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	Quit         key.Binding
	Down         key.Binding
	Up           key.Binding
	Toggle       key.Binding
	Left         key.Binding
	Right        key.Binding
	Top          key.Binding
	Bottom       key.Binding
	PageDown     key.Binding
	PageUp       key.Binding
	HalfDown     key.Binding
	HalfUp       key.Binding
	Parent       key.Binding
	NextSib      key.Binding
	PrevSib      key.Binding
	Collapse     key.Binding
	CollapseDeep key.Binding
	Expand       key.Binding
	ExpandDeep   key.Binding
	Search       key.Binding
	NextMatch    key.Binding
	PrevMatch    key.Binding
	Help         key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Search, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Toggle},
		{k.Top, k.Bottom, k.PageUp, k.PageDown, k.HalfUp, k.HalfDown},
		{k.Parent, k.NextSib, k.PrevSib, k.Collapse, k.CollapseDeep},
		{k.Expand, k.ExpandDeep, k.Search, k.NextMatch, k.PrevMatch, k.Help, k.Quit},
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
	Collapse: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "collapse siblings"),
	),
	CollapseDeep: key.NewBinding(
		key.WithKeys("C", "shift+c"),
		key.WithHelp("C", "deep collapse siblings"),
	),
	Expand: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "expand siblings"),
	),
	ExpandDeep: key.NewBinding(
		key.WithKeys("E", "shift+e"),
		key.WithHelp("E", "deep expand siblings"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	NextMatch: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next match"),
	),
	PrevMatch: key.NewBinding(
		key.WithKeys("p", "N"),
		key.WithHelp("p/N", "previous match"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

// KeyBindings returns help text describing all available key bindings.
func KeyBindings() string {
	type row struct {
		keys string
		desc string
	}

	keyMapType := reflect.TypeFor[keyMap]()
	keyMapValue := reflect.ValueOf(keys)
	bindingType := reflect.TypeFor[key.Binding]()

	rows := make([]row, 0, keyMapType.NumField())
	maxKeysWidth := len("Keys")
	for i := 0; i < keyMapType.NumField(); i++ {
		field := keyMapType.Field(i)
		if field.Type != bindingType {
			continue
		}

		binding, ok := keyMapValue.Field(i).Interface().(key.Binding)
		if !ok || !binding.Enabled() {
			continue
		}

		help := binding.Help()
		if help.Desc == "" {
			continue
		}

		keys := formatKeys(binding.Keys())
		maxKeysWidth = max(maxKeysWidth, len(keys))
		rows = append(rows, row{keys: keys, desc: help.Desc})
	}

	var out strings.Builder
	fmt.Fprintf(&out, "%-*s  %s\n", maxKeysWidth, "Keys", "Action")
	fmt.Fprintf(&out, "%-*s  %s\n", maxKeysWidth, strings.Repeat("-", maxKeysWidth), strings.Repeat("-", len("Action")))
	for _, row := range rows {
		fmt.Fprintf(&out, "%-*s  %s\n", maxKeysWidth, row.keys, row.desc)
	}

	return strings.TrimRight(out.String(), "\n")
}

func formatKeys(keys []string) string {
	seen := make(map[string]struct{}, len(keys))
	formatted := make([]string, 0, len(keys))
	for _, keyName := range keys {
		if keyName == " " {
			keyName = "space"
		}
		if _, ok := seen[keyName]; ok {
			continue
		}
		seen[keyName] = struct{}{}
		formatted = append(formatted, keyName)
	}
	return strings.Join(formatted, ", ")
}
