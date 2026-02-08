package sudoku

import "fmt"

type cell struct {
	x, y int
	v    int
}

type grid = [9][9]cell

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

func (m Model) isSolved() bool {
	for y := range GRIDSIZE {
		for x := range GRIDSIZE {
			c := m.grid[y][x]
			if c.v == 0 {
				return false
			}
			if hasConflict(m, c, x, y) {
				return false
			}
		}
	}
	return true
}
