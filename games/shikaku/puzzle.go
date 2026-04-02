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

const unownedRegionID = -1

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

// ValidRectangleForClue reports whether rect is a legal placement for clueID.
func (p *Puzzle) ValidRectangleForClue(rect Rectangle, clueID int) bool {
	clue := p.FindClueByID(clueID)
	if clue == nil {
		return false
	}
	if rect.Area() != clue.Value {
		return false
	}
	if rect.X < 0 || rect.Y < 0 || rect.X+rect.W > p.Width || rect.Y+rect.H > p.Height {
		return false
	}
	if !rect.Contains(clue.X, clue.Y) {
		return false
	}
	if p.Overlaps(rect, clueID) {
		return false
	}

	clueCount := 0
	for _, c := range p.Clues {
		if rect.Contains(c.X, c.Y) {
			clueCount++
			if c.ID != clueID {
				return false
			}
		}
	}

	return clueCount == 1
}

// CandidateRectanglesForClue returns every legal rectangle for the clue.
func (p *Puzzle) CandidateRectanglesForClue(clueID int) []Rectangle {
	clue := p.FindClueByID(clueID)
	if clue == nil {
		return nil
	}

	area := clue.Value
	candidates := make([]Rectangle, 0, area)
	for h := 1; h <= area; h++ {
		if area%h != 0 {
			continue
		}
		w := area / h
		if w > p.Width || h > p.Height {
			continue
		}

		minX := max(0, clue.X-w+1)
		maxX := min(clue.X, p.Width-w)
		minY := max(0, clue.Y-h+1)
		maxY := min(clue.Y, p.Height-h)
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				rect := Rectangle{ClueID: clueID, X: x, Y: y, W: w, H: h}
				if p.ValidRectangleForClue(rect, clueID) {
					candidates = append(candidates, rect)
				}
			}
		}
	}

	return candidates
}

func (p *Puzzle) encasedRectangleForClue(clueID int) (Rectangle, bool) {
	clue := p.FindClueByID(clueID)
	if clue == nil || p.FindRectangleForClue(clueID) != nil {
		return Rectangle{}, false
	}
	if p.CellOwner(clue.X, clue.Y) != unownedRegionID {
		return Rectangle{}, false
	}

	cells := p.unownedRegionFrom(clue.X, clue.Y)
	if len(cells) == 0 {
		return Rectangle{}, false
	}

	minX, maxX := cells[0][0], cells[0][0]
	minY, maxY := cells[0][1], cells[0][1]
	for _, cell := range cells[1:] {
		minX = min(minX, cell[0])
		maxX = max(maxX, cell[0])
		minY = min(minY, cell[1])
		maxY = max(maxY, cell[1])
	}

	rect := Rectangle{
		ClueID: clueID,
		X:      minX,
		Y:      minY,
		W:      maxX - minX + 1,
		H:      maxY - minY + 1,
	}
	if rect.Area() != len(cells) {
		return Rectangle{}, false
	}
	if !p.ValidRectangleForClue(rect, clueID) {
		return Rectangle{}, false
	}
	return rect, true
}

func (p *Puzzle) forcedRectangleForClue(clueID int) (Rectangle, bool) {
	clue := p.FindClueByID(clueID)
	if clue == nil || p.FindRectangleForClue(clueID) != nil {
		return Rectangle{}, false
	}

	if clue.Value == 1 {
		rect := Rectangle{
			ClueID: clueID,
			X:      clue.X,
			Y:      clue.Y,
			W:      1,
			H:      1,
		}
		if p.ValidRectangleForClue(rect, clueID) {
			return rect, true
		}
		return Rectangle{}, false
	}

	return p.encasedRectangleForClue(clueID)
}

func (p *Puzzle) unownedRegionFrom(x, y int) [][2]int {
	if x < 0 || x >= p.Width || y < 0 || y >= p.Height || p.CellOwner(x, y) != unownedRegionID {
		return nil
	}

	visited := make(map[[2]int]struct{}, p.Width*p.Height)
	queue := [][2]int{{x, y}}
	visited[[2]int{x, y}] = struct{}{}
	cells := make([][2]int, 0, p.Width*p.Height)

	for len(queue) > 0 {
		cell := queue[0]
		queue = queue[1:]
		cells = append(cells, cell)

		cx, cy := cell[0], cell[1]
		neighbors := [][2]int{
			{cx - 1, cy},
			{cx + 1, cy},
			{cx, cy - 1},
			{cx, cy + 1},
		}
		for _, next := range neighbors {
			nx, ny := next[0], next[1]
			key := [2]int{nx, ny}
			if nx < 0 || nx >= p.Width || ny < 0 || ny >= p.Height {
				continue
			}
			if p.CellOwner(nx, ny) != unownedRegionID {
				continue
			}
			if _, ok := visited[key]; ok {
				continue
			}
			visited[key] = struct{}{}
			queue = append(queue, key)
		}
	}

	return cells
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

// CluesInRect returns all clues contained within the given rectangle bounds.
func (p *Puzzle) CluesInRect(r Rectangle) []*Clue {
	var result []*Clue
	for i := range p.Clues {
		if r.Contains(p.Clues[i].X, p.Clues[i].Y) {
			result = append(result, &p.Clues[i])
		}
	}
	return result
}

// autoPlaceForcedRectangles places uniquely forced rectangles until stable.
func (p *Puzzle) autoPlaceForcedRectangles() bool {
	changedAny := false
	for {
		changed := false
		for _, c := range p.Clues {
			rect, ok := p.forcedRectangleForClue(c.ID)
			if !ok {
				continue
			}

			p.SetRectangle(rect)
			changed = true
			changedAny = true
		}
		if !changed {
			return changedAny
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
