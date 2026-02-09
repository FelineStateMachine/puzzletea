package wordsearch

import "encoding/json"

type Save struct {
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Grid           string `json:"grid"`
	Words          []Word `json:"words"`
	CursorX        int    `json:"cursor_x"`
	CursorY        int    `json:"cursor_y"`
	Selection      int    `json:"selection"`
	SelectionX     int    `json:"selection_x"`
	SelectionY     int    `json:"selection_y"`
	Won            bool   `json:"won"`
}

func ImportModel(data []byte) (*Model, error) {
	var exported Save
	if err := json.Unmarshal(data, &exported); err != nil {
		return nil, err
	}

	return &Model{
		width:          exported.Width,
		height:         exported.Height,
		grid:           newGrid(state(exported.Grid)),
		words:          exported.Words,
		cursor:         cursor{x: exported.CursorX, y: exported.CursorY},
		selection:      selectionState(exported.Selection),
		selectionStart: cursor{x: exported.SelectionX, y: exported.SelectionY},
		keys:           DefaultKeyMap,
		won:            exported.Won,
	}, nil
}
