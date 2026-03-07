package sudokurgb

import "math/rand/v2"

type quotaState struct {
	row [gridSize][valueCount + 1]uint8
	col [gridSize][valueCount + 1]uint8
	box [gridSize][valueCount + 1]uint8
}

type pos struct{ x, y int }

func newRNG() *rand.Rand {
	return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
}

func boxIndex(x, y int) int {
	return (y/3)*3 + (x / 3)
}

func buildQuotaState(g *grid) (quotaState, bool) {
	var state quotaState
	for i := range gridSize {
		for value := 1; value <= valueCount; value++ {
			state.row[i][value] = houseQuota
			state.col[i][value] = houseQuota
			state.box[i][value] = houseQuota
		}
	}

	for y := range gridSize {
		for x := range gridSize {
			value := g[y][x].v
			if value == 0 {
				continue
			}
			if value < 1 || value > valueCount {
				return quotaState{}, false
			}

			box := boxIndex(x, y)
			if state.row[y][value] == 0 || state.col[x][value] == 0 || state.box[box][value] == 0 {
				return quotaState{}, false
			}

			state.row[y][value]--
			state.col[x][value]--
			state.box[box][value]--
		}
	}

	return state, true
}

func canPlace(state *quotaState, value, x, y int) bool {
	box := boxIndex(x, y)
	return state.row[y][value] > 0 && state.col[x][value] > 0 && state.box[box][value] > 0
}

func placeValue(g *grid, state *quotaState, value, x, y int) {
	box := boxIndex(x, y)
	g[y][x].v = value
	state.row[y][value]--
	state.col[x][value]--
	state.box[box][value]--
}

func clearValue(g *grid, state *quotaState, value, x, y int) {
	box := boxIndex(x, y)
	g[y][x].v = 0
	state.row[y][value]++
	state.col[x][value]++
	state.box[box][value]++
}

func candidateMaskForCell(state *quotaState, x, y int) uint8 {
	var mask uint8
	for value := 1; value <= valueCount; value++ {
		if canPlace(state, value, x, y) {
			mask |= 1 << value
		}
	}
	return mask
}

func countCandidateBits(mask uint8) int {
	count := 0
	for value := 1; value <= valueCount; value++ {
		if mask&(1<<value) != 0 {
			count++
		}
	}
	return count
}

func findMostConstrainedEmptyCell(g *grid, state *quotaState) (x, y int, candidates uint8, found bool) {
	bestCount := valueCount + 1
	bestX, bestY := 0, 0
	for y := range gridSize {
		for x := range gridSize {
			if g[y][x].v != 0 {
				continue
			}

			mask := candidateMaskForCell(state, x, y)
			candidateCount := countCandidateBits(mask)
			if candidateCount == 0 {
				return x, y, 0, true
			}
			if !found || candidateCount < bestCount {
				found = true
				bestCount = candidateCount
				bestX, bestY = x, y
				candidates = mask
				if candidateCount == 1 {
					return bestX, bestY, candidates, true
				}
			}
		}
	}
	return bestX, bestY, candidates, found
}

func randomCandidateOrder(mask uint8, rng *rand.Rand) []int {
	values := make([]int, 0, valueCount)
	for value := 1; value <= valueCount; value++ {
		if mask&(1<<value) != 0 {
			values = append(values, value)
		}
	}
	rng.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
	return values
}

func fillGridRec(g *grid, rng *rand.Rand, state *quotaState) bool {
	x, y, candidates, found := findMostConstrainedEmptyCell(g, state)
	if !found {
		return true
	}
	if candidates == 0 {
		return false
	}

	for _, value := range randomCandidateOrder(candidates, rng) {
		placeValue(g, state, value, x, y)
		if fillGridRec(g, rng, state) {
			return true
		}
		clearValue(g, state, value, x, y)
	}

	return false
}

func fillGrid(g *grid) bool {
	return fillGridSeeded(g, newRNG())
}

func fillGridSeeded(g *grid, rng *rand.Rand) bool {
	state, ok := buildQuotaState(g)
	if !ok {
		return false
	}
	return fillGridRec(g, rng, &state)
}

func countSolutions(g *grid, limit int) int {
	if limit <= 0 {
		return 0
	}

	state, ok := buildQuotaState(g)
	if !ok {
		return 0
	}
	return countSolutionsRec(g, limit, &state)
}

func countSolutionsRec(g *grid, limit int, state *quotaState) int {
	if limit <= 0 {
		return 0
	}

	x, y, candidates, found := findMostConstrainedEmptyCell(g, state)
	if !found {
		return 1
	}
	if candidates == 0 {
		return 0
	}

	count := 0
	for value := 1; value <= valueCount; value++ {
		if candidates&(1<<value) == 0 {
			continue
		}
		placeValue(g, state, value, x, y)
		count += countSolutionsRec(g, limit-count, state)
		clearValue(g, state, value, x, y)
		if count >= limit {
			return count
		}
	}
	return count
}

func GenerateProvidedCells(mode SudokuRGBMode) []cell {
	return GenerateProvidedCellsSeeded(mode, newRNG())
}

func clueRemovalScore(g *grid, x, y int) int {
	score := 0
	for i := range gridSize {
		if i != x && g[y][i].v == 0 {
			score++
		}
		if i != y && g[i][x].v == 0 {
			score++
		}
	}

	boxX, boxY := (x/3)*3, (y/3)*3
	for dy := range 3 {
		for dx := range 3 {
			cx, cy := boxX+dx, boxY+dy
			if (cx != x || cy != y) && g[cy][cx].v == 0 {
				score++
			}
		}
	}
	return score
}

func pickRemovalCandidateIndex(g *grid, positions []pos, rng *rand.Rand) int {
	bestIndex := 0
	bestScore := clueRemovalScore(g, positions[0].x, positions[0].y)
	for i := 1; i < len(positions); i++ {
		score := clueRemovalScore(g, positions[i].x, positions[i].y)
		if score < bestScore || (score == bestScore && rng.IntN(2) == 0) {
			bestScore = score
			bestIndex = i
		}
	}
	return bestIndex
}

func collectFilledCells(g *grid) []cell {
	cells := make([]cell, 0, gridSize*gridSize)
	for y := range gridSize {
		for x := range gridSize {
			if g[y][x].v != 0 {
				cells = append(cells, g[y][x])
			}
		}
	}
	return cells
}

func GenerateProvidedCellsSeeded(mode SudokuRGBMode, rng *rand.Rand) []cell {
	bestCells := []cell{}
	bestCount := gridSize*gridSize + 1

	for attempt := 0; attempt < 96; attempt++ {
		g := newGrid(nil)
		if !fillGridSeeded(&g, rng) {
			continue
		}

		positions := make([]pos, 0, gridSize*gridSize)
		for y := range gridSize {
			for x := range gridSize {
				positions = append(positions, pos{x: x, y: y})
			}
		}
		rng.Shuffle(len(positions), func(i, j int) {
			positions[i], positions[j] = positions[j], positions[i]
		})

		filled := gridSize * gridSize
		for len(positions) > 0 && filled > mode.ProvidedCount {
			idx := pickRemovalCandidateIndex(&g, positions, rng)
			p := positions[idx]
			positions[idx] = positions[len(positions)-1]
			positions = positions[:len(positions)-1]

			saved := g[p.y][p.x].v
			g[p.y][p.x].v = 0
			if countSolutions(&g, 2) != 1 {
				g[p.y][p.x].v = saved
				continue
			}
			filled--
		}

		cells := collectFilledCells(&g)
		if len(cells) == mode.ProvidedCount {
			return cells
		}
		if len(cells) < bestCount {
			bestCount = len(cells)
			bestCells = cells
		}
	}

	return bestCells
}

func isValid(g *grid, value, x, y int) bool {
	if value < 1 || value > valueCount {
		return false
	}

	rowCount, colCount, boxCount := 0, 0, 0
	boxX, boxY := (x/3)*3, (y/3)*3
	for i := range gridSize {
		if i != x && g[y][i].v == value {
			rowCount++
		}
		if i != y && g[i][x].v == value {
			colCount++
		}
	}
	for dy := range 3 {
		for dx := range 3 {
			cx, cy := boxX+dx, boxY+dy
			if cx == x && cy == y {
				continue
			}
			if g[cy][cx].v == value {
				boxCount++
			}
		}
	}

	return rowCount < houseQuota && colCount < houseQuota && boxCount < houseQuota
}
