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

	var partitions []genRect

	if !partitionGrid(grid, width, height, maxRectSize, &partitions, rng) {
		return nil
	}

	// Place clues: pick a random cell within each partition.
	clues := make([]Clue, len(partitions))
	for i, r := range partitions {
		cx := r.x + rng.IntN(r.w)
		cy := r.y + rng.IntN(r.h)
		clues[i] = Clue{
			ID:    i,
			X:     cx,
			Y:     cy,
			Value: r.w * r.h,
		}
	}

	return &Puzzle{
		Width:  width,
		Height: height,
		Clues:  clues,
	}
}

// partitionGrid recursively fills the grid with non-overlapping rectangles.
func partitionGrid(grid [][]int, width, height, maxRectSize int, partitions *[]genRect, rng *rand.Rand) bool {
	// Find first uncovered cell (top-left scan).
	fx, fy := -1, -1
	for y := range height {
		for x := range width {
			if grid[y][x] == -1 {
				fx, fy = x, y
				break
			}
		}
		if fx >= 0 {
			break
		}
	}

	if fx < 0 {
		return true // all cells covered
	}

	type dim struct{ w, h int }
	var candidates []dim

	// Enumerate all rectangle sizes that fit at (fx, fy).
	for w := 1; w <= width-fx; w++ {
		for h := 1; h <= height-fy; h++ {
			if w*h > maxRectSize {
				continue
			}
			if fits(grid, fx, fy, w, h) {
				candidates = append(candidates, dim{w, h})
			}
		}
	}

	// Shuffle for randomness.
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	id := len(*partitions)
	for _, c := range candidates {
		// Place rectangle.
		place(grid, fx, fy, c.w, c.h, id)
		*partitions = append(*partitions, genRect{fx, fy, c.w, c.h})

		if partitionGrid(grid, width, height, maxRectSize, partitions, rng) {
			return true
		}

		// Backtrack.
		*partitions = (*partitions)[:len(*partitions)-1]
		clearGrid(grid, fx, fy, c.w, c.h)
	}

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
