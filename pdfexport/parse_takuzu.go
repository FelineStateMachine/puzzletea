package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type takuzuSave struct {
	Size     int    `json:"size"`
	State    string `json:"state"`
	Provided string `json:"provided"`
}

func ParseTakuzuPrintData(saveData []byte) (*TakuzuData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save takuzuSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode takuzu save: %w", err)
	}

	stateRows := splitNormalizedLines(save.State)
	providedRows := splitNormalizedLines(save.Provided)

	size := save.Size
	if size <= 0 {
		size = max(len(stateRows), len(providedRows))
	}
	if size <= 0 {
		return nil, nil
	}

	givens := make([][]string, size)
	for y := 0; y < size; y++ {
		givens[y] = make([]string, size)

		var stateRunes []rune
		if y < len(stateRows) {
			stateRunes = []rune(stateRows[y])
		}

		var providedRunes []rune
		if y < len(providedRows) {
			providedRunes = []rune(providedRows[y])
		}

		for x := 0; x < size; x++ {
			if x >= len(providedRunes) || providedRunes[x] != '#' {
				continue
			}
			if x >= len(stateRunes) {
				continue
			}
			if stateRunes[x] != '0' && stateRunes[x] != '1' {
				continue
			}
			givens[y][x] = string(stateRunes[x])
		}
	}

	return &TakuzuData{
		Size:                size,
		Givens:              givens,
		HorizontalRelations: make([][]string, size),
		VerticalRelations:   make([][]string, max(size-1, 0)),
		GroupEveryTwo:       true,
	}, nil
}
