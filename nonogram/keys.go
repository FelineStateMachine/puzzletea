package nonogram

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	game.CursorKeyMap
	FillTile  key.Binding
	MarkTile  key.Binding
	ClearTile key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,

	FillTile: key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "Fill"),
	),
	MarkTile: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "Mark"),
	),
	ClearTile: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("bkspc", "Clear"),
	),
}

func (m *Model) updateKeyBindings() {
	m.keys.Up.SetEnabled(m.cursor.Y > 0)
	m.keys.Down.SetEnabled(m.cursor.Y < m.height-1)
	m.keys.Left.SetEnabled(m.cursor.X > 0)
	m.keys.Right.SetEnabled(m.cursor.X < m.width-1)

	cursorVal := m.grid[m.cursor.Y][m.cursor.X]
	switch cursorVal {
	case filledTile:
		m.keys.FillTile.SetEnabled(false)
		m.keys.MarkTile.SetEnabled(true)
		m.keys.ClearTile.SetEnabled(true)
	case emptyTile:
		m.keys.FillTile.SetEnabled(true)
		m.keys.MarkTile.SetEnabled(true)
		m.keys.ClearTile.SetEnabled(false)
	case markedTile:
		m.keys.FillTile.SetEnabled(true)
		m.keys.MarkTile.SetEnabled(false)
		m.keys.ClearTile.SetEnabled(true)
	}
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right},
		{m.keys.FillTile, m.keys.MarkTile, m.keys.ClearTile},
	}
}
