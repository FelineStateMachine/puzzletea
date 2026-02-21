package hashiwokakero

import (
	"math/rand/v2"
	"testing"
)

func testBridgeLayoutByMode(t *testing.T, mode HashiMode) Puzzle {
	t.Helper()

	for attempt := range 64 {
		rng := rand.New(rand.NewPCG(uint64(mode.Width*100+attempt+1), uint64(mode.Height*1000+attempt+1)))
		islandCount := mode.MinIslands + rng.IntN(mode.MaxIslands-mode.MinIslands+1)
		islands := placeIslandsSeeded(mode.Width, mode.Height, islandCount, rng)
		if islands == nil {
			continue
		}

		p := Puzzle{
			Width:   mode.Width,
			Height:  mode.Height,
			Islands: islands,
		}

		if !buildSpanningTreeSeeded(&p, rng) {
			continue
		}
		addExtraBridgesSeeded(&p, rng)
		return p
	}

	t.Fatalf("failed to build bridge layout for %s", mode.Title())
	return Puzzle{}
}

func bridgeAxisSpan(p *Puzzle, b Bridge) (x1, y1, x2, y2 int, horizontal, ok bool) {
	a := p.FindIslandByID(b.Island1)
	c := p.FindIslandByID(b.Island2)
	if a == nil || c == nil {
		return 0, 0, 0, 0, false, false
	}

	if a.Y == c.Y {
		return a.X, a.Y, c.X, c.Y, true, true
	}
	if a.X == c.X {
		return a.X, a.Y, c.X, c.Y, false, true
	}
	return 0, 0, 0, 0, false, false
}

func rangeContainsInterior(v, bound1, bound2 int) bool {
	min, max := bound1, bound2
	if min > max {
		min, max = max, min
	}
	return v > min && v < max
}

func bridgesCross(p *Puzzle, a, b Bridge) bool {
	if a.Island1 == b.Island1 || a.Island1 == b.Island2 || a.Island2 == b.Island1 || a.Island2 == b.Island2 {
		return false
	}

	ax1, ay1, ax2, ay2, aHorizontal, aok := bridgeAxisSpan(p, a)
	bx1, by1, bx2, by2, bHorizontal, bok := bridgeAxisSpan(p, b)
	if !aok || !bok || aHorizontal == bHorizontal {
		return false
	}

	hx1, hy, hx2 := ax1, ay1, ax2
	vx, vy1, vy2 := bx1, by1, by2
	if !aHorizontal {
		hx1, hy, hx2 = bx1, by1, bx2
		vx, vy1, vy2 = ax1, ay1, ay2
	}

	return rangeContainsInterior(vx, hx1, hx2) && rangeContainsInterior(hy, vy1, vy2)
}

func assertNoBridgeCrossings(t *testing.T, p *Puzzle) {
	t.Helper()

	for i := 0; i < len(p.Bridges); i++ {
		for j := i + 1; j < len(p.Bridges); j++ {
			if bridgesCross(p, p.Bridges[i], p.Bridges[j]) {
				t.Fatalf("bridge %d crosses bridge %d", i, j)
			}
		}
	}
}

func TestGeneratePuzzleRequiredCountsByBoardSize(t *testing.T) {
	for modeIndex, mode := range benchmarkHashiModesByBoardSize() {
		mode := mode
		modeIndex := modeIndex

		t.Run(mode.Title(), func(t *testing.T) {
			puzzle, err := GeneratePuzzleSeeded(mode, rand.New(rand.NewPCG(uint64(modeIndex+900), uint64(modeIndex+901))))
			if err != nil {
				t.Fatalf("GeneratePuzzleSeeded failed: %v", err)
			}
			if puzzle.Width != mode.Width || puzzle.Height != mode.Height {
				t.Fatalf("dimensions = %dx%d, want %dx%d", puzzle.Width, puzzle.Height, mode.Width, mode.Height)
			}
			if len(puzzle.Bridges) != 0 {
				t.Fatalf("expected cleared bridges, got %d", len(puzzle.Bridges))
			}
			for _, island := range puzzle.Islands {
				if island.Required < 1 || island.Required > 8 {
					t.Fatalf("island %d required=%d out of range", island.ID, island.Required)
				}
			}
		})
	}
}

func TestBridgeLayoutConstraintInvariantsByBoardSize(t *testing.T) {
	for _, mode := range benchmarkHashiModesByBoardSize() {
		mode := mode
		t.Run(mode.Title(), func(t *testing.T) {
			p := testBridgeLayoutByMode(t, mode)

			if !p.IsConnected() {
				t.Fatal("expected connected bridge layout")
			}

			seen := make(map[[2]int]bool, len(p.Bridges))
			for _, b := range p.Bridges {
				if b.Count < 1 || b.Count > 2 {
					t.Fatalf("bridge count out of range: %d", b.Count)
				}
				key := [2]int{b.Island1, b.Island2}
				if seen[key] {
					t.Fatalf("duplicate bridge pair: %v", key)
				}
				seen[key] = true

				i1 := p.FindIslandByID(b.Island1)
				i2 := p.FindIslandByID(b.Island2)
				if i1 == nil || i2 == nil {
					t.Fatalf("bridge references unknown island: %+v", b)
				}
				if !isDirectlyConnectable(&p, *i1, *i2) {
					t.Fatalf("bridge between non-directly-connectable islands: %+v", b)
				}
			}

			assertNoBridgeCrossings(t, &p)

			for _, island := range p.Islands {
				count := p.BridgeCount(island.ID)
				if count < 1 || count > 8 {
					t.Fatalf("island %d bridge count=%d out of range", island.ID, count)
				}
			}
		})
	}
}
