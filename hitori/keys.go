package hitori

import (
	"github.com/FelineStateMachine/puzzletea/game"

	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	game.CursorKeyMap
	ShadeCell  key.Binding
	CircleCell key.Binding
	ClearCell  key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,
	ShadeCell: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "Shade"),
	),
	CircleCell: key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "Circle"),
	),
	ClearCell: key.NewBinding(
		key.WithKeys("backspace", "delete"),
		key.WithHelp("bkspc", "Clear"),
	),
}

func (m *Model) updateKeyBindings() {
	m.keys.Up.SetEnabled(m.cursor.Y > 0)
	m.keys.Down.SetEnabled(m.cursor.Y < m.size-1)
	m.keys.Left.SetEnabled(m.cursor.X > 0)
	m.keys.Right.SetEnabled(m.cursor.X < m.size-1)

	solved := m.solved
	m.keys.ShadeCell.SetEnabled(!solved)
	m.keys.CircleCell.SetEnabled(!solved)
	m.keys.ClearCell.SetEnabled(!solved)
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.ShadeCell, m.keys.CircleCell, m.keys.ClearCell},
	}
}
