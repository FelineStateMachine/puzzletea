package rippleeffect

import (
	"charm.land/bubbles/v2/key"
	"github.com/FelineStateMachine/puzzletea/game"
)

type KeyMap struct {
	game.CursorKeyMap
	FillValue key.Binding
	Clear     key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,
	FillValue: key.NewBinding(
		key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("1-9", "Fill"),
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

	locked := m.givens[m.cursor.Y][m.cursor.X] != 0
	enabled := !locked && !m.solved
	m.keys.FillValue.SetEnabled(enabled)
	m.keys.Clear.SetEnabled(enabled)
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.FillValue, m.keys.Clear},
	}
}
