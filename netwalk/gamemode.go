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
	Size         int
	TargetActive int
}

var (
	_ game.Mode          = NetwalkMode{}
	_ game.Spawner       = NetwalkMode{}
	_ game.SeededSpawner = NetwalkMode{}
)

func NewMode(title, desc string, size, targetActive int) NetwalkMode {
	return NetwalkMode{
		BaseMode:     game.NewBaseMode(title, desc),
		Size:         size,
		TargetActive: targetActive,
	}
}

func (m NetwalkMode) Spawn() (game.Gamer, error) {
	p, err := Generate(m.Size, m.TargetActive)
	if err != nil {
		return nil, err
	}
	return New(m, p)
}

func (m NetwalkMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	p, err := GenerateSeeded(m.Size, m.TargetActive, rng)
	if err != nil {
		return nil, err
	}
	return New(m, p)
}

var Modes = []game.Mode{
	NewMode("Mini 5x5", "Compact 5×5 network. Good first deduction pass.", 5, 8),
	NewMode("Easy 7x7", "7×7 board with a modest tree and clear local constraints.", 7, 14),
	NewMode("Medium 9x9", "Larger network with more branches and ambiguous elbows.", 9, 22),
	NewMode("Hard 11x11", "Dense mid-size network that rewards global checking.", 11, 30),
	NewMode("Expert 13x13", "Longer branch interactions and more disconnected-looking scrambles.", 13, 40),
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
