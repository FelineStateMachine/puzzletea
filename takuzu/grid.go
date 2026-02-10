package takuzu

import (
	"strings"
)

const (
	emptyCell rune = '.'
	zeroCell  rune = '0'
	oneCell   rune = '1'
)

type (
	grid  [][]rune
	state string
)

func newGrid(s state) grid {
	rows := strings.Split(string(s), "\n")
	g := make(grid, len(rows))
	for i, row := range rows {
		g[i] = []rune(row)
	}
	return g
}

func (g grid) String() string {
	rows := make([]string, len(g))
	for i, row := range g {
		rows[i] = string(row)
	}
	return strings.Join(rows, "\n")
}

func createEmptyState(size int) state {
	var b strings.Builder
	b.Grow((size + 1) * size)
	for y := range size {
		for range size {
			b.WriteRune(emptyCell)
		}
		if y < size-1 {
			b.WriteRune('\n')
		}
	}
	return state(b.String())
}

func (g grid) clone() grid {
	c := make(grid, len(g))
	for i, row := range g {
		c[i] = make([]rune, len(row))
		copy(c[i], row)
	}
	return c
}
