package fillomino

import (
	"fmt"
	"math/rand/v2"
)

type Puzzle struct {
	Width    int
	Height   int
	Givens   grid
	Solution grid
}

type region struct {
	cells []point
	size  int
}

func GeneratePuzzle(width, height, maxRegion int, givenRatio float64) (Puzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GeneratePuzzleSeeded(width, height, maxRegion, givenRatio, rng)
}

func GeneratePuzzleSeeded(width, height, maxRegion int, givenRatio float64, rng *rand.Rand) (Puzzle, error) {
	const maxAttempts = 12
	for attempt := 0; attempt < maxAttempts; attempt++ {
		solution, _, err := generateSolution(width, height, maxRegion, rng)
		if err != nil {
			continue
		}

		givens := buildStartingGivens(solution)
		givens = removeClues(givens, maxRegion, givenRatio, rng)
		if countSolutions(givens, maxRegion, 2) != 1 {
			continue
		}

		return Puzzle{
			Width:    width,
			Height:   height,
			Givens:   givens,
			Solution: solution,
		}, nil
	}

	return Puzzle{}, fmt.Errorf("generate fillomino puzzle %dx%d: exceeded retry budget", width, height)
}

func generateSolution(width, height, maxRegion int, rng *rand.Rand) (grid, []region, error) {
	assigned := newGrid(width, height)
	used := make([][]bool, height)
	for y := range height {
		used[y] = make([]bool, width)
	}

	regions := make([]region, 0, width*height/2)
	if !fillRegions(assigned, used, &regions, maxRegion, rng) {
		return nil, nil, fmt.Errorf("unable to fill solved grid")
	}
	return assigned, regions, nil
}

func fillRegions(solution grid, used [][]bool, regions *[]region, maxRegion int, rng *rand.Rand) bool {
	start, ok := firstUnused(used)
	if !ok {
		return true
	}

	candidates := rectangleCandidates(used, start, maxRegion, rng)
	for _, candidate := range candidates {
		if touchesEqualRegion(solution, candidate.cells, candidate.size) {
			continue
		}

		applyShape(solution, used, candidate.cells, candidate.size, true)
		*regions = append(*regions, candidate)
		if fillRegions(solution, used, regions, maxRegion, rng) {
			return true
		}
		*regions = (*regions)[:len(*regions)-1]
		applyShape(solution, used, candidate.cells, candidate.size, false)
	}

	return false
}

func firstUnused(used [][]bool) (point, bool) {
	for y := range len(used) {
		for x := range len(used[y]) {
			if !used[y][x] {
				return point{x: x, y: y}, true
			}
		}
	}
	return point{}, false
}

func rectangleCandidates(used [][]bool, start point, maxRegion int, rng *rand.Rand) []region {
	height := len(used)
	width := len(used[0])

	var candidates []region
	for h := 1; start.y+h <= height; h++ {
		for w := 1; start.x+w <= width; w++ {
			size := w * h
			if size > maxRegion {
				break
			}
			cells := make([]point, 0, size)
			valid := true
			for y := start.y; y < start.y+h && valid; y++ {
				for x := start.x; x < start.x+w; x++ {
					if used[y][x] {
						valid = false
						break
					}
					cells = append(cells, point{x: x, y: y})
				}
			}
			if !valid {
				continue
			}
			candidates = append(candidates, region{cells: cells, size: size})
		}
	}
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	return candidates
}

func touchesEqualRegion(solution grid, shape []point, size int) bool {
	shapeCells := make(map[point]struct{}, len(shape))
	for _, cell := range shape {
		shapeCells[cell] = struct{}{}
	}

	for _, cell := range shape {
		for _, next := range orthogonalNeighbors(cell, len(solution[0]), len(solution)) {
			if _, ok := shapeCells[next]; ok {
				continue
			}
			if solution[next.y][next.x] == size {
				return true
			}
		}
	}

	return false
}

func applyShape(solution grid, used [][]bool, shape []point, size int, place bool) {
	value := 0
	if place {
		value = size
	}
	for _, cell := range shape {
		used[cell.y][cell.x] = place
		solution[cell.y][cell.x] = value
	}
}

func buildStartingGivens(solution grid) grid {
	return cloneGrid(solution)
}

func removeClues(givens grid, maxRegion int, givenRatio float64, rng *rand.Rand) grid {
	width := len(givens[0])
	height := len(givens)
	target := int(float64(width*height) * givenRatio)
	if target < 1 {
		target = 1
	}

	cells := make([]point, 0, width*height)
	for y := range height {
		for x := range width {
			if givens[y][x] != 0 {
				cells = append(cells, point{x: x, y: y})
			}
		}
	}
	rng.Shuffle(len(cells), func(i, j int) {
		cells[i], cells[j] = cells[j], cells[i]
	})

	remaining := len(cells)
	for _, cell := range cells {
		if remaining <= target {
			break
		}
		value := givens[cell.y][cell.x]
		givens[cell.y][cell.x] = 0
		if countSolutions(givens, maxRegion, 2) != 1 {
			givens[cell.y][cell.x] = value
			continue
		}
		remaining--
	}
	return givens
}
