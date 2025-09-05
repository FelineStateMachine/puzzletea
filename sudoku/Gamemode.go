package sudoku

import (
	"github.com/FelineStateMachine/puzzletea/game"
)

type SudokuMode struct {
	title         string
	description   string
	ProvidedCount int
}

var _ game.Mode = SudokuMode{} // Verify that T implements I.

func (n SudokuMode) Title() string       { return "sudoku\t" + n.title }
func (n SudokuMode) Description() string { return n.description }
func (n SudokuMode) FilterValue() string { return "sudoku " + n.title + " " + n.description }

func NewMode(title, description string, providedCount int) SudokuMode {
	return SudokuMode{
		title:         title,
		description:   description,
		ProvidedCount: providedCount,
	}
}
