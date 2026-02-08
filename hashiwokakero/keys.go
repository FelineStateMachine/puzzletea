package hashiwokakero

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Cancel key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up", "w"),
		key.WithHelp("↑/w/k", "Up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down", "s"),
		key.WithHelp("↓/s/j", "Down"),
	),
	Left: key.NewBinding(
		key.WithKeys("h", "left", "a"),
		key.WithHelp("←/a/h", "Left"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right", "d"),
		key.WithHelp("→/d/l", "Right"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter/space", "Select"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "Cancel"),
	),
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.Select, m.keys.Cancel},
	}
}
