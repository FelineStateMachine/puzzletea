package spellpuzzle

import (
	"charm.land/bubbles/v2/key"
	"github.com/FelineStateMachine/puzzletea/game"
)

type KeyMap struct {
	Left    key.Binding
	Right   key.Binding
	Shuffle key.Binding
	Submit  key.Binding
	Back    key.Binding
}

var DefaultKeyMap = KeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "a", "h"),
		key.WithHelp("left", "Prev letter"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "d", "l"),
		key.WithHelp("right", "Next letter"),
	),
	Shuffle: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "Shuffle letters"),
	),
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Submit word"),
	),
	Back: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("bkspc", "Delete"),
	),
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Left, m.keys.Right, m.keys.Shuffle, m.keys.Submit, m.keys.Back},
		{game.DefaultCursorKeyMap.Up, game.DefaultCursorKeyMap.Down},
	}
}
