package nonogram

import (
	"encoding/json"
	"fmt"
)

type Save struct {
	State    string               `json:"state"`
	Width    int                  `json:"width"`
	Height   int                  `json:"height"`
	RowHints TomographyDefinition `json:"row-hints"`
	ColHints TomographyDefinition `json:"col-hints"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		RowHints: m.rowHints,
		ColHints: m.colHints,
		State:    m.grid.String(),
		Width:    m.width,
		Height:   m.height,
	}
	jsonData, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal save data: %w", err)
	}
	return jsonData, nil
}

func ImportModel(data []byte) (*Model, error) {
	var save Save
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, err
	}
	g := newGrid(state(save.State))
	curr := generateTomography(g)
	return &Model{
		width:        save.Width,
		height:       save.Height,
		rowHints:     save.RowHints,
		colHints:     save.ColHints,
		grid:         g,
		keys:         DefaultKeyMap,
		currentHints: curr,
		solved:       curr.rows.equal(save.RowHints) && curr.cols.equal(save.ColHints),
	}, nil
}
