package sudoku

import "math/rand/v2"

// isValid checks whether placing val at (x, y) in the grid is valid.
func isValid(g *grid, val, x, y int) bool {
	for i := range GRIDSIZE {
		if i != x && g[y][i].v == val {
			return false
		}
		if i != y && g[i][x].v == val {
			return false
		}
	}
	boxX, boxY := (x/3)*3, (y/3)*3
	for by := boxY; by < boxY+3; by++ {
		for bx := boxX; bx < boxX+3; bx++ {
			if (bx != x || by != y) && g[by][bx].v == val {
				return false
			}
		}
	}
	return true
}

// fillGrid fills an empty grid with a random valid sudoku solution using backtracking.
func fillGrid(g *grid) bool {
	for y := range GRIDSIZE {
		for x := range GRIDSIZE {
			if g[y][x].v == 0 {
				order := rand.Perm(GRIDSIZE)
				for _, i := range order {
					val := i + 1
					if isValid(g, val, x, y) {
						g[y][x].v = val
						if fillGrid(g) {
							return true
						}
						g[y][x].v = 0
					}
				}
				return false
			}
		}
	}
	return true
}

// countSolutions counts solutions of the grid up to limit using backtracking.
func countSolutions(g *grid, limit int) int {
	for y := range GRIDSIZE {
		for x := range GRIDSIZE {
			if g[y][x].v == 0 {
				count := 0
				for val := 1; val <= GRIDSIZE; val++ {
					if isValid(g, val, x, y) {
						g[y][x].v = val
						count += countSolutions(g, limit-count)
						g[y][x].v = 0
						if count >= limit {
							return count
						}
					}
				}
				return count
			}
		}
	}
	return 1
}

// GenerateProvidedCells generates a random sudoku puzzle with the number of
// provided cells determined by the mode's ProvidedCount.
func GenerateProvidedCells(m SudokuMode) []cell {
	g := newGrid(nil)
	fillGrid(&g)

	// Build shuffled list of all positions
	type pos struct{ x, y int }
	positions := make([]pos, 0, GRIDSIZE*GRIDSIZE)
	for y := range GRIDSIZE {
		for x := range GRIDSIZE {
			positions = append(positions, pos{x, y})
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Iteratively remove cells, ensuring unique solution
	filled := GRIDSIZE * GRIDSIZE
	for _, p := range positions {
		if filled <= m.ProvidedCount {
			break
		}
		saved := g[p.y][p.x].v
		g[p.y][p.x].v = 0
		if countSolutions(&g, 2) != 1 {
			g[p.y][p.x].v = saved
		} else {
			filled--
		}
	}

	// Collect remaining filled cells as provided hints
	var cells []cell
	for y := range GRIDSIZE {
		for x := range GRIDSIZE {
			if g[y][x].v != 0 {
				cells = append(cells, g[y][x])
			}
		}
	}
	return cells
}
