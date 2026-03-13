package pdfexport

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type nurikabeSave struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Clues  string `json:"clues"`
}

func ParseNurikabePrintData(saveData []byte) (*NurikabeData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save nurikabeSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode nurikabe save: %w", err)
	}

	width := save.Width
	height := save.Height
	if width <= 0 || height <= 0 {
		return nil, nil
	}

	clues, err := parseNurikabeClues(save.Clues, width, height)
	if err != nil {
		return nil, err
	}

	return &NurikabeData{
		Width:  width,
		Height: height,
		Clues:  clues,
	}, nil
}

func parseNurikabeClues(raw string, width, height int) ([][]int, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid clue dimensions: %dx%d", width, height)
	}

	clues := make([][]int, height)
	for y := range height {
		clues[y] = make([]int, width)
	}

	rows := splitNormalizedLines(raw)
	for y := 0; y < len(rows) && y < height; y++ {
		parts := strings.Split(rows[y], ",")
		for x := 0; x < len(parts) && x < width; x++ {
			token := strings.TrimSpace(parts[x])
			if token == "" {
				continue
			}
			value, err := strconv.Atoi(token)
			if err != nil {
				return nil, fmt.Errorf("invalid clue value %q at (%d,%d): %w", token, x, y, err)
			}
			if value < 0 {
				return nil, fmt.Errorf("negative clue value %d at (%d,%d)", value, x, y)
			}
			clues[y][x] = value
		}
	}

	return clues, nil
}
