package lightsout

import (
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

// Register the import function for save games.
func init() {
	game.Register("Lights Out", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

// Mode represents a specific difficulty configuration for Lights Out.
type Mode struct {
	game.BaseMode
	Width, Height int
}

// Compile-time assertions
var (
	_ game.Mode          = Mode{}
	_ game.Spawner       = Mode{}
	_ game.SeededSpawner = Mode{}
)

// NewMode creates a new game mode.
func NewMode(title, desc string, w, h int) Mode {
	return Mode{
		BaseMode: game.NewBaseMode(title, desc),
		Width:    w,
		Height:   h,
	}
}

// Spawn creates a new game instance for this mode.
func (m Mode) Spawn() (game.Gamer, error) {
	return New(m.Width, m.Height)
}

func (m Mode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	return NewSeeded(m.Width, m.Height, rng)
}

// Modes defines the available difficulty levels.
var Modes = []list.Item{
	NewMode("Easy", "3x3 grid", 3, 3),
	NewMode("Medium", "5x5 grid", 5, 5),
	NewMode("Hard", "7x7 grid", 7, 7),
	NewMode("Extreme", "9x9 grid", 9, 9),
}
