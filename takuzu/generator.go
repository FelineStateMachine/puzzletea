package takuzu

import "math/rand/v2"

type cellPos struct{ x, y int }

// generateComplete fills an empty grid with a valid Takuzu solution using
// backtracking with retry on failure.
func generateComplete(size int) grid {
	g := newGrid(createEmptyState(size))

	const maxRetries = 10
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Reset grid
		for y := range size {
			for x := range size {
				g[y][x] = emptyCell
			}
		}

		if fillGrid(g, size) {
			return g
		}
	}

	// Last attempt - use more exhaustive search
	for y := range size {
		for x := range size {
			g[y][x] = emptyCell
		}
	}
	if fillGrid(g, size) {
		return g
	}

	panic("failed to generate complete grid")
}

func fillGrid(g grid, size int) bool {
	order := make([]cellPos, 0, size*size)
	for y := range size {
		for x := range size {
			if g[y][x] == emptyCell {
				order = append(order, cellPos{x, y})
			}
		}
	}

	rand.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	for y := range size {
		for x := range size {
			if g[y][x] != emptyCell {
				continue
			}
			vals := [2]rune{zeroCell, oneCell}
			if rand.IntN(2) == 0 {
				vals[0], vals[1] = vals[1], vals[0]
			}
			for _, v := range vals {
				if canPlace(g, size, x, y, v) {
					g[y][x] = v
					if fillGrid(g, size) {
						return true
					}
					g[y][x] = emptyCell
				}
			}
			return false
		}
	}
	return true
}

// generatePuzzle removes cells from a complete grid to create a puzzle with a unique solution.
func generatePuzzle(complete grid, size int, prefilled float64) (puzzle grid, provided [][]bool) {
	puzzle = complete.clone()
	provided = make([][]bool, size)
	for y := range size {
		provided[y] = make([]bool, size)
		for x := range size {
			provided[y][x] = true
		}
	}

	target := int(prefilled * float64(size*size))

	positions := make([]cellPos, 0, size*size)
	for y := range size {
		for x := range size {
			positions = append(positions, cellPos{x, y})
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	filled := size * size
	for _, p := range positions {
		if filled <= target {
			break
		}
		saved := puzzle[p.y][p.x]
		puzzle[p.y][p.x] = emptyCell
		if countSolutions(puzzle, size, 2) != 1 {
			puzzle[p.y][p.x] = saved
		} else {
			provided[p.y][p.x] = false
			filled--
		}
	}

	return puzzle, provided
}

// countSolutions counts solutions of the grid up to limit using backtracking.
func countSolutions(g grid, size, limit int) int {
	for y := range size {
		for x := range size {
			if g[y][x] != emptyCell {
				continue
			}
			count := 0
			for _, v := range [2]rune{zeroCell, oneCell} {
				if canPlace(g, size, x, y, v) {
					g[y][x] = v
					count += countSolutions(g, size, limit-count)
					g[y][x] = emptyCell
					if count >= limit {
						return count
					}
				}
			}
			return count
		}
	}
	if hasUniqueLines(g, size) {
		return 1
	}
	return 0
}

// checkConstraints verifies no-triples and equal-count rules for every row and column.
func checkConstraints(g grid, size int) bool {
	half := size / 2
	for i := range size {
		zeroRow, oneRow := 0, 0
		zeroCol, oneCol := 0, 0
		for j := range size {
			switch g[i][j] {
			case zeroCell:
				zeroRow++
			case oneCell:
				oneRow++
			}
			switch g[j][i] {
			case zeroCell:
				zeroCol++
			case oneCell:
				oneCol++
			}
			if j >= 2 && g[i][j] == g[i][j-1] && g[i][j] == g[i][j-2] {
				return false
			}
			if j >= 2 && g[j][i] == g[j-1][i] && g[j][i] == g[j-2][i] {
				return false
			}
		}
		if zeroRow != half || oneRow != half {
			return false
		}
		if zeroCol != half || oneCol != half {
			return false
		}
	}
	return true
}

// hasUniqueLines checks that all rows are unique and all columns are unique.
func hasUniqueLines(g grid, size int) bool {
	for i := range size {
		for j := i + 1; j < size; j++ {
			if rowEqual(g[i], g[j]) {
				return false
			}
		}
	}
	for i := range size {
		for j := i + 1; j < size; j++ {
			if colEqual(g, size, i, j) {
				return false
			}
		}
	}
	return true
}

func rowEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func colEqual(g grid, size, c1, c2 int) bool {
	for r := range size {
		if g[r][c1] != g[r][c2] {
			return false
		}
	}
	return true
}

// canPlace checks whether placing val at (x,y) would violate Takuzu constraints.
func canPlace(g grid, size, x, y int, val rune) bool {
	if x >= 2 && g[y][x-1] == val && g[y][x-2] == val {
		return false
	}
	if x >= 1 && x < size-1 && g[y][x-1] == val && g[y][x+1] == val {
		return false
	}
	if x <= size-3 && g[y][x+1] == val && g[y][x+2] == val {
		return false
	}

	if y >= 2 && g[y-1][x] == val && g[y-2][x] == val {
		return false
	}
	if y >= 1 && y < size-1 && g[y-1][x] == val && g[y+1][x] == val {
		return false
	}
	if y <= size-3 && g[y+1][x] == val && g[y+2][x] == val {
		return false
	}

	half := size / 2
	count := 0
	for _, c := range g[y] {
		if c == val {
			count++
		}
	}
	if count >= half {
		return false
	}

	count = 0
	for r := range size {
		if g[r][x] == val {
			count++
		}
	}
	if count >= half {
		return false
	}

	if rowFilledExcept(g, y, x, size) {
		for other := range size {
			if other != y && rowFilled(g, other, size) && rowEqualWith(g[y], g[other], x, val) {
				return false
			}
		}
	}

	if colFilledExcept(g, x, y, size) {
		for other := range size {
			if other != x && colFilled(g, other, size) && colEqualWith(g, size, x, other, y, val) {
				return false
			}
		}
	}

	return true
}

func rowFilledExcept(g grid, y, skip, size int) bool {
	for x := range size {
		if x != skip && g[y][x] == emptyCell {
			return false
		}
	}
	return true
}

func colFilledExcept(g grid, x, skip, size int) bool {
	for y := range size {
		if y != skip && g[y][x] == emptyCell {
			return false
		}
	}
	return true
}

func rowEqualWith(a, b []rune, overrideIdx int, overrideVal rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		av := a[i]
		if i == overrideIdx {
			av = overrideVal
		}
		if av != b[i] {
			return false
		}
	}
	return true
}

func colEqualWith(g grid, size, c1, c2, overrideRow int, overrideVal rune) bool {
	for r := range size {
		v1 := g[r][c1]
		if r == overrideRow {
			v1 = overrideVal
		}
		if v1 != g[r][c2] {
			return false
		}
	}
	return true
}

func rowFilled(g grid, y, size int) bool {
	for x := range size {
		if g[y][x] == emptyCell {
			return false
		}
	}
	return true
}

func colFilled(g grid, x, size int) bool {
	for y := range size {
		if g[y][x] == emptyCell {
			return false
		}
	}
	return true
}
