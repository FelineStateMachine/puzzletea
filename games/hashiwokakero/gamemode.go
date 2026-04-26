package hashiwokakero

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

type HashiMode struct {
	game.BaseMode
	Width      int
	Height     int
	MinIslands int
	MaxIslands int
}

var (
	_ game.Mode          = HashiMode{}
	_ game.Spawner       = HashiMode{}
	_ game.SeededSpawner = HashiMode{}
	_ game.EloSpawner    = HashiMode{}
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
	puzzle, err := GeneratePuzzle(h)
	if err != nil {
		return nil, err
	}
	return New(h, puzzle), nil
}

func (h HashiMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	puzzle, err := GeneratePuzzleSeeded(h, rng)
	if err != nil {
		return nil, err
	}
	return New(h, puzzle), nil
}

var Modes = []game.Mode{
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

var ModeDefinitions = hashiModeDefinitions(Modes)

func hashiModeDefinitions(modes []game.Mode) []puzzle.ModeDef {
	defs := gameentry.BuildModeDefs(modes)
	presets := []difficulty.Elo{250, 500, 800, 700, 1100, 1500, 1300, 1700, 2100, 1900, 2400, 2900}
	for i, mode := range modes {
		if i >= len(defs) || i >= len(presets) {
			break
		}
		if _, ok := mode.(game.EloSpawner); !ok {
			continue
		}
		elo := presets[i]
		defs[i].PresetElo = &elo
	}
	return defs
}

var Definition = puzzle.NewDefinition(puzzle.DefinitionSpec{
	Name:         "Hashiwokakero",
	Description:  "Connect the islands with bridges.",
	Aliases:      []string{"hashi", "bridges"},
	Modes:        ModeDefinitions,
	DailyModeIDs: puzzle.SelectModeIDsByIndex(ModeDefinitions, 3, 1),
})

var Entry = gameentry.NewEntry(gameentry.EntrySpec{
	Definition: Definition,
	Help:       HelpContent,
	Import:     game.AdaptImport(ImportModel),
	Modes:      Modes,
	Print:      PDFPrintAdapter,
})
