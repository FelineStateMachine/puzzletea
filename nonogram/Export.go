package nonogram

import (
	"encoding/json"
	"fmt"
	"math"
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
	curr := generateTomography(m.grid)
	solved := curr.rows.equal(m.rowHints) && curr.cols.equal(m.colHints)
	save := Save{
		RowHints: m.rowHints,
		ColHints: m.colHints,
		Solved:   solved,
		State:    m.grid.String(),
		Width:    m.width,
		Height:   m.height,
	}
	jsonData, err := json.Marshal(save)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal save data:\n%v", save)
	}
	return jsonData, nil
}

func (m Model) GetTomography() Hints {
	return generateTomography(m.grid)
}

func GenerateRandomTomography(mode NonogramMode) Hints {
	return generateTomography(newGrid(generateRandomState(mode.Height, mode.Width)))
}

func generateRandomState(h, w int) state {
	if h <= 0 || w <= 0 {
		return ""
	}

	var b strings.Builder
	b.Grow((w + 1) * h)

	for i := range h {
		rowWeight := float32(math.Pow(rand.Float64(), 3))
		for range w {
			filled := rand.Float32() > rowWeight // Percentage chance to fill.
			if filled {
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
