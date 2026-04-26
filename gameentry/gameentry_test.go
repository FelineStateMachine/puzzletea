package gameentry

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

type stubMode struct {
	game.BaseMode
}

func (m stubMode) Spawn() (game.Gamer, error) {
	return nil, errors.New("not implemented")
}

type seededStubMode struct {
	stubMode
}

func (m seededStubMode) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	return nil, errors.New("not implemented")
}

type cancellableEloStubMode struct {
	stubMode
	calledWithCanceledContext bool
}

func (m *cancellableEloStubMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	return nil, difficulty.Report{}, errors.New("non-context spawn called")
}

func (m *cancellableEloStubMode) SpawnEloContext(ctx context.Context, seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := ctx.Err(); err != nil {
		m.calledWithCanceledContext = true
		return nil, difficulty.Report{}, err
	}
	return nil, difficulty.Report{}, nil
}

func TestBuildModeDefsMarksSeededModes(t *testing.T) {
	modes := []game.Mode{
		stubMode{BaseMode: game.NewBaseMode("Beginner", "Easy board")},
		seededStubMode{stubMode{BaseMode: game.NewBaseMode("Expert", "Hard board")}},
	}

	defs := BuildModeDefs(modes)
	if len(defs) != 2 {
		t.Fatalf("len(BuildModeDefs) = %d, want 2", len(defs))
	}
	if defs[0].Seeded {
		t.Fatal("defs[0].Seeded = true, want false")
	}
	if !defs[1].Seeded {
		t.Fatal("defs[1].Seeded = false, want true")
	}
}

func TestNewEntryKeepsMetadataAndRuntimeModesAligned(t *testing.T) {
	modes := []game.Mode{
		seededStubMode{stubMode{BaseMode: game.NewBaseMode("Medium", "Balanced board")}},
	}
	definition := puzzle.NewDefinition(puzzle.DefinitionSpec{
		Name:         "Sudoku",
		Modes:        BuildModeDefs(modes),
		DailyModeIDs: []puzzle.ModeID{"medium"},
	})

	entry := NewEntry(EntrySpec{
		Definition: definition,
		Help:       "Rules",
		Import:     func([]byte) (game.Gamer, error) { return nil, nil },
		Modes:      modes,
	})

	if got, want := entry.Definition.Name, "Sudoku"; got != want {
		t.Fatalf("Definition.Name = %q, want %q", got, want)
	}
	if got, want := len(entry.Modes), 1; got != want {
		t.Fatalf("len(Modes) = %d, want %d", got, want)
	}
	if entry.Modes[0].Seeded == nil {
		t.Fatal("Modes[0].Seeded = nil, want seeded spawner")
	}
	if entry.Print != nil {
		t.Fatal("Print = non-nil, want nil")
	}
}

func TestVariantEloSpawnerPreservesCancellation(t *testing.T) {
	preset := difficulty.Elo(1200)
	mode := &cancellableEloStubMode{
		stubMode: stubMode{BaseMode: game.NewBaseMode("Hard", "Hard board")},
	}
	entry := NewEntry(EntrySpec{
		Definition: puzzle.NewDefinition(puzzle.DefinitionSpec{
			Name: "Test",
			Modes: []puzzle.ModeDef{puzzle.NewModeDef(puzzle.ModeSpec{
				Title:       mode.Title(),
				Description: mode.Description(),
				PresetElo:   &preset,
			})},
		}),
		Modes: []game.Mode{mode},
	})
	if len(entry.Variants) != 1 {
		t.Fatalf("len(Variants) = %d, want 1", len(entry.Variants))
	}
	spawner, ok := entry.Variants[0].Elo.(game.CancellableEloSpawner)
	if !ok {
		t.Fatal("variant Elo spawner does not implement game.CancellableEloSpawner")
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, err := spawner.SpawnEloContext(ctx, "seed", preset)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SpawnEloContext error = %v, want context.Canceled", err)
	}
	if !mode.calledWithCanceledContext {
		t.Fatal("underlying mode did not receive the canceled context")
	}
}

func TestVariantSeededSpawnerPreservesCancellation(t *testing.T) {
	preset := difficulty.Elo(1200)
	mode := &cancellableEloStubMode{
		stubMode: stubMode{BaseMode: game.NewBaseMode("Hard", "Hard board")},
	}
	entry := NewEntry(EntrySpec{
		Definition: puzzle.NewDefinition(puzzle.DefinitionSpec{
			Name: "Test",
			Modes: []puzzle.ModeDef{puzzle.NewModeDef(puzzle.ModeSpec{
				Title:       mode.Title(),
				Description: mode.Description(),
				PresetElo:   &preset,
			})},
		}),
		Modes: []game.Mode{mode},
	})
	spawner, ok := entry.Variants[0].Seeded.(game.CancellableSeededSpawner)
	if !ok {
		t.Fatal("variant seeded spawner does not implement game.CancellableSeededSpawner")
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := spawner.SpawnSeededContext(ctx, rand.New(rand.NewPCG(1, 2)))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SpawnSeededContext error = %v, want context.Canceled", err)
	}
	if !mode.calledWithCanceledContext {
		t.Fatal("underlying mode did not receive the canceled context")
	}
}
