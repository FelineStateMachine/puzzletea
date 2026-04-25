package netwalk

import (
	"encoding/json"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Size             int    `json:"size"`
	Masks            string `json:"masks"`
	Rotations        string `json:"rotations"`
	InitialRotations string `json:"initial_rotations"`
	Kinds            string `json:"kinds"`
	Locks            string `json:"locks"`
	ModeTitle        string `json:"mode_title"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Size:             m.puzzle.Size,
		Masks:            encodeMaskRows(m.puzzle.Tiles),
		Rotations:        encodeRotationRows(m.puzzle.Tiles, false),
		InitialRotations: encodeRotationRows(m.puzzle.Tiles, true),
		Kinds:            encodeKindRows(m.puzzle.Tiles),
		Locks:            encodeLockRows(m.puzzle.Tiles),
		ModeTitle:        m.modeTitle,
	}
	data, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("marshal netwalk save: %w", err)
	}
	return data, nil
}

func ImportModel(data []byte) (*Model, error) {
	var save Save
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, fmt.Errorf("unmarshal netwalk save: %w", err)
	}

	if save.Locks == "" {
		save.Locks = encodeRows(save.Size, func(_, _ int) byte { return '.' })
	}
	puzzle, err := decodePuzzle(save.Size, save.Masks, save.Rotations, save.InitialRotations, save.Kinds, save.Locks)
	if err != nil {
		return nil, err
	}

	cursor := puzzle.firstActive()
	m := &Model{
		puzzle:    puzzle,
		cursor:    game.Cursor{X: cursor.X, Y: cursor.Y},
		keys:      DefaultKeyMap,
		modeTitle: save.ModeTitle,
	}
	m.recompute()
	return m, nil
}
