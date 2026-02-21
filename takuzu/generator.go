package takuzu

import "math/rand/v2"

type cellPos struct{ x, y int }

type lineStats struct {
	rowZero   []int
	rowOne    []int
	colZero   []int
	colOne    []int
	rowFilled []int
	colFilled []int
}

type mrvChoice struct {
	x, y  int
	vals  [2]rune
	count int
}

func newLineStats(g grid, size int) *lineStats {
	stats := &lineStats{
		rowZero:   make([]int, size),
		rowOne:    make([]int, size),
		colZero:   make([]int, size),
		colOne:    make([]int, size),
		rowFilled: make([]int, size),
		colFilled: make([]int, size),
	}

	for y := range size {
		for x := range size {
			val := g[y][x]
			if val == emptyCell {
				continue
			}

			stats.rowFilled[y]++
			stats.colFilled[x]++
			switch val {
			case zeroCell:
				stats.rowZero[y]++
				stats.colZero[x]++
			case oneCell:
				stats.rowOne[y]++
				stats.colOne[x]++
			}
		}
	}

	return stats
}

func (s *lineStats) apply(x, y int, val rune, delta int) {
	s.rowFilled[y] += delta
	s.colFilled[x] += delta

	switch val {
	case zeroCell:
		s.rowZero[y] += delta
		s.colZero[x] += delta
	case oneCell:
		s.rowOne[y] += delta
		s.colOne[x] += delta
	}
}

// generateComplete fills an empty grid with a valid Takuzu solution using
// backtracking with retry on failure.
func generateComplete(size int) grid {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return generateCompleteSeeded(size, rng)
}

func generateCompleteSeeded(size int, rng *rand.Rand) grid {
	g := newGrid(createEmptyState(size))

	const maxRetries = 10
	for attempt := 0; attempt < maxRetries; attempt++ {
		for y := range size {
			for x := range size {
				g[y][x] = emptyCell
			}
		}

		if fillGridSeeded(g, size, rng) {
			return g
		}
	}

	// Last attempt - use more exhaustive search
	for y := range size {
		for x := range size {
			g[y][x] = emptyCell
		}
	}
	if fillGridSeeded(g, size, rng) {
		return g
	}

	panic("failed to generate complete grid")
}

func fillGrid(g grid, size int) bool {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return fillGridSeeded(g, size, rng)
}

func fillGridSeeded(g grid, size int, rng *rand.Rand) bool {
	stats := newLineStats(g, size)
	return fillGridSeededWithStats(g, size, rng, stats)
}

func fillGridSeededWithStats(g grid, size int, rng *rand.Rand, stats *lineStats) bool {
	choice := selectMRVCell(g, size, stats, rng)
	if choice.x < 0 {
		return true
	}
	if choice.count == 0 {
		return false
	}

	if choice.count == 2 && rng.IntN(2) == 0 {
		choice.vals[0], choice.vals[1] = choice.vals[1], choice.vals[0]
	}

	for i := range choice.count {
		v := choice.vals[i]
		g[choice.y][choice.x] = v
		stats.apply(choice.x, choice.y, v, 1)
		if fillGridSeededWithStats(g, size, rng, stats) {
			return true
		}
		stats.apply(choice.x, choice.y, v, -1)
		g[choice.y][choice.x] = emptyCell
	}

	return false
}

func selectMRVCell(g grid, size int, stats *lineStats, rng *rand.Rand) mrvChoice {
	choice := mrvChoice{x: -1, y: -1, count: 3}
	tieCount := 0

	for y := range size {
		for x := range size {
			if g[y][x] != emptyCell {
				continue
			}

			vals, count := cellCandidates(g, size, x, y, stats)
			if count == 0 {
				return mrvChoice{x: x, y: y}
			}

			if count < choice.count {
				choice = mrvChoice{x: x, y: y, vals: vals, count: count}
				tieCount = 1
				continue
			}

			if count == choice.count {
				tieCount++
				if rng != nil && rng.IntN(tieCount) == 0 {
					choice = mrvChoice{x: x, y: y, vals: vals, count: count}
				}
			}
		}
	}

	if choice.x < 0 {
		return mrvChoice{x: -1, y: -1, count: 0}
	}

	return choice
}

func cellCandidates(g grid, size, x, y int, stats *lineStats) ([2]rune, int) {
	var vals [2]rune
	count := 0

	if canPlaceWithStats(g, size, x, y, zeroCell, stats) {
		vals[count] = zeroCell
		count++
	}
	if canPlaceWithStats(g, size, x, y, oneCell, stats) {
		vals[count] = oneCell
		count++
	}

	return vals, count
}

// generatePuzzle removes cells from a complete grid to create a puzzle with a unique solution.
func generatePuzzle(complete grid, size int, prefilled float64) (puzzle grid, provided [][]bool) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return generatePuzzleSeeded(complete, size, prefilled, rng)
}

func generatePuzzleSeeded(complete grid, size int, prefilled float64, rng *rand.Rand) (puzzle grid, provided [][]bool) {
	puzzle = complete.clone()
	provided = make([][]bool, size)
	for y := range size {
		provided[y] = make([]bool, size)
		for x := range size {
			provided[y][x] = true
		}
	}

	target := int(prefilled * float64(size*size))

	rowProvided := make([]int, size)
	colProvided := make([]int, size)
	for i := range size {
		rowProvided[i] = size
		colProvided[i] = size
	}

	candidates := make([]cellPos, 0, size*size)
	for y := range size {
		for x := range size {
			candidates = append(candidates, cellPos{x, y})
		}
	}
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	stats := newLineStats(puzzle, size)
	filled := size * size
	for len(candidates) > 0 && filled > target {
		bestIdx := pickRemovalCandidate(candidates, puzzle, rowProvided, colProvided, rng)
		p := candidates[bestIdx]
		last := len(candidates) - 1
		candidates[bestIdx] = candidates[last]
		candidates = candidates[:last]

		if puzzle[p.y][p.x] == emptyCell {
			continue
		}

		saved := puzzle[p.y][p.x]
		puzzle[p.y][p.x] = emptyCell
		stats.apply(p.x, p.y, saved, -1)
		if countSolutionsWithStats(puzzle, size, 2, stats) != 1 {
			puzzle[p.y][p.x] = saved
			stats.apply(p.x, p.y, saved, 1)
		} else {
			provided[p.y][p.x] = false
			filled--
			rowProvided[p.y]--
			colProvided[p.x]--
		}
	}

	return puzzle, provided
}

func pickRemovalCandidate(candidates []cellPos, puzzle grid, rowProvided, colProvided []int, rng *rand.Rand) int {
	bestIdx := 0
	bestScore := clueRemovalScore(candidates[0], puzzle, rowProvided, colProvided)

	for i := 1; i < len(candidates); i++ {
		score := clueRemovalScore(candidates[i], puzzle, rowProvided, colProvided)
		if score > bestScore || (score == bestScore && rng.IntN(2) == 0) {
			bestIdx = i
			bestScore = score
		}
	}

	return bestIdx
}

func clueRemovalScore(p cellPos, puzzle grid, rowProvided, colProvided []int) int {
	size := len(puzzle)
	score := rowProvided[p.y] + colProvided[p.x]

	if p.x > 0 && puzzle[p.y][p.x-1] != emptyCell {
		score++
	}
	if p.x < size-1 && puzzle[p.y][p.x+1] != emptyCell {
		score++
	}
	if p.y > 0 && puzzle[p.y-1][p.x] != emptyCell {
		score++
	}
	if p.y < size-1 && puzzle[p.y+1][p.x] != emptyCell {
		score++
	}

	return score
}

// countSolutions counts solutions of the grid up to limit using backtracking.
func countSolutions(g grid, size, limit int) int {
	stats := newLineStats(g, size)
	return countSolutionsWithStats(g, size, limit, stats)
}

func countSolutionsWithStats(g grid, size, limit int, stats *lineStats) int {
	choice := selectMRVCell(g, size, stats, nil)
	if choice.x < 0 {
		if hasUniqueLines(g, size) {
			return 1
		}
		return 0
	}
	if choice.count == 0 {
		return 0
	}

	total := 0
	for i := range choice.count {
		v := choice.vals[i]
		g[choice.y][choice.x] = v
		stats.apply(choice.x, choice.y, v, 1)
		total += countSolutionsWithStats(g, size, limit-total, stats)
		stats.apply(choice.x, choice.y, v, -1)
		g[choice.y][choice.x] = emptyCell
		if total >= limit {
			return total
		}
	}

	return total
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
	stats := newLineStats(g, size)
	return canPlaceWithStats(g, size, x, y, val, stats)
}

func canPlaceWithStats(g grid, size, x, y int, val rune, stats *lineStats) bool {
	if g[y][x] != emptyCell {
		return false
	}

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
	if val == zeroCell {
		if stats.rowZero[y] >= half || stats.colZero[x] >= half {
			return false
		}
	} else if val == oneCell {
		if stats.rowOne[y] >= half || stats.colOne[x] >= half {
			return false
		}
	}

	if stats.rowFilled[y] == size-1 {
		for other := range size {
			if other != y && stats.rowFilled[other] == size && rowEqualWith(g[y], g[other], x, val) {
				return false
			}
		}
	}

	if stats.colFilled[x] == size-1 {
		for other := range size {
			if other != x && stats.colFilled[other] == size && colEqualWith(g, size, x, other, y, val) {
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
