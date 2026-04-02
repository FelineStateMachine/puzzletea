package takuzuplus

import (
	"strings"

	"github.com/FelineStateMachine/puzzletea/games/takuzu"
)

const (
	emptyCell = takuzu.EmptyCell
	zeroCell  = takuzu.ZeroCell
	oneCell   = takuzu.OneCell

	relationNone rune = '.'
	relationSame rune = '='
	relationDiff rune = 'x'
)

type grid [][]rune

type relations struct {
	horizontal [][]rune
	vertical   [][]rune
}

func newGridFromState(s string) grid {
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

func (g grid) clone() grid {
	clone := make(grid, len(g))
	for y, row := range g {
		clone[y] = make([]rune, len(row))
		copy(clone[y], row)
	}
	return clone
}

func newRelations(size int) relations {
	r := relations{
		horizontal: make([][]rune, size),
		vertical:   make([][]rune, max(size-1, 0)),
	}
	for y := range size {
		r.horizontal[y] = make([]rune, max(size-1, 0))
		for x := range r.horizontal[y] {
			r.horizontal[y][x] = relationNone
		}
	}
	for y := range size - 1 {
		r.vertical[y] = make([]rune, size)
		for x := range r.vertical[y] {
			r.vertical[y][x] = relationNone
		}
	}
	return r
}

func (r relations) clone() relations {
	clone := newRelations(len(r.horizontal))
	for y := range r.horizontal {
		copy(clone.horizontal[y], r.horizontal[y])
	}
	for y := range r.vertical {
		copy(clone.vertical[y], r.vertical[y])
	}
	return clone
}

func serializeRuneRows(rows [][]rune) string {
	var b strings.Builder
	for y, row := range rows {
		b.WriteString(string(row))
		if y < len(rows)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func deserializeRuneRows(s string, rows, cols int) [][]rune {
	if rows <= 0 {
		return [][]rune{}
	}

	parsed := strings.Split(s, "\n")
	out := make([][]rune, rows)
	for y := range rows {
		out[y] = make([]rune, cols)
		for x := range cols {
			out[y][x] = relationNone
		}
		if y >= len(parsed) {
			continue
		}
		runes := []rune(parsed[y])
		for x := 0; x < cols && x < len(runes); x++ {
			out[y][x] = runes[x]
		}
	}
	return out
}

func countRelations(r relations) int {
	total := 0
	for _, row := range r.horizontal {
		for _, value := range row {
			if value != relationNone {
				total++
			}
		}
	}
	for _, row := range r.vertical {
		for _, value := range row {
			if value != relationNone {
				total++
			}
		}
	}
	return total
}
