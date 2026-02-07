package wordsearch

// Direction represents one of 8 possible directions a word can be placed
type Direction int

const (
	Right Direction = iota
	Down
	DownRight
	DownLeft
	Left
	Up
	UpRight
	UpLeft
)

// Vector returns the dx, dy unit vector for a direction
func (d Direction) Vector() (int, int) {
	switch d {
	case Right:
		return 1, 0
	case Down:
		return 0, 1
	case DownRight:
		return 1, 1
	case DownLeft:
		return -1, 1
	case Left:
		return -1, 0
	case Up:
		return 0, -1
	case UpRight:
		return 1, -1
	case UpLeft:
		return -1, -1
	default:
		return 0, 0
	}
}

// Position represents a coordinate in the grid
type Position struct {
	X, Y int
}

// Word represents a hidden word in the puzzle
type Word struct {
	Text      string
	Start     Position
	End       Position
	Direction Direction
	Found     bool
}

// Contains checks if the given position is part of this word
func (w *Word) Contains(pos Position) bool {
	dx, dy := w.Direction.Vector()
	x, y := w.Start.X, w.Start.Y

	for i := 0; i < len(w.Text); i++ {
		if x == pos.X && y == pos.Y {
			return true
		}
		x += dx
		y += dy
	}
	return false
}

// Positions returns all positions occupied by this word
func (w *Word) Positions() []Position {
	positions := make([]Position, 0, len(w.Text))
	dx, dy := w.Direction.Vector()
	x, y := w.Start.X, w.Start.Y

	for i := 0; i < len(w.Text); i++ {
		positions = append(positions, Position{X: x, Y: y})
		x += dx
		y += dy
	}
	return positions
}
