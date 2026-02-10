package takuzu

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the keybindings for Takuzu.
type KeyMap struct {
	game.CursorKeyMap
	PlaceZero key.Binding
	PlaceOne  key.Binding
	Clear     key.Binding
}

// DefaultKeyMap provides the standard keybindings.
var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,
	PlaceZero: key.NewBinding(
		key.WithKeys("z", "0"),
		key.WithHelp("z/0", "Place ●"),
	),
	PlaceOne: key.NewBinding(
		key.WithKeys("x", "1"),
		key.WithHelp("x/1", "Place ○"),
	),
	Clear: key.NewBinding(
		key.WithKeys("backspace", "delete"),
		key.WithHelp("bkspc", "Clear"),
	),
}

func (m *Model) updateKeyBindings() {
	m.keys.Up.SetEnabled(m.cursor.Y > 0)
	m.keys.Down.SetEnabled(m.cursor.Y < m.size-1)
	m.keys.Left.SetEnabled(m.cursor.X > 0)
	m.keys.Right.SetEnabled(m.cursor.X < m.size-1)

	onProvided := m.provided[m.cursor.Y][m.cursor.X]
	solved := m.solved
	m.keys.PlaceZero.SetEnabled(!onProvided && !solved)
	m.keys.PlaceOne.SetEnabled(!onProvided && !solved)
	m.keys.Clear.SetEnabled(!onProvided && !solved)
}

// GetFullHelp implements game.Gamer.
func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.PlaceZero, m.keys.PlaceOne, m.keys.Clear},
	}
}
