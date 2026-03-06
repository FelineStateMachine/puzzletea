package fillomino

import (
	"encoding/json"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	State        string `json:"state"`
	Provided     string `json:"provided"`
	ModeTitle    string `json:"mode_title"`
	MaxCellValue int    `json:"max_cell_value,omitempty"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Width:        m.width,
		Height:       m.height,
		State:        encodeGrid(m.grid),
		Provided:     serializeProvided(m.provided),
		ModeTitle:    m.modeTitle,
		MaxCellValue: m.maxCellValue,
	}
	data, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("marshal fillomino save: %w", err)
	}
	return data, nil
}

func ImportModel(data []byte) (*Model, error) {
	var save Save
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, fmt.Errorf("unmarshal fillomino save: %w", err)
	}
	if save.Width <= 0 || save.Height <= 0 {
		return nil, fmt.Errorf("invalid fillomino size %dx%d", save.Width, save.Height)
	}

	state, err := parseGrid(save.State, save.Width, save.Height)
	if err != nil {
		return nil, fmt.Errorf("decode fillomino state: %w", err)
	}
	provided := deserializeProvided(save.Provided, save.Width, save.Height)
	for y := range save.Height {
		for x := range save.Width {
			if provided[y][x] && state[y][x] == 0 {
				return nil, fmt.Errorf("provided cell at (%d,%d) is empty", x, y)
			}
		}
	}

	m := &Model{
		width:        save.Width,
		height:       save.Height,
		grid:         state,
		initialGrid:  initialGridForImport(state, provided),
		provided:     provided,
		cursor:       game.Cursor{},
		keys:         DefaultKeyMap,
		modeTitle:    save.ModeTitle,
		maxCellValue: restoredMaxCellValue(save.MaxCellValue, state),
	}
	m.recompute()
	return m, nil
}

func restoredMaxCellValue(savedMax int, state grid) int {
	if savedMax > 0 {
		return savedMax
	}

	return max(1, maxCellValue(state))
}

func initialGridForImport(state grid, provided [][]bool) grid {
	initial := newGrid(len(state[0]), len(state))
	for y := range len(state) {
		for x := range len(state[y]) {
			if provided[y][x] {
				initial[y][x] = state[y][x]
			}
		}
	}
	return initial
}
