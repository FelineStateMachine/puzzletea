package schedule

import (
	"math/rand/v2"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
)

type testSeededSpawner struct{}

func (testSeededSpawner) Spawn() (game.Gamer, error)                 { return nil, nil }
func (testSeededSpawner) SpawnSeeded(*rand.Rand) (game.Gamer, error) { return nil, nil }

func TestSelectBySeedDeterministic(t *testing.T) {
	entries := []Entry{
		{Spawner: testSeededSpawner{}, GameType: "Sudoku", Mode: "Easy"},
		{Spawner: testSeededSpawner{}, GameType: "Nonogram", Mode: "Mini"},
		{Spawner: testSeededSpawner{}, GameType: "Takuzu", Mode: "Beginner"},
	}

	first, ok := SelectBySeed("seed-1", entries)
	if !ok {
		t.Fatal("SelectBySeed() = false, want true")
	}
	second, ok := SelectBySeed("seed-1", entries)
	if !ok {
		t.Fatal("SelectBySeed() = false, want true")
	}
	if first.GameType != second.GameType || first.Mode != second.Mode {
		t.Fatalf("selection mismatch: (%q,%q) vs (%q,%q)",
			first.GameType, first.Mode, second.GameType, second.Mode)
	}
}

func TestSelectBySeedStableForUnchangedWinners(t *testing.T) {
	entries := []Entry{
		{Spawner: testSeededSpawner{}, GameType: "Sudoku", Mode: "Easy"},
		{Spawner: testSeededSpawner{}, GameType: "Nonogram", Mode: "Mini"},
		{Spawner: testSeededSpawner{}, GameType: "Takuzu", Mode: "Beginner"},
	}

	seeds := []string{"seed-a", "seed-b", "seed-c", "seed-d", "seed-e"}
	original := make(map[string]Entry, len(seeds))
	for _, seed := range seeds {
		selected, ok := SelectBySeed(seed, entries)
		if !ok {
			t.Fatalf("SelectBySeed(%q) = false, want true", seed)
		}
		original[seed] = selected
	}

	synthetic := Entry{Spawner: testSeededSpawner{}, GameType: "Synthetic", Mode: "Synthetic"}
	extended := append(append([]Entry(nil), entries...), synthetic)
	for _, seed := range seeds {
		selected, ok := SelectBySeed(seed, extended)
		if !ok {
			t.Fatalf("SelectBySeed(%q) = false, want true", seed)
		}
		if selected.GameType == synthetic.GameType && selected.Mode == synthetic.Mode {
			continue
		}
		want := original[seed]
		if selected.GameType != want.GameType || selected.Mode != want.Mode {
			t.Fatalf("seed %q changed from (%q,%q) to (%q,%q)",
				seed, want.GameType, want.Mode, selected.GameType, selected.Mode)
		}
	}
}
