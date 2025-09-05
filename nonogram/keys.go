package nonogram

import (
	"github.com/charmbracelet/bubbles/key"
)

type CursorKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
}

type KeyMap struct {
	CursorKeyMap
	FillTile  key.Binding
	MarkTile  key.Binding
	ClearTile key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: CursorKeyMap{
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
	},

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
		key.WithHelp("⌫", "Clear"),
	),
}

func (m *Model) updateKeyBindinds() {
	m.keys.Up.SetEnabled(m.cursor.y > 0)
	m.keys.Down.SetEnabled(m.cursor.y < m.height-1)
	m.keys.Left.SetEnabled(m.cursor.x > 0)
	m.keys.Right.SetEnabled(m.cursor.x < m.width-1)

	cursorVal := m.grid[m.cursor.y][m.cursor.x]
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
