// Package shikaku implements the Shikaku rectangle-partition puzzle game.
package shikaku

// Clue represents a numbered cell in the grid.
type Clue struct {
	ID    int `json:"id"`
	X     int `json:"x"`
	Y     int `json:"y"`
	Value int `json:"value"` // required rectangle area
}

// Rectangle represents a player-placed rectangle covering grid cells.
type Rectangle struct {
	ClueID int `json:"clue_id"`
	X      int `json:"x"` // top-left
	Y      int `json:"y"`
	W      int `json:"w"` // dimensions
	H      int `json:"h"`
}

// Area returns the area of the rectangle.
func (r Rectangle) Area() int { return r.W * r.H }

// Contains reports whether the rectangle covers the cell at (cx, cy).
func (r Rectangle) Contains(cx, cy int) bool {
	return cx >= r.X && cx < r.X+r.W && cy >= r.Y && cy < r.Y+r.H
}

// Puzzle holds the full state of a Shikaku game.
type Puzzle struct {
	Width, Height int
	Clues         []Clue
	Rectangles    []Rectangle // player-placed
	clueIndex     map[[2]int]int
}

// FindClueAt returns the clue at position (x, y), or nil.
func (p *Puzzle) FindClueAt(x, y int) *Clue {
	if p.clueIndex == nil {
		p.clueIndex = make(map[[2]int]int, len(p.Clues))
		for i, c := range p.Clues {
			p.clueIndex[[2]int{c.X, c.Y}] = i
		}
	}
	if idx, ok := p.clueIndex[[2]int{x, y}]; ok {
		return &p.Clues[idx]
	}
	return nil
}

// FindClueByID returns the clue with the given ID, or nil.
func (p *Puzzle) FindClueByID(id int) *Clue {
	if id >= 0 && id < len(p.Clues) && p.Clues[id].ID == id {
		return &p.Clues[id]
	}
	for i := range p.Clues {
		if p.Clues[i].ID == id {
			return &p.Clues[i]
		}
	}
	return nil
}

// FindRectangleForClue returns the rectangle placed for the given clue ID, or nil.
func (p *Puzzle) FindRectangleForClue(clueID int) *Rectangle {
	for i := range p.Rectangles {
		if p.Rectangles[i].ClueID == clueID {
			return &p.Rectangles[i]
		}
	}
	return nil
}

// CellOwner returns the clue ID of the rectangle covering (x, y), or -1.
func (p *Puzzle) CellOwner(x, y int) int {
	for _, r := range p.Rectangles {
		if r.Contains(x, y) {
			return r.ClueID
		}
	}
	return -1
}

// Overlaps reports whether the given rectangle overlaps any existing rectangle,
// optionally excluding the rectangle belonging to excludeClueID.
func (p *Puzzle) Overlaps(rect Rectangle, excludeClueID int) bool {
	for _, r := range p.Rectangles {
		if r.ClueID == excludeClueID {
			continue
		}
		if rectsOverlap(rect, r) {
			return true
		}
	}
	return false
}

func rectsOverlap(a, b Rectangle) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

// SetRectangle places or replaces the rectangle for a clue.
func (p *Puzzle) SetRectangle(rect Rectangle) {
	for i, r := range p.Rectangles {
		if r.ClueID == rect.ClueID {
			p.Rectangles[i] = rect
			return
		}
	}
	p.Rectangles = append(p.Rectangles, rect)
}

// RemoveRectangle removes the rectangle for the given clue ID.
func (p *Puzzle) RemoveRectangle(clueID int) {
	for i, r := range p.Rectangles {
		if r.ClueID == clueID {
			p.Rectangles = append(p.Rectangles[:i], p.Rectangles[i+1:]...)
			return
		}
	}
}

// IsSolved reports whether the puzzle is complete:
// every cell is covered exactly once and each rectangle contains
// exactly one clue whose value equals the rectangle's area.
func (p *Puzzle) IsSolved() bool {
	if len(p.Rectangles) != len(p.Clues) {
		return false
	}

	// Check each rectangle has exactly one clue with matching area.
	for _, r := range p.Rectangles {
		clue := p.FindClueByID(r.ClueID)
		if clue == nil {
			return false
		}
		if r.Area() != clue.Value {
			return false
		}
		if !r.Contains(clue.X, clue.Y) {
			return false
		}
		// Check no other clue is inside this rectangle.
		clueCount := 0
		for _, c := range p.Clues {
			if r.Contains(c.X, c.Y) {
				clueCount++
			}
		}
		if clueCount != 1 {
			return false
		}
	}

	// Check all cells are covered exactly once (no gaps, no overlaps).
	covered := make([][]bool, p.Height)
	for y := range p.Height {
		covered[y] = make([]bool, p.Width)
	}
	for _, r := range p.Rectangles {
		for dy := range r.H {
			for dx := range r.W {
				cx, cy := r.X+dx, r.Y+dy
				if cx < 0 || cx >= p.Width || cy < 0 || cy >= p.Height {
					return false
				}
				if covered[cy][cx] {
					return false // overlap
				}
				covered[cy][cx] = true
			}
		}
	}
	for y := range p.Height {
		for x := range p.Width {
			if !covered[y][x] {
				return false
			}
		}
	}

	return true
}
