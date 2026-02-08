package sudoku

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
	FillValue key.Binding
	ClearCell key.Binding
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
	FillValue: key.NewBinding(
		key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("[1-9]", "Fill Cell"),
	),
	ClearCell: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("bkspc", "Clear Cell Contents"),
	),
}

func (m *Model) updateKeyBindinds() {
	m.keys.Up.SetEnabled(m.cursor.y > 0)
	m.keys.Down.SetEnabled(m.cursor.y < 8)
	m.keys.Left.SetEnabled(m.cursor.x > 0)
	m.keys.Right.SetEnabled(m.cursor.x < 8)
	allowInteraction := !m.IsProvidedCell()
	m.keys.FillValue.SetEnabled(allowInteraction)
	m.keys.ClearCell.SetEnabled(allowInteraction)
}

func (m Model) IsProvidedCell() bool {
	v := false
	for _, hint := range m.provided {
		if m.cursor.x == hint.x && m.cursor.y == hint.y {
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
