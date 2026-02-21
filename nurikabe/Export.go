package nurikabe

import (
	"encoding/json"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Clues     string `json:"clues"`
	Marks     string `json:"marks"`
	ModeTitle string `json:"mode_title"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Width:     m.width,
		Height:    m.height,
		Clues:     serializeClues(m.clues),
		Marks:     m.marks.String(),
		ModeTitle: m.modeTitle,
	}
	data, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nurikabe save: %w", err)
	}
	return data, nil
}

func ImportModel(data []byte) (*Model, error) {
	var s Save
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nurikabe save: %w", err)
	}
	if s.Width <= 0 || s.Height <= 0 {
		return nil, fmt.Errorf("invalid save dimensions: %dx%d", s.Width, s.Height)
	}

	clues, err := parseClues(s.Clues, s.Width, s.Height)
	if err != nil {
		return nil, fmt.Errorf("invalid clue data: %w", err)
	}
	if err := validateClues(clues, s.Width, s.Height); err != nil {
		return nil, err
	}

	marks, err := parseGrid(s.Marks, s.Width, s.Height)
	if err != nil {
		return nil, fmt.Errorf("invalid mark data: %w", err)
	}

	for y := range s.Height {
		for x := range s.Width {
			if clues[y][x] > 0 {
				marks[y][x] = islandCell
			}
		}
	}

	initial := newGrid(s.Width, s.Height, unknownCell)
	for y := range s.Height {
		for x := range s.Width {
			if clues[y][x] > 0 {
				initial[y][x] = islandCell
			}
		}
	}

	m := &Model{
		width:        s.Width,
		height:       s.Height,
		clues:        clues,
		marks:        marks,
		initialMarks: initial,
		cursor:       game.Cursor{X: 0, Y: 0},
		keys:         DefaultKeyMap,
		modeTitle:    s.ModeTitle,
		dragTarget:   unknownCell,
	}
	m.recomputeDerivedState()
	return m, nil
}
