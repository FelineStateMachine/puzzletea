package gameentry

import (
	"errors"
	"math/rand/v2"
	"testing"

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
