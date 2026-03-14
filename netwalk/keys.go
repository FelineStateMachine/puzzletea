package netwalk

import (
	"charm.land/bubbles/v2/key"
	"github.com/FelineStateMachine/puzzletea/game"
)

type KeyMap struct {
	game.CursorKeyMap
	Rotate     key.Binding
	RotateBack key.Binding
	Lock       key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,
	Rotate: key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter/space", "Rotate"),
	),
	RotateBack: key.NewBinding(
		key.WithKeys("backspace", "shift+space"),
		key.WithHelp("bkspc", "Rotate back"),
	),
	Lock: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "Toggle lock"),
	),
}

func (m *Model) updateKeyBindings() {
	m.keys.Up.SetEnabled(m.cursor.Y > 0)
	m.keys.Down.SetEnabled(m.cursor.Y < m.puzzle.Size-1)
	m.keys.Left.SetEnabled(m.cursor.X > 0)
	m.keys.Right.SetEnabled(m.cursor.X < m.puzzle.Size-1)

	current := m.puzzle.Tiles[m.cursor.Y][m.cursor.X]
	canAct := !m.state.solved && isActive(current)
	m.keys.Rotate.SetEnabled(canAct && !current.Locked)
	m.keys.RotateBack.SetEnabled(canAct && !current.Locked)
	m.keys.Lock.SetEnabled(canAct)
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.Rotate, m.keys.RotateBack, m.keys.Lock},
	}
}
