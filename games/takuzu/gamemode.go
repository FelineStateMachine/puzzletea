package takuzu

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

type TakuzuMode struct {
	game.BaseMode
	Size      int
	Prefilled float64
}

var (
	_ game.Mode          = TakuzuMode{}
	_ game.Spawner       = TakuzuMode{}
	_ game.SeededSpawner = TakuzuMode{}
	_ game.EloSpawner    = TakuzuMode{}
)

func NewMode(title, desc string, size int, prefilled float64) TakuzuMode {
	return TakuzuMode{
		BaseMode:  game.NewBaseMode(title, desc),
		Size:      size,
		Prefilled: prefilled,
	}
}

func (t TakuzuMode) Spawn() (game.Gamer, error) {
	complete := generateComplete(t.Size)
	puzzle, provided := generatePuzzle(complete, t.Size, t.Prefilled)
	return New(t, puzzle, provided)
}

func (t TakuzuMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	complete := generateCompleteSeeded(t.Size, rng)
	puzzle, provided := generatePuzzleSeeded(complete, t.Size, t.Prefilled, rng)
	return New(t, puzzle, provided)
}

var Modes = []game.Mode{
	NewMode("Beginner", "6×6 grid, ~50% clues. Doubles and sandwich patterns.", 6, 0.50),
	NewMode("Easy", "6×6 grid, ~40% clues. Counting required.", 6, 0.40),
	NewMode("Medium", "8×8 grid, ~40% clues. Larger grid, moderate deduction.", 8, 0.40),
	NewMode("Tricky", "10×10 grid, ~38% clues. Uniqueness rule needed.", 10, 0.38),
	NewMode("Hard", "10×10 grid, ~32% clues. Long deduction chains.", 10, 0.32),
	NewMode("Very Hard", "12×12 grid, ~30% clues. Deep logic required.", 12, 0.30),
	NewMode("Extreme", "14×14 grid, ~28% clues. Maximum challenge.", 14, 0.28),
}

var ModeDefinitions = takuzuModeDefinitions(Modes)

func takuzuModeDefinitions(modes []game.Mode) []puzzle.ModeDef {
	defs := gameentry.BuildModeDefs(modes)
	presets := []difficulty.Elo{300, 700, 1100, 1600, 2100, 2500, 2900}
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
	Name:         "Takuzu",
	Description:  "Fill the grid with ● and ○. No 3 in a row.",
	Aliases:      []string{"binairo", "binary"},
	Modes:        ModeDefinitions,
	DailyModeIDs: puzzle.SelectModeIDsByIndex(ModeDefinitions, 2, 3),
})

var Entry = gameentry.NewEntry(gameentry.EntrySpec{
	Definition: Definition,
	Help:       HelpContent,
	Import:     game.AdaptImport(ImportModel),
	Modes:      Modes,
	Print:      PDFPrintAdapter,
})
