package sudoku

import "math/rand/v2"

const allDigitsMask uint16 = 0x3FE

type solverMasks struct {
	row [gridSize]uint16
	col [gridSize]uint16
	box [gridSize]uint16
}

// newRNG creates a fresh non-deterministic RNG.
func newRNG() *rand.Rand {
	return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
}

func valueBit(val int) uint16 {
	return uint16(1 << val)
}

func boxIndex(x, y int) int {
	return (y/3)*3 + (x / 3)
}

func buildMasks(g *grid) (solverMasks, bool) {
	var masks solverMasks
	for y := range gridSize {
		for x := range gridSize {
			val := g[y][x].v
			if val == 0 {
				continue
			}
			if val < 0 || val > gridSize {
				return solverMasks{}, false
			}

			bit := valueBit(val)
			box := boxIndex(x, y)
			if masks.row[y]&bit != 0 || masks.col[x]&bit != 0 || masks.box[box]&bit != 0 {
				return solverMasks{}, false
			}

			masks.row[y] |= bit
			masks.col[x] |= bit
			masks.box[box] |= bit
		}
	}
	return masks, true
}

func canPlace(masks *solverMasks, val, x, y int) bool {
	bit := valueBit(val)
	box := boxIndex(x, y)
	return masks.row[y]&bit == 0 && masks.col[x]&bit == 0 && masks.box[box]&bit == 0
}

func placeValue(g *grid, masks *solverMasks, val, x, y int) {
	bit := valueBit(val)
	box := boxIndex(x, y)
	g[y][x].v = val
	masks.row[y] |= bit
	masks.col[x] |= bit
	masks.box[box] |= bit
}

func clearValue(g *grid, masks *solverMasks, val, x, y int) {
	bit := valueBit(val)
	box := boxIndex(x, y)
	g[y][x].v = 0
	masks.row[y] &^= bit
	masks.col[x] &^= bit
	masks.box[box] &^= bit
}

func candidatesForCell(masks *solverMasks, x, y int) uint16 {
	used := masks.row[y] | masks.col[x] | masks.box[boxIndex(x, y)]
	return allDigitsMask &^ used
}

func findFirstEmptyCell(g *grid) (x, y int, found bool) {
	for y := range gridSize {
		for x := range gridSize {
			if g[y][x].v == 0 {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}

func fillGridRec(g *grid, rng *rand.Rand, masks *solverMasks) bool {
	x, y, found := findFirstEmptyCell(g)
	if !found {
		return true
	}

	order := rng.Perm(gridSize)
	for _, i := range order {
		val := i + 1
		if !canPlace(masks, val, x, y) {
			continue
		}

		placeValue(g, masks, val, x, y)
		if fillGridRec(g, rng, masks) {
			return true
		}
		clearValue(g, masks, val, x, y)
	}

	return false
}

func countCandidateBits(mask uint16) int {
	count := 0
	for val := 1; val <= gridSize; val++ {
		if mask&valueBit(val) != 0 {
			count++
		}
	}
	return count
}

func findMostConstrainedEmptyCell(g *grid, masks *solverMasks) (x, y int, candidates uint16, found bool) {
	bestCount := gridSize + 1
	bestX, bestY := 0, 0
	for y := range gridSize {
		for x := range gridSize {
			if g[y][x].v != 0 {
				continue
			}

			mask := candidatesForCell(masks, x, y)
			candidateCount := countCandidateBits(mask)
			if candidateCount == 0 {
				return x, y, 0, true
			}

			if !found || candidateCount < bestCount {
				bestCount = candidateCount
				candidates = mask
				bestX, bestY = x, y
				found = true
			}
			if bestCount == 1 {
				return bestX, bestY, candidates, true
			}
		}
	}

	return bestX, bestY, candidates, found
}

// isValid checks whether placing val at (x, y) in the grid is valid.
func isValid(g *grid, val, x, y int) bool {
	for i := range gridSize {
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
	rng := newRNG()
	return fillGridSeeded(g, rng)
}

func fillGridSeeded(g *grid, rng *rand.Rand) bool {
	masks, ok := buildMasks(g)
	if !ok {
		return false
	}
	return fillGridRec(g, rng, &masks)
}

// countSolutions counts solutions of the grid up to limit using backtracking.
// Returns 0 immediately if the grid already contains conflicting values.
func countSolutions(g *grid, limit int) int {
	if limit <= 0 {
		return 0
	}

	masks, ok := buildMasks(g)
	if !ok {
		return 0
	}
	return countSolutionsRec(g, limit, &masks)
}

// countSolutionsRec is the recursive backtracking helper for countSolutions.
func countSolutionsRec(g *grid, limit int, masks *solverMasks) int {
	if limit <= 0 {
		return 0
	}

	x, y, candidateMask, found := findMostConstrainedEmptyCell(g, masks)
	if !found {
		return 1
	}
	if candidateMask == 0 {
		return 0
	}

	count := 0
	for val := 1; val <= gridSize; val++ {
		bit := valueBit(val)
		if candidateMask&bit == 0 {
			continue
		}

		placeValue(g, masks, val, x, y)
		count += countSolutionsRec(g, limit-count, masks)
		clearValue(g, masks, val, x, y)
		if count >= limit {
			return count
		}
	}
	return count
}

// GenerateProvidedCells generates a random sudoku puzzle with the number of
// provided cells determined by the mode's ProvidedCount.
func GenerateProvidedCells(m SudokuMode) []cell {
	return GenerateProvidedCellsSeeded(m, newRNG())
}

type pos struct{ x, y int }

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
	for by := boxY; by < boxY+3; by++ {
		for bx := boxX; bx < boxX+3; bx++ {
			if (bx != x || by != y) && g[by][bx].v == 0 {
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

func GenerateProvidedCellsSeeded(m SudokuMode, rng *rand.Rand) []cell {
	g := newGrid(nil)
	fillGridSeeded(&g, rng)

	// Build shuffled list of all positions
	positions := make([]pos, 0, gridSize*gridSize)
	for y := range gridSize {
		for x := range gridSize {
			positions = append(positions, pos{x, y})
		}
	}
	rng.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Iteratively remove cells in low-impact-first order while preserving uniqueness.
	filled := gridSize * gridSize
	for len(positions) > 0 && filled > m.ProvidedCount {
		idx := pickRemovalCandidateIndex(&g, positions, rng)
		p := positions[idx]
		positions[idx] = positions[len(positions)-1]
		positions = positions[:len(positions)-1]

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
	for y := range gridSize {
		for x := range gridSize {
			if g[y][x].v != 0 {
				cells = append(cells, g[y][x])
			}
		}
	}
	return cells
}
