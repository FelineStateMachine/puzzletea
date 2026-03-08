package sudokurgb

import (
	_ "embed"
	"math/rand/v2"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

type SudokuRGBMode struct {
	game.BaseMode
	ProvidedCount int
}

var (
	_ game.Mode          = SudokuRGBMode{}
	_ game.Spawner       = SudokuRGBMode{}
	_ game.SeededSpawner = SudokuRGBMode{}
)

func NewMode(title, description string, providedCount int) SudokuRGBMode {
	return SudokuRGBMode{
		BaseMode:      game.NewBaseMode(title, description),
		ProvidedCount: providedCount,
	}
}

func (s SudokuRGBMode) Spawn() (game.Gamer, error) {
	return New(s, GenerateProvidedCells(s))
}

func (s SudokuRGBMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	return New(s, GenerateProvidedCellsSeeded(s, rng))
}

var Modes = []list.Item{
	NewMode("Beginner", "60 clues. Gentle intro to RGB quota logic.", 60),
	NewMode("Easy", "54 clues. Early rows and boxes resolve quickly.", 54),
	NewMode("Medium", "48 clues. Mixed row, column, and box pressure.", 48),
	NewMode("Hard", "42 clues. More ambiguous houses and cross-checking.", 42),
	NewMode("Expert", "36 clues. Sparse givens require deeper scanning.", 36),
	NewMode("Diabolical", "30 clues. Tightest clue budget in the launch set.", 30),
}

var DailyModes = []list.Item{
	Modes[1], // Easy
	Modes[2], // Medium
}

var Definition = game.Definition{
	Name:        "Sudoku RGB",
	Description: "Fill the board with RGB symbols so each row, column, and 3x3 box contains {1,1,1,2,2,2,3,3,3}. [1,2,3] maps to [▲,■,●]. Inspired by Sudoku Ripeto.",
	Aliases:     []string{"rgb sudoku", "ripeto", "sudoku ripeto"},
	Modes:       Modes,
	DailyModes:  DailyModes,
	Help:        HelpContent,
	Import:      func(data []byte) (game.Gamer, error) { return ImportModel(data) },
}
