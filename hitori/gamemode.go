package hitori

import (
	_ "embed"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/gamereg"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

//go:embed help.md
var HelpContent string

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

func NewMode(title, desc string, size int, blackRatio float64) HitoriMode {
	return HitoriMode{
		BaseMode:   game.NewBaseMode(title, desc),
		Size:       size,
		BlackRatio: blackRatio,
	}
}

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

var Modes = []game.Mode{
	NewMode("Mini", "5\u00d75 grid, gentle introduction.", 5, 0.32),
	NewMode("Easy", "6\u00d76 grid, straightforward logic.", 6, 0.32),
	NewMode("Medium", "8\u00d78 grid, moderate challenge.", 8, 0.30),
	NewMode("Tricky", "9\u00d79 grid, requires careful deduction.", 9, 0.30),
	NewMode("Hard", "10\u00d710 grid, advanced logic chains.", 10, 0.30),
	NewMode("Expert", "12\u00d712 grid, maximum challenge.", 12, 0.28),
}

var ModeDefinitions = gamereg.BuildModeDefs(Modes)

var Definition = puzzle.NewDefinition(puzzle.DefinitionSpec{
	Name:         "Hitori",
	Description:  "Shade the cells to eliminate duplicates.",
	Modes:        ModeDefinitions,
	DailyModeIDs: puzzle.SelectModeIDsByIndex(ModeDefinitions, 1, 2),
})

var Entry = gamereg.NewEntry(gamereg.EntrySpec{
	Definition: Definition,
	Help:       HelpContent,
	Import:     game.AdaptImport(ImportModel),
	Modes:      Modes,
})
