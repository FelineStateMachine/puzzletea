package fillomino

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
	MaxRegion  int
	GivenRatio float64
}

var (
	_ game.Mode          = Mode{}
	_ game.Spawner       = Mode{}
	_ game.SeededSpawner = Mode{}
)

func NewMode(title, description string, size, maxRegion int, givenRatio float64) Mode {
	return Mode{
		BaseMode:   game.NewBaseMode(title, description),
		Size:       size,
		MaxRegion:  maxRegion,
		GivenRatio: givenRatio,
	}
}

func (m Mode) Spawn() (game.Gamer, error) {
	puzzle, err := GeneratePuzzle(m.Size, m.Size, m.MaxRegion, m.GivenRatio)
	if err != nil {
		return nil, err
	}
	return New(m, puzzle)
}

func (m Mode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	puzzle, err := GeneratePuzzleSeeded(m.Size, m.Size, m.MaxRegion, m.GivenRatio, rng)
	if err != nil {
		return nil, err
	}
	return New(m, puzzle)
}

var Modes = []list.Item{
	NewMode("Mini 5x5", "Small board with generous clues.", 5, 5, 0.70),
	NewMode("Easy 6x6", "Compact board with simple regions.", 6, 6, 0.66),
	NewMode("Medium 8x8", "Balanced deduction and region growth.", 8, 7, 0.60),
	NewMode("Hard 10x10", "Larger board with denser interactions.", 10, 8, 0.56),
	NewMode("Expert 12x12", "Wide board with long deduction chains.", 12, 9, 0.52),
}

var DailyModes = []list.Item{
	Modes[1],
	Modes[2],
	Modes[3],
}

var Definition = game.Definition{
	Name:        "Fillomino",
	Description: "Grow numbered regions to their exact sizes.",
	Aliases:     []string{"polyomino", "regions"},
	Modes:       Modes,
	DailyModes:  DailyModes,
	Help:        HelpContent,
	Import:      func(data []byte) (game.Gamer, error) { return ImportModel(data) },
}
