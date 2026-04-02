package pdfexport

import (
	"fmt"
	"strconv"
	"strings"
)

func splitNormalizedLines(raw string) []string {
	normalized := strings.ReplaceAll(strings.ReplaceAll(raw, "\r\n", "\n"), "\r", "\n")
	if strings.TrimSpace(normalized) == "" {
		return nil
	}
	return strings.Split(normalized, "\n")
}

func parseNumberGrid(encoded string, width, height int) ([][]int, error) {
	rows := strings.Split(strings.TrimSpace(encoded), "\n")
	grid := make([][]int, height)
	for y := range height {
		grid[y] = make([]int, width)
		if y >= len(rows) {
			continue
		}

		fields := strings.Fields(rows[y])
		if len(fields) == 0 {
			fields = splitCompactFields(rows[y])
		}
		if len(fields) != width {
			return nil, fmt.Errorf("decode number grid: invalid row width")
		}
		for x := range width {
			if fields[x] == "." {
				continue
			}
			value, err := strconv.Atoi(fields[x])
			if err != nil {
				return nil, fmt.Errorf("decode number grid: %w", err)
			}
			grid[y][x] = value
		}
	}
	return grid, nil
}

func parseProvidedMask(encoded string, width, height int) [][]bool {
	rows := strings.Split(encoded, "\n")
	mask := make([][]bool, height)
	for y := range height {
		mask[y] = make([]bool, width)
		if y >= len(rows) {
			continue
		}
		for x := 0; x < width && x < len(rows[y]); x++ {
			mask[y][x] = rows[y][x] == '#'
		}
	}
	return mask
}

func splitCompactFields(row string) []string {
	fields := make([]string, 0, len(row))
	for _, r := range row {
		fields = append(fields, string(r))
	}
	return fields
}
