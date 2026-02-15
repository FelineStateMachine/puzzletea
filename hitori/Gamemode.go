package hitori

import (
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

func init() {
	game.Register("Hitori", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

// HitoriMode represents a specific difficulty configuration for Hitori.
type HitoriMode struct {
	game.BaseMode
	Size       int
	BlackRatio float64
}

var (
	_ game.Mode          = HitoriMode{}
	_ game.Spawner       = HitoriMode{}
	_ game.SeededSpawner = HitoriMode{}
)

// NewMode creates a new Hitori game mode.
func NewMode(title, desc string, size int, blackRatio float64) HitoriMode {
	return HitoriMode{
		BaseMode:   game.NewBaseMode(title, desc),
		Size:       size,
		BlackRatio: blackRatio,
	}
}

// Spawn creates a new game instance for this mode.
func (h HitoriMode) Spawn() (game.Gamer, error) {
	puzzle, err := Generate(h.Size, h.BlackRatio)
	if err != nil {
		return nil, err
	}
	return New(h, puzzle)
}

func (h HitoriMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	puzzle, err := GenerateSeeded(h.Size, h.BlackRatio, rng)
	if err != nil {
		return nil, err
	}
	return New(h, puzzle)
}

// Modes defines the available difficulty levels.
// Difficulty is controlled by grid size and the ratio of black cells.
// Smaller grids with fewer black cells are easier to deduce; larger grids
// with more black cells require deeper chains of logic.
var Modes = []list.Item{
	NewMode("Mini", "5\u00d75 grid, gentle introduction.", 5, 0.32),
	NewMode("Easy", "6\u00d76 grid, straightforward logic.", 6, 0.32),
	NewMode("Medium", "8\u00d78 grid, moderate challenge.", 8, 0.30),
	NewMode("Tricky", "9\u00d79 grid, requires careful deduction.", 9, 0.30),
	NewMode("Hard", "10\u00d710 grid, advanced logic chains.", 10, 0.30),
	NewMode("Expert", "12\u00d712 grid, maximum challenge.", 12, 0.28),
}
