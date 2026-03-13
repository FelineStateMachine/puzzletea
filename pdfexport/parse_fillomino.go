package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type fillominoSave struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	State    string `json:"state"`
	Provided string `json:"provided"`
}

func ParseFillominoPrintData(saveData []byte) (*FillominoData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save fillominoSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode fillomino save: %w", err)
	}
	if save.Width <= 0 || save.Height <= 0 {
		return nil, nil
	}

	state, err := parseNumberGrid(save.State, save.Width, save.Height)
	if err != nil {
		return nil, err
	}
	provided := parseProvidedMask(save.Provided, save.Width, save.Height)

	givens := make([][]int, save.Height)
	for y := 0; y < save.Height; y++ {
		givens[y] = make([]int, save.Width)
		for x := 0; x < save.Width; x++ {
			if provided[y][x] {
				givens[y][x] = state[y][x]
			}
		}
	}

	return &FillominoData{
		Width:  save.Width,
		Height: save.Height,
		Givens: givens,
	}, nil
}
