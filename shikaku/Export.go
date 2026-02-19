package shikaku

import (
	"encoding/json"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
)

// Save represents the serialized state of a Shikaku game.
type Save struct {
	Width      int         `json:"width"`
	Height     int         `json:"height"`
	Clues      []Clue      `json:"clues"`
	Rectangles []Rectangle `json:"rectangles"`
	ModeTitle  string      `json:"mode_title"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Width:      m.puzzle.Width,
		Height:     m.puzzle.Height,
		Clues:      m.puzzle.Clues,
		Rectangles: m.puzzle.Rectangles,
		ModeTitle:  m.modeTitle,
	}
	data, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal shikaku save: %w", err)
	}
	return data, nil
}

// ImportModel reconstructs a Model from saved JSON data.
func ImportModel(data []byte) (*Model, error) {
	var s Save
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shikaku save: %w", err)
	}

	p := Puzzle{
		Width:      s.Width,
		Height:     s.Height,
		Clues:      s.Clues,
		Rectangles: s.Rectangles,
	}
	p.autoPlaceSingles()

	return &Model{
		puzzle:    p,
		cursor:    game.Cursor{X: 0, Y: 0},
		keys:      DefaultKeyMap,
		modeTitle: s.ModeTitle,
	}, nil
}
