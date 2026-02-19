package wordsearch

import (
	_ "embed"
	"math/rand/v2"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

func init() {
	game.Register("Word Search", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

type WordSearchMode struct {
	game.BaseMode
	Width       int
	Height      int
	WordCount   int
	MinWordLen  int
	MaxWordLen  int
	AllowedDirs []Direction
}

var (
	_ game.Mode          = WordSearchMode{}
	_ game.Spawner       = WordSearchMode{}
	_ game.SeededSpawner = WordSearchMode{}
)

func NewMode(title, description string, width, height, wordCount, minLen, maxLen int, allowedDirs []Direction) WordSearchMode {
	return WordSearchMode{
		BaseMode:    game.NewBaseMode(title, description),
		Width:       width,
		Height:      height,
		WordCount:   wordCount,
		MinWordLen:  minLen,
		MaxWordLen:  maxLen,
		AllowedDirs: allowedDirs,
	}
}

func (w WordSearchMode) Spawn() (game.Gamer, error) {
	grid, words := GenerateWordSearch(w.Width, w.Height, w.WordCount, w.MinWordLen, w.MaxWordLen, w.AllowedDirs)
	return New(w, grid, words)
}

func (w WordSearchMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	grid, words := GenerateWordSearchSeeded(w.Width, w.Height, w.WordCount, w.MinWordLen, w.MaxWordLen, w.AllowedDirs, rng)
	return New(w, grid, words)
}

var Modes = []list.Item{
	NewMode("Easy 10x10", "Find 6 words in a 10x10 grid.", 10, 10, 6, 3, 5, []Direction{Right, Down, DownRight}),
	NewMode("Medium 15x15", "Find 10 words in a 15x15 grid.", 15, 15, 10, 4, 7, []Direction{Right, Down, DownRight, DownLeft, Left, Up}),
	NewMode("Hard 20x20", "Find 15 words in a 20x20 grid.", 20, 20, 15, 5, 10, []Direction{Right, Down, DownRight, DownLeft, Left, Up, UpRight, UpLeft}),
}

var DailyModes = []list.Item{
	Modes[0], // Easy 10x10
}
