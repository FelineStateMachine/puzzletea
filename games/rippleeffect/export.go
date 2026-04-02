package rippleeffect

import (
	"encoding/json"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	State     string `json:"state"`
	Givens    string `json:"givens"`
	Cages     []Cage `json:"cages"`
	ModeTitle string `json:"mode_title"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Width:     m.width,
		Height:    m.height,
		State:     encodeGrid(m.grid),
		Givens:    encodeGrid(m.givens),
		Cages:     m.geo.cages,
		ModeTitle: m.modeTitle,
	}
	data, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("marshal ripple effect save: %w", err)
	}
	return data, nil
}

func ImportModel(data []byte) (*Model, error) {
	var save Save
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, fmt.Errorf("unmarshal ripple effect save: %w", err)
	}

	geo, err := buildGeometry(save.Width, save.Height, save.Cages)
	if err != nil {
		return nil, fmt.Errorf("build ripple effect geometry: %w", err)
	}

	state, err := parseGrid(save.State, save.Width, save.Height)
	if err != nil {
		return nil, fmt.Errorf("decode ripple effect state: %w", err)
	}
	givens, err := parseGrid(save.Givens, save.Width, save.Height)
	if err != nil {
		return nil, fmt.Errorf("decode ripple effect givens: %w", err)
	}

	for y := range save.Height {
		for x := range save.Width {
			if givens[y][x] != 0 && state[y][x] != givens[y][x] {
				return nil, fmt.Errorf("given at (%d,%d) does not match current state", x, y)
			}
			if givens[y][x] != 0 && !valueAllowed(geo, givens, point{x: x, y: y}, givens[y][x]) {
				return nil, fmt.Errorf("invalid given at (%d,%d)", x, y)
			}
		}
	}

	m := &Model{
		width:       save.Width,
		height:      save.Height,
		geo:         geo,
		grid:        state,
		initialGrid: cloneGrid(givens),
		givens:      givens,
		cursor:      game.Cursor{},
		keys:        DefaultKeyMap,
		modeTitle:   save.ModeTitle,
	}
	m.recompute()
	return m, nil
}
