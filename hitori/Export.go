package hitori

import (
	"encoding/json"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Size      int    `json:"size"`
	State     string `json:"state"`
	Provided  string `json:"provided"`
	ModeTitle string `json:"mode_title"`
}

func (m Model) GetSave() ([]byte, error) {
	s := Save{
		Size:      m.size,
		State:     m.grid.String(),
		Provided:  serializeProvided(m.provided),
		ModeTitle: m.modeTitle,
	}
	return json.Marshal(s)
}

func ImportModel(data []byte) (*Model, error) {
	var s Save
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	grid := newGrid(state(s.State))
	provided := deserializeProvided(s.Provided, s.Size)

	m := &Model{
		size:        s.Size,
		grid:        grid,
		initialGrid: grid.clone(),
		provided:    provided,
		cursor:      game.Cursor{X: s.Size / 2, Y: s.Size / 2},
		keys:        DefaultKeyMap,
		modeTitle:   s.ModeTitle,
	}
	m.solved = m.checkSolved()
	return m, nil
}

func serializeProvided(p [][]bool) string {
	result := make([]rune, 0, len(p)*len(p[0])+len(p)-1)
	for y, row := range p {
		for _, val := range row {
			if val {
				result = append(result, '#')
			} else {
				result = append(result, '.')
			}
		}
		if y < len(p)-1 {
			result = append(result, '\n')
		}
	}
	return string(result)
}

func deserializeProvided(data string, size int) [][]bool {
	result := make([][]bool, size)
	for i := range result {
		result[i] = make([]bool, size)
	}

	row := 0
	col := 0
	for _, ch := range data {
		if ch == '\n' {
			row++
			col = 0
			continue
		}
		if row < size && col < size {
			result[row][col] = (ch == '#')
			col++
		}
	}
	return result
}
