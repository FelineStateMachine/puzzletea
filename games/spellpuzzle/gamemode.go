package spellpuzzle

import (
	_ "embed"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/gameentry"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

//go:embed help.md
var HelpContent string

type SpellPuzzleMode struct {
	game.BaseMode
	BankSize       int
	BoardWordCount int
	MinBonusWords  int
}

var (
	_ game.Mode          = SpellPuzzleMode{}
	_ game.Spawner       = SpellPuzzleMode{}
	_ game.SeededSpawner = SpellPuzzleMode{}
	_ game.EloSpawner    = SpellPuzzleMode{}
)

func NewMode(title, description string, bankSize, boardWords, minBonusWords int) SpellPuzzleMode {
	return SpellPuzzleMode{
		BaseMode:       game.NewBaseMode(title, description),
		BankSize:       bankSize,
		BoardWordCount: boardWords,
		MinBonusWords:  minBonusWords,
	}
}

func (m SpellPuzzleMode) Spawn() (game.Gamer, error) {
	puzzle, err := GeneratePuzzle(m)
	if err != nil {
		return nil, err
	}
	return New(m, puzzle)
}

func (m SpellPuzzleMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	puzzle, err := GeneratePuzzleSeeded(m, rng)
	if err != nil {
		return nil, err
	}
	return New(m, puzzle)
}

var Modes = []game.Mode{
	NewMode("Beginner", "6 letters, 4 board words, gentle intro.", 6, 4, 3),
	NewMode("Easy", "7 letters, 6 board words, balanced start.", 7, 6, 4),
	NewMode("Medium", "8 letters, 8 board words, denser layout.", 8, 8, 6),
	NewMode("Hard", "9 letters, 9 board words, largest launch board.", 9, 9, 8),
}

var ModeDefinitions = spellPuzzleModeDefinitions(Modes)

func spellPuzzleModeDefinitions(modes []game.Mode) []puzzle.ModeDef {
	defs := gameentry.BuildModeDefs(modes)
	presets := []difficulty.Elo{0, 600, 1500, 2400}
	for i := range defs {
		if i >= len(presets) {
			break
		}
		elo := presets[i]
		defs[i].PresetElo = &elo
	}
	return defs
}

var Definition = puzzle.NewDefinition(puzzle.DefinitionSpec{
	Name:         "Spell Puzzle",
	Description:  "Connect letters to fill a crossword with bonus anagrams.",
	Aliases:      []string{"spell", "spellpuzzle"},
	Modes:        ModeDefinitions,
	DailyModeIDs: puzzle.SelectModeIDsByIndex(ModeDefinitions, 0),
})

var Entry = gameentry.NewEntry(gameentry.EntrySpec{
	Definition: Definition,
	Help:       HelpContent,
	Import:     game.AdaptImport(ImportModel),
	Modes:      Modes,
	Print:      PDFPrintAdapter,
})
