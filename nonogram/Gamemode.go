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
	NewMode("Mini", "5x5 grid, ~65% filled. Quick puzzle, straightforward hints.", 5, 5, 0.65),
	NewMode("Pocket", "5x5 grid, ~50% filled. Compact but balanced.", 5, 5, 0.50),
	NewMode("Teaser", "5x5 grid, ~35% filled. Small but tricky.", 5, 5, 0.35),
	// 10x10
	NewMode("Standard", "10x10 grid, ~67% filled. Classic size, dense hints.", 10, 10, 0.67),
	NewMode("Classic", "10x10 grid, ~52% filled. The typical nonogram experience.", 10, 10, 0.52),
	NewMode("Tricky", "10x10 grid, ~37% filled. Sparse hints require reasoning.", 10, 10, 0.37),
	// 15x15
	NewMode("Large", "15x15 grid, ~69% filled. Bigger grid, constraining hints.", 15, 15, 0.69),
	NewMode("Grand", "15x15 grid, ~54% filled. A substantial challenge.", 15, 15, 0.54),
	// 20x20
	NewMode("Epic", "20x20 grid, ~71% filled. A epic undertaking.", 20, 20, 0.71),
	NewMode("Massive", "20x20 grid, ~56% filled. Truly massive puzzle.", 20, 20, 0.56),
}
