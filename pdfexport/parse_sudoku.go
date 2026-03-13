package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type sudokuSave struct {
	Provided []sudokuCell `json:"provided"`
}

type sudokuCell struct {
	X int `json:"x"`
	Y int `json:"y"`
	V int `json:"v"`
}

func ParseSudokuPrintData(saveData []byte) (*SudokuData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save sudokuSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode sudoku save: %w", err)
	}

	var givens [9][9]int
	for _, cell := range save.Provided {
		if !isSudokuCellInBounds(cell.X, cell.Y) {
			continue
		}
		if cell.V < 1 || cell.V > 9 {
			continue
		}
		givens[cell.Y][cell.X] = cell.V
	}

	return &SudokuData{Givens: givens}, nil
}

func ParseSudokuRGBPrintData(saveData []byte) (*SudokuData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save sudokuSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode sudoku rgb save: %w", err)
	}

	var givens [9][9]int
	for _, cell := range save.Provided {
		if !isSudokuCellInBounds(cell.X, cell.Y) {
			continue
		}
		if cell.V < 1 || cell.V > 3 {
			continue
		}
		givens[cell.Y][cell.X] = cell.V
	}

	return &SudokuData{Givens: givens}, nil
}

func isSudokuCellInBounds(x, y int) bool {
	return x >= 0 && x < 9 && y >= 0 && y < 9
}
