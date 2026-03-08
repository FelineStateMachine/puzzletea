package takuzuplus

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Save struct {
	Size                int    `json:"size"`
	State               string `json:"state"`
	Provided            string `json:"provided"`
	ModeTitle           string `json:"mode_title"`
	HorizontalRelations string `json:"horizontal_relations"`
	VerticalRelations   string `json:"vertical_relations"`
}

func (m Model) GetSave() ([]byte, error) {
	save := Save{
		Size:                m.size,
		State:               m.grid.String(),
		Provided:            serializeProvided(m.provided),
		ModeTitle:           m.modeTitle,
		HorizontalRelations: serializeRuneRows(m.relations.horizontal),
		VerticalRelations:   serializeRuneRows(m.relations.vertical),
	}
	data, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal takuzu+ save: %w", err)
	}
	return data, nil
}

func ImportModel(data []byte) (*Model, error) {
	var s Save
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal takuzu+ save: %w", err)
	}
	if s.Size <= 0 {
		return nil, fmt.Errorf("invalid grid size in save: %d", s.Size)
	}

	g := newGridFromState(s.State)
	provided := deserializeProvided(s.Provided, s.Size)
	rels := relations{
		horizontal: deserializeRuneRows(s.HorizontalRelations, s.Size, max(s.Size-1, 0)),
		vertical:   deserializeRuneRows(s.VerticalRelations, max(s.Size-1, 0), s.Size),
	}

	initial := g.clone()
	for y := range s.Size {
		for x := range s.Size {
			if !provided[y][x] {
				initial[y][x] = emptyCell
			}
		}
	}

	m := &Model{
		size:        s.Size,
		grid:        g,
		initialGrid: initial,
		provided:    provided,
		relations:   rels,
		cursor:      game.Cursor{X: 0, Y: 0},
		keys:        DefaultKeyMap,
		modeTitle:   s.ModeTitle,
	}
	m.solved = m.checkSolved()
	return m, nil
}

func serializeProvided(p [][]bool) string {
	var b strings.Builder
	for y, row := range p {
		for _, v := range row {
			if v {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		if y < len(p)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func deserializeProvided(s string, size int) [][]bool {
	rows := strings.Split(s, "\n")
	p := make([][]bool, size)
	for y := range size {
		p[y] = make([]bool, size)
		if y >= len(rows) {
			continue
		}
		for x := range size {
			if x < len(rows[y]) && rows[y][x] == '#' {
				p[y][x] = true
			}
		}
	}
	return p
}
