package nurikabe

import (
	"fmt"
	"strconv"
	"strings"
)

type cellState rune

const (
	unknownCell        cellState = '?'
	seaCell            cellState = '~'
	islandCell         cellState = 'o'
	legacyRequiredLand cellState = '!'
)

type (
	grid     [][]cellState
	clueGrid [][]int
)

type Puzzle struct {
	Width, Height int
	Clues         clueGrid
}

func newGrid(width, height int, fill cellState) grid {
	g := make(grid, height)
	for y := range height {
		g[y] = make([]cellState, width)
		for x := range width {
			g[y][x] = fill
		}
	}
	return g
}

func (g grid) clone() grid {
	c := make(grid, len(g))
	for y, row := range g {
		c[y] = make([]cellState, len(row))
		copy(c[y], row)
	}
	return c
}

func (g grid) String() string {
	rows := make([]string, len(g))
	for y, row := range g {
		var b strings.Builder
		b.Grow(len(row))
		for _, c := range row {
			b.WriteRune(rune(c))
		}
		rows[y] = b.String()
	}
	return strings.Join(rows, "\n")
}

func parseGrid(s string, width, height int) (grid, error) {
	rows := strings.Split(s, "\n")
	if len(rows) != height {
		return nil, fmt.Errorf("grid row count mismatch: got %d, want %d", len(rows), height)
	}

	g := make(grid, height)
	for y := range height {
		runes := []rune(rows[y])
		if len(runes) != width {
			return nil, fmt.Errorf("grid width mismatch at row %d: got %d, want %d", y, len(runes), width)
		}
		g[y] = make([]cellState, width)
		for x, r := range runes {
			switch cellState(r) {
			case unknownCell, seaCell, islandCell:
				g[y][x] = cellState(r)
			case legacyRequiredLand:
				// Backward compatibility: old required marks now map to island.
				g[y][x] = islandCell
			default:
				return nil, fmt.Errorf("invalid grid symbol %q at (%d,%d)", r, x, y)
			}
		}
	}

	return g, nil
}

func serializeClues(clues clueGrid) string {
	rows := make([]string, len(clues))
	for y, row := range clues {
		parts := make([]string, len(row))
		for x, v := range row {
			if v < 0 {
				v = 0
			}
			parts[x] = strconv.Itoa(v)
		}
		rows[y] = strings.Join(parts, ",")
	}
	return strings.Join(rows, "\n")
}

func parseClues(s string, width, height int) (clueGrid, error) {
	rows := strings.Split(strings.TrimSpace(s), "\n")
	if height == 0 || width == 0 {
		return nil, fmt.Errorf("invalid clue dimensions: %dx%d", width, height)
	}
	if len(rows) != height {
		return nil, fmt.Errorf("clue row count mismatch: got %d, want %d", len(rows), height)
	}

	clues := make(clueGrid, height)
	for y := range height {
		parts := strings.Split(rows[y], ",")
		if len(parts) != width {
			return nil, fmt.Errorf("clue width mismatch at row %d: got %d, want %d", y, len(parts), width)
		}
		clues[y] = make([]int, width)
		for x, p := range parts {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				return nil, fmt.Errorf("invalid clue value %q at (%d,%d): %w", p, x, y, err)
			}
			if v < 0 {
				return nil, fmt.Errorf("negative clue value %d at (%d,%d)", v, x, y)
			}
			clues[y][x] = v
		}
	}

	return clues, nil
}

func cloneClues(c clueGrid) clueGrid {
	out := make(clueGrid, len(c))
	for y, row := range c {
		out[y] = make([]int, len(row))
		copy(out[y], row)
	}
	return out
}

func validateClues(clues clueGrid, width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d", width, height)
	}
	if len(clues) != height {
		return fmt.Errorf("clue row count mismatch: got %d, want %d", len(clues), height)
	}
	for y, row := range clues {
		if len(row) != width {
			return fmt.Errorf("clue width mismatch at row %d: got %d, want %d", y, len(row), width)
		}
		for x, v := range row {
			if v < 0 {
				return fmt.Errorf("negative clue at (%d,%d)", x, y)
			}
		}
	}
	return nil
}

func isClueCell(clues clueGrid, x, y int) bool {
	return y >= 0 && y < len(clues) && x >= 0 && x < len(clues[y]) && clues[y][x] > 0
}

func countClues(clues clueGrid) int {
	total := 0
	for _, row := range clues {
		for _, v := range row {
			if v > 0 {
				total++
			}
		}
	}
	return total
}
