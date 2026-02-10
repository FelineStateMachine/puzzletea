package nonogram

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

func init() {
	game.Register("Nonogram", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

type NonogramMode struct {
	game.BaseMode
	Width, Height int
	Density       float64 // target fill percentage, 0.0–1.0
}

var (
	_ game.Mode    = NonogramMode{} // compile-time interface check
	_ game.Spawner = NonogramMode{} // compile-time interface check
)

func NewMode(title, description string, width, height int, density float64) NonogramMode {
	return NonogramMode{
		BaseMode: game.NewBaseMode(title, description),
		Width:    width,
		Height:   height,
		Density:  density,
	}
}

func (n NonogramMode) Spawn() (game.Gamer, error) {
	hints := GenerateRandomTomography(n)
	return New(n, hints)
}

var Modes = []list.Item{
	// 5x5
	NewMode("Easy 5x5", "5x5 grid, ~35% filled. Simple hints.", 5, 5, 0.35),
	NewMode("Medium 5x5", "5x5 grid, ~50% filled. Balanced challenge.", 5, 5, 0.50),
	NewMode("Hard 5x5", "5x5 grid, ~65% filled. Dense hints.", 5, 5, 0.65),
	// 10x10
	NewMode("Easy 10x10", "10x10 grid, ~35% filled. Simple hints.", 10, 10, 0.35),
	NewMode("Medium 10x10", "10x10 grid, ~50% filled. Balanced challenge.", 10, 10, 0.50),
	NewMode("Hard 10x10", "10x10 grid, ~65% filled. Dense hints.", 10, 10, 0.65),
	// 15x15
	NewMode("Easy 15x15", "15x15 grid, ~35% filled. Simple hints.", 15, 15, 0.35),
	NewMode("Medium 15x15", "15x15 grid, ~50% filled. Balanced challenge.", 15, 15, 0.50),
	NewMode("Hard 15x15", "15x15 grid, ~65% filled. Dense hints.", 15, 15, 0.65),
	// 20x20
	NewMode("Easy 20x20", "20x20 grid, ~35% filled. Simple hints.", 20, 20, 0.35),
	NewMode("Medium 20x20", "20x20 grid, ~50% filled. Balanced challenge.", 20, 20, 0.50),
	NewMode("Hard 20x20", "20x20 grid, ~65% filled. Dense hints.", 20, 20, 0.65),
}
