package rippleeffect

import (
	"fmt"
	"math/rand/v2"
)

type Puzzle struct {
	Width    int
	Height   int
	Cages    []Cage
	Givens   grid
	Solution grid
}

func GeneratePuzzle(width, height, maxCage int, givenRatio float64) (Puzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GeneratePuzzleSeeded(width, height, maxCage, givenRatio, rng)
}

func GeneratePuzzleSeeded(width, height, maxCage int, givenRatio float64, rng *rand.Rand) (Puzzle, error) {
	const maxAttempts = 24
	for attempt := 0; attempt < maxAttempts; attempt++ {
		cages := generateCages(width, height, maxCage, rng)
		geo, err := buildGeometry(width, height, cages)
		if err != nil {
			continue
		}

		solution, ok := generateSolution(geo, rng)
		if !ok {
			continue
		}

		givens := removeClues(geo, solution, givenRatio, rng)
		if countSolutions(geo, givens, 2) != 1 {
			continue
		}

		return Puzzle{
			Width:    width,
			Height:   height,
			Cages:    cages,
			Givens:   givens,
			Solution: solution,
		}, nil
	}

	return Puzzle{}, fmt.Errorf("generate ripple effect puzzle %dx%d: exceeded retry budget", width, height)
}

func generateCages(width, height, maxCage int, rng *rand.Rand) []Cage {
	assigned := make([][]bool, height)
	for y := range height {
		assigned[y] = make([]bool, width)
	}

	cages := make([]Cage, 0, width*height/2)
	for {
		start, ok := firstUnassigned(assigned)
		if !ok {
			return cages
		}

		target := chooseCageSize(maxCage, remainingUnassigned(assigned), rng)
		shape := growCageShape(start, target, assigned, rng)
		for _, cell := range shape {
			assigned[cell.y][cell.x] = true
		}

		cells := make([]Cell, len(shape))
		for i, cell := range shape {
			cells[i] = Cell{X: cell.x, Y: cell.y}
		}
		cages = append(cages, Cage{
			ID:    len(cages),
			Size:  len(cells),
			Cells: cells,
		})
	}
}

func firstUnassigned(assigned [][]bool) (point, bool) {
	for y := range len(assigned) {
		for x := range len(assigned[y]) {
			if !assigned[y][x] {
				return point{x: x, y: y}, true
			}
		}
	}
	return point{}, false
}

func remainingUnassigned(assigned [][]bool) int {
	count := 0
	for y := range len(assigned) {
		for x := range len(assigned[y]) {
			if !assigned[y][x] {
				count++
			}
		}
	}
	return count
}

func chooseCageSize(maxCage, remaining int, rng *rand.Rand) int {
	if remaining < maxCage {
		maxCage = remaining
	}
	weights := []int{0, 2, 4, 5, 4, 3, 2, 1, 1, 1}
	totalWeight := 0
	for size := 1; size <= maxCage; size++ {
		totalWeight += weights[min(size, len(weights)-1)]
	}
	if totalWeight == 0 {
		return 1
	}

	pick := rng.IntN(totalWeight)
	for size := 1; size <= maxCage; size++ {
		pick -= weights[min(size, len(weights)-1)]
		if pick < 0 {
			return size
		}
	}
	return 1
}

func growCageShape(start point, target int, assigned [][]bool, rng *rand.Rand) []point {
	shape := []point{start}
	inShape := map[point]struct{}{start: {}}
	height := len(assigned)
	width := len(assigned[0])

	for len(shape) < target {
		frontier := make([]point, 0, 4)
		seen := make(map[point]struct{})
		for _, cell := range shape {
			for _, next := range orthogonalNeighbors(cell, width, height) {
				if assigned[next.y][next.x] {
					continue
				}
				if _, ok := inShape[next]; ok {
					continue
				}
				if _, ok := seen[next]; ok {
					continue
				}
				frontier = append(frontier, next)
				seen[next] = struct{}{}
			}
		}
		if len(frontier) == 0 {
			break
		}
		next := frontier[rng.IntN(len(frontier))]
		shape = append(shape, next)
		inShape[next] = struct{}{}
	}

	return shape
}

func generateSolution(geo *geometry, rng *rand.Rand) (grid, bool) {
	state := newGrid(geo.width, geo.height)
	if fillSolution(geo, state, rng) {
		return state, true
	}
	return nil, false
}

func fillSolution(geo *geometry, state grid, rng *rand.Rand) bool {
	cell, candidates, ok := chooseNextFillCell(geo, state, rng)
	if !ok {
		return true
	}
	if len(candidates) == 0 {
		return false
	}

	for _, value := range candidates {
		state[cell.y][cell.x] = value
		if fillSolution(geo, state, rng) {
			return true
		}
		state[cell.y][cell.x] = 0
	}

	return false
}

func chooseNextFillCell(geo *geometry, state grid, rng *rand.Rand) (point, []int, bool) {
	best := point{}
	bestCandidates := []int(nil)
	bestCount := 10
	ties := 0

	for y := range geo.height {
		for x := range geo.width {
			if state[y][x] != 0 {
				continue
			}
			cell := point{x: x, y: y}
			candidates := candidatesForCell(geo, state, cell)
			if len(candidates) < bestCount {
				best = cell
				bestCandidates = candidates
				bestCount = len(candidates)
				ties = 1
				continue
			}
			if len(candidates) == bestCount {
				ties++
				if rng.IntN(ties) == 0 {
					best = cell
					bestCandidates = candidates
				}
			}
		}
	}

	if bestCount == 10 {
		return point{}, nil, false
	}

	rng.Shuffle(len(bestCandidates), func(i, j int) {
		bestCandidates[i], bestCandidates[j] = bestCandidates[j], bestCandidates[i]
	})
	return best, bestCandidates, true
}

func removeClues(geo *geometry, solution grid, givenRatio float64, rng *rand.Rand) grid {
	givens := cloneGrid(solution)
	totalCells := geo.width * geo.height
	target := int(float64(totalCells) * givenRatio)
	if target < len(geo.cages) {
		target = len(geo.cages)
	}

	remainingByCage := make([]int, len(geo.cages))
	for cageIdx, cells := range geo.cageCells {
		remainingByCage[cageIdx] = len(cells)
	}

	cells := make([]point, 0, totalCells)
	for y := range geo.height {
		for x := range geo.width {
			cells = append(cells, point{x: x, y: y})
		}
	}
	rng.Shuffle(len(cells), func(i, j int) {
		cells[i], cells[j] = cells[j], cells[i]
	})

	remaining := totalCells
	for _, cell := range cells {
		if remaining <= target {
			break
		}

		cageIdx := geo.cageGrid[cell.y][cell.x]
		if remainingByCage[cageIdx] <= 1 {
			continue
		}

		value := givens[cell.y][cell.x]
		givens[cell.y][cell.x] = 0
		remainingByCage[cageIdx]--
		if countSolutions(geo, givens, 2) != 1 {
			givens[cell.y][cell.x] = value
			remainingByCage[cageIdx]++
			continue
		}
		remaining--
	}

	return givens
}
