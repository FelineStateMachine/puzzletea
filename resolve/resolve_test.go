package resolve

import (
	"strings"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/games/lightsout"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
)

var testEntries = registry.Entries()

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "Sudoku", "sudoku"},
		{"all caps", "SUDOKU", "sudoku"},
		{"hyphens to spaces", "lights-out", "lights out"},
		{"underscores to spaces", "word_search", "word search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.input); got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestModeEntry(t *testing.T) {
	sudokuEntry, ok := registry.Resolve("Sudoku")
	if !ok {
		t.Fatal("missing sudoku entry")
	}

	t.Run("empty name returns default (first) mode", func(t *testing.T) {
		mode, err := ModeEntry(sudokuEntry, "")
		if err != nil {
			t.Fatal(err)
		}
		if mode.Spawner == nil || mode.Definition.Title != "Beginner" {
			t.Fatalf("default mode = (%v, %q), want Beginner", mode.Spawner != nil, mode.Definition.Title)
		}
	})

	t.Run("case insensitive match", func(t *testing.T) {
		mode, err := ModeEntry(sudokuEntry, "easy")
		if err != nil {
			t.Fatal(err)
		}
		if mode.Definition.Title != "Easy" {
			t.Fatalf("title = %q, want Easy", mode.Definition.Title)
		}
	})

	t.Run("unknown mode", func(t *testing.T) {
		_, err := ModeEntry(sudokuEntry, "impossible")
		if err == nil || !strings.Contains(err.Error(), "unknown mode") {
			t.Fatalf("error = %v, want unknown mode", err)
		}
	})

	t.Run("empty modes list", func(t *testing.T) {
		emptyEntry := registry.Entry{
			Definition: puzzle.Definition{Name: "Empty"},
		}
		_, err := ModeEntry(emptyEntry, "any")
		if err == nil || !strings.Contains(err.Error(), "no available modes") {
			t.Fatalf("error = %v, want no available modes", err)
		}
	})
}

func TestModeNames(t *testing.T) {
	entry, ok := registry.Resolve("Lights Out")
	if !ok {
		t.Fatal("missing lights out entry")
	}
	names := ModeNames(entry)
	expected := []string{"Easy", "Medium", "Hard", "Extreme"}
	if len(names) != len(expected) {
		t.Fatalf("len(ModeNames) = %d, want %d", len(names), len(expected))
	}
	for i, want := range expected {
		if names[i] != want {
			t.Fatalf("mode[%d] = %q, want %q", i, names[i], want)
		}
	}
}

func TestRNGFromString(t *testing.T) {
	rng1 := RNGFromString("hello-world")
	rng2 := RNGFromString("hello-world")
	for i := range 10 {
		if v1, v2 := rng1.Uint64(), rng2.Uint64(); v1 != v2 {
			t.Fatalf("draw %d: got %d and %d for same seed", i, v1, v2)
		}
	}
}

func TestSeededMode(t *testing.T) {
	s1, gt1, m1, err := SeededMode("test-seed-42", testEntries)
	if err != nil {
		t.Fatal(err)
	}
	s2, gt2, m2, err := SeededMode("test-seed-42", testEntries)
	if err != nil {
		t.Fatal(err)
	}
	if s1 == nil || s2 == nil || gt1 != gt2 || m1 != m2 {
		t.Fatalf("seeded mode mismatch: (%v,%q,%q) vs (%v,%q,%q)", s1 != nil, gt1, m1, s2 != nil, gt2, m2)
	}

	_, _, _, err = SeededMode("test", nil)
	if err == nil {
		t.Fatal("expected error for empty entries")
	}

	emptyEntries := []registry.Entry{{
		Definition: puzzle.Definition{Name: "Empty"},
	}}
	_, _, _, err = SeededMode("test", emptyEntries)
	if err == nil {
		t.Fatal("expected error when no seeded modes available")
	}
}

func TestSeededModeStableOnEntryChange(t *testing.T) {
	seeds := []string{"alpha", "bravo", "charlie", "delta"}
	type selection struct{ gameType, mode string }
	original := make([]selection, len(seeds))
	for i, seed := range seeds {
		_, gt, m, err := SeededMode(seed, testEntries)
		if err != nil {
			t.Fatal(err)
		}
		original[i] = selection{gt, m}
	}

	extended := append([]registry.Entry(nil), testEntries...)
	extended = append(extended, registry.Entry{
		Definition: puzzle.Definition{
			ID:    puzzle.CanonicalGameID("Synthetic"),
			Name:  "Synthetic",
			Modes: []puzzle.ModeDef{{ID: puzzle.CanonicalModeID("Easy"), Title: "Easy", Seeded: true}},
		},
		Modes: []registry.ModeEntry{{
			Definition: puzzle.ModeDef{ID: puzzle.CanonicalModeID("Easy"), Title: "Easy", Seeded: true},
			Spawner:    lightsout.Modes[0].(game.Spawner),
			Seeded:     lightsout.Modes[0].(game.SeededSpawner),
		}},
	})

	for i, seed := range seeds {
		_, gt, m, err := SeededMode(seed, extended)
		if err != nil {
			t.Fatal(err)
		}
		if gt == "Synthetic" {
			continue
		}
		if gt != original[i].gameType || m != original[i].mode {
			t.Fatalf("seed %q changed from (%q,%q) to (%q,%q)", seed, original[i].gameType, original[i].mode, gt, m)
		}
	}
}
