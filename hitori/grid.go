package hitori

type (
	grid  [][]rune
	state string
)

const (
	emptyCell  = ' '
	shadedCell = '#'
)

func newGrid(s state) grid {
	if s == "" {
		return nil
	}
	rows := [][]rune{}
	current := []rune{}
	for _, r := range s {
		if r == '\n' {
			if len(current) > 0 {
				rows = append(rows, current)
				current = []rune{}
			}
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		rows = append(rows, current)
	}
	return rows
}

func (g grid) String() string {
	var result []rune
	for y, row := range g {
		result = append(result, row...)
		if y < len(g)-1 {
			result = append(result, '\n')
		}
	}
	return string(result)
}

func createEmptyState(size int) state {
	lines := make([]rune, 0, size*(size+1))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			lines = append(lines, emptyCell)
		}
		if y < size-1 {
			lines = append(lines, '\n')
		}
	}
	return state(lines)
}

func (g grid) clone() grid {
	result := make(grid, len(g))
	for y := range g {
		result[y] = make([]rune, len(g[y]))
		copy(result[y], g[y])
	}
	return result
}
