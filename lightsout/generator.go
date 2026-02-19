package lightsout

import (
	"math/rand/v2"
)

// Generate creates a new w x h grid with a solvable puzzle.
// It starts with all lights off and applies random toggles.
func Generate(w, h int) [][]bool {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GenerateSeeded(w, h, rng)
}

func GenerateSeeded(w, h int, rng *rand.Rand) [][]bool {
	grid := make([][]bool, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]bool, w)
	}

	// Iterate through every cell and randomly decide whether to "click" it.
	// Since clicking a cell twice cancels out, and the order of clicks doesn't matter,
	// iterating once through all cells covers all reachable states from "all off".
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if rng.IntN(2) == 1 {
				Toggle(grid, x, y)
			}
		}
	}

	// Ensure the puzzle isn't already solved (all off).
	if IsSolved(grid) {
		// Toggle one random cell to ensure at least some lights are on.
		Toggle(grid, rng.IntN(w), rng.IntN(h))
	}

	return grid
}

// Toggle switches the state of the cell at (x,y) and its immediate neighbors.
// It assumes grid is rectangular.
func Toggle(grid [][]bool, x, y int) {
	h := len(grid)
	if h == 0 {
		return
	}
	w := len(grid[0])

	safeToggle := func(c, r int) {
		if c >= 0 && c < w && r >= 0 && r < h {
			grid[r][c] = !grid[r][c]
		}
	}

	safeToggle(x, y)
	safeToggle(x, y-1)
	safeToggle(x, y+1)
	safeToggle(x-1, y)
	safeToggle(x+1, y)
}

func IsSolved(grid [][]bool) bool {
	for _, row := range grid {
		for _, cell := range row {
			if cell {
				return false
			}
		}
	}
	return true
}
