package takuzu

import (
	_ "embed"
	"math/rand/v2"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

func init() {
	game.Register("Takuzu", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

// TakuzuMode represents a specific difficulty configuration for Takuzu.
type TakuzuMode struct {
	game.BaseMode
	Size      int
	Prefilled float64
}

var (
	_ game.Mode          = TakuzuMode{}
	_ game.Spawner       = TakuzuMode{}
	_ game.SeededSpawner = TakuzuMode{}
)

// NewMode creates a new Takuzu game mode.
func NewMode(title, desc string, size int, prefilled float64) TakuzuMode {
	return TakuzuMode{
		BaseMode:  game.NewBaseMode(title, desc),
		Size:      size,
		Prefilled: prefilled,
	}
}

// Spawn creates a new game instance for this mode.
func (t TakuzuMode) Spawn() (game.Gamer, error) {
	complete := generateComplete(t.Size)
	puzzle, provided := generatePuzzle(complete, t.Size, t.Prefilled)
	return New(t, puzzle, provided)
}

func (t TakuzuMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	complete := generateCompleteSeeded(t.Size, rng)
	puzzle, provided := generatePuzzleSeeded(complete, t.Size, t.Prefilled, rng)
	return New(t, puzzle, provided)
}

// Modes defines the available difficulty levels.
// Difficulty is controlled by two axes: grid size and clue density.
// Smaller grids with more clues need only basic pattern recognition (doubles,
// sandwich patterns). Larger grids with fewer clues require counting, uniqueness
// elimination, and deeper chains of deduction.
var Modes = []list.Item{
	NewMode("Beginner", "6×6 grid, ~50% clues. Doubles and sandwich patterns.", 6, 0.50),
	NewMode("Easy", "6×6 grid, ~40% clues. Counting required.", 6, 0.40),
	NewMode("Medium", "8×8 grid, ~40% clues. Larger grid, moderate deduction.", 8, 0.40),
	NewMode("Tricky", "10×10 grid, ~38% clues. Uniqueness rule needed.", 10, 0.38),
	NewMode("Hard", "10×10 grid, ~32% clues. Long deduction chains.", 10, 0.32),
	NewMode("Very Hard", "12×12 grid, ~30% clues. Deep logic required.", 12, 0.30),
	NewMode("Extreme", "14×14 grid, ~28% clues. Maximum challenge.", 14, 0.28),
}

// DailyModes is the subset of Modes eligible for daily puzzle rotation.
var DailyModes = []list.Item{
	Modes[2], // Medium 8x8
	Modes[3], // Tricky 10x10
}
