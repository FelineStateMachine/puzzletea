package nonogram

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
)

type Save struct {
	Solved   bool                 `json:"solved"`
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
		Solved:   m.solved,
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

func (m Model) GetTomography() Hints {
	return generateTomography(m.grid)
}

func GenerateRandomTomography(mode NonogramMode) Hints {
	for {
		s := generateRandomState(mode.Height, mode.Width, mode.Density)
		g := newGrid(s)
		hints := generateTomography(g)
		if isValidPuzzle(hints, mode.Height, mode.Width) {
			return hints
		}
	}
}

func generateRandomState(h, w int, density float64) state {
	if h <= 0 || w <= 0 {
		return ""
	}

	density = max(0.1, min(0.9, density))

	var b strings.Builder
	b.Grow((w + 1) * h)

	for i := range h {
		rowDensity := density + (rand.Float64()-0.5)*0.3
		rowDensity = max(0.05, min(0.95, rowDensity))
		for range w {
			if rand.Float64() < rowDensity {
				b.WriteRune(filledTile)
			} else {
				b.WriteRune(emptyTile)
			}
		}
		if i < h-1 {
			b.WriteRune('\n')
		}
	}

	return state(b.String())
}

func isValidPuzzle(hints Hints, height, width int) bool {
	for _, rh := range hints.rows {
		if len(rh) == 1 && (rh[0] == 0 || rh[0] == width) {
			return false
		}
	}
	for _, ch := range hints.cols {
		if len(ch) == 1 && (ch[0] == 0 || ch[0] == height) {
			return false
		}
	}
	return true
}
