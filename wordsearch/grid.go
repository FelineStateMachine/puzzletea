package wordsearch

import "strings"

// grid is a 2D array of runes representing the letter grid
type grid [][]rune

// state is the string serialization of the grid
type state string

// newGrid creates a grid from a serialized state
func newGrid(s state) grid {
	if s == "" {
		return grid{}
	}

	lines := strings.Split(string(s), "\n")
	g := make(grid, len(lines))

	for i, line := range lines {
		g[i] = []rune(line)
	}

	return g
}

// String serializes the grid to a state string
func (g grid) String() string {
	if len(g) == 0 {
		return ""
	}

	lines := make([]string, len(g))
	for i, row := range g {
		lines[i] = string(row)
	}

	return strings.Join(lines, "\n")
}

// createEmptyGrid creates a grid of the given dimensions filled with spaces
func createEmptyGrid(height, width int) grid {
	g := make(grid, height)
	for i := range g {
		g[i] = make([]rune, width)
		for j := range g[i] {
			g[i][j] = ' '
		}
	}
	return g
}

// Get returns the rune at the given position, or 0 if out of bounds
func (g grid) Get(x, y int) rune {
	if y < 0 || y >= len(g) || x < 0 || x >= len(g[y]) {
		return 0
	}
	return g[y][x]
}

// Set sets the rune at the given position if in bounds
func (g grid) Set(x, y int, r rune) {
	if y >= 0 && y < len(g) && x >= 0 && x < len(g[y]) {
		g[y][x] = r
	}
}
