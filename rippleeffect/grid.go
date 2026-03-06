package rippleeffect

import (
	"errors"
	"strconv"
	"strings"
)

var errInvalidGridEncoding = errors.New("invalid ripple effect grid encoding")

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
			fields = splitCompactFields(rows[y])
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

func encodeGrid(g grid) string {
	rows := make([]string, 0, len(g))
	for _, row := range g {
		fields := make([]string, len(row))
		for x, value := range row {
			if value == 0 {
				fields[x] = "."
				continue
			}
			fields[x] = strconv.Itoa(value)
		}
		rows = append(rows, strings.Join(fields, " "))
	}
	return strings.Join(rows, "\n")
}

func splitCompactFields(row string) []string {
	fields := make([]string, 0, len(row))
	for _, r := range row {
		fields = append(fields, string(r))
	}
	return fields
}
