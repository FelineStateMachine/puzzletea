package lightsout

import (
	_ "embed"
	"math/rand/v2"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

func init() {
	game.Register("Lights Out", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

type Mode struct {
	game.BaseMode
	Width, Height int
}

var (
	_ game.Mode          = Mode{}
	_ game.Spawner       = Mode{}
	_ game.SeededSpawner = Mode{}
)

func NewMode(title, desc string, w, h int) Mode {
	return Mode{
		BaseMode: game.NewBaseMode(title, desc),
		Width:    w,
		Height:   h,
	}
}

func (m Mode) Spawn() (game.Gamer, error) {
	return New(m.Width, m.Height)
}

func (m Mode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	return NewSeeded(m.Width, m.Height, rng)
}

var Modes = []list.Item{
	NewMode("Easy", "3x3 grid", 3, 3),
	NewMode("Medium", "5x5 grid", 5, 5),
	NewMode("Hard", "7x7 grid", 7, 7),
	NewMode("Extreme", "9x9 grid", 9, 9),
}

var DailyModes = []list.Item{
	Modes[1], // Medium 5x5
	Modes[2], // Hard 7x7
}
