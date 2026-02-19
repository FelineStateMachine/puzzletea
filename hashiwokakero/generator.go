package hashiwokakero

import (
	"errors"
	"math/rand/v2"
)

const maxGenerateAttempts = 200

// GeneratePuzzle creates a solvable hashiwokakero puzzle for the given mode.
func GeneratePuzzle(mode HashiMode) (Puzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GeneratePuzzleSeeded(mode, rng)
}

func GeneratePuzzleSeeded(mode HashiMode, rng *rand.Rand) (Puzzle, error) {
	islandCount := mode.MinIslands + rng.IntN(mode.MaxIslands-mode.MinIslands+1)

	for range maxGenerateAttempts {
		p := tryGenerateSeeded(mode.Width, mode.Height, islandCount, rng)
		if p != nil {
			return *p, nil
		}
	}
	return Puzzle{}, errors.New("failed to generate puzzle after maximum attempts")
}

func tryGenerateSeeded(width, height, islandCount int, rng *rand.Rand) *Puzzle {
	islands := placeIslandsSeeded(width, height, islandCount, rng)
	if islands == nil {
		return nil
	}

	p := &Puzzle{
		Width:   width,
		Height:  height,
		Islands: islands,
	}

	if !buildSpanningTreeSeeded(p, rng) {
		return nil
	}

	addExtraBridgesSeeded(p, rng)

	for i := range p.Islands {
		p.Islands[i].Required = p.BridgeCount(p.Islands[i].ID)
	}

	for _, isl := range p.Islands {
		if isl.Required < 1 {
			return nil
		}
	}

	p.Bridges = nil
	p.invalidateCache()

	return p
}

// placeIslands randomly places islands on the grid.
// Islands must share rows or columns with at least one other island to allow bridges.
func placeIslands(width, height, count int) []Island {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return placeIslandsSeeded(width, height, count, rng)
}

func placeIslandsSeeded(width, height, count int, rng *rand.Rand) []Island {
	if count <= 0 {
		return nil
	}

	type pos struct{ x, y int }
	occupied := make(map[pos]bool)
	islands := make([]Island, 0, count)

	for i := range count {
		placed := false
		for range 200 {
			x := rng.IntN(width)
			y := rng.IntN(height)
			p := pos{x, y}

			if occupied[p] {
				continue
			}

			// Ensure minimum spacing of 2 cells from existing islands
			// (unless they share a row/column, which is required for bridges)
			tooClose := false
			for _, other := range islands {
				dx := abs(x - other.X)
				dy := abs(y - other.Y)
				// Adjacent diagonal — too close and can't bridge
				if dx <= 1 && dy <= 1 && !(dx == 0 || dy == 0) {
					tooClose = true
					break
				}
				// Same position as existing (already checked by occupied map)
			}
			if tooClose {
				continue
			}

			occupied[p] = true
			islands = append(islands, Island{ID: i, X: x, Y: y})
			placed = true
			break
		}
		if !placed {
			return nil // couldn't place enough islands
		}
	}

	return islands
}

// buildSpanningTreeSeeded connects all islands using a randomized approach.
// Returns false if unable to connect all islands.
func buildSpanningTreeSeeded(p *Puzzle, rng *rand.Rand) bool {
	if len(p.Islands) <= 1 {
		return true
	}

	type pair struct{ id1, id2 int }
	var pairs []pair

	for i := range len(p.Islands) {
		for j := i + 1; j < len(p.Islands); j++ {
			a := p.Islands[i]
			b := p.Islands[j]

			if a.X != b.X && a.Y != b.Y {
				continue // not aligned
			}

			if isDirectlyConnectable(p, a, b) {
				pairs = append(pairs, pair{a.ID, b.ID})
			}
		}
	}

	rng.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

	// Union-Find to build spanning tree
	parent := make(map[int]int)
	rank := make(map[int]int)
	for _, isl := range p.Islands {
		parent[isl.ID] = isl.ID
	}

	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b int) bool {
		ra, rb := find(a), find(b)
		if ra == rb {
			return false
		}
		if rank[ra] < rank[rb] {
			ra, rb = rb, ra
		}
		parent[rb] = ra
		if rank[ra] == rank[rb] {
			rank[ra]++
		}
		return true
	}

	edgesAdded := 0
	for _, pr := range pairs {
		if find(pr.id1) == find(pr.id2) {
			continue
		}
		// Check if this bridge would cross existing ones
		if p.WouldCross(pr.id1, pr.id2) {
			continue
		}
		count := 1 + rng.IntN(2) // 1 or 2 bridges
		p.SetBridge(pr.id1, pr.id2, count)
		union(pr.id1, pr.id2)
		edgesAdded++
		if edgesAdded == len(p.Islands)-1 {
			break
		}
	}

	return p.IsConnected()
}

// addExtraBridgesSeeded adds additional bridges beyond the spanning tree for complexity.
func addExtraBridgesSeeded(p *Puzzle, rng *rand.Rand) {
	var pairs []struct{ id1, id2 int }

	for i := 0; i < len(p.Islands); i++ {
		for j := i + 1; j < len(p.Islands); j++ {
			a := p.Islands[i]
			b := p.Islands[j]

			if a.X != b.X && a.Y != b.Y {
				continue
			}

			if !isDirectlyConnectable(p, a, b) {
				continue
			}

			existing := p.GetBridge(a.ID, b.ID)
			if existing != nil && existing.Count >= 2 {
				continue // already maxed
			}

			pairs = append(pairs, struct{ id1, id2 int }{a.ID, b.ID})
		}
	}

	rng.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

	// Add bridges to ~30% of eligible pairs
	limit := len(pairs) / 3
	added := 0
	for _, pr := range pairs {
		if added >= limit {
			break
		}

		existing := p.GetBridge(pr.id1, pr.id2)
		if existing != nil && existing.Count >= 2 {
			continue
		}

		if existing == nil && p.WouldCross(pr.id1, pr.id2) {
			continue
		}

		newCount := 1
		if existing != nil {
			newCount = existing.Count + 1
		}

		// Don't let any island exceed 8 bridges
		addAmount := newCount
		if existing != nil {
			addAmount = newCount - existing.Count
		}
		if p.BridgeCount(pr.id1)+addAmount > 8 || p.BridgeCount(pr.id2)+addAmount > 8 {
			continue
		}

		p.SetBridge(pr.id1, pr.id2, newCount)
		added++
	}
}

// isDirectlyConnectable checks if two aligned islands can be connected
// (no other island sits between them).
func isDirectlyConnectable(p *Puzzle, a, b Island) bool {
	if a.X == b.X {
		// Vertical alignment
		minY, maxY := a.Y, b.Y
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		for _, isl := range p.Islands {
			if isl.X == a.X && isl.Y > minY && isl.Y < maxY {
				return false
			}
		}
		return true
	}
	if a.Y == b.Y {
		// Horizontal alignment
		minX, maxX := a.X, b.X
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for _, isl := range p.Islands {
			if isl.Y == a.Y && isl.X > minX && isl.X < maxX {
				return false
			}
		}
		return true
	}
	return false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
