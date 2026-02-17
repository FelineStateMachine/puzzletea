package lightsout

import (
	"charm.land/bubbles/v2/key"
	"github.com/FelineStateMachine/puzzletea/game"
)

type KeyMap struct {
	game.CursorKeyMap
	Toggle key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,

	Toggle: key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter/space", "Toggle"),
	),
}

func (m *Model) updateKeyBindings() {
	m.keys.Up.SetEnabled(m.cursor.Y > 0)
	m.keys.Down.SetEnabled(m.cursor.Y < m.height-1)
	m.keys.Left.SetEnabled(m.cursor.X > 0)
	m.keys.Right.SetEnabled(m.cursor.X < m.width-1)
	m.keys.Toggle.SetEnabled(!m.IsSolved())
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.Toggle},
	}
}
