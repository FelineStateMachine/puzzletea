package takuzuplus

import (
	_ "embed"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

type TakuzuPlusMode struct {
	game.BaseMode
	Size      int
	Prefilled float64
	profile   relationProfile
}

var (
	_ game.Mode          = TakuzuPlusMode{}
	_ game.Spawner       = TakuzuPlusMode{}
	_ game.SeededSpawner = TakuzuPlusMode{}
)

func NewMode(title, desc string, size int, prefilled float64, profile relationProfile) TakuzuPlusMode {
	return TakuzuPlusMode{
		BaseMode:  game.NewBaseMode(title, desc),
		Size:      size,
		Prefilled: prefilled,
		profile:   profile,
	}
}

func (t TakuzuPlusMode) Spawn() (game.Gamer, error) {
	complete := generateComplete(t.Size)
	puzzle, provided, rels := generatePuzzle(complete, t.Size, t.Prefilled, t.profile)
	return New(t, puzzle, provided, rels)
}

func (t TakuzuPlusMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	complete := generateCompleteSeeded(t.Size, rng)
	puzzle, provided, rels := generatePuzzleSeeded(complete, t.Size, t.Prefilled, t.profile, rng)
	return New(t, puzzle, provided, rels)
}

var modeProfiles = relationProfiles()

var Modes = []game.Mode{
	NewMode("Beginner", "6×6 grid, relation clues ease early logic.", 6, 0.50, modeProfiles[0]),
	NewMode("Easy", "6×6 grid with additive = and x clues.", 6, 0.40, modeProfiles[1]),
	NewMode("Medium", "8×8 grid, mixed Takuzu and relation deductions.", 8, 0.40, modeProfiles[2]),
	NewMode("Tricky", "10×10 grid, uniqueness and relation clues interact.", 10, 0.38, modeProfiles[3]),
	NewMode("Hard", "10×10 grid, longer deduction chains with relation clues.", 10, 0.32, modeProfiles[4]),
	NewMode("Very Hard", "12×12 grid, sparse givens plus relation clues.", 12, 0.30, modeProfiles[5]),
	NewMode("Extreme", "14×14 grid, maximum size with additive relation clues.", 14, 0.28, modeProfiles[6]),
}

var Definition = game.NewDefinition(game.DefinitionSpec{
	Name:             "Takuzu+",
	Description:      "Fill the grid with ● and ○ using some relational clues. No 3 in a row.",
	Aliases:          []string{"takuzu plus", "binario+", "binario plus"},
	Modes:            Modes,
	DailyModeIndexes: []int{2, 3},
	Help:             HelpContent,
	Import:           game.AdaptImport(ImportModel),
})
