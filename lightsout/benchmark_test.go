package lightsout

import (
	"math/rand/v2"
	"testing"
)

func skipBenchmarkInShortMode(b *testing.B) {
	b.Helper()
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}
}

func BenchmarkGenerateByMode(b *testing.B) {
	skipBenchmarkInShortMode(b)

	cases := []struct {
		name string
		w    int
		h    int
	}{
		{name: "easy_3x3", w: 3, h: 3},
		{name: "medium_5x5", w: 5, h: 5},
		{name: "hard_7x7", w: 7, h: 7},
		{name: "extreme_9x9", w: 9, h: 9},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			rng := rand.New(rand.NewPCG(11, 12))
			b.ReportAllocs()

			for range b.N {
				_ = GenerateSeeded(tt.w, tt.h, rng)
			}
		})
	}
}

func BenchmarkToggle(b *testing.B) {
	skipBenchmarkInShortMode(b)

	cases := []struct {
		name string
		w    int
		h    int
		x    int
		y    int
	}{
		{name: "easy_center", w: 3, h: 3, x: 1, y: 1},
		{name: "medium_center", w: 5, h: 5, x: 2, y: 2},
		{name: "hard_center", w: 7, h: 7, x: 3, y: 3},
		{name: "extreme_center", w: 9, h: 9, x: 4, y: 4},
		{name: "extreme_corner", w: 9, h: 9, x: 0, y: 0},
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			g := makeGrid(tt.h, tt.w, false)
			b.ReportAllocs()

			for range b.N {
				Toggle(g, tt.x, tt.y)
			}
		})
	}
}
