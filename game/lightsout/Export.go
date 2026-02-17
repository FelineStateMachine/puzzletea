package lightsout

import (
	"encoding/json"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Grid        [][]bool `json:"grid"`
	InitialGrid [][]bool `json:"initial_grid,omitempty"`
	CursorX     int      `json:"cx"`
	CursorY     int      `json:"cy"`
	ModeTitle   string   `json:"mode_title"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Grid:        m.grid,
		InitialGrid: m.initialGrid,
		CursorX:     m.cursor.X,
		CursorY:     m.cursor.Y,
		ModeTitle:   m.modeTitle,
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

	// Fall back to a copy of the current grid for saves that predate
	// the initialGrid field.
	initial := s.InitialGrid
	if len(initial) == 0 {
		initial = make([][]bool, h)
		for y := range h {
			initial[y] = make([]bool, w)
			copy(initial[y], s.Grid[y])
		}
	}

	return &Model{
		grid:        s.Grid,
		initialGrid: initial,
		width:       w,
		height:      h,
		cursor:      game.Cursor{X: s.CursorX, Y: s.CursorY},
		modeTitle:   s.ModeTitle,
		keys:        DefaultKeyMap,
	}, nil
}
