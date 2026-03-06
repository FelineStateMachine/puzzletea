package rippleeffect

import "fmt"

type Cell struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Cage struct {
	ID    int    `json:"id"`
	Size  int    `json:"size"`
	Cells []Cell `json:"cells"`
}

type boundaryMask uint8

const (
	boundaryTop boundaryMask = 1 << iota
	boundaryRight
	boundaryBottom
	boundaryLeft
)

func (m boundaryMask) has(flag boundaryMask) bool {
	return m&flag != 0
}

type geometry struct {
	width      int
	height     int
	cages      []Cage
	cageGrid   [][]int
	cageCells  [][]point
	cageSizes  []int
	boundaries [][]boundaryMask
}

func buildGeometry(width, height int, cages []Cage) (*geometry, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid ripple effect size %dx%d", width, height)
	}
	if len(cages) == 0 {
		return nil, fmt.Errorf("ripple effect requires at least one cage")
	}

	cageGrid := make([][]int, height)
	for y := range height {
		cageGrid[y] = make([]int, width)
		for x := range width {
			cageGrid[y][x] = -1
		}
	}

	idSeen := make(map[int]struct{}, len(cages))
	cageCells := make([][]point, len(cages))
	cageSizes := make([]int, len(cages))
	for cageIdx, cage := range cages {
		if _, exists := idSeen[cage.ID]; exists {
			return nil, fmt.Errorf("duplicate cage id %d", cage.ID)
		}
		idSeen[cage.ID] = struct{}{}

		size := cage.Size
		if size == 0 {
			size = len(cage.Cells)
			cages[cageIdx].Size = size
		}
		if size != len(cage.Cells) {
			return nil, fmt.Errorf("cage %d size %d does not match %d cells", cage.ID, cage.Size, len(cage.Cells))
		}
		if size <= 0 || size > 9 {
			return nil, fmt.Errorf("cage %d has invalid size %d", cage.ID, size)
		}

		cells := make([]point, 0, len(cage.Cells))
		for _, cell := range cage.Cells {
			if cell.X < 0 || cell.X >= width || cell.Y < 0 || cell.Y >= height {
				return nil, fmt.Errorf("cage %d has out-of-bounds cell (%d,%d)", cage.ID, cell.X, cell.Y)
			}
			if cageGrid[cell.Y][cell.X] >= 0 {
				return nil, fmt.Errorf("cell (%d,%d) belongs to multiple cages", cell.X, cell.Y)
			}
			cageGrid[cell.Y][cell.X] = cageIdx
			cells = append(cells, point{x: cell.X, y: cell.Y})
		}
		if !cellsConnected(cells) {
			return nil, fmt.Errorf("cage %d is not connected", cage.ID)
		}

		cageCells[cageIdx] = cells
		cageSizes[cageIdx] = size
	}

	for y := range height {
		for x := range width {
			if cageGrid[y][x] < 0 {
				return nil, fmt.Errorf("cell (%d,%d) is not assigned to any cage", x, y)
			}
		}
	}

	return &geometry{
		width:      width,
		height:     height,
		cages:      cages,
		cageGrid:   cageGrid,
		cageCells:  cageCells,
		cageSizes:  cageSizes,
		boundaries: buildBoundaryMasks(cageGrid),
	}, nil
}

func cellsConnected(cells []point) bool {
	if len(cells) <= 1 {
		return true
	}

	cellSet := make(map[point]struct{}, len(cells))
	for _, cell := range cells {
		cellSet[cell] = struct{}{}
	}

	queue := []point{cells[0]}
	visited := map[point]struct{}{cells[0]: {}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, next := range orthogonalNeighbors(curr, 1<<30, 1<<30) {
			if _, ok := cellSet[next]; !ok {
				continue
			}
			if _, ok := visited[next]; ok {
				continue
			}
			visited[next] = struct{}{}
			queue = append(queue, next)
		}
	}

	return len(visited) == len(cells)
}

func buildBoundaryMasks(cageGrid [][]int) [][]boundaryMask {
	height := len(cageGrid)
	width := len(cageGrid[0])
	masks := make([][]boundaryMask, height)
	for y := range height {
		masks[y] = make([]boundaryMask, width)
		for x := range width {
			cage := cageGrid[y][x]
			if y == 0 || cageGrid[y-1][x] != cage {
				masks[y][x] |= boundaryTop
			}
			if x == width-1 || cageGrid[y][x+1] != cage {
				masks[y][x] |= boundaryRight
			}
			if y == height-1 || cageGrid[y+1][x] != cage {
				masks[y][x] |= boundaryBottom
			}
			if x == 0 || cageGrid[y][x-1] != cage {
				masks[y][x] |= boundaryLeft
			}
		}
	}
	return masks
}

func orthogonalNeighbors(p point, width, height int) []point {
	neighbors := make([]point, 0, 4)
	if p.x > 0 {
		neighbors = append(neighbors, point{x: p.x - 1, y: p.y})
	}
	if p.x+1 < width {
		neighbors = append(neighbors, point{x: p.x + 1, y: p.y})
	}
	if p.y > 0 {
		neighbors = append(neighbors, point{x: p.x, y: p.y - 1})
	}
	if p.y+1 < height {
		neighbors = append(neighbors, point{x: p.x, y: p.y + 1})
	}
	return neighbors
}
