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

type shapeBias int

const (
	shapeBiasNeutral shapeBias = iota
	shapeBiasCompact
	shapeBiasWinding
)

type generationProfile struct {
	cageWeights       []int
	frontierSamples   int
	shapeBias         shapeBias
	minGivensByCage   []int
	maxSingletonCages int
}

func GeneratePuzzle(width, height, maxCage int, givenRatio float64) (Puzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GeneratePuzzleSeeded(width, height, maxCage, givenRatio, rng)
}

func GeneratePuzzleSeeded(width, height, maxCage int, givenRatio float64, rng *rand.Rand) (Puzzle, error) {
	return generatePuzzleSeededWithProfile(width, height, maxCage, givenRatio, defaultGenerationProfile(maxCage), rng)
}

func generatePuzzleSeededWithProfile(width, height, maxCage int, givenRatio float64, profile generationProfile, rng *rand.Rand) (Puzzle, error) {
	const maxAttempts = 48
	profile = profile.withDefaults(maxCage)
	for range maxAttempts {
		cages := generateCages(width, height, maxCage, profile, rng)
		cages = normalizeSingletonCages(width, height, maxCage, cages, profile, rng)
		geo, err := buildGeometry(width, height, cages)
		if err != nil {
			continue
		}

		solution, ok := generateSolution(geo, rng)
		if !ok {
			continue
		}

		givens := removeClues(geo, solution, givenRatio, profile, rng)
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

func defaultGenerationProfile(maxCage int) generationProfile {
	return generationProfile{
		cageWeights:       defaultCageWeights(maxCage),
		frontierSamples:   1,
		shapeBias:         shapeBiasNeutral,
		minGivensByCage:   make([]int, maxCage+1),
		maxSingletonCages: -1,
	}
}

func defaultCageWeights(maxCage int) []int {
	base := []int{0, 2, 4, 5, 4, 3, 2, 1, 1, 1}
	weights := make([]int, maxCage+1)
	for size := range len(weights) {
		if size == 0 {
			continue
		}
		weights[size] = base[min(size, len(base)-1)]
	}
	return weights
}

func (p generationProfile) withDefaults(maxCage int) generationProfile {
	if len(p.cageWeights) == 0 {
		p.cageWeights = defaultCageWeights(maxCage)
	}
	if p.frontierSamples <= 0 {
		p.frontierSamples = 1
	}
	if len(p.minGivensByCage) == 0 {
		p.minGivensByCage = make([]int, maxCage+1)
	}
	if len(p.minGivensByCage) <= maxCage {
		p.minGivensByCage = growIntSlice(p.minGivensByCage, maxCage+1)
	}
	return p
}

func growIntSlice(values []int, size int) []int {
	grown := make([]int, size)
	copy(grown, values)
	return grown
}

func normalizeSingletonCages(width, height, maxCage int, cages []Cage, profile generationProfile, rng *rand.Rand) []Cage {
	if profile.maxSingletonCages < 0 {
		return cages
	}

	normalized := append([]Cage(nil), cages...)
	for countSingletonCages(normalized) > profile.maxSingletonCages {
		singletons := singletonCageIndices(normalized)
		if len(singletons) == 0 {
			break
		}

		singletonIdx := singletons[rng.IntN(len(singletons))]
		targetIdx, ok := chooseSingletonMergeTarget(width, height, maxCage, normalized, singletonIdx, rng)
		if !ok {
			break
		}

		cell := normalized[singletonIdx].Cells[0]
		normalized[targetIdx].Cells = append(normalized[targetIdx].Cells, cell)
		normalized[targetIdx].Size = len(normalized[targetIdx].Cells)
		normalized = append(normalized[:singletonIdx], normalized[singletonIdx+1:]...)
	}

	return renumberCages(normalized)
}

func countSingletonCages(cages []Cage) int {
	count := 0
	for _, cage := range cages {
		if len(cage.Cells) == 1 {
			count++
		}
	}
	return count
}

func singletonCageIndices(cages []Cage) []int {
	indices := make([]int, 0)
	for idx, cage := range cages {
		if len(cage.Cells) == 1 {
			indices = append(indices, idx)
		}
	}
	return indices
}

func chooseSingletonMergeTarget(width, height, maxCage int, cages []Cage, singletonIdx int, rng *rand.Rand) (int, bool) {
	singleton := cages[singletonIdx]
	if len(singleton.Cells) != 1 {
		return 0, false
	}

	cageGrid := buildCageIndexGrid(width, height, cages)
	cell := singleton.Cells[0]
	neighbors := make([]int, 0, 4)
	seen := make(map[int]struct{}, 4)
	for _, next := range orthogonalNeighbors(point{x: cell.X, y: cell.Y}, width, height) {
		cageIdx := cageGrid[next.y][next.x]
		if cageIdx < 0 || cageIdx == singletonIdx {
			continue
		}
		if _, exists := seen[cageIdx]; exists {
			continue
		}
		seen[cageIdx] = struct{}{}
		neighbors = append(neighbors, cageIdx)
	}
	if len(neighbors) == 0 {
		return 0, false
	}

	allowed := make([]int, 0, len(neighbors))
	for _, idx := range neighbors {
		if len(cages[idx].Cells)+1 <= maxCage {
			allowed = append(allowed, idx)
		}
	}
	if len(allowed) > 0 {
		neighbors = allowed
	}

	minSize := len(cages[neighbors[0]].Cells)
	best := make([]int, 0, len(neighbors))
	for _, idx := range neighbors {
		size := len(cages[idx].Cells)
		if size < minSize {
			minSize = size
			best = best[:0]
		}
		if size == minSize {
			best = append(best, idx)
		}
	}

	return best[rng.IntN(len(best))], true
}

func buildCageIndexGrid(width, height int, cages []Cage) [][]int {
	grid := make([][]int, height)
	for y := range height {
		grid[y] = make([]int, width)
		for x := range width {
			grid[y][x] = -1
		}
	}

	for cageIdx, cage := range cages {
		for _, cell := range cage.Cells {
			grid[cell.Y][cell.X] = cageIdx
		}
	}

	return grid
}

func renumberCages(cages []Cage) []Cage {
	renumbered := make([]Cage, len(cages))
	for idx, cage := range cages {
		renumbered[idx] = Cage{
			ID:    idx,
			Size:  len(cage.Cells),
			Cells: append([]Cell(nil), cage.Cells...),
		}
	}
	return renumbered
}

func generateCages(width, height, maxCage int, profile generationProfile, rng *rand.Rand) []Cage {
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

		target := chooseCageSize(maxCage, remainingUnassigned(assigned), profile.cageWeights, rng)
		shape := growCageShape(start, target, assigned, profile, rng)
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

func chooseCageSize(maxCage, remaining int, weights []int, rng *rand.Rand) int {
	if remaining < maxCage {
		maxCage = remaining
	}
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

func growCageShape(start point, target int, assigned [][]bool, profile generationProfile, rng *rand.Rand) []point {
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
		next := chooseFrontierPoint(frontier, inShape, assigned, profile, rng)
		shape = append(shape, next)
		inShape[next] = struct{}{}
	}

	return shape
}

func chooseFrontierPoint(frontier []point, inShape map[point]struct{}, assigned [][]bool, profile generationProfile, rng *rand.Rand) point {
	if len(frontier) == 1 || profile.shapeBias == shapeBiasNeutral {
		return frontier[rng.IntN(len(frontier))]
	}

	sampleCount := min(profile.frontierSamples, len(frontier))
	best := frontier[rng.IntN(len(frontier))]
	bestScore := frontierScore(best, inShape, assigned)

	for pick := 1; pick < sampleCount; pick++ {
		candidate := frontier[rng.IntN(len(frontier))]
		score := frontierScore(candidate, inShape, assigned)
		if prefersFrontierScore(profile.shapeBias, score, bestScore) {
			best = candidate
			bestScore = score
		}
	}

	return best
}

func frontierScore(candidate point, inShape map[point]struct{}, assigned [][]bool) int {
	score := 0
	for _, neighbor := range orthogonalNeighbors(candidate, len(assigned[0]), len(assigned)) {
		if assigned[neighbor.y][neighbor.x] {
			continue
		}
		if _, ok := inShape[neighbor]; ok {
			score++
		}
	}
	return score
}

func prefersFrontierScore(bias shapeBias, candidate, current int) bool {
	switch bias {
	case shapeBiasCompact:
		return candidate > current
	case shapeBiasWinding:
		return candidate < current
	default:
		return false
	}
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

func removeClues(geo *geometry, solution grid, givenRatio float64, profile generationProfile, rng *rand.Rand) grid {
	givens := cloneGrid(solution)
	totalCells := geo.width * geo.height
	target := int(float64(totalCells) * givenRatio)

	remainingByCage := make([]int, len(geo.cages))
	minimumByCage := make([]int, len(geo.cages))
	minimumTotal := 0
	for cageIdx, cells := range geo.cageCells {
		remainingByCage[cageIdx] = len(cells)
		minimumByCage[cageIdx] = minGivensForCage(profile, len(cells))
		minimumTotal += minimumByCage[cageIdx]
	}

	if target < minimumTotal {
		target = minimumTotal
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
		if remainingByCage[cageIdx] <= minimumByCage[cageIdx] {
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

func minGivensForCage(profile generationProfile, cageSize int) int {
	if cageSize <= 0 {
		return 0
	}
	if cageSize >= len(profile.minGivensByCage) {
		return max(0, min(profile.minGivensByCage[len(profile.minGivensByCage)-1], cageSize))
	}
	return max(0, min(profile.minGivensByCage[cageSize], cageSize))
}
