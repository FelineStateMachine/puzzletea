package sudoku

import (
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

func init() {
	game.Register("Sudoku", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

type SudokuMode struct {
	game.BaseMode
	ProvidedCount int
}

var (
	_ game.Mode          = SudokuMode{} // compile-time interface check
	_ game.Spawner       = SudokuMode{} // compile-time interface check
	_ game.SeededSpawner = SudokuMode{} // compile-time interface check
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

var Modes = []list.Item{
	NewMode("Beginner", "45–52 clues. Single Candidate / Scanning.", 45),
	NewMode("Easy", "38–44 clues. Naked Singles.", 38),
	NewMode("Medium", "32–37 clues. Hidden Pairs / Pointing.", 32),
	NewMode("Hard", "27–31 clues. Box-Line Reduction / Triples.", 27),
	NewMode("Expert", "22–26 clues. X-Wing / Y-Wing.", 22),
	NewMode("Diabolical", "17–21 clues. Swordfish / XY-Chains.", 17),
}

// DailyModes is the subset of Modes eligible for daily puzzle rotation.
var DailyModes = []list.Item{
	Modes[1], // Easy
	Modes[2], // Medium
}
