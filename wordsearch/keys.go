package wordsearch

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Cancel key.Binding
	Quit   key.Binding
	Help   key.Binding
}

func newKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k", "w"),
			key.WithHelp("↑/k/w", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j", "s"),
			key.WithHelp("↓/j/s", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h", "a"),
			key.WithHelp("←/h/a", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l", "d"),
			key.WithHelp("→/l/d", "move right"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select start/end"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel selection"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Select, k.Cancel}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Cancel, k.Help},
	}
}
