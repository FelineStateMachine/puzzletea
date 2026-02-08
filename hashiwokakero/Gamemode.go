package hashiwokakero

import (
	"github.com/FelineStateMachine/puzzletea/game"
)

// HashiMode defines a hashiwokakero difficulty/configuration.
type HashiMode struct {
	title       string
	description string
	Width       int
	Height      int
	MinIslands  int
	MaxIslands  int
}

var _ game.Mode = HashiMode{} // compile-time interface check

func (h HashiMode) Title() string       { return "hashiwokakero\t" + h.title }
func (h HashiMode) Description() string { return h.description }
func (h HashiMode) FilterValue() string {
	return "hashiwokakero bridges hashi " + h.title + " " + h.description
}

func NewMode(title, description string, width, height, minIslands, maxIslands int) HashiMode {
	return HashiMode{
		title:       title,
		description: description,
		Width:       width,
		Height:      height,
		MinIslands:  minIslands,
		MaxIslands:  maxIslands,
	}
}
