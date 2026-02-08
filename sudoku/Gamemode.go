package sudoku

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/list"
)

type SudokuMode struct {
	game.BaseMode
	ProvidedCount int
}

var _ game.Mode = SudokuMode{}    // compile-time interface check
var _ game.Spawner = SudokuMode{} // compile-time interface check

func NewMode(title, description string, providedCount int) SudokuMode {
	return SudokuMode{
		BaseMode:      game.NewBaseMode(title, description),
		ProvidedCount: providedCount,
	}
}

func (s SudokuMode) Spawn() (game.Gamer, error) {
	return New(s, GenerateProvidedCells(s))
}

var Modes = []list.Item{
	NewMode("Easy - 38 Provided Cells", "A random sudoku with at least 38 cells provided to start.", 38),
	NewMode("Hard - 26 Provided Cells", "A random sudoku with at least 26 cells provided to start.", 26),
}
