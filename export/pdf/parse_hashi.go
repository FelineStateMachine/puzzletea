package pdfexport

import (
	"encoding/json"
	"fmt"
	"strings"
)

type hashiSave struct {
	Width   int           `json:"width"`
	Height  int           `json:"height"`
	Islands []hashiIsland `json:"islands"`
}

type hashiIsland struct {
	X        int `json:"x"`
	Y        int `json:"y"`
	Required int `json:"required"`
}

func ParseHashiPrintData(saveData []byte) (*HashiData, error) {
	if len(strings.TrimSpace(string(saveData))) == 0 {
		return nil, nil
	}

	var save hashiSave
	if err := json.Unmarshal(saveData, &save); err != nil {
		return nil, fmt.Errorf("decode hashiwokakero save: %w", err)
	}
	if save.Width <= 0 || save.Height <= 0 {
		return nil, nil
	}

	islands := make([]HashiIsland, 0, len(save.Islands))
	for _, island := range save.Islands {
		if island.X < 0 || island.X >= save.Width || island.Y < 0 || island.Y >= save.Height {
			continue
		}
		if island.Required <= 0 {
			continue
		}
		islands = append(islands, HashiIsland(island))
	}

	return &HashiData{
		Width:   save.Width,
		Height:  save.Height,
		Islands: islands,
	}, nil
}
