package fillomino

import (
	"strconv"
	"strings"
)

type point struct {
	x int
	y int
}

type grid [][]int

func newGrid(width, height int) grid {
	cells := make(grid, height)
	for y := range height {
		cells[y] = make([]int, width)
	}
	return cells
}

func cloneGrid(src grid) grid {
	if src == nil {
		return nil
	}

	height := len(src)
	width := 0
	if height > 0 {
		width = len(src[0])
	}

	dst := newGrid(width, height)
	for y := range height {
		copy(dst[y], src[y])
	}
	return dst
}

func parseGrid(encoded string, width, height int) (grid, error) {
	cells := newGrid(width, height)
	rows := strings.Split(strings.TrimSpace(encoded), "\n")
	for y := range height {
		if y >= len(rows) {
			continue
		}

		fields := strings.Fields(rows[y])
		if len(fields) == 0 {
			fields = splitCompactRow(rows[y])
		}
		if len(fields) != width {
			return nil, errInvalidGridEncoding
		}
		for x := range width {
			if fields[x] == "." {
				continue
			}
			value, err := strconv.Atoi(fields[x])
			if err != nil {
				return nil, errInvalidGridEncoding
			}
			cells[y][x] = value
		}
	}
	return cells, nil
}

func splitCompactRow(row string) []string {
	items := make([]string, 0, len(row))
	for _, r := range row {
		items = append(items, string(r))
	}
	return items
}

func encodeGrid(g grid) string {
	var rows []string
	for y, row := range g {
		items := make([]string, len(row))
		for x, value := range row {
			if value == 0 {
				items[x] = "."
				continue
			}
			items[x] = strconv.Itoa(value)
		}
		rows = append(rows, strings.Join(items, " "))
		if y == len(g)-1 {
			break
		}
	}
	return strings.Join(rows, "\n")
}

func newProvidedMask(givens grid) [][]bool {
	mask := make([][]bool, len(givens))
	for y := range len(givens) {
		mask[y] = make([]bool, len(givens[y]))
		for x := range len(givens[y]) {
			mask[y][x] = givens[y][x] != 0
		}
	}
	return mask
}

func serializeProvided(mask [][]bool) string {
	var rows []string
	for _, row := range mask {
		var line strings.Builder
		for _, value := range row {
			if value {
				line.WriteByte('#')
				continue
			}
			line.WriteByte('.')
		}
		rows = append(rows, line.String())
	}
	return strings.Join(rows, "\n")
}

func deserializeProvided(encoded string, width, height int) [][]bool {
	mask := make([][]bool, height)
	rows := strings.Split(encoded, "\n")
	for y := range height {
		mask[y] = make([]bool, width)
		if y >= len(rows) {
			continue
		}
		for x := range width {
			if x < len(rows[y]) && rows[y][x] == '#' {
				mask[y][x] = true
			}
		}
	}
	return mask
}

func maxCellValue(g grid) int {
	best := 0
	for y := range len(g) {
		for x := range len(g[y]) {
			if g[y][x] > best {
				best = g[y][x]
			}
		}
	}
	return best
}
