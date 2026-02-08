package nonogram

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

type NonogramMode struct {
	game.BaseMode
	Height, Width int
}

var _ game.Mode = NonogramMode{}    // compile-time interface check
var _ game.Spawner = NonogramMode{} // compile-time interface check

func NewMode(title, description string, Height, Width int) NonogramMode {
	return NonogramMode{
		BaseMode: game.NewBaseMode(title, description),
		Height:   Height,
		Width:    Width,
	}
}

func (n NonogramMode) Spawn() (game.Gamer, error) {
	hints := GenerateRandomTomography(n)
	return New(n, hints)
}

var Modes = []list.Item{
	NewMode("Easy - 5x5", "A random nonogram on a five by five board.", 5, 5),
	NewMode("Medium - 10x10", "A random nonogram on a ten by ten board.", 10, 10),
	NewMode("Hard - 15x15", "A random nonogram on a fifteen by fifteen board.", 15, 15),
	NewMode("Extra - 5x10", "A random nonogram on a five by ten board.", 5, 10),
}
