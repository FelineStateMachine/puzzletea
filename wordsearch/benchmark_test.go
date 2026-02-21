package wordsearch

import (
	"math/rand/v2"
	"testing"
)

type wordSearchBenchCase struct {
	name        string
	width       int
	height      int
	wordCount   int
	minWordLen  int
	maxWordLen  int
	allowedDirs []Direction
}

func BenchmarkGenerateWordSearchByMode(b *testing.B) {
	cases := []wordSearchBenchCase{
		{
			name:        "easy_10x10",
			width:       10,
			height:      10,
			wordCount:   6,
			minWordLen:  3,
			maxWordLen:  5,
			allowedDirs: []Direction{Right, Down, DownRight},
		},
		{
			name:        "medium_15x15",
			width:       15,
			height:      15,
			wordCount:   10,
			minWordLen:  4,
			maxWordLen:  7,
			allowedDirs: []Direction{Right, Down, DownRight, DownLeft, Left, Up},
		},
		{
			name:        "hard_20x20",
			width:       20,
			height:      20,
			wordCount:   15,
			minWordLen:  5,
			maxWordLen:  10,
			allowedDirs: []Direction{Right, Down, DownRight, DownLeft, Left, Up, UpRight, UpLeft},
		},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			rng := rand.New(rand.NewPCG(1, 2))
			var aggregate placementStats
			b.ReportAllocs()

			for range b.N {
				_, _, stats := generateWordSearchSeededWithStats(
					tt.width,
					tt.height,
					tt.wordCount,
					tt.minWordLen,
					tt.maxWordLen,
					tt.allowedDirs,
					rng,
				)
				aggregate.TargetWords += stats.TargetWords
				aggregate.PlacedWords += stats.PlacedWords
				aggregate.FailedWords += stats.FailedWords
				aggregate.RandomAttempts += stats.RandomAttempts
				aggregate.FallbackAttempts += stats.FallbackAttempts
				aggregate.FallbackUsed += stats.FallbackUsed
				aggregate.FallbackPlaced += stats.FallbackPlaced
			}

			b.ReportMetric(aggregate.SuccessRate()*100, "place_pct")
			b.ReportMetric(aggregate.AttemptsPerWord(), "attempts/word")
			if aggregate.TargetWords > 0 {
				fallbackRate := float64(aggregate.FallbackUsed) * 100 / float64(aggregate.TargetWords)
				b.ReportMetric(fallbackRate, "fallback_pct")
			}
		})
	}
}

func BenchmarkTryPlaceWord(b *testing.B) {
	cases := []struct {
		name        string
		width       int
		height      int
		text        string
		allowedDirs []Direction
	}{
		{
			name:        "easy_10x10",
			width:       10,
			height:      10,
			text:        "HELLO",
			allowedDirs: []Direction{Right, Down, DownRight},
		},
		{
			name:        "medium_15x15",
			width:       15,
			height:      15,
			text:        "PUZZLE",
			allowedDirs: []Direction{Right, Down, DownRight, DownLeft, Left, Up},
		},
		{
			name:        "hard_20x20",
			width:       20,
			height:      20,
			text:        "WORDSEARCH",
			allowedDirs: []Direction{Right, Down, DownRight, DownLeft, Left, Up, UpRight, UpLeft},
		},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			rng := rand.New(rand.NewPCG(3, 4))
			g := createEmptyGrid(tt.height, tt.width)
			b.ReportAllocs()

			for range b.N {
				w := tryPlaceWordSeeded(g, tt.text, tt.allowedDirs, 100, rng)
				if w == nil {
					b.Fatalf("tryPlaceWordSeeded returned nil for %q", tt.text)
				}
				clearWord(g, *w)
			}
		})
	}
}

func clearWord(g grid, w Word) {
	dx, dy := w.Direction.Vector()
	for i := range len(w.Text) {
		x := w.Start.X + dx*i
		y := w.Start.Y + dy*i
		g.Set(x, y, ' ')
	}
}
