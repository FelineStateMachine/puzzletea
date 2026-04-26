package cmd

import (
	"math/rand/v2"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
)

type testSpawnGamer struct{}

func (testSpawnGamer) GetDebugInfo() string                 { return "" }
func (testSpawnGamer) GetFullHelp() [][]key.Binding         { return nil }
func (testSpawnGamer) GetSave() ([]byte, error)             { return []byte(`{}`), nil }
func (testSpawnGamer) IsSolved() bool                       { return false }
func (testSpawnGamer) Reset() game.Gamer                    { return testSpawnGamer{} }
func (testSpawnGamer) SetTitle(string) game.Gamer           { return testSpawnGamer{} }
func (testSpawnGamer) Init() tea.Cmd                        { return nil }
func (testSpawnGamer) View() string                         { return "" }
func (testSpawnGamer) Update(tea.Msg) (game.Gamer, tea.Cmd) { return testSpawnGamer{}, nil }

type testLegacySpawner struct {
	spawned      bool
	seeded       bool
	seededCalled bool
}

func (s *testLegacySpawner) Spawn() (game.Gamer, error) {
	s.spawned = true
	return testSpawnGamer{}, nil
}

func (s *testLegacySpawner) SpawnSeeded(*rand.Rand) (game.Gamer, error) {
	s.seeded = true
	s.seededCalled = true
	return testSpawnGamer{}, nil
}

type testEloSpawner struct {
	testLegacySpawner
	called bool
	seed   string
	elo    difficulty.Elo
}

func (s *testEloSpawner) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	s.called = true
	s.seed = seed
	s.elo = elo
	return testSpawnGamer{}, difficulty.Report{
		TargetElo:  elo,
		ActualElo:  elo,
		Confidence: difficulty.ConfidenceHigh,
	}, nil
}

func TestSpawnFromModeUsesPresetEloWhenDifficultyOmitted(t *testing.T) {
	preset := difficulty.Elo(1200)
	spawner := &testEloSpawner{}
	mode := registry.ModeEntry{
		Definition: puzzle.ModeDef{Title: "Preset", PresetElo: &preset},
		Spawner:    spawner,
		Elo:        spawner,
	}

	_, report, err := spawnFromMode(mode, "", nil, "generated-name")
	if err != nil {
		t.Fatalf("spawnFromMode returned error: %v", err)
	}
	if !spawner.called {
		t.Fatal("expected SpawnElo to be called")
	}
	if spawner.seed != "generated-name" {
		t.Fatalf("Elo seed = %q, want generated save name", spawner.seed)
	}
	if spawner.elo != preset || report.TargetElo != preset || report.Confidence == "" {
		t.Fatalf("Elo/report = (%d, %#v), want preset report", spawner.elo, report)
	}
}

func TestSpawnFromVariantUsesDefaultEloWhenDifficultyOmitted(t *testing.T) {
	defaultElo := difficulty.Elo(1200)
	spawner := &testEloSpawner{}
	variant := registry.VariantEntry{
		Definition: puzzle.VariantDef{Title: "Rule", DefaultElo: defaultElo},
		Elo:        spawner,
	}

	_, report, err := spawnFromVariant(variant, "", nil, "generated-name")
	if err != nil {
		t.Fatalf("spawnFromVariant returned error: %v", err)
	}
	if !spawner.called {
		t.Fatal("expected SpawnElo to be called")
	}
	if spawner.seed != "generated-name" {
		t.Fatalf("Elo seed = %q, want generated save name", spawner.seed)
	}
	if spawner.elo != defaultElo || report.TargetElo != defaultElo || report.Confidence == "" {
		t.Fatalf("Elo/report = (%d, %#v), want default report", spawner.elo, report)
	}
}

func TestSpawnFromVariantExplicitDifficultyOverridesDefault(t *testing.T) {
	explicit := difficulty.Elo(2200)
	spawner := &testEloSpawner{}
	variant := registry.VariantEntry{
		Definition: puzzle.VariantDef{Title: "Rule", DefaultElo: 800},
		Elo:        spawner,
	}

	_, _, err := spawnFromVariant(variant, "user-seed", &explicit, "generated-name")
	if err != nil {
		t.Fatalf("spawnFromVariant returned error: %v", err)
	}
	if spawner.elo != explicit {
		t.Fatalf("Elo = %d, want explicit %d", spawner.elo, explicit)
	}
	if spawner.seed != "user-seed" {
		t.Fatalf("seed = %q, want explicit seed", spawner.seed)
	}
}

func TestSpawnFromModeExplicitDifficultyOverridesPreset(t *testing.T) {
	preset := difficulty.Elo(800)
	explicit := difficulty.Elo(2200)
	spawner := &testEloSpawner{}
	mode := registry.ModeEntry{
		Definition: puzzle.ModeDef{Title: "Preset", PresetElo: &preset},
		Spawner:    spawner,
		Elo:        spawner,
	}

	_, _, err := spawnFromMode(mode, "user-seed", &explicit, "generated-name")
	if err != nil {
		t.Fatalf("spawnFromMode returned error: %v", err)
	}
	if spawner.elo != explicit {
		t.Fatalf("Elo = %d, want explicit %d", spawner.elo, explicit)
	}
	if spawner.seed != "user-seed" {
		t.Fatalf("seed = %q, want explicit seed", spawner.seed)
	}
}

func TestSpawnFromModeFallsBackWithoutElo(t *testing.T) {
	spawner := &testLegacySpawner{}
	mode := registry.ModeEntry{
		Definition: puzzle.ModeDef{Title: "Legacy"},
		Spawner:    spawner,
		Seeded:     spawner,
	}

	if _, _, err := spawnFromMode(mode, "", nil, "generated-name"); err != nil {
		t.Fatalf("random spawnFromMode returned error: %v", err)
	}
	if !spawner.spawned {
		t.Fatal("expected legacy Spawn to be called")
	}
}
