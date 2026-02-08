package hashiwokakero

import (
	"encoding/json"
	"fmt"
)

type Save struct {
	Solved  bool     `json:"solved"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	Islands []Island `json:"islands"`
	Bridges []Bridge `json:"bridges"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Solved:  m.puzzle.IsSolved(),
		Width:   m.puzzle.Width,
		Height:  m.puzzle.Height,
		Islands: m.puzzle.Islands,
		Bridges: m.puzzle.Bridges,
	}
	jsonData, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal save data: %v", err)
	}
	return jsonData, nil
}
