package fillomino

import (
	"errors"
	"math"
)

var errInvalidGridEncoding = errors.New("invalid fillomino grid encoding")

type component struct {
	value    int
	cells    []point
	frontier map[point]struct{}
}

type validationResult struct {
	solved    bool
	conflicts [][]bool
}

func validateGridState(g grid) validationResult {
	height := len(g)
	width := 0
	if height > 0 {
		width = len(g[0])
	}

	conflicts := make([][]bool, height)
	for y := range height {
		conflicts[y] = make([]bool, width)
	}

	visited := make([][]bool, height)
	for y := range height {
		visited[y] = make([]bool, width)
	}

	complete := true
	for y := range height {
		for x := range width {
			if g[y][x] == 0 {
				complete = false
				continue
			}
			if visited[y][x] {
				continue
			}

			comp := buildComponent(g, point{x: x, y: y}, visited)
			markComponentConflicts(g, comp, complete, conflicts)
		}
	}

	solved := complete && !hasConflicts(conflicts)
	return validationResult{
		solved:    solved,
		conflicts: conflicts,
	}
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

func buildComponent(g grid, start point, visited [][]bool) component {
	value := g[start.y][start.x]
	queue := []point{start}
	visited[start.y][start.x] = true
	comp := component{
		value:    value,
		frontier: make(map[point]struct{}),
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		comp.cells = append(comp.cells, curr)

		for _, next := range orthogonalNeighbors(curr, len(g[0]), len(g)) {
			switch {
			case g[next.y][next.x] == value && !visited[next.y][next.x]:
				visited[next.y][next.x] = true
				queue = append(queue, next)
			case g[next.y][next.x] == 0:
				comp.frontier[next] = struct{}{}
			}
		}
	}

	return comp
}

func markComponentConflicts(g grid, comp component, assumeComplete bool, conflicts [][]bool) {
	size := len(comp.cells)
	value := comp.value
	if value <= 0 {
		for _, cell := range comp.cells {
			conflicts[cell.y][cell.x] = true
		}
		return
	}

	if size > value || size+len(comp.frontier) < value {
		for _, cell := range comp.cells {
			conflicts[cell.y][cell.x] = true
		}
		for cell := range comp.frontier {
			conflicts[cell.y][cell.x] = true
		}
		return
	}

	if assumeComplete && size != value {
		for _, cell := range comp.cells {
			conflicts[cell.y][cell.x] = true
		}
	}
}

func orthogonalNeighbors(p point, width, height int) []point {
	neighbors := make([]point, 0, 4)
	if p.x > 0 {
		neighbors = append(neighbors, point{x: p.x - 1, y: p.y})
	}
	if p.x+1 < width {
		neighbors = append(neighbors, point{x: p.x + 1, y: p.y})
	}
	if p.y > 0 {
		neighbors = append(neighbors, point{x: p.x, y: p.y - 1})
	}
	if p.y+1 < height {
		neighbors = append(neighbors, point{x: p.x, y: p.y + 1})
	}
	return neighbors
}

func countSolutions(givens grid, maxValue, limit int) int {
	working := cloneGrid(givens)
	return searchSolutions(working, maxValue, limit)
}

func searchSolutions(g grid, maxValue, limit int) int {
	if limit <= 0 {
		return 0
	}
	if !isPartialGridValid(g) {
		return 0
	}

	cell, candidates, ok := chooseNextCell(g, maxValue)
	if !ok {
		if validateGridState(g).solved {
			return 1
		}
		return 0
	}

	total := 0
	for _, value := range candidates {
		g[cell.y][cell.x] = value
		total += searchSolutions(g, maxValue, limit-total)
		if total >= limit {
			g[cell.y][cell.x] = 0
			return total
		}
		g[cell.y][cell.x] = 0
	}
	return total
}

func chooseNextCell(g grid, maxValue int) (point, []int, bool) {
	best := point{}
	bestCandidates := []int(nil)
	bestCount := math.MaxInt

	for y := range len(g) {
		for x := range len(g[y]) {
			if g[y][x] != 0 {
				continue
			}

			candidates := candidatesForCell(g, point{x: x, y: y}, maxValue)
			if len(candidates) == 0 {
				return point{x: x, y: y}, nil, true
			}
			if len(candidates) < bestCount {
				best = point{x: x, y: y}
				bestCandidates = candidates
				bestCount = len(candidates)
				if bestCount == 1 {
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

func candidatesForCell(g grid, cell point, maxValue int) []int {
	candidates := make([]int, 0, maxValue)
	for value := 1; value <= maxValue; value++ {
		g[cell.y][cell.x] = value
		if isPartialGridValid(g) {
			candidates = append(candidates, value)
		}
		g[cell.y][cell.x] = 0
	}
	return candidates
}

func isPartialGridValid(g grid) bool {
	height := len(g)
	if height == 0 {
		return true
	}
	width := len(g[0])

	visited := make([][]bool, height)
	for y := range height {
		visited[y] = make([]bool, width)
	}

	for y := range height {
		for x := range width {
			if g[y][x] == 0 || visited[y][x] {
				continue
			}

			comp := buildComponent(g, point{x: x, y: y}, visited)
			size := len(comp.cells)
			if size > comp.value {
				return false
			}
			if size+len(comp.frontier) < comp.value {
				return false
			}
			if len(comp.frontier) == 0 && size != comp.value {
				return false
			}
		}
	}

	return true
}
