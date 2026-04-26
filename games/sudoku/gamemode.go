package sudoku

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

type SudokuMode struct {
	game.BaseMode
	ProvidedCount int
}

var (
	_ game.Mode          = SudokuMode{}
	_ game.Spawner       = SudokuMode{}
	_ game.SeededSpawner = SudokuMode{}
	_ game.EloSpawner    = SudokuMode{}
)

func NewMode(title, description string, providedCount int) SudokuMode {
	return SudokuMode{
		BaseMode:      game.NewBaseMode(title, description),
		ProvidedCount: providedCount,
	}
}

func (s SudokuMode) Spawn() (game.Gamer, error) {
	return New(s, GenerateProvidedCells(s))
}

func (s SudokuMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	return New(s, GenerateProvidedCellsSeeded(s, rng))
}

var Modes = []game.Mode{
	NewMode("Beginner", "45–52 clues. Single Candidate / Scanning.", 45),
	NewMode("Easy", "38–44 clues. Naked Singles.", 38),
	NewMode("Medium", "32–37 clues. Hidden Pairs / Pointing.", 32),
	NewMode("Hard", "27–31 clues. Box-Line Reduction / Triples.", 27),
	NewMode("Expert", "22–26 clues. X-Wing / Y-Wing.", 22),
	NewMode("Diabolical", "17–21 clues. Swordfish / XY-Chains.", 17),
}

var ModeDefinitions = sudokuModeDefinitions(Modes)

func sudokuModeDefinitions(modes []game.Mode) []puzzle.ModeDef {
	defs := gameentry.BuildModeDefs(modes)
	presets := []difficulty.Elo{0, 600, 1200, 1800, 2400, 3000}
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
	Name:         "Sudoku",
	Description:  "Fill the 9x9 grid following sudoku rules.",
	Modes:        ModeDefinitions,
	DailyModeIDs: puzzle.SelectModeIDsByIndex(ModeDefinitions, 1, 2),
})

var Entry = gameentry.NewEntry(gameentry.EntrySpec{
	Definition: Definition,
	Help:       HelpContent,
	Import:     game.AdaptImport(ImportModel),
	Modes:      Modes,
	Print:      PDFPrintAdapter,
})
