package resolve

import (
	"strings"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/wordsearch"

	"github.com/charmbracelet/bubbles/list"
)

var testCategories = []list.Item{
	game.Category{Name: "Hashiwokakero", Desc: "Connect islands with bridges.", Modes: hashiwokakero.Modes},
	game.Category{Name: "Hitori", Desc: "Shade cells to eliminate duplicates.", Modes: hitori.Modes},
	game.Category{Name: "Lights Out", Desc: "Turn off all the lights.", Modes: lightsout.Modes},
	game.Category{Name: "Nonogram", Desc: "Fill cells to match row and column hints.", Modes: nonogram.Modes},
	game.Category{Name: "Sudoku", Desc: "Fill the 9x9 grid following sudoku rules.", Modes: sudoku.Modes},
	game.Category{Name: "Takuzu", Desc: "Fill the grid with ● and ○.", Modes: takuzu.Modes},
	game.Category{Name: "Word Search", Desc: "Find hidden words in a letter grid.", Modes: wordsearch.Modes},
}

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

func TestCategory(t *testing.T) {
	// Exact and case-insensitive matches.
	exactTests := []struct {
		name     string
		input    string
		wantName string
	}{
		{"exact Nonogram", "Nonogram", "Nonogram"},
		{"exact Sudoku", "Sudoku", "Sudoku"},
		{"exact Lights Out", "Lights Out", "Lights Out"},
		{"exact Word Search", "Word Search", "Word Search"},
		{"exact Hashiwokakero", "Hashiwokakero", "Hashiwokakero"},
		{"exact Takuzu", "Takuzu", "Takuzu"},
		{"lowercase", "nonogram", "Nonogram"},
		{"all caps", "SUDOKU", "Sudoku"},
		{"hyphenated", "lights-out", "Lights Out"},
		{"underscored", "lights_out", "Lights Out"},
		{"mixed case hyphen", "Word-Search", "Word Search"},
	}

	for _, tt := range exactTests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := Category(tt.input, testCategories)
			if err != nil {
				t.Fatalf("Category(%q) error: %v", tt.input, err)
			}
			if cat.Name != tt.wantName {
				t.Errorf("Category(%q).Name = %q, want %q", tt.input, cat.Name, tt.wantName)
			}
		})
	}

	// All aliases.
	aliasTests := []struct {
		alias    string
		wantName string
	}{
		{"hashi", "Hashiwokakero"},
		{"bridges", "Hashiwokakero"},
		{"hitori", "Hitori"},
		{"lights", "Lights Out"},
		{"ws", "Word Search"},
		{"words", "Word Search"},
		{"wordsearch", "Word Search"},
		{"binairo", "Takuzu"},
		{"binary", "Takuzu"},
		{"nonogram", "Nonogram"},
		{"sudoku", "Sudoku"},
		{"takuzu", "Takuzu"},
	}

	for _, tt := range aliasTests {
		t.Run("alias_"+tt.alias, func(t *testing.T) {
			cat, err := Category(tt.alias, testCategories)
			if err != nil {
				t.Fatalf("Category(%q) error: %v", tt.alias, err)
			}
			if cat.Name != tt.wantName {
				t.Errorf("Category(%q).Name = %q, want %q", tt.alias, cat.Name, tt.wantName)
			}
		})
	}

	// Error cases.
	t.Run("unknown game", func(t *testing.T) {
		_, err := Category("chess", testCategories)
		if err == nil {
			t.Fatal("expected error for unknown game")
		}
		if !strings.Contains(err.Error(), "unknown game") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "unknown game")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := Category("", testCategories)
		if err == nil {
			t.Fatal("expected error for empty string")
		}
	})
}

func TestMode(t *testing.T) {
	// Find a known category with modes for testing.
	sudokuCat, err := Category("Sudoku", testCategories)
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
		emptyCat := game.Category{Name: "Empty", Modes: []list.Item{}}
		_, _, err := Mode(emptyCat, "any")
		if err == nil {
			t.Fatal("expected error for empty modes")
		}
		if !strings.Contains(err.Error(), "no available modes") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "no available modes")
		}
	})
}

func TestCategoryNames(t *testing.T) {
	names := CategoryNames(testCategories)
	if len(names) != len(testCategories) {
		t.Fatalf("len(CategoryNames()) = %d, want %d", len(names), len(testCategories))
	}

	// Verify all expected games are present.
	expected := map[string]bool{
		"Hashiwokakero": true,
		"Hitori":        true,
		"Lights Out":    true,
		"Nonogram":      true,
		"Sudoku":        true,
		"Takuzu":        true,
		"Word Search":   true,
	}
	for _, n := range names {
		if !expected[n] {
			t.Errorf("unexpected category name: %q", n)
		}
		delete(expected, n)
	}
	for n := range expected {
		t.Errorf("missing category name: %q", n)
	}
}

func TestModeNames(t *testing.T) {
	cat, err := Category("Lights Out", testCategories)
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
