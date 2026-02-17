package hashiwokakero

import (
	"encoding/json"
	"fmt"
)

type Save struct {
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	Islands   []Island `json:"islands"`
	Bridges   []Bridge `json:"bridges"`
	ModeTitle string   `json:"mode_title,omitempty"`
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
		modeTitle:    save.ModeTitle,
		keys:         DefaultKeyMap,
	}, nil
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Width:     m.puzzle.Width,
		Height:    m.puzzle.Height,
		Islands:   m.puzzle.Islands,
		Bridges:   m.puzzle.Bridges,
		ModeTitle: m.modeTitle,
	}
	jsonData, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal save data: %w", err)
	}
	return jsonData, nil
}
