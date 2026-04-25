package netwalk

import (
	_ "embed"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/gameentry"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

//go:embed help.md
var HelpContent string

type NetwalkMode struct {
	game.BaseMode
	Size      int
	FillRatio float64
	Profile   generateProfile
}

var (
	_ game.Mode          = NetwalkMode{}
	_ game.Spawner       = NetwalkMode{}
	_ game.SeededSpawner = NetwalkMode{}
)

func NewMode(title, desc string, size int, fillRatio float64, profile generateProfile) NetwalkMode {
	return NetwalkMode{
		BaseMode:  game.NewBaseMode(title, desc),
		Size:      size,
		FillRatio: fillRatio,
		Profile:   profile,
	}
}

func (m NetwalkMode) Spawn() (game.Gamer, error) {
	p, err := GenerateSeededWithDensity(
		m.Size,
		m.FillRatio,
		m.Profile,
		rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	)
	if err != nil {
		return nil, err
	}
	return New(m, p)
}

func (m NetwalkMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	p, err := GenerateSeededWithDensity(m.Size, m.FillRatio, m.Profile, rng)
	if err != nil {
		return nil, err
	}
	return New(m, p)
}

var (
	miniProfile = generateProfile{
		ParentDegreeWeights:    [5]int{16, 16, 6, -8, -12},
		OrthogonalPackedWeight: 2,
		DiagonalPackedWeight:   1,
		SpanGrowthWeight:       12,
		MinSpanRatio:           0.55,
	}
	easyProfile = generateProfile{
		ParentDegreeWeights:    [5]int{14, 14, 8, -4, -8},
		OrthogonalPackedWeight: 4,
		DiagonalPackedWeight:   2,
		SpanGrowthWeight:       14,
		MinSpanRatio:           0.62,
	}
	mediumProfile = generateProfile{
		ParentDegreeWeights:    [5]int{10, 10, 12, 2, -4},
		OrthogonalPackedWeight: 6,
		DiagonalPackedWeight:   3,
		SpanGrowthWeight:       14,
		MinSpanRatio:           0.70,
	}
	hardProfile = generateProfile{
		ParentDegreeWeights:    [5]int{6, 6, 16, 8, 0},
		OrthogonalPackedWeight: 8,
		DiagonalPackedWeight:   4,
		SpanGrowthWeight:       16,
		MinSpanRatio:           0.78,
	}
	expertProfile = generateProfile{
		ParentDegreeWeights:    [5]int{4, 4, 18, 10, 2},
		OrthogonalPackedWeight: 10,
		DiagonalPackedWeight:   5,
		SpanGrowthWeight:       18,
		MinSpanRatio:           0.84,
	}
)

var Modes = []game.Mode{
	NewMode("Mini 5x5", "Compact 5×5 network with a denser starter tree.", 5, 0.50, miniProfile),
	NewMode("Easy 7x7", "7×7 board with fuller coverage and gentle local tangles.", 7, 0.57, easyProfile),
	NewMode("Medium 9x9", "Balanced 9×9 network with tighter clusters and more branching.", 9, 0.64, mediumProfile),
	NewMode("Hard 11x11", "Dense 11×11 board that packs branches into close local interactions.", 11, 0.72, hardProfile),
	NewMode("Expert 13x13", "Large, crowded network with heavy branching and frequent near-miss tangles.", 13, 0.78, expertProfile),
}

var ModeDefinitions = gameentry.BuildModeDefs(Modes)

var Definition = puzzle.NewDefinition(puzzle.DefinitionSpec{
	Name:         "Netwalk",
	Description:  "Rotate network tiles until every computer connects to the server.",
	Aliases:      []string{"network"},
	Modes:        ModeDefinitions,
	DailyModeIDs: puzzle.SelectModeIDsByIndex(ModeDefinitions, 1, 2),
})

var Entry = gameentry.NewEntry(gameentry.EntrySpec{
	Definition: Definition,
	Help:       HelpContent,
	Import:     game.AdaptImport(ImportModel),
	Modes:      Modes,
	Print:      PDFPrintAdapter,
})
