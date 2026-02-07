package main

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/sudoku"

	// Implementations of the Gamer interface should interface here.
	nonogram "github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
)

func (m model) SpawnGame(mode game.Mode) (game.Gamer, error) {
	switch mode := mode.(type) {
	case nonogram.NonogramMode:
		hints := nonogram.GenerateRandomTomography(mode)
		return nonogram.New(mode, hints)
	case sudoku.SudokuMode:
		return sudoku.New(mode, sudoku.GenerateProvidedCells(mode))
	case wordsearch.WordSearchMode:
		grid, words := wordsearch.GenerateWordSearch(mode.Width, mode.Height, mode.WordCount, mode.MinWordLen, mode.MaxWordLen, mode.AllowedDirs)
		return wordsearch.New(mode, grid, words), nil
	default:
		return nil, fmt.Errorf("Faled to create game.\nUnimplemented game type `%v`.", mode)
	}
}
