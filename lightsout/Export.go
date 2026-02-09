package lightsout

import (
	"encoding/json"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Grid      [][]bool `json:"grid"`
	CursorX   int      `json:"cx"`
	CursorY   int      `json:"cy"`
	ModeTitle string   `json:"mode_title"`
	Solved    bool     `json:"solved"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Grid:      m.grid,
		CursorX:   m.cursor.X,
		CursorY:   m.cursor.Y,
		ModeTitle: m.modeTitle,
		Solved:    m.IsSolved(),
	}
	jsonData, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lightsout save: %w", err)
	}
	return jsonData, nil
}

func ImportModel(data []byte) (*Model, error) {
	var s Save
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lightsout save: %w", err)
	}

	if len(s.Grid) == 0 {
		return nil, fmt.Errorf("empty grid in save")
	}

	h := len(s.Grid)
	w := len(s.Grid[0])

	return &Model{
		grid:      s.Grid,
		width:     w,
		height:    h,
		cursor:    game.Cursor{X: s.CursorX, Y: s.CursorY},
		modeTitle: s.ModeTitle,
		keys:      DefaultKeyMap,
	}, nil
}
