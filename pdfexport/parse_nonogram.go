package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type nonogramSave struct {
	State    string  `json:"state"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	RowHints [][]int `json:"row-hints"`
	ColHints [][]int `json:"col-hints"`
}

func ParseNonogramPrintData(saveData []byte) (*NonogramData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save nonogramSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode nonogram save: %w", err)
	}

	stateRows := splitNonogramStateRows(save.State)

	width := save.Width
	if width <= 0 {
		width = len(save.ColHints)
	}
	if width <= 0 {
		width = maxRuneWidth(stateRows)
	}

	height := save.Height
	if height <= 0 {
		height = len(save.RowHints)
	}
	if height <= 0 {
		height = len(stateRows)
	}

	if width <= 0 || height <= 0 {
		return nil, nil
	}

	return &NonogramData{
		Width:    width,
		Height:   height,
		RowHints: normalizeNonogramHintRows(save.RowHints, height),
		ColHints: normalizeNonogramHintRows(save.ColHints, width),
		Grid:     normalizeNonogramStateGrid(stateRows, width, height),
	}, nil
}

func splitNonogramStateRows(raw string) []string {
	normalized := strings.ReplaceAll(strings.ReplaceAll(raw, "\r\n", "\n"), "\r", "\n")
	if normalized == "" {
		return nil
	}
	return strings.Split(normalized, "\n")
}

func maxRuneWidth(rows []string) int {
	maxWidth := 0
	for _, row := range rows {
		if n := len([]rune(row)); n > maxWidth {
			maxWidth = n
		}
	}
	return maxWidth
}

func normalizeNonogramHintRows(src [][]int, size int) [][]int {
	if size <= 0 {
		return nil
	}

	normalized := make([][]int, size)
	for i := range size {
		if i >= len(src) {
			normalized[i] = []int{0}
			continue
		}

		filtered := make([]int, 0, len(src[i]))
		for _, value := range src[i] {
			if value > 0 {
				filtered = append(filtered, value)
			}
		}
		if len(filtered) == 0 {
			filtered = []int{0}
		}
		normalized[i] = filtered
	}

	return normalized
}

func normalizeNonogramStateGrid(rows []string, width, height int) [][]string {
	if width <= 0 || height <= 0 {
		return nil
	}

	grid := make([][]string, height)
	for y := range height {
		grid[y] = make([]string, width)
		for x := range width {
			grid[y][x] = " "
		}

		if y >= len(rows) {
			continue
		}

		runes := []rune(rows[y])
		for x := 0; x < width && x < len(runes); x++ {
			if runes[x] == ' ' {
				continue
			}
			grid[y][x] = string(runes[x])
		}
	}

	return grid
}
