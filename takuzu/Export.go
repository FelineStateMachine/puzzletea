package takuzu

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
)

// Save represents the serialized state of a Takuzu game.
type Save struct {
	Size      int    `json:"size"`
	State     string `json:"state"`
	Provided  string `json:"provided"`
	ModeTitle string `json:"mode_title"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Size:      m.size,
		State:     m.grid.String(),
		Provided:  serializeProvided(m.provided),
		ModeTitle: m.modeTitle,
	}
	data, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal takuzu save: %w", err)
	}
	return data, nil
}

// ImportModel reconstructs a Model from saved JSON data.
func ImportModel(data []byte) (*Model, error) {
	var s Save
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal takuzu save: %w", err)
	}
	if s.Size <= 0 {
		return nil, fmt.Errorf("invalid grid size in save: %d", s.Size)
	}

	g := newGrid(state(s.State))
	provided := deserializeProvided(s.Provided, s.Size)

	m := &Model{
		size:      s.Size,
		grid:      g,
		provided:  provided,
		cursor:    game.Cursor{X: 0, Y: 0},
		keys:      DefaultKeyMap,
		modeTitle: s.ModeTitle,
	}
	m.solved = m.checkSolved()
	return m, nil
}

// serializeProvided encodes the provided matrix as a string of '.' (not provided) and '#' (provided).
func serializeProvided(p [][]bool) string {
	var b strings.Builder
	for y, row := range p {
		for _, v := range row {
			if v {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		if y < len(p)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// deserializeProvided decodes a provided-matrix string back into [][]bool.
func deserializeProvided(s string, size int) [][]bool {
	rows := strings.Split(s, "\n")
	p := make([][]bool, size)
	for y := range size {
		p[y] = make([]bool, size)
		if y < len(rows) {
			for x := range size {
				if x < len(rows[y]) && rows[y][x] == '#' {
					p[y][x] = true
				}
			}
		}
	}
	return p
}
