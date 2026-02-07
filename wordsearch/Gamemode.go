package wordsearch

import "github.com/FelineStateMachine/puzzletea/game"

// WordSearchMode implements game.Mode for word search puzzles
type WordSearchMode struct {
	title       string
	description string
	Width       int
	Height      int
	WordCount   int
	MinWordLen  int
	MaxWordLen  int
	AllowedDirs []Direction
}

// NewMode creates a new WordSearchMode with the given parameters
func NewMode(title, description string, width, height, wordCount, minLen, maxLen int, allowedDirs []Direction) game.Mode {
	return WordSearchMode{
		title:       title,
		description: description,
		Width:       width,
		Height:      height,
		WordCount:   wordCount,
		MinWordLen:  minLen,
		MaxWordLen:  maxLen,
		AllowedDirs: allowedDirs,
	}
}

func (m WordSearchMode) Title() string {
	return m.title
}

func (m WordSearchMode) Description() string {
	return m.description
}

func (m WordSearchMode) FilterValue() string {
	return m.title
}
