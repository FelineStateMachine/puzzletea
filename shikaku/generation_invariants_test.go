package shikaku

import (
	"math/rand/v2"
	"testing"
)

func TestPartitionGridCoverageBySize(t *testing.T) {
	for modeIndex, mode := range benchmarkShikakuModes() {
		mode := mode
		modeIndex := modeIndex

		t.Run(mode.Title(), func(t *testing.T) {
			grid := makeUncoveredGenerationGrid(mode.Width, mode.Height)
			partitions := make([]genRect, 0, mode.Width*mode.Height/2)

			ok := partitionGrid(
				grid,
				mode.Width,
				mode.Height,
				mode.MaxRectSize,
				&partitions,
				rand.New(rand.NewPCG(uint64(modeIndex+700), uint64(modeIndex+701))),
			)
			if !ok {
				t.Fatal("partitionGrid returned false")
			}
			if len(partitions) == 0 {
				t.Fatal("expected at least one partition")
			}

			covered := make([][]bool, mode.Height)
			for y := range mode.Height {
				covered[y] = make([]bool, mode.Width)
			}

			totalArea := 0
			for _, r := range partitions {
				if r.w <= 0 || r.h <= 0 {
					t.Fatalf("invalid partition size: %+v", r)
				}
				if r.x < 0 || r.y < 0 || r.x+r.w > mode.Width || r.y+r.h > mode.Height {
					t.Fatalf("partition out of bounds: %+v", r)
				}
				if r.w*r.h > mode.MaxRectSize {
					t.Fatalf("partition area %d exceeds max %d", r.w*r.h, mode.MaxRectSize)
				}

				totalArea += r.w * r.h
				for dy := range r.h {
					for dx := range r.w {
						x, y := r.x+dx, r.y+dy
						if covered[y][x] {
							t.Fatalf("overlap at (%d,%d)", x, y)
						}
						covered[y][x] = true
					}
				}
			}

			if totalArea != mode.Width*mode.Height {
				t.Fatalf("partition area sum=%d, want %d", totalArea, mode.Width*mode.Height)
			}

			for y := range mode.Height {
				for x := range mode.Width {
					if !covered[y][x] {
						t.Fatalf("uncovered cell at (%d,%d)", x, y)
					}
					if grid[y][x] < 0 {
						t.Fatalf("grid cell (%d,%d) left uncovered in generation grid", x, y)
					}
				}
			}
		})
	}
}

func TestGeneratePuzzleClueInvariantsBySize(t *testing.T) {
	for modeIndex, mode := range benchmarkShikakuModes() {
		mode := mode
		modeIndex := modeIndex

		t.Run(mode.Title(), func(t *testing.T) {
			puzzle, err := GeneratePuzzleSeeded(
				mode.Width,
				mode.Height,
				mode.MaxRectSize,
				rand.New(rand.NewPCG(uint64(modeIndex+800), uint64(modeIndex+801))),
			)
			if err != nil {
				t.Fatalf("GeneratePuzzleSeeded failed: %v", err)
			}
			if puzzle.Width != mode.Width || puzzle.Height != mode.Height {
				t.Fatalf("dimensions = %dx%d, want %dx%d", puzzle.Width, puzzle.Height, mode.Width, mode.Height)
			}
			if len(puzzle.Rectangles) != 0 {
				t.Fatalf("expected no pre-placed rectangles, got %d", len(puzzle.Rectangles))
			}

			seenPos := make(map[[2]int]bool, len(puzzle.Clues))
			seenID := make(map[int]bool, len(puzzle.Clues))
			sum := 0

			for _, clue := range puzzle.Clues {
				if clue.X < 0 || clue.X >= mode.Width || clue.Y < 0 || clue.Y >= mode.Height {
					t.Fatalf("clue out of bounds: %+v", clue)
				}
				if clue.Value < 1 || clue.Value > mode.MaxRectSize {
					t.Fatalf("clue value out of range: %+v", clue)
				}

				pos := [2]int{clue.X, clue.Y}
				if seenPos[pos] {
					t.Fatalf("duplicate clue position: (%d,%d)", clue.X, clue.Y)
				}
				seenPos[pos] = true

				if seenID[clue.ID] {
					t.Fatalf("duplicate clue ID: %d", clue.ID)
				}
				seenID[clue.ID] = true

				sum += clue.Value
			}

			if sum != mode.Width*mode.Height {
				t.Fatalf("sum of clue values=%d, want %d", sum, mode.Width*mode.Height)
			}
			if len(seenID) != len(puzzle.Clues) {
				t.Fatal("clue IDs are not unique")
			}
		})
	}
}
