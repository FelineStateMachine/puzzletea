package hashiwokakero

import (
	"math/rand/v2"
	"testing"
)

func benchmarkHashiModesByBoardSize() []HashiMode {
	return []HashiMode{
		NewMode("Hard 7x7", "benchmark", 7, 7, 17, 20),
		NewMode("Hard 9x9", "benchmark", 9, 9, 28, 32),
		NewMode("Hard 11x11", "benchmark", 11, 11, 42, 48),
		NewMode("Hard 13x13", "benchmark", 13, 13, 59, 68),
	}
}

func BenchmarkGeneratePuzzleByBoardSize(b *testing.B) {
	for modeIndex, mode := range benchmarkHashiModesByBoardSize() {
		mode := mode
		modeIndex := modeIndex
		b.Run(mode.Title(), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var (
					puzzle Puzzle
					err    error
				)
				for attempt := range 32 {
					rng := rand.New(rand.NewPCG(uint64(i+1), uint64(modeIndex*10+attempt+1)))
					puzzle, err = GeneratePuzzleSeeded(mode, rng)
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

func BenchmarkBuildSpanningTree(b *testing.B) {
	for modeIndex, mode := range benchmarkHashiModesByBoardSize() {
		mode := mode
		modeIndex := modeIndex
		b.Run(mode.Title(), func(b *testing.B) {
			base, err := GeneratePuzzleSeeded(mode, rand.New(rand.NewPCG(uint64(modeIndex+50), uint64(modeIndex+60))))
			if err != nil {
				b.Fatalf("seed puzzle generation failed: %v", err)
			}

			islands := make([]Island, len(base.Islands))
			copy(islands, base.Islands)

			b.ReportAllocs()
			b.ResetTimer()
			successes := 0
			for i := 0; i < b.N; i++ {
				ok := false
				for attempt := range 32 {
					islandCopy := make([]Island, len(islands))
					copy(islandCopy, islands)

					p := Puzzle{
						Width:   mode.Width,
						Height:  mode.Height,
						Islands: islandCopy,
					}

					ok = buildSpanningTreeSeeded(
						&p,
						rand.New(rand.NewPCG(uint64(i+1), uint64(modeIndex*100+attempt+1))),
					)
					if ok {
						successes++
						break
					}
				}
			}
			if successes == 0 {
				b.Fatal("buildSpanningTreeSeeded never succeeded")
			}
		})
	}
}
