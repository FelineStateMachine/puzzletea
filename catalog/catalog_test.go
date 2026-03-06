package catalog

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/lightsout"
)

func TestValidateAll(t *testing.T) {
	if err := Validate(All); err != nil {
		t.Fatalf("Validate(All) error = %v", err)
	}
}

func TestResolveSupportsCanonicalNamesAndAliases(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Sudoku", want: "Sudoku"},
		{input: "hashi", want: "Hashiwokakero"},
		{input: "lights", want: "Lights Out"},
		{input: "wordsearch", want: "Word Search"},
	}

	for _, tt := range tests {
		def, ok := Resolve(tt.input)
		if !ok {
			t.Fatalf("Resolve(%q) returned !ok", tt.input)
		}
		if def.Name != tt.want {
			t.Fatalf("Resolve(%q).Name = %q, want %q", tt.input, def.Name, tt.want)
		}
	}
}

func TestCategoriesAndNamesPreserveCatalogOrder(t *testing.T) {
	names := Names()
	categories := Categories()
	if len(names) != len(All) {
		t.Fatalf("len(Names()) = %d, want %d", len(names), len(All))
	}
	if len(categories) != len(All) {
		t.Fatalf("len(Categories()) = %d, want %d", len(categories), len(All))
	}
	for i, def := range All {
		if names[i] != def.Name {
			t.Fatalf("Names()[%d] = %q, want %q", i, names[i], def.Name)
		}
		if categories[i].Name != def.Name {
			t.Fatalf("Categories()[%d].Name = %q, want %q", i, categories[i].Name, def.Name)
		}
	}
}

func TestImportRoundTrip(t *testing.T) {
	g, err := lightsout.New(3, 3)
	if err != nil {
		t.Fatal(err)
	}
	save, err := g.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := Import("Lights Out", save)
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}
	if loaded == nil {
		t.Fatal("Import returned nil game")
	}
}

func TestDailyEntriesHaveMetadata(t *testing.T) {
	entries := DailyEntries()
	if len(entries) == 0 {
		t.Fatal("DailyEntries returned no entries")
	}
	for i, entry := range entries {
		if entry.Spawner == nil {
			t.Fatalf("DailyEntries()[%d].Spawner is nil", i)
		}
		if entry.GameType == "" || entry.Mode == "" {
			t.Fatalf("DailyEntries()[%d] missing metadata", i)
		}
	}
}
