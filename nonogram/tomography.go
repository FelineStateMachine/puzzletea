package nonogram

type TomographyDefinition [][]int

type Hints struct {
	rows, cols TomographyDefinition
}

func generateTomography(g grid) Hints {
	height := len(g)
	width := 0
	if height > 0 {
		for _, row := range g {
			if len(row) > width {
				width = len(row)
			}
		}
	}

	rowHints := make([][]int, height)
	for y, row := range g {
		count := 0
		var rowHint []int
		for _, tile := range row {
			if tile == filledTile {
				count++
			} else {
				if count > 0 {
					rowHint = append(rowHint, count)
				}
				count = 0
			}
		}
		if count > 0 {
			rowHint = append(rowHint, count)
		}
		if len(rowHint) == 0 {
			rowHint = append(rowHint, 0)
		}
		rowHints[y] = rowHint
	}

	colHints := make([][]int, width)
	for x := range width {
		count := 0
		var colHint []int
		for y := range height {
			r := emptyTile
			if x < len(g[y]) {
				r = g[y][x]
			}

			if r == filledTile {
				count++
			} else {
				if count > 0 {
					colHint = append(colHint, count)
				}
				count = 0
			}
		}
		if count > 0 {
			colHint = append(colHint, count)
		}
		if len(colHint) == 0 {
			colHint = append(colHint, 0)
		}
		colHints[x] = colHint
	}

	return Hints{
		rowHints,
		colHints,
	}
}

func (t TomographyDefinition) RequiredLen() int {
	var maxLen int
	for _, r := range t {
		if len(r) > maxLen {
			maxLen = len(r)
		}
	}
	return maxLen
}

func (t TomographyDefinition) equal(o TomographyDefinition) bool {
	if len(t) != len(o) {
		return false
	}
	for i, r := range t {
		if len(r) != len(o[i]) {
			return false
		}
		for j, v := range r {
			if v != o[i][j] {
				return false
			}
		}
	}
	return true
}
