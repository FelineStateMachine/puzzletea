package hitori

import (
	"encoding/json"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
)

// Save represents the serialized state of a Hitori game.
type Save struct {
	Size      int    `json:"size"`
	Numbers   string `json:"numbers"`
	Marks     string `json:"marks"`
	ModeTitle string `json:"mode_title"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Size:      m.size,
		Numbers:   m.numbers.String(),
		Marks:     serializeMarks(m.marks),
		ModeTitle: m.modeTitle,
	}
	data, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hitori save: %w", err)
	}
	return data, nil
}

// ImportModel reconstructs a Model from saved JSON data.
func ImportModel(data []byte) (*Model, error) {
	var s Save
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hitori save: %w", err)
	}
	if s.Size <= 0 {
		return nil, fmt.Errorf("invalid grid size in save: %d", s.Size)
	}

	numbers := newGrid(state(s.Numbers))
	marks := deserializeMarks(s.Marks, s.Size)

	m := &Model{
		size:      s.Size,
		numbers:   numbers,
		marks:     marks,
		cursor:    game.Cursor{X: 0, Y: 0},
		keys:      DefaultKeyMap,
		modeTitle: s.ModeTitle,
	}
	m.initialMarks = newMarks(s.Size)
	m.solved = m.checkSolved()
	return m, nil
}
