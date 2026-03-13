package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type shikakuSave struct {
	Width  int           `json:"width"`
	Height int           `json:"height"`
	Clues  []shikakuClue `json:"clues"`
}

type shikakuClue struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Value int `json:"value"`
}

func ParseShikakuPrintData(saveData []byte) (*ShikakuData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save shikakuSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode shikaku save: %w", err)
	}
	if save.Width <= 0 || save.Height <= 0 {
		return nil, nil
	}

	clues := make([][]int, save.Height)
	for y := 0; y < save.Height; y++ {
		clues[y] = make([]int, save.Width)
	}

	for _, clue := range save.Clues {
		if clue.X < 0 || clue.X >= save.Width || clue.Y < 0 || clue.Y >= save.Height {
			continue
		}
		if clue.Value <= 0 {
			continue
		}
		clues[clue.Y][clue.X] = clue.Value
	}

	return &ShikakuData{
		Width:  save.Width,
		Height: save.Height,
		Clues:  clues,
	}, nil
}
