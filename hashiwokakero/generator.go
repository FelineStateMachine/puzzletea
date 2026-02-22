package hashiwokakero

import (
	"errors"
	"math/rand/v2"
)

const maxGenerateAttempts = 200

type islandPair struct {
	id1 int
	id2 int
}

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

	pairs := connectableIslandPairs(p)
	if !buildSpanningTreeWithPairsSeeded(p, pairs, rng) {
		return nil
	}

	addExtraBridgesWithPairsSeeded(p, pairs, rng)

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
	return buildSpanningTreeWithPairsSeeded(p, connectableIslandPairs(p), rng)
}

func connectableIslandPairs(p *Puzzle) []islandPair {
	pairs := make([]islandPair, 0, len(p.Islands)*2)
	for i := range len(p.Islands) {
		for j := i + 1; j < len(p.Islands); j++ {
			a := p.Islands[i]
			b := p.Islands[j]

			if a.X != b.X && a.Y != b.Y {
				continue // not aligned
			}

			if isDirectlyConnectable(p, a, b) {
				pairs = append(pairs, islandPair{id1: a.ID, id2: b.ID})
			}
		}
	}
	return pairs
}

func buildSpanningTreeWithPairsSeeded(p *Puzzle, pairs []islandPair, rng *rand.Rand) bool {
	if len(p.Islands) <= 1 {
		return true
	}
	if len(pairs) == 0 {
		return false
	}

	order := make([]islandPair, len(pairs))
	copy(order, pairs)

	rng.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
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
	for _, pr := range order {
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
	addExtraBridgesWithPairsSeeded(p, connectableIslandPairs(p), rng)
}

func addExtraBridgesWithPairsSeeded(p *Puzzle, pairs []islandPair, rng *rand.Rand) {
	eligible := make([]islandPair, 0, len(pairs))
	degrees := make(map[int]int, len(p.Islands))
	for _, bridge := range p.Bridges {
		degrees[bridge.Island1] += bridge.Count
		degrees[bridge.Island2] += bridge.Count
	}

	for _, pr := range pairs {
		existing := p.GetBridge(pr.id1, pr.id2)
		if existing != nil && existing.Count >= 2 {
			continue // already maxed
		}
		eligible = append(eligible, pr)
	}

	rng.Shuffle(len(eligible), func(i, j int) {
		eligible[i], eligible[j] = eligible[j], eligible[i]
	})

	// Add bridges to ~30% of eligible pairs
	limit := len(eligible) / 3
	added := 0
	for _, pr := range eligible {
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
		if degrees[pr.id1]+addAmount > 8 || degrees[pr.id2]+addAmount > 8 {
			continue
		}

		p.SetBridge(pr.id1, pr.id2, newCount)
		degrees[pr.id1] += addAmount
		degrees[pr.id2] += addAmount
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
