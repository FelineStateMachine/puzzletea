package wordsearch

import (
	"encoding/json"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Grid       string `json:"grid"`
	Words      []Word `json:"words"`
	CursorX    int    `json:"cursor_x"`
	CursorY    int    `json:"cursor_y"`
	Selection  int    `json:"selection"`
	SelectionX int    `json:"selection_x"`
	SelectionY int    `json:"selection_y"`
	Solved     bool   `json:"solved"`
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
		cursor:         game.Cursor{X: exported.CursorX, Y: exported.CursorY},
		selection:      selectionState(exported.Selection),
		selectionStart: game.Cursor{X: exported.SelectionX, Y: exported.SelectionY},
		keys:           DefaultKeyMap,
		solved:         exported.Solved,
	}, nil
}
