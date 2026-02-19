package shikaku

import (
	"charm.land/bubbles/v2/key"
	"github.com/FelineStateMachine/puzzletea/game"
)

type KeyMap struct {
	game.CursorKeyMap
	ShrinkUp    key.Binding
	ShrinkDown  key.Binding
	ShrinkLeft  key.Binding
	ShrinkRight key.Binding
	Select      key.Binding
	Cancel      key.Binding
	Delete      key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,
	ShrinkUp: key.NewBinding(
		key.WithKeys("shift+up", "K", "W"),
		key.WithHelp("shift+↑", "Shrink up"),
	),
	ShrinkDown: key.NewBinding(
		key.WithKeys("shift+down", "J", "S"),
		key.WithHelp("shift+↓", "Shrink down"),
	),
	ShrinkLeft: key.NewBinding(
		key.WithKeys("shift+left", "H", "A"),
		key.WithHelp("shift+←", "Shrink left"),
	),
	ShrinkRight: key.NewBinding(
		key.WithKeys("shift+right", "L", "D"),
		key.WithHelp("shift+→", "Shrink right"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter/space", "Select/Confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "Cancel"),
	),
	Delete: key.NewBinding(
		key.WithKeys("backspace", "delete"),
		key.WithHelp("bkspc", "Delete"),
	),
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.ShrinkUp, m.keys.ShrinkDown, m.keys.ShrinkLeft, m.keys.ShrinkRight},
		{m.keys.Select, m.keys.Cancel, m.keys.Delete},
	}
}
