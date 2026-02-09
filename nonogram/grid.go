package nonogram

import (
	"strings"
)

type grid [][]rune
type state string

func newGrid(s state) grid {
	rows := strings.Split(string(s), "\n")

	grid := make([][]rune, len(rows))

	for i, col := range rows {
		grid[i] = []rune(col)
	}

	return grid
}

func (g grid) String() string {
	rows := make([]string, len(g))

	for i, row := range g {
		rows[i] = string(row)
	}

	grid := strings.Join(rows, "\n")
	return grid
}

func createEmptyState(h, w int) state {
	if h <= 0 || w <= 0 {
		return ""
	}

	var b strings.Builder
	b.Grow((w + 1) * h)

	for i := range h {
		for range w {
			b.WriteRune(emptyTile)
		}
		if i < h-1 {
			b.WriteRune('\n')
		}
	}

	return state(b.String())
}
