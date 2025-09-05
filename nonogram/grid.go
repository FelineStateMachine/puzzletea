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

// loadSave - Parses optional param save. If none, uses default.
func loadSave(def state, optionalSave ...string) state {
	var s state
	//TODO: use default state as state and layer save runes onto it. this would allow save strings to be malformed a bit.
	if len(optionalSave) > 0 {
		s = state(optionalSave[0])
	} else {
		s = def
	}
	return s
}
