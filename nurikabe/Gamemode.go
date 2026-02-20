package nurikabe

import (
	"context"
	_ "embed"
	"fmt"
	"math/rand/v2"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

//go:embed help.md
var HelpContent string

func init() {
	game.Register("Nurikabe", func(data []byte) (game.Gamer, error) {
		return ImportModel(data)
	})
}

type NurikabeMode struct {
	game.BaseMode
	Width         int
	Height        int
	ClueDensity   float64
	MaxIslandSize int
}

var (
	_ game.Mode                     = NurikabeMode{}
	_ game.Spawner                  = NurikabeMode{}
	_ game.SeededSpawner            = NurikabeMode{}
	_ game.CancellableSpawner       = NurikabeMode{}
	_ game.CancellableSeededSpawner = NurikabeMode{}
)

func NewMode(title, desc string, width, height int, clueDensity float64, maxIslandSize int) NurikabeMode {
	return NurikabeMode{
		BaseMode:      game.NewBaseMode(title, desc),
		Width:         width,
		Height:        height,
		ClueDensity:   clueDensity,
		MaxIslandSize: maxIslandSize,
	}
}

func (n NurikabeMode) Spawn() (game.Gamer, error) {
	return n.SpawnContext(context.Background())
}

func (n NurikabeMode) SpawnContext(ctx context.Context) (game.Gamer, error) {
	p, err := GenerateWithContext(ctx, n)
	if err != nil {
		return nil, err
	}
	return New(n, p)
}

func (n NurikabeMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	return n.SpawnSeededContext(context.Background(), rng)
}

func (n NurikabeMode) SpawnSeededContext(ctx context.Context, rng *rand.Rand) (game.Gamer, error) {
	if rng == nil {
		return nil, fmt.Errorf("nil RNG")
	}
	p, err := GenerateSeededWithContext(ctx, n, rng)
	if err != nil {
		return nil, err
	}
	return New(n, p)
}

var Modes = []list.Item{
	NewMode("Mini", "5x5 grid, gentle introduction.", 5, 5, 0.28, 5),
	NewMode("Easy", "7x7 grid, balanced logic.", 7, 7, 0.24, 7),
	NewMode("Medium", "9x9 grid, moderate deduction.", 9, 9, 0.20, 9),
	NewMode("Hard", "11x11 grid, lower clue density.", 11, 11, 0.16, 11),
	NewMode("Expert", "12x12 grid, sparse clues and long chains.", 12, 12, 0.14, 12),
}

var DailyModes = []list.Item{
	Modes[1], // Easy
	Modes[2], // Medium
}
