package sudoku

import "fmt"

type cell struct {
	x, y int
	v    int
}

type grid = [9][9]cell

func newGrid(provided []cell) grid {
	var g grid
	for i := range gridSize {
		g[i] = [gridSize]cell{}

		for j := range gridSize {
			g[i][j].x = j
			g[i][j].y = i
		}
	}
	for _, hint := range provided {
		g[hint.y][hint.x].v = hint.v
	}
	return g
}

func (c cell) String() string {
	return fmt.Sprintf("r%dc%d=%d", c.y, c.x, c.v)
}

func (m Model) isSolved() bool {
	conflicts := computeConflicts(m.grid)
	return isSolvedWith(m.grid, conflicts)
}

func isSolvedWith(g grid, conflicts [gridSize][gridSize]bool) bool {
	for y := range gridSize {
		for x := range gridSize {
			if g[y][x].v == 0 || conflicts[y][x] {
				return false
			}
		}
	}
	return true
}
