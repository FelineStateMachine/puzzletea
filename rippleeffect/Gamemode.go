package rippleeffect

import (
	_ "embed"
	"math/rand/v2"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

type Mode struct {
	game.BaseMode
	Size       int
	MaxCage    int
	GivenRatio float64
}

var (
	_ game.Mode          = Mode{}
	_ game.Spawner       = Mode{}
	_ game.SeededSpawner = Mode{}
)

func NewMode(title, description string, size, maxCage int, givenRatio float64) Mode {
	return Mode{
		BaseMode:   game.NewBaseMode(title, description),
		Size:       size,
		MaxCage:    maxCage,
		GivenRatio: givenRatio,
	}
}

func (m Mode) Spawn() (game.Gamer, error) {
	puzzle, err := GeneratePuzzle(m.Size, m.Size, m.MaxCage, m.GivenRatio)
	if err != nil {
		return nil, err
	}
	return New(m, puzzle)
}

func (m Mode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	puzzle, err := GeneratePuzzleSeeded(m.Size, m.Size, m.MaxCage, m.GivenRatio, rng)
	if err != nil {
		return nil, err
	}
	return New(m, puzzle)
}

var Modes = []list.Item{
	NewMode("Mini 5x5", "Small board with compact cages.", 5, 3, 0.76),
	NewMode("Easy 6x6", "Gentle introduction to ripple spacing.", 6, 3, 0.72),
	NewMode("Medium 7x7", "Balanced cages and row pressure.", 7, 4, 0.68),
	NewMode("Hard 8x8", "Denser ripple interactions.", 8, 4, 0.64),
	NewMode("Expert 9x9", "Wide board with longer deduction chains.", 9, 5, 0.60),
}

var DailyModes = []list.Item{
	Modes[1],
	Modes[2],
	Modes[3],
}

var Definition = game.Definition{
	Name:        "Ripple Effect",
	Description: "Place digits in cages without violating ripple distance.",
	Aliases:     []string{"ripple"},
	Modes:       Modes,
	DailyModes:  DailyModes,
	Help:        HelpContent,
	Import:      func(data []byte) (game.Gamer, error) { return ImportModel(data) },
}
