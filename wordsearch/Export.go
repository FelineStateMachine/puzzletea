package wordsearch

import "encoding/json"

type exportData struct {
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

func exportModel(m Model) ([]byte, error) {
	data := exportData{
		Width:      m.width,
		Height:     m.height,
		Grid:       m.grid.String(),
		Words:      m.words,
		CursorX:    m.cursor.x,
		CursorY:    m.cursor.y,
		Selection:  int(m.selection),
		SelectionX: m.selectionStart.x,
		SelectionY: m.selectionStart.y,
		Won:        m.won,
	}

	return json.Marshal(data)
}

func ImportModel(data []byte) (*Model, error) {
	var exported exportData
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
		keys:           newKeyMap(),
		won:            exported.Won,
	}, nil
}
