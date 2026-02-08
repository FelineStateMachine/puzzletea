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
}

// FindIslandAt returns the island at (x,y), or nil.
func (p *Puzzle) FindIslandAt(x, y int) *Island {
	for i := range p.Islands {
		if p.Islands[i].X == x && p.Islands[i].Y == y {
			return &p.Islands[i]
		}
	}
	return nil
}

// FindIslandByID returns the island with the given ID, or nil.
func (p *Puzzle) FindIslandByID(id int) *Island {
	for i := range p.Islands {
		if p.Islands[i].ID == id {
			return &p.Islands[i]
		}
	}
	return nil
}

// GetBridge returns the bridge between two islands (by ID), or nil.
func (p *Puzzle) GetBridge(id1, id2 int) *Bridge {
	a, b := id1, id2
	if a > b {
		a, b = b, a
	}
	for i := range p.Bridges {
		if p.Bridges[i].Island1 == a && p.Bridges[i].Island2 == b {
			return &p.Bridges[i]
		}
	}
	return nil
}

// BridgeCount returns the total number of bridges connected to an island.
func (p *Puzzle) BridgeCount(islandID int) int {
	count := 0
	for _, b := range p.Bridges {
		if b.Island1 == islandID || b.Island2 == islandID {
			count += b.Count
		}
	}
	return count
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
		// Check for island at this position
		if isl := p.FindIslandAt(x, y); isl != nil {
			return isl
		}
		// Check if a perpendicular bridge crosses this cell
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

	// Determine direction of new bridge
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

	if count == 0 {
		// Remove bridge
		for i := range p.Bridges {
			if p.Bridges[i].Island1 == a && p.Bridges[i].Island2 == b {
				p.Bridges = append(p.Bridges[:i], p.Bridges[i+1:]...)
				return
			}
		}
		return
	}

	// Update or create bridge
	for i := range p.Bridges {
		if p.Bridges[i].Island1 == a && p.Bridges[i].Island2 == b {
			p.Bridges[i].Count = count
			return
		}
	}
	p.Bridges = append(p.Bridges, Bridge{Island1: a, Island2: b, Count: count})
}

// CellContent returns what occupies the given grid cell.
func (p *Puzzle) CellContent(x, y int) cellInfo {
	// Check islands first
	for _, isl := range p.Islands {
		if isl.X == x && isl.Y == y {
			return cellInfo{Kind: cellIsland, IslandID: isl.ID}
		}
	}

	// Check bridges
	for i, b := range p.Bridges {
		isl1 := p.FindIslandByID(b.Island1)
		isl2 := p.FindIslandByID(b.Island2)
		if isl1 == nil || isl2 == nil {
			continue
		}

		if isl1.Y == isl2.Y && y == isl1.Y {
			// Horizontal bridge
			minX, maxX := isl1.X, isl2.X
			if minX > maxX {
				minX, maxX = maxX, minX
			}
			if x > minX && x < maxX {
				return cellInfo{Kind: cellBridgeH, BridgeIdx: i, BridgeCount: b.Count}
			}
		} else if isl1.X == isl2.X && x == isl1.X {
			// Vertical bridge
			minY, maxY := isl1.Y, isl2.Y
			if minY > maxY {
				minY, maxY = maxY, minY
			}
			if y > minY && y < maxY {
				return cellInfo{Kind: cellBridgeV, BridgeIdx: i, BridgeCount: b.Count}
			}
		}
	}

	return cellInfo{Kind: cellEmpty}
}

// IsConnected checks if all islands form a single connected component via bridges.
func (p *Puzzle) IsConnected() bool {
	if len(p.Islands) <= 1 {
		return true
	}
	if len(p.Bridges) == 0 {
		return false
	}

	visited := make(map[int]bool)
	queue := []int{p.Islands[0].ID}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true

		for _, b := range p.Bridges {
			if b.Island1 == cur && !visited[b.Island2] {
				queue = append(queue, b.Island2)
			}
			if b.Island2 == cur && !visited[b.Island1] {
				queue = append(queue, b.Island1)
			}
		}
	}

	return len(visited) == len(p.Islands)
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
