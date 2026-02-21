package hashiwokakero

// Island represents a numbered node on the grid.
type Island struct {
	ID       int
	X, Y     int
	Required int // target bridge count (1-8)
}

// Bridge represents a connection between two islands.
type Bridge struct {
	Island1, Island2 int // island IDs (Island1 < Island2)
	Count            int // 1 or 2
}

// cellKind describes what occupies a grid cell.
type cellKind int

const (
	cellEmpty cellKind = iota
	cellIsland
	cellBridgeH // horizontal bridge segment
	cellBridgeV // vertical bridge segment
)

// cellInfo is the rendering info for a single grid cell.
type cellInfo struct {
	Kind        cellKind
	IslandID    int // valid when Kind == cellIsland
	BridgeIdx   int // index into Puzzle.Bridges, valid for bridge cells
	BridgeCount int // 1 or 2
}

// Puzzle holds the full game state.
type Puzzle struct {
	Width, Height int
	Islands       []Island
	Bridges       []Bridge
	cellCache     [][]cellInfo   // lazily built, invalidated on bridge changes
	posIndex      map[[2]int]int // (x,y) → island slice index; lazily built
	bridgeCounts  []int          // per-island bridge count; lazily built
	bridgeIndex   map[[2]int]int // (island1,island2) -> Bridges index; lazily built
}

// FindIslandAt returns the island at (x,y), or nil.
func (p *Puzzle) FindIslandAt(x, y int) *Island {
	if p.posIndex == nil {
		p.buildPosIndex()
	}
	if idx, ok := p.posIndex[[2]int{x, y}]; ok {
		return &p.Islands[idx]
	}
	return nil
}

func (p *Puzzle) buildPosIndex() {
	p.posIndex = make(map[[2]int]int, len(p.Islands))
	for i, isl := range p.Islands {
		p.posIndex[[2]int{isl.X, isl.Y}] = i
	}
}

func bridgeKey(id1, id2 int) [2]int {
	a, b := id1, id2
	if a > b {
		a, b = b, a
	}
	return [2]int{a, b}
}

func (p *Puzzle) buildBridgeIndex() {
	p.bridgeIndex = make(map[[2]int]int, len(p.Bridges))
	for i, b := range p.Bridges {
		p.bridgeIndex[bridgeKey(b.Island1, b.Island2)] = i
	}
}

// FindIslandByID returns the island with the given ID, or nil.
func (p *Puzzle) FindIslandByID(id int) *Island {
	if id >= 0 && id < len(p.Islands) && p.Islands[id].ID == id {
		return &p.Islands[id]
	}
	// Fallback for non-sequential IDs
	for i := range p.Islands {
		if p.Islands[i].ID == id {
			return &p.Islands[i]
		}
	}
	return nil
}

// GetBridge returns the bridge between two islands (by ID), or nil.
func (p *Puzzle) GetBridge(id1, id2 int) *Bridge {
	if p.bridgeIndex == nil {
		p.buildBridgeIndex()
	}

	if idx, ok := p.bridgeIndex[bridgeKey(id1, id2)]; ok {
		if idx >= 0 && idx < len(p.Bridges) {
			return &p.Bridges[idx]
		}
	}
	return nil
}

// BridgeCount returns the total number of bridges connected to an island.
func (p *Puzzle) BridgeCount(islandID int) int {
	if p.bridgeCounts == nil {
		p.buildBridgeCounts()
	}
	if islandID >= 0 && islandID < len(p.bridgeCounts) {
		return p.bridgeCounts[islandID]
	}
	return 0
}

func (p *Puzzle) buildBridgeCounts() {
	p.bridgeCounts = make([]int, len(p.Islands))
	for _, b := range p.Bridges {
		if b.Island1 >= 0 && b.Island1 < len(p.bridgeCounts) {
			p.bridgeCounts[b.Island1] += b.Count
		}
		if b.Island2 >= 0 && b.Island2 < len(p.bridgeCounts) {
			p.bridgeCounts[b.Island2] += b.Count
		}
	}
}

// FindAdjacentIsland casts a ray from the island at fromID in direction (dx,dy)
// and returns the first island hit, or nil. dx/dy must be (0,1),(0,-1),(1,0),(-1,0).
// Stops if a perpendicular bridge crosses the path.
func (p *Puzzle) FindAdjacentIsland(fromID, dx, dy int) *Island {
	from := p.FindIslandByID(fromID)
	if from == nil {
		return nil
	}
	x, y := from.X+dx, from.Y+dy
	for x >= 0 && x < p.Width && y >= 0 && y < p.Height {
		if isl := p.FindIslandAt(x, y); isl != nil {
			return isl
		}
		ci := p.CellContent(x, y)
		if dx == 0 && ci.Kind == cellBridgeH {
			return nil // horizontal bridge blocks vertical ray
		}
		if dy == 0 && ci.Kind == cellBridgeV {
			return nil // vertical bridge blocks horizontal ray
		}
		x += dx
		y += dy
	}
	return nil
}

// WouldCross checks if adding a bridge between id1 and id2 would cross any existing bridge.
func (p *Puzzle) WouldCross(id1, id2 int) bool {
	isl1 := p.FindIslandByID(id1)
	isl2 := p.FindIslandByID(id2)
	if isl1 == nil || isl2 == nil {
		return true
	}

	horizontal := isl1.Y == isl2.Y
	vertical := isl1.X == isl2.X
	if !horizontal && !vertical {
		return true // not a valid bridge direction
	}

	if horizontal {
		y := isl1.Y
		minX, maxX := isl1.X, isl2.X
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for x := minX + 1; x < maxX; x++ {
			ci := p.CellContent(x, y)
			if ci.Kind == cellIsland || ci.Kind == cellBridgeV {
				return true
			}
		}
	} else {
		x := isl1.X
		minY, maxY := isl1.Y, isl2.Y
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		for y := minY + 1; y < maxY; y++ {
			ci := p.CellContent(x, y)
			if ci.Kind == cellIsland || ci.Kind == cellBridgeH {
				return true
			}
		}
	}
	return false
}

// SetBridge sets or removes the bridge between two islands.
// count=0 removes the bridge. count must be 0, 1, or 2.
func (p *Puzzle) SetBridge(id1, id2, count int) {
	a, b := id1, id2
	if a > b {
		a, b = b, a
	}

	p.invalidateBridgeCaches()
	if p.bridgeIndex == nil {
		p.buildBridgeIndex()
	}
	key := bridgeKey(a, b)

	if count == 0 {
		// Remove bridge
		idx, ok := p.bridgeIndex[key]
		if !ok {
			return
		}

		p.Bridges = append(p.Bridges[:idx], p.Bridges[idx+1:]...)
		delete(p.bridgeIndex, key)
		for i := idx; i < len(p.Bridges); i++ {
			p.bridgeIndex[bridgeKey(p.Bridges[i].Island1, p.Bridges[i].Island2)] = i
		}
		return
	}

	// Update or create bridge
	if idx, ok := p.bridgeIndex[key]; ok {
		p.Bridges[idx].Count = count
		return
	}

	p.Bridges = append(p.Bridges, Bridge{Island1: a, Island2: b, Count: count})
	p.bridgeIndex[key] = len(p.Bridges) - 1
}

// CellContent returns what occupies the given grid cell, using a lazily-built cache.
func (p *Puzzle) CellContent(x, y int) cellInfo {
	if p.cellCache == nil {
		p.rebuildCellCache()
	}
	return p.cellCache[y][x]
}

// rebuildCellCache recomputes the full cell grid from islands and bridges.
func (p *Puzzle) rebuildCellCache() {
	cache := make([][]cellInfo, p.Height)
	for y := range cache {
		cache[y] = make([]cellInfo, p.Width)
	}

	for _, isl := range p.Islands {
		cache[isl.Y][isl.X] = cellInfo{Kind: cellIsland, IslandID: isl.ID}
	}

	for i, b := range p.Bridges {
		isl1 := p.FindIslandByID(b.Island1)
		isl2 := p.FindIslandByID(b.Island2)
		if isl1 == nil || isl2 == nil {
			continue
		}

		if isl1.Y == isl2.Y {
			// Horizontal bridge
			minX, maxX := isl1.X, isl2.X
			if minX > maxX {
				minX, maxX = maxX, minX
			}
			for x := minX + 1; x < maxX; x++ {
				cache[isl1.Y][x] = cellInfo{Kind: cellBridgeH, BridgeIdx: i, BridgeCount: b.Count}
			}
		} else if isl1.X == isl2.X {
			// Vertical bridge
			minY, maxY := isl1.Y, isl2.Y
			if minY > maxY {
				minY, maxY = maxY, minY
			}
			for y := minY + 1; y < maxY; y++ {
				cache[y][isl1.X] = cellInfo{Kind: cellBridgeV, BridgeIdx: i, BridgeCount: b.Count}
			}
		}
	}

	p.cellCache = cache
}

// invalidateCache marks all lazy caches as stale.
func (p *Puzzle) invalidateCache() {
	p.cellCache = nil
	p.posIndex = nil
	p.bridgeCounts = nil
	p.bridgeIndex = nil
}

// invalidateBridgeCaches marks bridge-derived caches as stale.
func (p *Puzzle) invalidateBridgeCaches() {
	p.cellCache = nil
	p.bridgeCounts = nil
}

// IsConnected checks if all islands form a single connected component via bridges.
func (p *Puzzle) IsConnected() bool {
	n := len(p.Islands)
	if n <= 1 {
		return true
	}
	if len(p.Bridges) == 0 {
		return false
	}

	// Build adjacency list from bridges
	adj := make(map[int][]int, n)
	for _, b := range p.Bridges {
		adj[b.Island1] = append(adj[b.Island1], b.Island2)
		adj[b.Island2] = append(adj[b.Island2], b.Island1)
	}

	visited := make([]bool, n)
	queue := []int{p.Islands[0].ID}
	visited[p.Islands[0].ID] = true
	count := 1

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nb := range adj[cur] {
			if !visited[nb] {
				visited[nb] = true
				count++
				queue = append(queue, nb)
			}
		}
	}

	return count == n
}

// IsSolved returns true if the puzzle is completely and correctly solved.
func (p *Puzzle) IsSolved() bool {
	for _, isl := range p.Islands {
		if p.BridgeCount(isl.ID) != isl.Required {
			return false
		}
	}
	return p.IsConnected()
}
