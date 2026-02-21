package hitori

import (
	"errors"
	"math/rand/v2"
)

type cellPos struct{ x, y int }

// generateLatinSquare creates an NxN grid where each number 1..N appears
// exactly once in each row and column. Uses cyclic shifts with row and
// column shuffling.
func generateLatinSquare(size int) grid {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return generateLatinSquareSeeded(size, rng)
}

func generateLatinSquareSeeded(size int, rng *rand.Rand) grid {
	firstRow := make([]rune, size)
	for i := range size {
		firstRow[i] = rune('1' + i)
	}
	rng.Shuffle(len(firstRow), func(i, j int) {
		firstRow[i], firstRow[j] = firstRow[j], firstRow[i]
	})

	g := make(grid, size)
	for y := range size {
		g[y] = make([]rune, size)
		for x := range size {
			g[y][x] = firstRow[(x+y)%size]
		}
	}

	cols := make([]int, size)
	for i := range size {
		cols[i] = i
	}
	rng.Shuffle(len(cols), func(i, j int) {
		cols[i], cols[j] = cols[j], cols[i]
	})

	rowOrder := make([]int, size)
	for i := range size {
		rowOrder[i] = i
	}
	rng.Shuffle(len(rowOrder), func(i, j int) {
		rowOrder[i], rowOrder[j] = rowOrder[j], rowOrder[i]
	})

	result := make(grid, size)
	for y := range size {
		result[y] = make([]rune, size)
		for x := range size {
			result[y][x] = g[rowOrder[y]][cols[x]]
		}
	}

	return result
}

// generateValidMask determines which cells should be black. The mask
// satisfies: no adjacent black cells, and all white cells connected.
func generateValidMask(size int, blackRatio float64) [][]bool {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return generateValidMaskSeeded(size, blackRatio, rng)
}

func generateValidMaskSeeded(size int, blackRatio float64, rng *rand.Rand) [][]bool {
	mask := make([][]bool, size)
	for y := range size {
		mask[y] = make([]bool, size)
	}

	targetBlack := int(blackRatio * float64(size*size))

	positions := make([]cellPos, 0, size*size)
	for y := range size {
		for x := range size {
			positions = append(positions, cellPos{x, y})
		}
	}
	rng.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	blackCount := 0
	for _, p := range positions {
		if blackCount >= targetBlack {
			break
		}

		if hasOrthogonalNeighbor(mask, size, p.x, p.y) {
			continue
		}

		mask[p.y][p.x] = true

		if requiresConnectivityCheck(mask, size, p.x, p.y) &&
			!whiteCellsConnected(mask, size) {
			mask[p.y][p.x] = false
			continue
		}

		blackCount++
	}

	return mask
}

func hasOrthogonalNeighbor(mask [][]bool, size, x, y int) bool {
	if x > 0 && mask[y][x-1] {
		return true
	}
	if x < size-1 && mask[y][x+1] {
		return true
	}
	if y > 0 && mask[y-1][x] {
		return true
	}
	if y < size-1 && mask[y+1][x] {
		return true
	}
	return false
}

func requiresConnectivityCheck(mask [][]bool, size, x, y int) bool {
	neighbors := 0
	if x > 0 && !mask[y][x-1] {
		neighbors++
	}
	if x < size-1 && !mask[y][x+1] {
		neighbors++
	}
	if y > 0 && !mask[y-1][x] {
		neighbors++
	}
	if y < size-1 && !mask[y+1][x] {
		neighbors++
	}
	return neighbors > 1
}

func whiteCellsConnected(mask [][]bool, size int) bool {
	startX, startY := -1, -1
	whiteCount := 0
	for y := range size {
		for x := range size {
			if !mask[y][x] {
				whiteCount++
				if startX == -1 {
					startX, startY = x, y
				}
			}
		}
	}
	if whiteCount <= 1 {
		return true
	}

	visited := make([][]bool, size)
	for y := range size {
		visited[y] = make([]bool, size)
	}

	queue := []cellPos{{startX, startY}}
	visited[startY][startX] = true
	visitedCount := 1

	dirs := [4]cellPos{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, d := range dirs {
			nx, ny := curr.x+d.x, curr.y+d.y
			if nx >= 0 && nx < size && ny >= 0 && ny < size &&
				!visited[ny][nx] && !mask[ny][nx] {
				visited[ny][nx] = true
				visitedCount++
				queue = append(queue, cellPos{nx, ny})
			}
		}
	}

	return visitedCount == whiteCount
}

// constructPuzzleSeeded builds the initial puzzle grid from a Latin square and mask.
// White cells keep their original number. Black cells get a duplicate from
// a white cell in the same row.
func constructPuzzleSeeded(baseGrid grid, mask [][]bool, rng *rand.Rand) grid {
	size := len(baseGrid)
	puzzle := baseGrid.clone()

	for y := range size {
		for x := range size {
			if !mask[y][x] {
				continue
			}
			// Collect white cell numbers in same row.
			var rowNums []rune
			for cx := range size {
				if cx != x && !mask[y][cx] {
					rowNums = append(rowNums, baseGrid[y][cx])
				}
			}
			if len(rowNums) > 0 {
				puzzle[y][x] = rowNums[rng.IntN(len(rowNums))]
				continue
			}
			// Fallback: column numbers.
			var colNums []rune
			for cy := range size {
				if cy != y && !mask[cy][x] {
					colNums = append(colNums, baseGrid[cy][x])
				}
			}
			if len(colNums) > 0 {
				puzzle[y][x] = colNums[rng.IntN(len(colNums))]
			}
		}
	}
	return puzzle
}

// refinePuzzleSeeded tries different number assignments for black cells to achieve
// a unique solution. It iterates through each black cell trying all possible
// duplicate values, checking if the resulting puzzle has exactly one solution.
func refinePuzzleSeeded(puzzle grid, mask [][]bool, size int, rng *rand.Rand) (grid, bool) {
	// Collect black cell positions.
	var blackCells []cellPos
	for y := range size {
		for x := range size {
			if mask[y][x] {
				blackCells = append(blackCells, cellPos{x, y})
			}
		}
	}

	if len(blackCells) == 0 {
		return puzzle, countPuzzleSolutions(puzzle, size, 2) == 1
	}

	// Try to refine each black cell's value to reduce solution count.
	refined := puzzle.clone()
	currentCount := countPuzzleSolutions(refined, size, 3)
	for _, bc := range blackCells {
		// Collect candidate numbers for this black cell.
		candidates := cellCandidates(refined, mask, size, bc.x, bc.y)
		rankRefinementCandidates(refined, mask, size, bc.x, bc.y, candidates, rng)

		best := refined[bc.y][bc.x]
		bestCount := currentCount

		for _, num := range candidates {
			if num == best {
				continue
			}
			refined[bc.y][bc.x] = num
			c := countPuzzleSolutions(refined, size, 3)
			if c == 1 {
				return refined, true
			}
			if c > 0 && c < bestCount {
				best = num
				bestCount = c
			}
		}
		refined[bc.y][bc.x] = best
		currentCount = bestCount

		if currentCount == 1 {
			return refined, true
		}
	}

	return refined, countPuzzleSolutions(refined, size, 2) == 1
}

func rankRefinementCandidates(
	puzzle grid,
	mask [][]bool,
	size, x, y int,
	candidates []rune,
	rng *rand.Rand,
) {
	if len(candidates) <= 1 {
		return
	}

	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	for i := 1; i < len(candidates); i++ {
		key := candidates[i]
		keyScore := candidateConstraintScore(puzzle, mask, size, x, y, key)
		j := i - 1
		for ; j >= 0; j-- {
			score := candidateConstraintScore(puzzle, mask, size, x, y, candidates[j])
			if score >= keyScore {
				break
			}
			candidates[j+1] = candidates[j]
		}
		candidates[j+1] = key
	}
}

func candidateConstraintScore(puzzle grid, mask [][]bool, size, x, y int, num rune) int {
	score := 0
	for cx := range size {
		if cx != x && !mask[y][cx] && puzzle[y][cx] == num {
			score++
		}
	}
	for cy := range size {
		if cy != y && !mask[cy][x] && puzzle[cy][x] == num {
			score++
		}
	}
	return score
}

// cellCandidates returns all valid number choices for a black cell at (x,y):
// any number that appears among white cells in the same row or column.
func cellCandidates(puzzle grid, mask [][]bool, size, x, y int) []rune {
	seen := map[rune]bool{}
	for cx := range size {
		if cx != x && !mask[y][cx] {
			seen[puzzle[y][cx]] = true
		}
	}
	for cy := range size {
		if cy != y && !mask[cy][x] {
			seen[puzzle[cy][x]] = true
		}
	}
	result := make([]rune, 0, len(seen))
	for num := range seen {
		result = append(result, num)
	}
	return result
}

// solverState is used during backtracking to track cell assignments.
type solverState int

const (
	unknown solverState = iota
	white
	black
)

func countPuzzleSolutions(puzzle grid, size, limit int) int {
	st := make([][]solverState, size)
	for y := range size {
		st[y] = make([]solverState, size)
	}
	order := rowMajorOrder(size)
	return solveBT(puzzle, st, order, 0, limit)
}

func solveBT(
	puzzle grid,
	st [][]solverState,
	order []cellPos,
	pos, limit int,
) int {
	size := len(st)
	if pos == len(order) {
		marks := stateToMarks(st, size)
		if allWhiteConnected(marks, size) {
			return 1
		}
		return 0
	}

	nextIdx := pickNextOrderIndex(puzzle, st, order, pos, size)
	order[pos], order[nextIdx] = order[nextIdx], order[pos]

	x, y := order[pos].x, order[pos].y
	whiteOK := canBeWhite(puzzle, st, size, x, y)
	blackOK := canBeBlack(st, size, x, y)
	count := 0

	if !whiteOK && !blackOK {
		order[pos], order[nextIdx] = order[nextIdx], order[pos]
		return 0
	}

	if whiteOK {
		st[y][x] = white
		count += solveBT(puzzle, st, order, pos+1, limit-count)
		st[y][x] = unknown
		if count >= limit {
			order[pos], order[nextIdx] = order[nextIdx], order[pos]
			return count
		}
	}

	if blackOK {
		st[y][x] = black
		count += solveBT(puzzle, st, order, pos+1, limit-count)
		st[y][x] = unknown
		if count >= limit {
			order[pos], order[nextIdx] = order[nextIdx], order[pos]
			return count
		}
	}

	order[pos], order[nextIdx] = order[nextIdx], order[pos]
	return count
}

func rowMajorOrder(size int) []cellPos {
	order := make([]cellPos, 0, size*size)
	for y := range size {
		for x := range size {
			order = append(order, cellPos{x, y})
		}
	}
	return order
}

func pickNextOrderIndex(
	puzzle grid,
	st [][]solverState,
	order []cellPos,
	pos, size int,
) int {
	const lookaheadWindow = 2

	end := pos + lookaheadWindow
	if end > len(order) {
		end = len(order)
	}

	for i := pos; i < end; i++ {
		x, y := order[i].x, order[i].y
		whiteOK := canBeWhite(puzzle, st, size, x, y)
		blackOK := canBeBlack(st, size, x, y)

		options := 0
		if whiteOK {
			options++
		}
		if blackOK {
			options++
		}
		if options <= 1 {
			return i
		}
	}

	return pos
}

func canBeWhite(puzzle grid, st [][]solverState, size, x, y int) bool {
	num := puzzle[y][x]
	for i := range size {
		if i != x && st[y][i] == white && puzzle[y][i] == num {
			return false
		}
	}
	for i := range size {
		if i != y && st[i][x] == white && puzzle[i][x] == num {
			return false
		}
	}
	return true
}

func canBeBlack(st [][]solverState, size, x, y int) bool {
	if x > 0 && st[y][x-1] == black {
		return false
	}
	if x < size-1 && st[y][x+1] == black {
		return false
	}
	if y > 0 && st[y-1][x] == black {
		return false
	}
	if y < size-1 && st[y+1][x] == black {
		return false
	}
	return true
}

func stateToMarks(st [][]solverState, size int) [][]cellMark {
	marks := make([][]cellMark, size)
	for y := range size {
		marks[y] = make([]cellMark, size)
		for x := range size {
			if st[y][x] == black {
				marks[y][x] = shaded
			}
		}
	}
	return marks
}

// Generate creates a Hitori puzzle of the given size. Returns the puzzle
// numbers grid or an error if generation fails after all attempts.
func Generate(size int, blackRatio float64) (grid, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GenerateSeeded(size, blackRatio, rng)
}

func GenerateSeeded(size int, blackRatio float64, rng *rand.Rand) (grid, error) {
	const maxAttempts = 200

	for range maxAttempts {
		baseGrid := generateLatinSquareSeeded(size, rng)
		mask := generateValidMaskSeeded(size, blackRatio, rng)
		puzzle := constructPuzzleSeeded(baseGrid, mask, rng)

		// First try the random construction.
		if countPuzzleSolutions(puzzle, size, 2) == 1 {
			return puzzle, nil
		}

		// Refine by trying different values for black cells.
		refined, ok := refinePuzzleSeeded(puzzle, mask, size, rng)
		if ok {
			return refined, nil
		}
	}

	return nil, errors.New("failed to generate hitori puzzle after max attempts")
}
