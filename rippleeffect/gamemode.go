package rippleeffect

import (
	_ "embed"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

type Mode struct {
	game.BaseMode
	Size       int
	MaxCage    int
	GivenRatio float64
	profile    generationProfile
}

var (
	_ game.Mode          = Mode{}
	_ game.Spawner       = Mode{}
	_ game.SeededSpawner = Mode{}
)

func NewMode(title, description string, size, maxCage int, givenRatio float64) Mode {
	return NewModeWithProfile(title, description, size, maxCage, givenRatio, defaultGenerationProfile(maxCage))
}

func NewModeWithProfile(title, description string, size, maxCage int, givenRatio float64, profile generationProfile) Mode {
	return Mode{
		BaseMode:   game.NewBaseMode(title, description),
		Size:       size,
		MaxCage:    maxCage,
		GivenRatio: givenRatio,
		profile:    profile,
	}
}

func (m Mode) Spawn() (game.Gamer, error) {
	puzzle, err := m.generatePuzzle()
	if err != nil {
		return nil, err
	}
	return New(m, puzzle)
}

func (m Mode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	puzzle, err := m.generatePuzzleSeeded(rng)
	if err != nil {
		return nil, err
	}
	return New(m, puzzle)
}

var Modes = []game.Mode{
	NewModeWithProfile(
		"Mini 5x5",
		"Compact cages with extra anchors for quick local reads.",
		5,
		3,
		0.69,
		generationProfile{
			cageWeights:     []int{0, 5, 6, 3},
			frontierSamples: 3,
			shapeBias:       shapeBiasCompact,
			minGivensByCage: []int{0, 1, 1, 2},
		},
	),
	NewModeWithProfile(
		"Easy 6x6",
		"Gentle spacing logic with compact cages and steady clues.",
		6,
		3,
		0.64,
		generationProfile{
			cageWeights:     []int{0, 3, 5, 4},
			frontierSamples: 3,
			shapeBias:       shapeBiasCompact,
			minGivensByCage: []int{0, 1, 1, 2},
		},
	),
	NewModeWithProfile(
		"Medium 7x7",
		"Mixed cage shapes with a balanced clue spread.",
		7,
		4,
		0.59,
		generationProfile{
			cageWeights:     []int{0, 2, 4, 5, 3},
			frontierSamples: 2,
			shapeBias:       shapeBiasNeutral,
			minGivensByCage: []int{0, 1, 1, 1, 1},
		},
	),
	NewModeWithProfile(
		"Hard 8x8",
		"Longer cages and lighter anchors create broader ripple scans.",
		8,
		4,
		0.55,
		generationProfile{
			cageWeights:     []int{0, 1, 2, 4, 5},
			frontierSamples: 3,
			shapeBias:       shapeBiasWinding,
			minGivensByCage: []int{0, 1, 1, 1, 1},
		},
	),
	NewModeWithProfile(
		"Expert 9x9",
		"Sparser anchors and winding large cages push global deductions.",
		9,
		5,
		0.51,
		generationProfile{
			cageWeights:     []int{0, 1, 1, 3, 4, 5},
			frontierSamples: 4,
			shapeBias:       shapeBiasWinding,
			minGivensByCage: []int{0, 1, 1, 1, 1, 1},
		},
	),
}

var DailyModes = []game.Mode{
	Modes[1],
	Modes[2],
	Modes[3],
}

var Definition = game.Definition{
	Name:        "Ripple Effect",
	Description: "Fill the cages with sequential numbers without violating ripple distance.",
	Aliases:     []string{"ripple"},
	Modes:       Modes,
	DailyModes:  DailyModes,
	Help:        HelpContent,
	Import:      func(data []byte) (game.Gamer, error) { return ImportModel(data) },
}

func (m Mode) generatePuzzle() (Puzzle, error) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return m.generatePuzzleSeeded(rng)
}

func (m Mode) generatePuzzleSeeded(rng *rand.Rand) (Puzzle, error) {
	return generatePuzzleSeededWithProfile(m.Size, m.Size, m.MaxCage, m.GivenRatio, m.profile.withDefaults(m.MaxCage), rng)
}
