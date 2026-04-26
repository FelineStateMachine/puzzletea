package rippleeffect

import (
	"context"
	"math"
)

type validationResult struct {
	solved    bool
	conflicts [][]bool
}

func validateGridState(state grid, geo *geometry) validationResult {
	conflicts := newConflictGrid(geo.width, geo.height)
	complete := true

	for cageIdx, cells := range geo.cageCells {
		size := geo.cageSizes[cageIdx]
		seen := make(map[int]point, size)
		for _, cell := range cells {
			value := state[cell.y][cell.x]
			if value == 0 {
				complete = false
				continue
			}
			if value < 1 || value > size {
				conflicts[cell.y][cell.x] = true
				continue
			}
			if prior, exists := seen[value]; exists {
				conflicts[cell.y][cell.x] = true
				conflicts[prior.y][prior.x] = true
				continue
			}
			seen[value] = cell
		}
	}

	for y := range geo.height {
		for x := range geo.width {
			value := state[y][x]
			if value == 0 {
				continue
			}
			for step := 1; step < value; step++ {
				if x+step < geo.width && state[y][x+step] == value {
					conflicts[y][x] = true
					conflicts[y][x+step] = true
				}
				if y+step < geo.height && state[y+step][x] == value {
					conflicts[y][x] = true
					conflicts[y+step][x] = true
				}
			}
		}
	}

	return validationResult{
		solved:    complete && !hasConflicts(conflicts),
		conflicts: conflicts,
	}
}

func newConflictGrid(width, height int) [][]bool {
	conflicts := make([][]bool, height)
	for y := range height {
		conflicts[y] = make([]bool, width)
	}
	return conflicts
}

func hasConflicts(conflicts [][]bool) bool {
	for y := range len(conflicts) {
		for x := range len(conflicts[y]) {
			if conflicts[y][x] {
				return true
			}
		}
	}
	return false
}

func countSolutions(geo *geometry, givens grid, limit int) int {
	return countSolutionsContext(context.Background(), geo, givens, limit)
}

func countSolutionsContext(ctx context.Context, geo *geometry, givens grid, limit int) int {
	working := cloneGrid(givens)
	return searchSolutionsContext(ctx, geo, working, limit)
}

func searchSolutionsContext(ctx context.Context, geo *geometry, state grid, limit int) int {
	if limit <= 0 {
		return 0
	}
	if ctx.Err() != nil {
		return -1
	}

	cell, candidates, ok := chooseNextCell(geo, state)
	if !ok {
		if validateGridState(state, geo).solved {
			return 1
		}
		return 0
	}
	if len(candidates) == 0 {
		return 0
	}

	total := 0
	for _, value := range candidates {
		state[cell.y][cell.x] = value
		n := searchSolutionsContext(ctx, geo, state, limit-total)
		if n < 0 {
			state[cell.y][cell.x] = 0
			return n
		}
		total += n
		if total >= limit {
			state[cell.y][cell.x] = 0
			return total
		}
		state[cell.y][cell.x] = 0
	}

	return total
}

func chooseNextCell(geo *geometry, state grid) (point, []int, bool) {
	best := point{}
	bestCandidates := []int(nil)
	bestCount := math.MaxInt

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
				if bestCount <= 1 {
					return best, bestCandidates, true
				}
			}
		}
	}

	if bestCount == math.MaxInt {
		return point{}, nil, false
	}
	return best, bestCandidates, true
}

func candidatesForCell(geo *geometry, state grid, cell point) []int {
	cageIdx := geo.cageGrid[cell.y][cell.x]
	size := geo.cageSizes[cageIdx]
	candidates := make([]int, 0, size)
	for value := 1; value <= size; value++ {
		if valueAllowed(geo, state, cell, value) {
			candidates = append(candidates, value)
		}
	}
	return candidates
}

func valueAllowed(geo *geometry, state grid, cell point, value int) bool {
	if value <= 0 {
		return false
	}

	cageIdx := geo.cageGrid[cell.y][cell.x]
	if value > geo.cageSizes[cageIdx] {
		return false
	}

	for _, other := range geo.cageCells[cageIdx] {
		if other == cell {
			continue
		}
		if state[other.y][other.x] == value {
			return false
		}
	}

	for step := 1; step < value; step++ {
		if cell.x-step >= 0 && state[cell.y][cell.x-step] == value {
			return false
		}
		if cell.x+step < geo.width && state[cell.y][cell.x+step] == value {
			return false
		}
		if cell.y-step >= 0 && state[cell.y-step][cell.x] == value {
			return false
		}
		if cell.y+step < geo.height && state[cell.y+step][cell.x] == value {
			return false
		}
	}

	return true
}
