package shikaku

import (
	"errors"
	"math/rand/v2"
)

const maxGenerateAttempts = 100

// genRect is used during generation to track grid partitions.
type genRect struct {
	x, y, w, h int
}

type dim struct {
	w, h int
}

type templateCacheKey struct {
	x, y    int
	maxW    int
	maxH    int
	maxArea int
}

type failedStateKey struct {
	x, y      int
	maxW      int
	maxH      int
	remaining int
}

type partitionSearchState struct {
	templateCache map[templateCacheKey][]dim
	failedStates  map[failedStateKey]bool
}

// GeneratePuzzle creates a Shikaku puzzle with random RNG.
func GeneratePuzzle(width, height, maxRectSize int) (Puzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GeneratePuzzleSeeded(width, height, maxRectSize, rng)
}

// GeneratePuzzleSeeded creates a Shikaku puzzle with a deterministic RNG.
func GeneratePuzzleSeeded(width, height, maxRectSize int, rng *rand.Rand) (Puzzle, error) {
	for range maxGenerateAttempts {
		p := tryGenerate(width, height, maxRectSize, rng)
		if p != nil {
			return *p, nil
		}
	}
	return Puzzle{}, errors.New("failed to generate shikaku puzzle after maximum attempts")
}

func tryGenerate(width, height, maxRectSize int, rng *rand.Rand) *Puzzle {
	// Grid tracks which partition each cell belongs to (-1 = uncovered).
	grid := make([][]int, height)
	for y := range height {
		grid[y] = make([]int, width)
		for x := range width {
			grid[y][x] = -1
		}
	}

	partitions := make([]genRect, 0, width*height/2)
	if !partitionGrid(grid, width, height, maxRectSize, &partitions, rng) {
		return nil
	}

	// Place clues: pick a random cell within each partition.
	clues := make([]Clue, len(partitions))
	allSingles := true
	for i, r := range partitions {
		cx := r.x + rng.IntN(r.w)
		cy := r.y + rng.IntN(r.h)
		area := r.w * r.h
		if area > 1 {
			allSingles = false
		}
		clues[i] = Clue{
			ID:    i,
			X:     cx,
			Y:     cy,
			Value: area,
		}
	}
	if allSingles {
		return nil
	}

	return &Puzzle{
		Width:  width,
		Height: height,
		Clues:  clues,
	}
}

func isAnchorCell(grid [][]int, x, y int) bool {
	if grid[y][x] != -1 {
		return false
	}
	if x > 0 && grid[y][x-1] == -1 {
		return false
	}
	if y > 0 && grid[y-1][x] == -1 {
		return false
	}
	return true
}

func anchorExtents(grid [][]int, x, y, width, height int) (maxW, maxH int) {
	maxW = 0
	for cx := x; cx < width && grid[y][cx] == -1; cx++ {
		maxW++
	}
	maxH = 0
	for cy := y; cy < height && grid[cy][x] == -1; cy++ {
		maxH++
	}
	return maxW, maxH
}

func candidateTemplatesForAnchor(
	state *partitionSearchState,
	x,
	y,
	maxW,
	maxH,
	maxRectSize int,
) []dim {
	key := templateCacheKey{
		x:       x,
		y:       y,
		maxW:    maxW,
		maxH:    maxH,
		maxArea: maxRectSize,
	}
	if templates, ok := state.templateCache[key]; ok {
		return templates
	}

	templates := make([]dim, 0, maxW*maxH)
	for w := 1; w <= maxW; w++ {
		for h := 1; h <= maxH; h++ {
			if w*h > maxRectSize {
				continue
			}
			templates = append(templates, dim{w: w, h: h})
		}
	}

	state.templateCache[key] = templates
	return templates
}

func selectMostConstrainedAnchor(
	grid [][]int,
	width,
	height,
	maxRectSize int,
	state *partitionSearchState,
) (x, y, maxW, maxH int, candidates []dim, found bool) {
	bestCount := width*height + 1
	bestX, bestY := -1, -1
	bestW, bestH := 0, 0

	for y := range height {
		for x := range width {
			if !isAnchorCell(grid, x, y) {
				continue
			}

			curW, curH := anchorExtents(grid, x, y, width, height)
			templates := candidateTemplatesForAnchor(state, x, y, curW, curH, maxRectSize)
			valid := make([]dim, 0, len(templates))
			for _, c := range templates {
				if fits(grid, x, y, c.w, c.h) {
					valid = append(valid, c)
				}
			}

			if len(valid) == 0 {
				return x, y, curW, curH, nil, true
			}
			if len(valid) < bestCount {
				bestCount = len(valid)
				bestX, bestY = x, y
				bestW, bestH = curW, curH
				candidates = valid
				found = true
			}
			if bestCount == 1 {
				return bestX, bestY, bestW, bestH, candidates, true
			}
		}
	}

	if !found {
		return 0, 0, 0, 0, nil, false
	}
	return bestX, bestY, bestW, bestH, candidates, true
}

// partitionGrid recursively fills the grid with non-overlapping rectangles.
func partitionGrid(grid [][]int, width, height, maxRectSize int, partitions *[]genRect, rng *rand.Rand) bool {
	state := partitionSearchState{
		templateCache: make(map[templateCacheKey][]dim, width*height),
		failedStates:  make(map[failedStateKey]bool, width*height),
	}
	return partitionGridRec(grid, width, height, maxRectSize, partitions, rng, &state, width*height)
}

func partitionGridRec(
	grid [][]int,
	width,
	height,
	maxRectSize int,
	partitions *[]genRect,
	rng *rand.Rand,
	state *partitionSearchState,
	remaining int,
) bool {
	x, y, maxW, maxH, candidates, found := selectMostConstrainedAnchor(
		grid,
		width,
		height,
		maxRectSize,
		state,
	)
	if !found {
		return true // all cells covered
	}

	failKey := failedStateKey{x: x, y: y, maxW: maxW, maxH: maxH, remaining: remaining}
	if state.failedStates[failKey] {
		return false
	}
	if len(candidates) == 0 {
		state.failedStates[failKey] = true
		return false
	}

	order := make([]dim, len(candidates))
	copy(order, candidates)
	rng.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	id := len(*partitions)
	for _, c := range order {
		// Place rectangle.
		place(grid, x, y, c.w, c.h, id)
		*partitions = append(*partitions, genRect{x, y, c.w, c.h})

		if partitionGridRec(
			grid,
			width,
			height,
			maxRectSize,
			partitions,
			rng,
			state,
			remaining-(c.w*c.h),
		) {
			return true
		}

		// Backtrack.
		*partitions = (*partitions)[:len(*partitions)-1]
		clearGrid(grid, x, y, c.w, c.h)
	}

	state.failedStates[failKey] = true
	return false
}

// fits checks if a rectangle of size w*h can be placed at (x, y) without overlap.
func fits(grid [][]int, x, y, w, h int) bool {
	for dy := range h {
		for dx := range w {
			if grid[y+dy][x+dx] != -1 {
				return false
			}
		}
	}
	return true
}

// place marks cells in the grid with the given partition ID.
func place(grid [][]int, x, y, w, h, id int) {
	for dy := range h {
		for dx := range w {
			grid[y+dy][x+dx] = id
		}
	}
}

// clearGrid resets cells in the grid back to uncovered.
func clearGrid(grid [][]int, x, y, w, h int) {
	for dy := range h {
		for dx := range w {
			grid[y+dy][x+dx] = -1
		}
	}
}
