package sudoku

import (
	_ "embed"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/gamereg"
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

var ModeDefinitions = gamereg.BuildModeDefs(Modes)

var Definition = puzzle.NewDefinition(puzzle.DefinitionSpec{
	Name:         "Sudoku",
	Description:  "Fill the 9x9 grid following sudoku rules.",
	Modes:        ModeDefinitions,
	DailyModeIDs: puzzle.SelectModeIDsByIndex(ModeDefinitions, 1, 2),
})

var Entry = gamereg.NewEntry(gamereg.EntrySpec{
	Definition: Definition,
	Help:       HelpContent,
	Import:     game.AdaptImport(ImportModel),
	Modes:      Modes,
})
