package hashiwokakero

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

func init() {
	game.Register("Hashiwokakero", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

// HashiMode defines a hashiwokakero difficulty/configuration.
type HashiMode struct {
	game.BaseMode
	Width      int
	Height     int
	MinIslands int
	MaxIslands int
}

var (
	_ game.Mode    = HashiMode{} // compile-time interface check
	_ game.Spawner = HashiMode{} // compile-time interface check
)

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
	NewMode("Easy 7x7", "7x7 grid with 8-10 islands.", 7, 7, 8, 10),
	NewMode("Medium 7x7", "7x7 grid with 12-15 islands.", 7, 7, 12, 15),
	NewMode("Hard 7x7", "7x7 grid with 17-20 islands.", 7, 7, 17, 20),
	NewMode("Easy 9x9", "9x9 grid with 12-16 islands.", 9, 9, 12, 16),
	NewMode("Medium 9x9", "9x9 grid with 20-24 islands.", 9, 9, 20, 24),
	NewMode("Hard 9x9", "9x9 grid with 28-32 islands.", 9, 9, 28, 32),
	NewMode("Easy 11x11", "11x11 grid with 18-24 islands.", 11, 11, 18, 24),
	NewMode("Medium 11x11", "11x11 grid with 30-36 islands.", 11, 11, 30, 36),
	NewMode("Hard 11x11", "11x11 grid with 42-48 islands.", 11, 11, 42, 48),
	NewMode("Easy 13x13", "13x13 grid with 25-34 islands.", 13, 13, 25, 34),
	NewMode("Medium 13x13", "13x13 grid with 42-51 islands.", 13, 13, 42, 51),
	NewMode("Hard 13x13", "13x13 grid with 59-68 islands.", 13, 13, 59, 68),
}
