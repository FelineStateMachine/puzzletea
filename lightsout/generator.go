package lightsout

import (
	"math/rand/v2"
)

// Generate creates a new w x h grid with a solvable puzzle.
// It starts with all lights off and applies random toggles.
func Generate(w, h int) [][]bool {
	grid := make([][]bool, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]bool, w)
	}

	// Iterate through every cell and randomly decide whether to "click" it.
	// Since clicking a cell twice cancels out, and the order of clicks doesn't matter,
	// iterating once through all cells covers all reachable states from "all off".
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if rand.IntN(2) == 1 {
				Toggle(grid, x, y)
			}
		}
	}

	// Ensure the puzzle isn't already solved (all off).
	if IsSolved(grid) {
		// Toggle one random cell to ensure at least some lights are on.
		Toggle(grid, rand.IntN(w), rand.IntN(h))
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

	// Helper to safely toggle
	safeToggle := func(c, r int) {
		if c >= 0 && c < w && r >= 0 && r < h {
			grid[r][c] = !grid[r][c]
		}
	}

	safeToggle(x, y)   // Center
	safeToggle(x, y-1) // Up
	safeToggle(x, y+1) // Down
	safeToggle(x-1, y) // Left
	safeToggle(x+1, y) // Right
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
