package main

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/sudoku"

	// Implementations of the Gamer interface should interface here.
	nonogram "github.com/FelineStateMachine/puzzletea/nonogram"
)

func (m model) SpawnGame(mode game.Mode) (game.Gamer, error) {
	switch mode := mode.(type) {
	case nonogram.NonogramMode:
		hints := nonogram.GenerateRandomTomography(mode)
		return nonogram.New(mode, hints)
	case sudoku.SudokuMode:
		return sudoku.New(mode, sudoku.GenerateProvidedCells(mode))
	default:
		return nil, fmt.Errorf("Faled to create game.\nUnimplemented game type `%v`.", mode)
	}
}
