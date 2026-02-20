package nurikabe

import (
	"charm.land/bubbles/v2/key"
	"github.com/FelineStateMachine/puzzletea/game"
)

type KeyMap struct {
	game.CursorKeyMap
	SetSea    key.Binding
	SetIsland key.Binding
	Clear     key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,
	SetSea: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "Sea"),
	),
	SetIsland: key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "Island"),
	),
	Clear: key.NewBinding(
		key.WithKeys("backspace", "delete"),
		key.WithHelp("bkspc", "Clear"),
	),
}

func (m *Model) updateKeyBindings() {
	m.keys.Up.SetEnabled(m.cursor.Y > 0)
	m.keys.Down.SetEnabled(m.cursor.Y < m.height-1)
	m.keys.Left.SetEnabled(m.cursor.X > 0)
	m.keys.Right.SetEnabled(m.cursor.X < m.width-1)

	locked := isClueCell(m.clues, m.cursor.X, m.cursor.Y) || m.solved
	m.keys.SetSea.SetEnabled(!locked)
	m.keys.SetIsland.SetEnabled(!locked)
	m.keys.Clear.SetEnabled(!locked)
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.SetSea, m.keys.SetIsland, m.keys.Clear},
	}
}
