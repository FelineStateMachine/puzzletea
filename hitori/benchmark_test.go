package hitori

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

type hitoriBenchMode struct {
	size       int
	blackRatio float64
}

var hitoriBenchModes = []hitoriBenchMode{
	{size: 5, blackRatio: 0.32},
	{size: 6, blackRatio: 0.32},
	{size: 8, blackRatio: 0.30},
	{size: 9, blackRatio: 0.30},
	{size: 10, blackRatio: 0.30},
	{size: 12, blackRatio: 0.28},
}

func BenchmarkGenerateBySize(b *testing.B) {
	for _, mode := range hitoriBenchModes {
		name := fmt.Sprintf("%dx%d", mode.size, mode.size)
		b.Run(name, func(b *testing.B) {
			rng := rand.New(rand.NewPCG(uint64(mode.size), uint64(mode.size*23+5)))
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				base := generateLatinSquareSeeded(mode.size, rng)
				mask := generateValidMaskSeeded(mode.size, mode.blackRatio, rng)
				puzzle := constructPuzzleSeeded(base, mask, rng)
				if len(puzzle) != mode.size {
					b.Fatalf("unexpected puzzle size %d", mode.size)
				}
			}
		})
	}
}

func BenchmarkCountPuzzleSolutionsBySize(b *testing.B) {
	for _, mode := range hitoriBenchModes {
		name := fmt.Sprintf("%dx%d", mode.size, mode.size)
		b.Run(name, func(b *testing.B) {
			rng := rand.New(rand.NewPCG(uint64(mode.size*97), uint64(mode.size*31+7)))
			puzzle := buildSolverBenchPuzzle(mode.size, rng)

			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				count := countPuzzleSolutions(puzzle, mode.size, 2)
				if count < 1 {
					b.Fatalf("countPuzzleSolutions returned %d, want >= 1", count)
				}
			}
		})
	}
}

func buildSolverBenchPuzzle(size int, rng *rand.Rand) grid {
	puzzle := generateLatinSquareSeeded(size, rng)

	for y := range size {
		x := (y * 2) % size
		next := (x + 1) % size
		puzzle[y][next] = puzzle[y][x]
	}

	return puzzle
}
