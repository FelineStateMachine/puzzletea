package lightsout

import (
	_ "embed"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/gameentry"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

//go:embed help.md
var HelpContent string

type Mode struct {
	game.BaseMode
	Width, Height int
}

var (
	_ game.Mode          = Mode{}
	_ game.Spawner       = Mode{}
	_ game.SeededSpawner = Mode{}
	_ game.EloSpawner    = Mode{}
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

var Modes = []game.Mode{
	NewMode("Easy", "3x3 grid", 3, 3),
	NewMode("Medium", "5x5 grid", 5, 5),
	NewMode("Hard", "7x7 grid", 7, 7),
	NewMode("Extreme", "9x9 grid", 9, 9),
}

var ModeDefinitions = lightsOutModeDefinitions(Modes)

func lightsOutModeDefinitions(modes []game.Mode) []puzzle.ModeDef {
	defs := gameentry.BuildModeDefs(modes)
	presets := []difficulty.Elo{0, 1000, 2000, 3000}
	for i := range defs {
		if i >= len(presets) {
			break
		}
		elo := presets[i]
		defs[i].PresetElo = &elo
	}
	return defs
}

var Definition = puzzle.NewDefinition(puzzle.DefinitionSpec{
	Name:         "Lights Out",
	Description:  "Turn the lights off.",
	Aliases:      []string{"lights"},
	Modes:        ModeDefinitions,
	DailyModeIDs: puzzle.SelectModeIDsByIndex(ModeDefinitions, 1, 2),
})

var Entry = gameentry.NewEntry(gameentry.EntrySpec{
	Definition: Definition,
	Help:       HelpContent,
	Import:     game.AdaptImport(ImportModel),
	Modes:      Modes,
})
