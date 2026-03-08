package catalog_test

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/registry"
)

func TestValidateDefinitions(t *testing.T) {
	definitions := registry.Definitions()
	if err := catalog.Validate(definitions); err != nil {
		t.Fatalf("Validate(definitions) error = %v", err)
	}
}

func TestResolveSupportsCanonicalNamesAndAliases(t *testing.T) {
	idx := catalog.MustBuild(registry.Definitions())
	tests := []struct {
		input string
		want  string
	}{
		{input: "Sudoku", want: "Sudoku"},
		{input: "ripeto", want: "Sudoku RGB"},
		{input: "fillomino", want: "Fillomino"},
		{input: "hashi", want: "Hashiwokakero"},
		{input: "lights", want: "Lights Out"},
		{input: "ripple", want: "Ripple Effect"},
		{input: "wordsearch", want: "Word Search"},
	}

	for _, tt := range tests {
		def, ok := idx.Resolve(tt.input)
		if !ok {
			t.Fatalf("Resolve(%q) returned !ok", tt.input)
		}
		if def.Name != tt.want {
			t.Fatalf("Resolve(%q).Name = %q, want %q", tt.input, def.Name, tt.want)
		}
	}
}

func TestDefinitionsAndNamesPreserveRegistryOrder(t *testing.T) {
	definitions := registry.Definitions()
	idx := catalog.MustBuild(definitions)
	names := idx.Names()
	if len(names) != len(definitions) {
		t.Fatalf("len(Names()) = %d, want %d", len(names), len(definitions))
	}
	for i, def := range definitions {
		if names[i] != def.Name {
			t.Fatalf("Names()[%d] = %q, want %q", i, names[i], def.Name)
		}
	}
}

func TestDailyEntriesHaveMetadata(t *testing.T) {
	idx := catalog.MustBuild(registry.Definitions())
	entries := idx.DailyEntries()
	if len(entries) == 0 {
		t.Fatal("DailyEntries returned no entries")
	}
	for i, entry := range entries {
		if entry.GameID == "" || entry.GameName == "" || entry.ModeID == "" || entry.Mode == "" {
			t.Fatalf("DailyEntries()[%d] missing metadata", i)
		}
	}
}
