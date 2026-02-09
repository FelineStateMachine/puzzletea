package sudoku

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
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
		key.WithHelp("[1-9]", "Fill Cell"),
	),
	ClearCell: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("bkspc", "Clear Cell Contents"),
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
	v := false
	for _, hint := range m.provided {
		if m.cursor.X == hint.x && m.cursor.Y == hint.y {
			v = true
		}
	}
	return v
}

// GetFullHelp implements game.Gamer.
func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.FillValue, m.keys.ClearCell},
	}
}
