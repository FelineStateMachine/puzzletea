package hitori

import "strings"

type cellMark int

const (
	unmarked cellMark = iota
	shaded
	circled
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

func (g grid) clone() grid {
	c := make(grid, len(g))
	for i, row := range g {
		c[i] = make([]rune, len(row))
		copy(c[i], row)
	}
	return c
}

func newMarks(size int) [][]cellMark {
	marks := make([][]cellMark, size)
	for y := range size {
		marks[y] = make([]cellMark, size)
	}
	return marks
}

func cloneMarks(marks [][]cellMark) [][]cellMark {
	c := make([][]cellMark, len(marks))
	for y, row := range marks {
		c[y] = make([]cellMark, len(row))
		copy(c[y], row)
	}
	return c
}

func serializeMarks(marks [][]cellMark) string {
	var b strings.Builder
	for y, row := range marks {
		for _, m := range row {
			switch m {
			case shaded:
				b.WriteByte('X')
			case circled:
				b.WriteByte('O')
			default:
				b.WriteByte('.')
			}
		}
		if y < len(marks)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func deserializeMarks(s string, size int) [][]cellMark {
	rows := strings.Split(s, "\n")
	marks := make([][]cellMark, size)
	for y := range size {
		marks[y] = make([]cellMark, size)
		if y < len(rows) {
			for x, r := range rows[y] {
				if x >= size {
					break
				}
				switch r {
				case 'X':
					marks[y][x] = shaded
				case 'O':
					marks[y][x] = circled
				default:
					marks[y][x] = unmarked
				}
			}
		}
	}
	return marks
}
