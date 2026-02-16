package sudoku

import (
	"encoding/json"
	"fmt"
	"strings"
)

type exportCell struct {
	X int `json:"x"`
	Y int `json:"y"`
	V int `json:"v"`
}

type Save struct {
	Grid      string       `json:"grid"`
	Provided  []exportCell `json:"provided"`
	ModeTitle string       `json:"mode_title,omitempty"`
}

func gridToString(g grid) string {
	var b strings.Builder
	for y := range gridSize {
		for x := range gridSize {
			b.WriteByte(byte('0' + g[y][x].v))
		}
		if y < gridSize-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func ImportModel(data []byte) (*Model, error) {
	var save Save
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, err
	}

	provided := make([]cell, len(save.Provided))
	for i, ec := range save.Provided {
		provided[i] = cell{x: ec.X, y: ec.Y, v: ec.V}
	}

	g := newGrid(provided)

	// Parse save.Grid to restore user-entered values.
	rows := strings.Split(save.Grid, "\n")
	for y := 0; y < gridSize && y < len(rows); y++ {
		for x := 0; x < gridSize && x < len(rows[y]); x++ {
			v := int(rows[y][x] - '0')
			if v != 0 {
				g[y][x].v = v
			}
		}
	}

	return &Model{
		grid:         g,
		provided:     provided,
		providedGrid: buildProvidedGrid(provided),
		conflicts:    computeConflicts(g),
		modeTitle:    save.ModeTitle,
		keys:         DefaultKeyMap,
	}, nil
}

func (m Model) GetSave() ([]byte, error) {
	provided := make([]exportCell, len(m.provided))
	for i, c := range m.provided {
		provided[i] = exportCell{X: c.x, Y: c.y, V: c.v}
	}
	save := Save{
		Grid:      gridToString(m.grid),
		Provided:  provided,
		ModeTitle: m.modeTitle,
	}
	jsonData, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal save data: %w", err)
	}
	return jsonData, nil
}
