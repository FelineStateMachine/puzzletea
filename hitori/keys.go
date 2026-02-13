package hitori

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	game.CursorKeyMap
	Shade key.Binding
	Clear key.Binding
}

var DefaultKeyMap = KeyMap{
	CursorKeyMap: game.DefaultCursorKeyMap,
	Shade: key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "shade"),
	),
	Clear: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("bkspc", "clear"),
	),
}
