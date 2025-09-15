package sudoku

import "fmt"

type cell struct {
	x, y int
	v    int
}

type grid = [9][9]cell

func GenerateProvidedCells(m SudokuMode) []cell {
	// TODO: implement
	cells := []cell{
		{1, 2, 3},
		{3, 4, 5},
	}
	//create solved board
	// iteratively remove cells and ensure only 1 solution
	// backtrack if not unique
	// return known values
	return cells
}

func newGrid(provided []cell) grid {
	var g grid
	for i := range GRIDSIZE {
		g[i] = [GRIDSIZE]cell{}

		for j := range GRIDSIZE {
			g[i][j].x = j
			g[i][j].y = i
		}
	}
	for _, hint := range provided {
		g[hint.y][hint.x].v = hint.v
	}
	return g
}

// loadSave - Parses optional param save. If none, uses default.
func loadSave(def grid, optionalSave ...string) grid {
	var g grid
	if len(optionalSave) > 0 {
		// TODO load saved cell values.
	} else {
		g = def
	}
	return g
}

func (c cell) String() string {
	return fmt.Sprintf("r%dc%d=%d", c.x, c.y, c.v)
}
