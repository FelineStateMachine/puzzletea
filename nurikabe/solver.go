package nurikabe

import (
	"context"
	"errors"
	"fmt"
)

var errNodeLimit = errors.New("solver node limit exceeded")

type SolveStats struct {
	Nodes    int
	Branches int
	MaxDepth int
}

type point struct {
	x, y int
}

type islandComponent struct {
	cells     []point
	clueCount int
	clueValue int
	cluePos   point
}

var dirs = [4]point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

func CountSolutions(puzzle Puzzle, limit, nodeLimit int) (int, SolveStats, error) {
	return CountSolutionsContext(context.Background(), puzzle, limit, nodeLimit)
}

func CountSolutionsContext(ctx context.Context, puzzle Puzzle, limit, nodeLimit int) (int, SolveStats, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 {
		return 0, SolveStats{}, nil
	}
	if nodeLimit <= 0 {
		nodeLimit = 250000
	}
	if err := validateClues(puzzle.Clues, puzzle.Width, puzzle.Height); err != nil {
		return 0, SolveStats{}, err
	}

	state := newGrid(puzzle.Width, puzzle.Height, unknownCell)
	for y := range puzzle.Height {
		for x := range puzzle.Width {
			if puzzle.Clues[y][x] > 0 {
				state[y][x] = islandCell
			}
		}
	}

	stats := SolveStats{}
	solutions := 0
	var dfs func(grid, int) error
	dfs = func(g grid, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if stats.Nodes >= nodeLimit {
			return errNodeLimit
		}
		stats.Nodes++
		if depth > stats.MaxDepth {
			stats.MaxDepth = depth
		}

		if err := propagate(g, puzzle.Clues); err != nil {
			return nil
		}
		if err := validatePartial(g, puzzle.Clues); err != nil {
			return nil
		}

		x, y, ok := pickUnknown(g, puzzle.Clues)
		if !ok {
			if isSolvedGrid(g, puzzle.Clues) {
				solutions++
			}
			return nil
		}

		order := candidateOrder(g, puzzle.Clues, x, y)
		stats.Branches++
		for _, candidate := range order {
			if err := ctx.Err(); err != nil {
				return err
			}
			next := g.clone()
			next[y][x] = candidate
			if err := dfs(next, depth+1); err != nil {
				if errors.Is(err, errNodeLimit) ||
					errors.Is(err, context.Canceled) ||
					errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				continue
			}
			if solutions >= limit {
				return nil
			}
		}

		return nil
	}

	err := dfs(state, 0)
	if errors.Is(err, errNodeLimit) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return solutions, stats, err
	}
	return solutions, stats, nil
}

func propagate(g grid, clues clueGrid) error {
	height := len(g)
	if height == 0 {
		return nil
	}
	width := len(g[0])

	for {
		changed := false
		comps, compIdx := islandComponents(g, clues)

		for _, comp := range comps {
			if comp.clueCount == 1 && len(comp.cells) == comp.clueValue {
				for _, c := range comp.cells {
					for _, d := range dirs {
						nx, ny := c.x+d.x, c.y+d.y
						if nx < 0 || nx >= width || ny < 0 || ny >= height {
							continue
						}
						if g[ny][nx] == unknownCell {
							g[ny][nx] = seaCell
							changed = true
						}
					}
				}
			}
		}

		for y := range height {
			for x := range width {
				if g[y][x] != unknownCell {
					continue
				}
				seen := map[point]bool{}
				for _, d := range dirs {
					nx, ny := x+d.x, y+d.y
					if nx < 0 || nx >= width || ny < 0 || ny >= height {
						continue
					}
					if g[ny][nx] != islandCell {
						continue
					}
					idx := compIdx[ny][nx]
					if idx < 0 || idx >= len(comps) {
						continue
					}
					comp := comps[idx]
					if comp.clueCount == 1 {
						seen[comp.cluePos] = true
					}
				}
				if len(seen) >= 2 {
					g[y][x] = seaCell
					changed = true
				}
			}
		}

		if err := validatePartial(g, clues); err != nil {
			return err
		}
		if !changed {
			return nil
		}
	}
}

func validatePartial(g grid, clues clueGrid) error {
	if hasSeaSquare(g) {
		return fmt.Errorf("contains 2x2 sea block")
	}
	if !seaCanRemainConnected(g) {
		return fmt.Errorf("sea components cannot be connected")
	}

	comps, _ := islandComponents(g, clues)
	for _, comp := range comps {
		size := len(comp.cells)
		switch {
		case comp.clueCount > 1:
			return fmt.Errorf("island has multiple clues")
		case comp.clueCount == 1 && size > comp.clueValue:
			return fmt.Errorf("island exceeds clue size")
		case comp.clueCount == 0 && !componentCanReachClue(g, clues, comp):
			return fmt.Errorf("orphan island component")
		case comp.clueCount == 1 && maxReachableForComponent(g, clues, comp) < comp.clueValue:
			return fmt.Errorf("clue cannot reach required size")
		}
	}

	return nil
}

func isSolvedGrid(g grid, clues clueGrid) bool {
	if len(g) == 0 || len(g[0]) == 0 {
		return false
	}

	height := len(g)
	width := len(g[0])
	for y := range height {
		for x := range width {
			if clues[y][x] > 0 && g[y][x] != islandCell {
				return false
			}
			if g[y][x] == unknownCell {
				return false
			}
		}
	}

	if hasSeaSquare(g) {
		return false
	}
	if !isSeaConnected(g) {
		return false
	}

	comps, _ := islandComponents(g, clues)
	seenClues := 0
	for _, comp := range comps {
		if comp.clueCount != 1 {
			return false
		}
		if len(comp.cells) != comp.clueValue {
			return false
		}
		seenClues++
	}

	return seenClues == countClues(clues)
}

func pickUnknown(g grid, clues clueGrid) (x, y int, ok bool) {
	bestScore := -1
	for yy, row := range g {
		for xx := range row {
			if g[yy][xx] != unknownCell {
				continue
			}
			if clues[yy][xx] > 0 {
				continue
			}
			score := 0
			for _, d := range dirs {
				nx, ny := xx+d.x, yy+d.y
				if nx < 0 || ny < 0 || ny >= len(g) || nx >= len(row) {
					continue
				}
				switch g[ny][nx] {
				case islandCell:
					score += 3
				case seaCell:
					score += 1
				}
				if clues[ny][nx] > 0 {
					score += 4
				}
			}
			if score > bestScore {
				bestScore = score
				x, y = xx, yy
				ok = true
			}
		}
	}
	return x, y, ok
}

func candidateOrder(g grid, clues clueGrid, x, y int) []cellState {
	islandFirst := false
	for _, d := range dirs {
		nx, ny := x+d.x, y+d.y
		if nx < 0 || ny < 0 || ny >= len(g) || nx >= len(g[0]) {
			continue
		}
		if g[ny][nx] == islandCell || clues[ny][nx] > 0 {
			islandFirst = true
			break
		}
	}

	if islandFirst {
		return []cellState{islandCell, seaCell}
	}
	return []cellState{seaCell, islandCell}
}

func seaCanRemainConnected(g grid) bool {
	height := len(g)
	if height == 0 {
		return true
	}
	width := len(g[0])

	start := point{-1, -1}
	seaCount := 0
	for y := range height {
		for x := range width {
			if g[y][x] == seaCell {
				seaCount++
				if start.x == -1 {
					start = point{x, y}
				}
			}
		}
	}
	if seaCount <= 1 {
		return true
	}

	visited := make([][]bool, height)
	for y := range height {
		visited[y] = make([]bool, width)
	}
	queue := []point{start}
	visited[start.y][start.x] = true

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := p.x+d.x, p.y+d.y
			if nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			if visited[ny][nx] || g[ny][nx] == islandCell {
				continue
			}
			visited[ny][nx] = true
			queue = append(queue, point{nx, ny})
		}
	}

	for y := range height {
		for x := range width {
			if g[y][x] == seaCell && !visited[y][x] {
				return false
			}
		}
	}
	return true
}

func isSeaConnected(g grid) bool {
	height := len(g)
	if height == 0 {
		return false
	}
	width := len(g[0])

	start := point{-1, -1}
	seaCount := 0
	for y := range height {
		for x := range width {
			if g[y][x] == seaCell {
				seaCount++
				if start.x == -1 {
					start = point{x, y}
				}
			}
		}
	}
	if seaCount == 0 {
		return false
	}

	visited := make([][]bool, height)
	for y := range height {
		visited[y] = make([]bool, width)
	}

	queue := []point{start}
	visited[start.y][start.x] = true
	seen := 1

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := p.x+d.x, p.y+d.y
			if nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			if visited[ny][nx] || g[ny][nx] != seaCell {
				continue
			}
			visited[ny][nx] = true
			seen++
			queue = append(queue, point{nx, ny})
		}
	}

	return seen == seaCount
}

func hasSeaSquare(g grid) bool {
	height := len(g)
	if height == 0 {
		return false
	}
	width := len(g[0])
	for y := 0; y < height-1; y++ {
		for x := 0; x < width-1; x++ {
			if g[y][x] == seaCell && g[y][x+1] == seaCell && g[y+1][x] == seaCell && g[y+1][x+1] == seaCell {
				return true
			}
		}
	}
	return false
}

func islandComponents(g grid, clues clueGrid) ([]islandComponent, [][]int) {
	height := len(g)
	width := len(g[0])
	idx := make([][]int, height)
	for y := range height {
		idx[y] = make([]int, width)
		for x := range width {
			idx[y][x] = -1
		}
	}

	components := make([]islandComponent, 0)
	for y := range height {
		for x := range width {
			if g[y][x] != islandCell || idx[y][x] >= 0 {
				continue
			}

			compIndex := len(components)
			comp := islandComponent{cluePos: point{-1, -1}}
			queue := []point{{x, y}}
			idx[y][x] = compIndex

			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				comp.cells = append(comp.cells, p)
				if clues[p.y][p.x] > 0 {
					comp.clueCount++
					comp.clueValue = clues[p.y][p.x]
					comp.cluePos = p
				}

				for _, d := range dirs {
					nx, ny := p.x+d.x, p.y+d.y
					if nx < 0 || ny < 0 || nx >= width || ny >= height {
						continue
					}
					if g[ny][nx] != islandCell || idx[ny][nx] >= 0 {
						continue
					}
					idx[ny][nx] = compIndex
					queue = append(queue, point{nx, ny})
				}
			}

			components = append(components, comp)
		}
	}

	return components, idx
}

func componentCanReachClue(g grid, clues clueGrid, comp islandComponent) bool {
	height := len(g)
	width := len(g[0])
	seen := make([][]bool, height)
	for y := range height {
		seen[y] = make([]bool, width)
	}

	queue := make([]point, len(comp.cells))
	copy(queue, comp.cells)
	for _, c := range comp.cells {
		seen[c.y][c.x] = true
	}

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if clues[p.y][p.x] > 0 {
			return true
		}
		for _, d := range dirs {
			nx, ny := p.x+d.x, p.y+d.y
			if nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			if seen[ny][nx] || g[ny][nx] == seaCell {
				continue
			}
			seen[ny][nx] = true
			queue = append(queue, point{nx, ny})
		}
	}

	return false
}

func maxReachableForComponent(g grid, clues clueGrid, comp islandComponent) int {
	if comp.clueCount != 1 {
		return 0
	}

	height := len(g)
	width := len(g[0])
	seen := make([][]bool, height)
	for y := range height {
		seen[y] = make([]bool, width)
	}

	queue := make([]point, len(comp.cells))
	copy(queue, comp.cells)
	for _, c := range comp.cells {
		seen[c.y][c.x] = true
	}

	count := len(comp.cells)
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := p.x+d.x, p.y+d.y
			if nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			if seen[ny][nx] || g[ny][nx] == seaCell {
				continue
			}
			if clues[ny][nx] > 0 && (nx != comp.cluePos.x || ny != comp.cluePos.y) {
				continue
			}
			seen[ny][nx] = true
			count++
			queue = append(queue, point{nx, ny})
		}
	}

	return count
}

func hasAnyUnknown(g grid) bool {
	for _, row := range g {
		for _, c := range row {
			if c == unknownCell {
				return true
			}
		}
	}
	return false
}

func computeConflicts(marks grid, clues clueGrid) [][]bool {
	height := len(marks)
	width := len(marks[0])
	conflicts := make([][]bool, height)
	for y := range height {
		conflicts[y] = make([]bool, width)
	}

	for y := 0; y < height-1; y++ {
		for x := 0; x < width-1; x++ {
			if marks[y][x] == seaCell && marks[y][x+1] == seaCell && marks[y+1][x] == seaCell && marks[y+1][x+1] == seaCell {
				conflicts[y][x] = true
				conflicts[y][x+1] = true
				conflicts[y+1][x] = true
				conflicts[y+1][x+1] = true
			}
		}
	}

	comps, _ := islandComponents(marks, clues)
	for _, comp := range comps {
		bad := false
		size := len(comp.cells)
		if comp.clueCount > 1 {
			bad = true
		}
		if comp.clueCount == 1 && size > comp.clueValue {
			bad = true
		}
		if comp.clueCount == 0 && !componentCanReachClue(marks, clues, comp) {
			bad = true
		}
		if !hasAnyUnknown(marks) && comp.clueCount == 1 && size != comp.clueValue {
			bad = true
		}
		if bad {
			for _, c := range comp.cells {
				conflicts[c.y][c.x] = true
			}
		}
	}

	if !hasAnyUnknown(marks) {
		markDisconnectedSeaConflicts(marks, conflicts)
	}

	return conflicts
}

func markDisconnectedSeaConflicts(marks grid, conflicts [][]bool) {
	height := len(marks)
	width := len(marks[0])
	visited := make([][]bool, height)
	for y := range height {
		visited[y] = make([]bool, width)
	}

	var components [][]point
	for y := range height {
		for x := range width {
			if marks[y][x] != seaCell || visited[y][x] {
				continue
			}
			queue := []point{{x, y}}
			visited[y][x] = true
			component := []point{}
			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				component = append(component, p)
				for _, d := range dirs {
					nx, ny := p.x+d.x, p.y+d.y
					if nx < 0 || ny < 0 || nx >= width || ny >= height {
						continue
					}
					if visited[ny][nx] || marks[ny][nx] != seaCell {
						continue
					}
					visited[ny][nx] = true
					queue = append(queue, point{nx, ny})
				}
			}
			components = append(components, component)
		}
	}

	if len(components) <= 1 {
		return
	}

	largestIdx := 0
	largestSize := len(components[0])
	for i := 1; i < len(components); i++ {
		if len(components[i]) > largestSize {
			largestSize = len(components[i])
			largestIdx = i
		}
	}

	for i, component := range components {
		if i == largestIdx {
			continue
		}
		for _, c := range component {
			conflicts[c.y][c.x] = true
		}
	}
}
