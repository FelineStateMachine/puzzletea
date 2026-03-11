package spellpuzzle

import (
	"sort"

	"github.com/FelineStateMachine/puzzletea/game"
)

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type WordPlacement struct {
	Text        string      `json:"text"`
	Start       Position    `json:"start"`
	Orientation Orientation `json:"orientation"`
	Found       bool        `json:"found"`
}

func (w WordPlacement) Positions() []Position {
	positions := make([]Position, 0, len(w.Text))
	for i := range len(w.Text) {
		pos := w.Start
		if w.Orientation == Horizontal {
			pos.X += i
		} else {
			pos.Y += i
		}
		positions = append(positions, pos)
	}
	return positions
}

type boardCell struct {
	Letter     rune
	Occupied   bool
	Revealed   bool
	Horizontal bool
	Vertical   bool
	Owners     uint16
}

type board struct {
	Width  int
	Height int
	Cells  [][]boardCell
}

func buildBoard(placements []WordPlacement) board {
	if len(placements) == 0 {
		return board{}
	}

	minX, minY := placements[0].Start.X, placements[0].Start.Y
	maxX, maxY := placements[0].Start.X, placements[0].Start.Y
	for _, placement := range placements {
		for _, pos := range placement.Positions() {
			minX = min(minX, pos.X)
			minY = min(minY, pos.Y)
			maxX = max(maxX, pos.X)
			maxY = max(maxY, pos.Y)
		}
	}

	width := maxX - minX + 1
	height := maxY - minY + 1
	cells := make([][]boardCell, height)
	for y := range cells {
		cells[y] = make([]boardCell, width)
	}

	for placementIndex, placement := range placements {
		for i, letter := range placement.Text {
			x := placement.Start.X - minX
			y := placement.Start.Y - minY
			if placement.Orientation == Horizontal {
				x += i
			} else {
				y += i
			}

			cell := &cells[y][x]
			cell.Letter = letter
			cell.Occupied = true
			cell.Revealed = cell.Revealed || placement.Found
			if placement.Orientation == Horizontal {
				cell.Horizontal = true
			} else {
				cell.Vertical = true
			}
			cell.Owners |= 1 << placementIndex
		}
	}

	return board{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}

func (b board) cellAt(x, y int) boardCell {
	if y < 0 || y >= b.Height || x < 0 || x >= b.Width {
		return boardCell{}
	}
	return b.Cells[y][x]
}

func (b board) hasVerticalEdge(x, y int) bool {
	left := b.cellAt(x-1, y)
	right := b.cellAt(x, y)

	switch {
	case !left.Occupied && !right.Occupied:
		return false
	case !left.Occupied || !right.Occupied:
		return true
	default:
		return left.Owners&right.Owners == 0
	}
}

func (b board) hasHorizontalEdge(x, y int) bool {
	top := b.cellAt(x, y-1)
	bottom := b.cellAt(x, y)

	switch {
	case !top.Occupied && !bottom.Occupied:
		return false
	case !top.Occupied || !bottom.Occupied:
		return true
	default:
		return top.Owners&bottom.Owners == 0
	}
}

func (b board) bridgeFillState(bridge game.DynamicGridBridge) (fill, solved bool) {
	if bridge.Kind == game.DynamicGridBridgeJunction || bridge.Count < 2 {
		return false, false
	}

	var (
		sharedOwners uint16
		allRevealed  = true
	)
	for i := 0; i < bridge.Count; i++ {
		cell := b.cellAt(bridge.Cells[i].X, bridge.Cells[i].Y)
		if !cell.Occupied {
			return false, false
		}
		if i == 0 {
			sharedOwners = cell.Owners
		} else {
			sharedOwners &= cell.Owners
		}
		allRevealed = allRevealed && cell.Revealed
	}

	if sharedOwners == 0 {
		return false, false
	}
	return true, allRevealed
}

func normalizePlacements(placements []WordPlacement) []WordPlacement {
	if len(placements) == 0 {
		return nil
	}

	minX, minY := placements[0].Start.X, placements[0].Start.Y
	for _, placement := range placements {
		for _, pos := range placement.Positions() {
			minX = min(minX, pos.X)
			minY = min(minY, pos.Y)
		}
	}

	type placementKey struct {
		Text        string
		Start       Position
		Orientation Orientation
	}

	merged := make(map[placementKey]WordPlacement, len(placements))
	for _, placement := range placements {
		placement.Start.X -= minX
		placement.Start.Y -= minY
		key := placementKey{
			Text:        placement.Text,
			Start:       placement.Start,
			Orientation: placement.Orientation,
		}
		existing, ok := merged[key]
		if ok {
			existing.Found = existing.Found || placement.Found
			merged[key] = existing
			continue
		}
		merged[key] = placement
	}

	normalized := make([]WordPlacement, 0, len(merged))
	for _, placement := range merged {
		normalized = append(normalized, placement)
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		if len(normalized[i].Text) != len(normalized[j].Text) {
			return len(normalized[i].Text) > len(normalized[j].Text)
		}
		if normalized[i].Start.Y != normalized[j].Start.Y {
			return normalized[i].Start.Y < normalized[j].Start.Y
		}
		return normalized[i].Start.X < normalized[j].Start.X
	})

	return normalized
}
