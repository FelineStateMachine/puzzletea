package hitori

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

func init() {
	game.Register("Hitori", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

type HitoriMode struct {
	game.BaseMode
	Size      int
	Prefilled float64
}

var (
	_ game.Mode    = HitoriMode{}
	_ game.Spawner = HitoriMode{}
)

func NewMode(title, description string, size int, prefilled float64) HitoriMode {
	return HitoriMode{
		BaseMode:  game.NewBaseMode(title, description),
		Size:      size,
		Prefilled: prefilled,
	}
}

func (h HitoriMode) Spawn() (game.Gamer, error) {
	puzzle, provided, err := GeneratePuzzle(h)
	if err != nil {
		return nil, err
	}
	return New(h, puzzle, provided)
}

var Modes = []list.Item{
	NewMode("Tiny", "5x5 grid, 55% clues. Quick introduction.", 5, 0.55),
	NewMode("Easy", "6x6 grid, 50% clues. Simple patterns.", 6, 0.50),
	NewMode("Medium", "8x8 grid, 45% clues. Standard difficulty.", 8, 0.45),
	NewMode("Hard", "10x10 grid, 40% clues. Longer deductions.", 10, 0.40),
	NewMode("Expert", "12x12 grid, 35% clues. Complex logic.", 12, 0.35),
}
