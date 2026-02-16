package main

import "github.com/charmbracelet/bubbles/key"

type rootKeyMap struct {
	Quit      key.Binding
	MainMenu  key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Debug     key.Binding
	FullHelp  key.Binding
	ResetGame key.Binding
}

var rootKeys = rootKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
	),
	MainMenu: key.NewBinding(
		key.WithKeys("ctrl+n"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
	),
	Debug: key.NewBinding(
		key.WithKeys("ctrl+e"),
	),
	FullHelp: key.NewBinding(
		key.WithKeys("ctrl+h"),
	),
	ResetGame: key.NewBinding(
		key.WithKeys("ctrl+r"),
	),
}
