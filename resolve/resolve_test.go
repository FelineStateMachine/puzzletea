package resolve

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/lightsout"
)

var testDefinitions = catalog.All

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
		{"mixed separators", "Word-Search_Test", "word search test"},
		{"already normalized", "sudoku", "sudoku"},
		{"leading trailing space", "  sudoku  ", "sudoku"},
		{"empty string", "", ""},
		{"only hyphens", "---", ""},
		{"only underscores", "___", ""},
		{"mixed with spaces", " lights-out ", "lights out"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.input)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMode(t *testing.T) {
	// Find a known category with modes for testing.
	sudokuCat, err := catalog.Category("Sudoku")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty name returns default (first) mode", func(t *testing.T) {
		spawner, title, err := Mode(sudokuCat, "")
		if err != nil {
			t.Fatal(err)
		}
		if spawner == nil {
			t.Fatal("expected non-nil spawner")
		}
		// First Sudoku mode is "Beginner".
		if title != "Beginner" {
			t.Errorf("default mode title = %q, want %q", title, "Beginner")
		}
	})

	t.Run("exact mode title", func(t *testing.T) {
		_, title, err := Mode(sudokuCat, "Easy")
		if err != nil {
			t.Fatal(err)
		}
		if title != "Easy" {
			t.Errorf("title = %q, want %q", title, "Easy")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		_, title, err := Mode(sudokuCat, "easy")
		if err != nil {
			t.Fatal(err)
		}
		if title != "Easy" {
			t.Errorf("title = %q, want %q", title, "Easy")
		}
	})

	t.Run("unknown mode", func(t *testing.T) {
		_, _, err := Mode(sudokuCat, "impossible")
		if err == nil {
			t.Fatal("expected error for unknown mode")
		}
		if !strings.Contains(err.Error(), "unknown mode") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "unknown mode")
		}
	})

	t.Run("empty modes list", func(t *testing.T) {
		emptyCat := game.Category{Name: "Empty", Modes: nil}
		_, _, err := Mode(emptyCat, "any")
		if err == nil {
			t.Fatal("expected error for empty modes")
		}
		if !strings.Contains(err.Error(), "no available modes") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "no available modes")
		}
	})
}

func TestModeSupportsSudokuRGBAliasCategory(t *testing.T) {
	cat, err := catalog.Category("ripeto")
	if err != nil {
		t.Fatal(err)
	}
	if cat.Name != "Sudoku RGB" {
		t.Fatalf("category name = %q, want %q", cat.Name, "Sudoku RGB")
	}

	_, title, err := Mode(cat, "easy")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Easy" {
		t.Fatalf("mode title = %q, want %q", title, "Easy")
	}
}

func TestModeNames(t *testing.T) {
	cat, err := catalog.Category("Lights Out")
	if err != nil {
		t.Fatal(err)
	}

	names := ModeNames(cat)
	if len(names) != len(cat.Modes) {
		t.Fatalf("len(ModeNames) = %d, want %d", len(names), len(cat.Modes))
	}

	// Lights Out modes: Easy, Medium, Hard, Extreme.
	expected := []string{"Easy", "Medium", "Hard", "Extreme"}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("mode[%d] = %q, want %q", i, names[i], want)
		}
	}
}

// --- RNGFromString (P0) ---

func TestRNGFromString(t *testing.T) {
	t.Run("deterministic output", func(t *testing.T) {
		rng1 := RNGFromString("hello-world")
		rng2 := RNGFromString("hello-world")

		// Draw several values and confirm they match.
		for i := range 10 {
			v1 := rng1.Uint64()
			v2 := rng2.Uint64()
			if v1 != v2 {
				t.Fatalf("draw %d: got %d and %d for same seed", i, v1, v2)
			}
		}
	})

	t.Run("different seeds produce different output", func(t *testing.T) {
		rng1 := RNGFromString("seed-a")
		rng2 := RNGFromString("seed-b")

		// Extremely unlikely for two different seeds to produce the same first draw.
		if rng1.Uint64() == rng2.Uint64() {
			t.Error("different seeds produced identical first draw")
		}
	})

	t.Run("empty string is valid", func(t *testing.T) {
		rng := RNGFromString("")
		// Should not panic; just verify it produces output.
		_ = rng.Uint64()
	})
}

// --- SeededMode (P0) ---

func TestSeededMode(t *testing.T) {
	t.Run("deterministic selection", func(t *testing.T) {
		s1, gt1, m1, err := SeededMode("test-seed-42", testDefinitions)
		if err != nil {
			t.Fatal(err)
		}
		s2, gt2, m2, err := SeededMode("test-seed-42", testDefinitions)
		if err != nil {
			t.Fatal(err)
		}

		if gt1 != gt2 {
			t.Errorf("game types differ: %q vs %q", gt1, gt2)
		}
		if m1 != m2 {
			t.Errorf("mode titles differ: %q vs %q", m1, m2)
		}
		if s1 == nil || s2 == nil {
			t.Fatal("expected non-nil spawners")
		}
	})

	t.Run("returns valid metadata", func(t *testing.T) {
		spawner, gameType, modeTitle, err := SeededMode("test-seed", testDefinitions)
		if err != nil {
			t.Fatal(err)
		}
		if spawner == nil {
			t.Fatal("expected non-nil spawner")
		}
		if gameType == "" {
			t.Error("expected non-empty game type")
		}
		if modeTitle == "" {
			t.Error("expected non-empty mode title")
		}
	})

	t.Run("empty categories returns error", func(t *testing.T) {
		_, _, _, err := SeededMode("test", nil)
		if err == nil {
			t.Fatal("expected error for empty categories")
		}
	})

	t.Run("categories with no seeded spawners returns error", func(t *testing.T) {
		// Category with no modes at all.
		cats := []game.Definition{
			{Name: "Empty", Modes: nil, Import: func([]byte) (game.Gamer, error) { return nil, nil }},
		}
		_, _, _, err := SeededMode("test", cats)
		if err == nil {
			t.Fatal("expected error when no seeded modes available")
		}
	})

	t.Run("stable on category change", func(t *testing.T) {
		// Verify rendezvous hashing stability: adding a category should
		// only affect seeds where the new category's mode wins.
		seeds := []string{
			"alpha", "bravo", "charlie", "delta", "echo",
			"foxtrot", "golf", "hotel", "india", "juliet",
		}

		type selection struct {
			gameType string
			mode     string
		}

		// Record selections with original categories.
		original := make([]selection, len(seeds))
		for i, seed := range seeds {
			_, gt, m, err := SeededMode(seed, testDefinitions)
			if err != nil {
				t.Fatal(err)
			}
			original[i] = selection{gt, m}
		}

		// Add a synthetic category with one mode.
		extended := make([]game.Definition, len(testDefinitions))
		copy(extended, testDefinitions)
		extended = append(extended, game.Definition{
			Name:        "Synthetic",
			Description: "Fake game for testing",
			Modes:       []list.Item{lightsout.Modes[0]},
			Import:      func([]byte) (game.Gamer, error) { return nil, nil },
		})

		for i, seed := range seeds {
			_, gt, m, err := SeededMode(seed, extended)
			if err != nil {
				t.Fatal(err)
			}
			if gt == "Synthetic" {
				continue // this seed was "stolen" by the new entry
			}
			if gt != original[i].gameType || m != original[i].mode {
				t.Errorf("seed %q: selection changed from (%q,%q) to (%q,%q) after adding unrelated category",
					seed, original[i].gameType, original[i].mode, gt, m)
			}
		}
	})
}
