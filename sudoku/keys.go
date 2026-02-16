package sudoku

import (
	"charm.land/bubbles/v2/key"
	"github.com/FelineStateMachine/puzzletea/game"
)

type KeyMap struct {
	game.CursorKeyMap
	FillValue key.Binding
	ClearCell key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,
	FillValue: key.NewBinding(
		key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("1-9", "Fill"),
	),
	ClearCell: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("bkspc", "Clear"),
	),
}

func (m *Model) updateKeyBindings() {
	m.keys.Up.SetEnabled(m.cursor.Y > 0)
	m.keys.Down.SetEnabled(m.cursor.Y < 8)
	m.keys.Left.SetEnabled(m.cursor.X > 0)
	m.keys.Right.SetEnabled(m.cursor.X < 8)
	allowInteraction := !m.IsProvidedCell()
	m.keys.FillValue.SetEnabled(allowInteraction)
	m.keys.ClearCell.SetEnabled(allowInteraction)
}

func (m Model) IsProvidedCell() bool {
	return m.providedGrid[m.cursor.Y][m.cursor.X]
}

// GetFullHelp implements game.Gamer.
func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.FillValue, m.keys.ClearCell},
	}
}
