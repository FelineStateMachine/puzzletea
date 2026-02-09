package game

import "github.com/charmbracelet/bubbles/key"

// CursorKeyMap defines the shared cursor movement key bindings.
type CursorKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
}

// DefaultCursorKeyMap provides the standard cursor movement bindings.
var DefaultCursorKeyMap = CursorKeyMap{
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
}
