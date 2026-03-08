package app

import "charm.land/bubbles/v2/key"

type rootKeyMap struct {
	Quit      key.Binding
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
