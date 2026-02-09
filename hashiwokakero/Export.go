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

func ImportModel(data []byte) (*Model, error) {
	var save Save
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, err
	}

	cursorIsland := 0
	if len(save.Islands) > 0 {
		cursorIsland = save.Islands[0].ID
	}

	return &Model{
		puzzle: Puzzle{
			Width:   save.Width,
			Height:  save.Height,
			Islands: save.Islands,
			Bridges: save.Bridges,
		},
		cursorIsland: cursorIsland,
		keys:         DefaultKeyMap,
	}, nil
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
