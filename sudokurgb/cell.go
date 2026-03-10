package sudokurgb

import "fmt"

type cell struct {
	x, y int
	v    int
}

type grid = [gridSize][gridSize]cell

func newGrid(provided []cell) grid {
	var g grid
	for y := range gridSize {
		for x := range gridSize {
			g[y][x] = cell{x: x, y: y}
		}
	}
	for _, hint := range provided {
		g[hint.y][hint.x].v = hint.v
	}
	return g
}

func (c cell) String() string {
	return fmt.Sprintf("r%dc%d=%d", c.y, c.x, c.v)
}

func (m Model) isSolved() bool {
	return isSolvedWith(m.grid, m.analysis)
}

func isSolvedWith(g grid, analysis boardAnalysis) bool {
	for y := range gridSize {
		for x := range gridSize {
			if g[y][x].v == 0 {
				return false
			}
		}
	}

	if analysis.hasConflicts() {
		return false
	}

	for i := range gridSize {
		if !houseHasExactQuota(rowValues(g, i)) {
			return false
		}
		if !houseHasExactQuota(colValues(g, i)) {
			return false
		}
	}

	for box := range gridSize {
		if !houseHasExactQuota(boxValues(g, box)) {
			return false
		}
	}

	return true
}

func rowValues(g grid, y int) [gridSize]int {
	var values [gridSize]int
	for x := range gridSize {
		values[x] = g[y][x].v
	}
	return values
}

func colValues(g grid, x int) [gridSize]int {
	var values [gridSize]int
	for y := range gridSize {
		values[y] = g[y][x].v
	}
	return values
}

func boxValues(g grid, box int) [gridSize]int {
	var values [gridSize]int
	boxX, boxY := (box%3)*3, (box/3)*3
	index := 0
	for dy := range 3 {
		for dx := range 3 {
			values[index] = g[boxY+dy][boxX+dx].v
			index++
		}
	}
	return values
}

func houseHasExactQuota(values [gridSize]int) bool {
	var counts [valueCount + 1]int
	for _, value := range values {
		if value < 1 || value > valueCount {
			return false
		}
		counts[value]++
	}
	for value := 1; value <= valueCount; value++ {
		if counts[value] != houseQuota {
			return false
		}
	}
	return true
}
