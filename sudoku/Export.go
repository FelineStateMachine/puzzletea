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
	Solved   bool         `json:"solved"`
	Grid     string       `json:"grid"`
	Provided []exportCell `json:"provided"`
}

func gridToString(g grid) string {
	var b strings.Builder
	for y := range GRIDSIZE {
		for x := range GRIDSIZE {
			b.WriteByte(byte('0' + g[y][x].v))
		}
		if y < GRIDSIZE-1 {
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
	for y := 0; y < GRIDSIZE && y < len(rows); y++ {
		for x := 0; x < GRIDSIZE && x < len(rows[y]); x++ {
			v := int(rows[y][x] - '0')
			if v != 0 {
				g[y][x].v = v
			}
		}
	}

	return &Model{
		grid:     g,
		provided: provided,
		keys:     DefaultKeyMap,
	}, nil
}

func (m Model) GetSave() ([]byte, error) {
	provided := make([]exportCell, len(m.provided))
	for i, c := range m.provided {
		provided[i] = exportCell{X: c.x, Y: c.y, V: c.v}
	}
	save := Save{
		Solved:   m.isSolved(),
		Grid:     gridToString(m.grid),
		Provided: provided,
	}
	jsonData, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal save data: %v", err)
	}
	return jsonData, nil
}
