package game

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Cursor represents a 2D grid cursor position.
type Cursor struct {
	X, Y int
}

// Move updates the cursor position based on the key message and bounds.
// Returns true if the cursor moved.
func (c *Cursor) Move(keys CursorKeyMap, msg tea.KeyPressMsg, maxX, maxY int) bool {
	switch {
	case key.Matches(msg, keys.Up):
		if c.Y > 0 {
			c.Y--
			return true
		}
	case key.Matches(msg, keys.Down):
		if c.Y < maxY {
			c.Y++
			return true
		}
	case key.Matches(msg, keys.Left):
		if c.X > 0 {
			c.X--
			return true
		}
	case key.Matches(msg, keys.Right):
		if c.X < maxX {
			c.X++
			return true
		}
	}
	return false
}
