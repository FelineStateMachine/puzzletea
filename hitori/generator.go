package hitori

import (
	"errors"
	"math/rand"
)

func GeneratePuzzle(mode HitoriMode) (puzzle grid, provided [][]bool, err error) {
	size := mode.Size
	prefilled := mode.Prefilled

	for attempts := 0; attempts < 50; attempts++ {
		puzzle = generateSimplePuzzle(size, prefilled)
		if puzzle != nil {
			provided = make([][]bool, size)
			for i := range provided {
				provided[i] = make([]bool, size)
			}
			for y := 0; y < size; y++ {
				for x := 0; x < size; x++ {
					if puzzle[y][x] != emptyCell && puzzle[y][x] != shadedCell {
						provided[y][x] = true
					}
				}
			}
			return puzzle, provided, nil
		}
	}

	return nil, nil, errors.New("failed to generate hitori puzzle")
}

func generateSimplePuzzle(size int, prefilled float64) grid {
	puzzle := make(grid, size)
	for i := range puzzle {
		puzzle[i] = make([]rune, size)
		for j := range puzzle[i] {
			puzzle[i][j] = emptyCell
		}
	}

	for y := 0; y < size; y++ {
		used := make(map[rune]bool)
		avail := make([]int, 0, size)
		for x := 0; x < size; x++ {
			avail = append(avail, x)
		}
		rand.Shuffle(len(avail), func(i, j int) {
			avail[i], avail[j] = avail[j], avail[i]
		})

		rowVals := rand.Perm(size)
		idx := 0
		for _, x := range avail {
			val := rune(rowVals[idx] + 1)
			if !used[val] {
				puzzle[y][x] = val
				used[val] = true
			}
			idx++
		}
	}

	targetFilled := int(float64(size*size) * prefilled)
	cells := make([][2]int, 0, size*size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			cells = append(cells, [2]int{x, y})
		}
	}
	rand.Shuffle(len(cells), func(i, j int) {
		cells[i], cells[j] = cells[j], cells[i]
	})

	filled := size * size
	for _, cell := range cells {
		if filled <= targetFilled {
			break
		}
		x, y := cell[0], cell[1]
		if puzzle[y][x] != emptyCell {
			original := puzzle[y][x]
			puzzle[y][x] = shadedCell
			filled--

			if !hasValidSolution(puzzle, size) {
				puzzle[y][x] = original
				filled++
			}
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if puzzle[y][x] == shadedCell {
				puzzle[y][x] = emptyCell
			}
		}
	}

	return puzzle
}

func hasValidSolution(g grid, size int) bool {
	work := make(grid, size)
	for i := range work {
		work[i] = make([]rune, size)
		copy(work[i], g[i])
	}

	return solveHitori(work, size)
}

func solveHitori(g grid, size int) bool {
	var backtrack func(x, y int) bool
	backtrack = func(x, y int) bool {
		if y >= size {
			return isValidSolution(g, size)
		}

		nextX := x + 1
		nextY := y
		if nextX >= size {
			nextX = 0
			nextY = y + 1
		}

		if g[y][x] == shadedCell || g[y][x] == emptyCell {
			return backtrack(nextX, nextY)
		}

		original := g[y][x]

		if canShade(g, x, y, size) {
			g[y][x] = shadedCell
			if backtrack(nextX, nextY) {
				return true
			}
		}

		g[y][x] = emptyCell
		if backtrack(nextX, nextY) {
			return true
		}

		g[y][x] = original
		return false
	}

	return backtrack(0, 0)
}

func canShade(g grid, x, y, size int) bool {
	if x > 0 && g[y][x-1] == shadedCell {
		return false
	}
	if x+1 < size && g[y][x+1] == shadedCell {
		return false
	}
	if y > 0 && g[y-1][x] == shadedCell {
		return false
	}
	if y+1 < size && g[y+1][x] == shadedCell {
		return false
	}
	return true
}

func isValidSolution(g grid, size int) bool {
	if !checkConstraints(g, size) {
		return false
	}
	if !isConnected(g, size) {
		return false
	}
	return true
}

func checkConstraints(g grid, size int) bool {
	for y := 0; y < size; y++ {
		seen := make(map[rune]bool)
		for x := 0; x < size; x++ {
			val := g[y][x]
			if val == shadedCell || val == emptyCell {
				continue
			}
			if seen[val] {
				return false
			}
			seen[val] = true
		}
	}

	for x := 0; x < size; x++ {
		seen := make(map[rune]bool)
		for y := 0; y < size; y++ {
			val := g[y][x]
			if val == shadedCell || val == emptyCell {
				continue
			}
			if seen[val] {
				return false
			}
			seen[val] = true
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if g[y][x] == shadedCell {
				if x+1 < size && g[y][x+1] == shadedCell {
					return false
				}
				if y+1 < size && g[y+1][x] == shadedCell {
					return false
				}
			}
		}
	}

	return true
}

func isConnected(g grid, size int) bool {
	var startX, startY int
	found := false
	for y := 0; y < size && !found; y++ {
		for x := 0; x < size; x++ {
			if g[y][x] != shadedCell {
				startX, startY = x, y
				found = true
				break
			}
		}
	}
	if !found {
		return false
	}

	visited := make([][]bool, size)
	for i := range visited {
		visited[i] = make([]bool, size)
	}

	stack := [][2]int{{startX, startY}}
	visited[startY][startX] = true
	count := 0
	total := 0

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if g[y][x] != shadedCell {
				total++
			}
		}
	}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		count++

		x, y := curr[0], curr[1]
		if x > 0 && !visited[y][x-1] && g[y][x-1] != shadedCell {
			visited[y][x-1] = true
			stack = append(stack, [2]int{x - 1, y})
		}
		if x+1 < size && !visited[y][x+1] && g[y][x+1] != shadedCell {
			visited[y][x+1] = true
			stack = append(stack, [2]int{x + 1, y})
		}
		if y > 0 && !visited[y-1][x] && g[y-1][x] != shadedCell {
			visited[y-1][x] = true
			stack = append(stack, [2]int{x, y - 1})
		}
		if y+1 < size && !visited[y+1][x] && g[y+1][x] != shadedCell {
			visited[y+1][x] = true
			stack = append(stack, [2]int{x, y + 1})
		}
	}

	return count == total
}
