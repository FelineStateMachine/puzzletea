package nonogram

import (
	"github.com/FelineStateMachine/puzzletea/game"
)

type NonogramMode struct {
	title         string
	description   string
	Height, Width int
}

var _ game.Mode = NonogramMode{} // Verify that T implements I.

func (n NonogramMode) Title() string       { return "nonogram\t" + n.title }
func (n NonogramMode) Description() string { return n.description }
func (n NonogramMode) FilterValue() string { return "nonogram " + n.title + " " + n.description }

func NewMode(title, description string, Height, Width int) NonogramMode {
	return NonogramMode{
		title:       title,
		description: description,
		Height:      Height,
		Width:       Width,
	}
}
