package takuzu

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

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
	_ game.Mode    = TakuzuMode{}
	_ game.Spawner = TakuzuMode{}
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

// Modes defines the available difficulty levels.
var Modes = []list.Item{
	NewMode("Easy", "6×6 grid, ~50% clues.", 6, 0.50),
	NewMode("Medium", "8×8 grid, ~40% clues.", 8, 0.40),
	NewMode("Hard", "10×10 grid, ~35% clues.", 10, 0.35),
	NewMode("Expert", "12×12 grid, ~30% clues.", 12, 0.30),
}
