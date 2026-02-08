package hashiwokakero

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

// HashiMode defines a hashiwokakero difficulty/configuration.
type HashiMode struct {
	game.BaseMode
	Width      int
	Height     int
	MinIslands int
	MaxIslands int
}

var _ game.Mode = HashiMode{}    // compile-time interface check
var _ game.Spawner = HashiMode{} // compile-time interface check

func NewMode(title, description string, width, height, minIslands, maxIslands int) HashiMode {
	return HashiMode{
		BaseMode:   game.NewBaseMode(title, description),
		Width:      width,
		Height:     height,
		MinIslands: minIslands,
		MaxIslands: maxIslands,
	}
}

func (h HashiMode) Spawn() (game.Gamer, error) {
	puzzle := GeneratePuzzle(h)
	return New(h, puzzle), nil
}

var Modes = []list.Item{
	NewMode("Easy - 7x7", "Connect islands with bridges on a 7x7 grid.", 7, 7, 6, 8),
	NewMode("Medium - 9x9", "Connect islands with bridges on a 9x9 grid.", 9, 9, 10, 14),
	NewMode("Hard - 13x13", "Connect islands with bridges on a 13x13 grid.", 13, 13, 18, 24),
}
