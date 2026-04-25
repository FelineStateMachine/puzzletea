package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type takuzuPlusSave struct {
	Size                int    `json:"size"`
	State               string `json:"state"`
	Provided            string `json:"provided"`
	HorizontalRelations string `json:"horizontal_relations"`
	VerticalRelations   string `json:"vertical_relations"`
}

func ParseTakuzuPlusPrintData(saveData []byte) (*TakuzuData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save takuzuPlusSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode takuzu+ save: %w", err)
	}

	stateRows := splitNormalizedLines(save.State)
	providedRows := splitNormalizedLines(save.Provided)
	horizontalRows := splitNormalizedLines(save.HorizontalRelations)
	verticalRows := splitNormalizedLines(save.VerticalRelations)

	size := save.Size
	if size <= 0 {
		size = max(max(len(stateRows), len(providedRows)), max(len(horizontalRows), len(verticalRows)+1))
	}
	if size <= 0 {
		return nil, nil
	}

	givens := make([][]string, size)
	horizontal := make([][]string, size)
	vertical := make([][]string, max(size-1, 0))
	for y := 0; y < size; y++ {
		givens[y] = make([]string, size)
		horizontal[y] = make([]string, max(size-1, 0))

		var stateRunes []rune
		if y < len(stateRows) {
			stateRunes = []rune(stateRows[y])
		}
		var providedRunes []rune
		if y < len(providedRows) {
			providedRunes = []rune(providedRows[y])
		}
		var relationRunes []rune
		if y < len(horizontalRows) {
			relationRunes = []rune(horizontalRows[y])
		}

		for x := 0; x < size; x++ {
			if x < size-1 && x < len(relationRunes) && (relationRunes[x] == '=' || relationRunes[x] == 'x') {
				horizontal[y][x] = string(relationRunes[x])
			}
			if x >= len(providedRunes) || providedRunes[x] != '#' || x >= len(stateRunes) {
				continue
			}
			if stateRunes[x] != '0' && stateRunes[x] != '1' {
				continue
			}
			givens[y][x] = string(stateRunes[x])
		}
	}

	for y := 0; y < size-1; y++ {
		vertical[y] = make([]string, size)
		if y >= len(verticalRows) {
			continue
		}
		relationRunes := []rune(verticalRows[y])
		for x := 0; x < size && x < len(relationRunes); x++ {
			if relationRunes[x] == '=' || relationRunes[x] == 'x' {
				vertical[y][x] = string(relationRunes[x])
			}
		}
	}

	return &TakuzuData{
		Size:                size,
		Givens:              givens,
		HorizontalRelations: horizontal,
		VerticalRelations:   vertical,
		GroupEveryTwo:       true,
	}, nil
}
