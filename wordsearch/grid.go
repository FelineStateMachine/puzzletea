package wordsearch

import "strings"

type (
	grid  [][]rune
	state string
)

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

func (g grid) Get(x, y int) rune {
	if y < 0 || y >= len(g) || x < 0 || x >= len(g[y]) {
		return 0
	}
	return g[y][x]
}

func (g grid) Set(x, y int, r rune) {
	if y >= 0 && y < len(g) && x >= 0 && x < len(g[y]) {
		g[y][x] = r
	}
}
