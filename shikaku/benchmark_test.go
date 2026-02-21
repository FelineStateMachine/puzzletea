package shikaku

import (
	"math/rand/v2"
	"testing"
)

func benchmarkShikakuModes() []ShikakuMode {
	return []ShikakuMode{
		NewMode("Mini 5x5", "benchmark", 5, 5, 5),
		NewMode("Easy 7x7", "benchmark", 7, 7, 8),
		NewMode("Medium 8x8", "benchmark", 8, 8, 12),
		NewMode("Hard 10x10", "benchmark", 10, 10, 15),
		NewMode("Expert 12x12", "benchmark", 12, 12, 20),
	}
}

func makeUncoveredGenerationGrid(width, height int) [][]int {
	g := make([][]int, height)
	for y := range height {
		g[y] = make([]int, width)
		for x := range width {
			g[y][x] = -1
		}
	}
	return g
}

func BenchmarkGeneratePuzzleBySize(b *testing.B) {
	for modeIndex, mode := range benchmarkShikakuModes() {
		mode := mode
		modeIndex := modeIndex
		b.Run(mode.Title(), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var (
					puzzle Puzzle
					err    error
				)
				for attempt := range 3 {
					rng := rand.New(rand.NewPCG(uint64(i+1), uint64(modeIndex*10+attempt+1)))
					puzzle, err = GeneratePuzzleSeeded(mode.Width, mode.Height, mode.MaxRectSize, rng)
					if err == nil {
						break
					}
				}
				if err != nil {
					b.Fatalf("GeneratePuzzleSeeded failed: %v", err)
				}
				if puzzle.Width != mode.Width || puzzle.Height != mode.Height {
					b.Fatalf("wrong dimensions: got %dx%d, want %dx%d", puzzle.Width, puzzle.Height, mode.Width, mode.Height)
				}
			}
		})
	}
}

func BenchmarkPartitionGrid(b *testing.B) {
	for modeIndex, mode := range benchmarkShikakuModes() {
		mode := mode
		modeIndex := modeIndex
		b.Run(mode.Title(), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				rng := rand.New(rand.NewPCG(uint64(i+1), uint64(modeIndex+100)))
				grid := makeUncoveredGenerationGrid(mode.Width, mode.Height)
				partitions := make([]genRect, 0, mode.Width*mode.Height/2)
				ok := partitionGrid(grid, mode.Width, mode.Height, mode.MaxRectSize, &partitions, rng)
				if !ok {
					b.Fatalf("partitionGrid failed at iteration %d", i)
				}
			}
		})
	}
}
