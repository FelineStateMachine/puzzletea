package shikaku

import (
	_ "embed"
	"math/rand/v2"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

func init() {
	game.Register("Shikaku", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

type ShikakuMode struct {
	game.BaseMode
	Width       int
	Height      int
	MaxRectSize int
}

var (
	_ game.Mode          = ShikakuMode{}
	_ game.Spawner       = ShikakuMode{}
	_ game.SeededSpawner = ShikakuMode{}
)

func NewMode(title, description string, width, height, maxRectSize int) ShikakuMode {
	return ShikakuMode{
		BaseMode:    game.NewBaseMode(title, description),
		Width:       width,
		Height:      height,
		MaxRectSize: maxRectSize,
	}
}

func (s ShikakuMode) Spawn() (game.Gamer, error) {
	puzzle, err := GeneratePuzzle(s.Width, s.Height, s.MaxRectSize)
	if err != nil {
		return nil, err
	}
	return New(s, puzzle), nil
}

func (s ShikakuMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	puzzle, err := GeneratePuzzleSeeded(s.Width, s.Height, s.MaxRectSize, rng)
	if err != nil {
		return nil, err
	}
	return New(s, puzzle), nil
}

var Modes = []list.Item{
	NewMode("Mini 5x5", "5x5 grid, gentle introduction.", 5, 5, 5),
	NewMode("Easy 7x7", "7x7 grid, straightforward puzzles.", 7, 7, 8),
	NewMode("Medium 8x8", "8x8 grid, moderate challenge.", 8, 8, 12),
	NewMode("Hard 10x10", "10x10 grid, requires careful planning.", 10, 10, 15),
	NewMode("Expert 12x12", "12x12 grid, maximum challenge.", 12, 12, 20),
}

var DailyModes = []list.Item{
	Modes[1], // Easy 7x7
	Modes[2], // Medium 8x8
}
