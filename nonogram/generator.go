package nonogram

import (
	"context"
	"math/rand/v2"
	"strings"
	"time"
)

const maxAttempts = 100

func GenerateRandomTomography(mode NonogramMode) Hints {
	maxDim := max(mode.Width, mode.Height)
	timeout := time.Duration(maxDim) * time.Second

	for range maxAttempts {
		s := generateRandomState(mode.Height, mode.Width, mode.Density)
		g := newGrid(s)
		hints := generateTomography(g)
		if !isValidPuzzle(hints, mode.Height, mode.Width) {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		count := countSolutions(hints, mode.Width, mode.Height, 2, ctx)
		cancel()
		if count == 1 {
			return hints
		}
	}
	return Hints{}
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

type cellState int

const (
	cellUnknown cellState = iota
	cellEmpty
	cellFilled
)

type solutionGrid struct {
	cells [][]cellState
	w, h  int
}

func newSolutionGrid(w, h int) solutionGrid {
	cells := make([][]cellState, h)
	for y := range h {
		cells[y] = make([]cellState, w)
		for x := range w {
			cells[y][x] = cellUnknown
		}
	}
	return solutionGrid{cells: cells, w: w, h: h}
}

func (g *solutionGrid) clone() solutionGrid {
	cells := make([][]cellState, g.h)
	for y := range g.h {
		cells[y] = make([]cellState, g.w)
		copy(cells[y], g.cells[y])
	}
	return solutionGrid{cells: cells, w: g.w, h: g.h}
}

func (g *solutionGrid) set(x, y int, state cellState) bool {
	if g.cells[y][x] != cellUnknown && g.cells[y][x] != state {
		return false
	}
	g.cells[y][x] = state
	return true
}

func (g *solutionGrid) isComplete() bool {
	for y := range g.h {
		for x := range g.w {
			if g.cells[y][x] == cellUnknown {
				return false
			}
		}
	}
	return true
}

func countSolutions(hints Hints, w, h, limit int, ctx context.Context) int {
	g := newSolutionGrid(w, h)
	return countSolutionsRecursive(g, hints, limit, ctx)
}

func countSolutionsRecursive(g solutionGrid, hints Hints, limit int, ctx context.Context) int {
	select {
	case <-ctx.Done():
		return -1
	default:
	}

	propagated, valid := propagate(g, hints)
	if !valid {
		return 0
	}

	if propagated.isComplete() {
		return 1
	}

	for y := range propagated.h {
		for x := range propagated.w {
			if propagated.cells[y][x] == cellUnknown {
				count := 0

				g1 := propagated.clone()
				if g1.set(x, y, cellFilled) {
					n := countSolutionsRecursive(g1, hints, limit-count, ctx)
					if n < 0 {
						return n
					}
					count += n
					if count >= limit {
						return count
					}
				}

				g2 := propagated.clone()
				if g2.set(x, y, cellEmpty) {
					n := countSolutionsRecursive(g2, hints, limit-count, ctx)
					if n < 0 {
						return n
					}
					count += n
					if count >= limit {
						return count
					}
				}

				return count
			}
		}
	}

	return 0
}

func propagate(g solutionGrid, hints Hints) (solutionGrid, bool) {
	changed := true
	for changed {
		changed = false

		for y := range g.h {
			row := make([]cellState, g.w)
			for x := range g.w {
				row[x] = g.cells[y][x]
			}

			newRow, valid := propagateLine(row, hints.rows[y])
			if !valid {
				return g, false
			}

			for x := range g.w {
				if g.cells[y][x] == cellUnknown && newRow[x] != cellUnknown {
					g.cells[y][x] = newRow[x]
					changed = true
				}
			}
		}

		for x := range g.w {
			col := make([]cellState, g.h)
			for y := range g.h {
				col[y] = g.cells[y][x]
			}

			newCol, valid := propagateLine(col, hints.cols[x])
			if !valid {
				return g, false
			}

			for y := range g.h {
				if g.cells[y][x] == cellUnknown && newCol[y] != cellUnknown {
					g.cells[y][x] = newCol[y]
					changed = true
				}
			}
		}
	}

	return g, true
}

func propagateLine(line []cellState, hint []int) ([]cellState, bool) {
	n := len(line)
	possibilities := generateLinePossibilities(n, hint)

	filtered := make([][]cellState, 0, len(possibilities))
	for _, poss := range possibilities {
		if matchesLine(poss, line) {
			filtered = append(filtered, poss)
		}
	}

	if len(filtered) == 0 {
		return nil, false
	}

	result := make([]cellState, n)
	for i := range n {
		allFilled := true
		allEmpty := true
		for _, poss := range filtered {
			if poss[i] != cellFilled {
				allFilled = false
			}
			if poss[i] != cellEmpty {
				allEmpty = false
			}
		}
		if allFilled {
			result[i] = cellFilled
		} else if allEmpty {
			result[i] = cellEmpty
		} else {
			result[i] = cellUnknown
		}
	}

	return result, true
}

func generateLinePossibilities(length int, hint []int) [][]cellState {
	if len(hint) == 0 || (len(hint) == 1 && hint[0] == 0) {
		result := make([]cellState, length)
		for i := range result {
			result[i] = cellEmpty
		}
		return [][]cellState{result}
	}

	totalFilled := 0
	for _, h := range hint {
		totalFilled += h
	}
	minLength := totalFilled + len(hint) - 1
	if minLength > length {
		return nil
	}

	return generateLinePossibilitiesRecursive(length, hint, 0, 0, make([]cellState, 0, length))
}

func generateLinePossibilitiesRecursive(length int, hint []int, hintIdx, pos int, current []cellState) [][]cellState {
	if hintIdx == len(hint) {
		for len(current) < length {
			current = append(current, cellEmpty)
		}
		return [][]cellState{append([]cellState(nil), current...)}
	}

	blockLen := hint[hintIdx]
	remainingHints := hint[hintIdx+1:]
	remainingFilled := 0
	for _, h := range remainingHints {
		remainingFilled += h
	}
	minRemaining := remainingFilled + len(remainingHints)

	maxStart := length - blockLen - minRemaining

	var results [][]cellState

	for start := pos; start <= maxStart; start++ {
		newCurrent := append([]cellState(nil), current...)

		for i := len(current); i < start; i++ {
			newCurrent = append(newCurrent, cellEmpty)
		}

		for i := 0; i < blockLen; i++ {
			newCurrent = append(newCurrent, cellFilled)
		}

		if hintIdx < len(hint)-1 {
			newCurrent = append(newCurrent, cellEmpty)
		}
		results = append(results, generateLinePossibilitiesRecursive(length, hint, hintIdx+1, len(newCurrent), newCurrent)...)
	}

	return results
}

func matchesLine(possibility, line []cellState) bool {
	if len(possibility) != len(line) {
		return false
	}
	for i, p := range possibility {
		if line[i] != cellUnknown && line[i] != p {
			return false
		}
	}
	return true
}
