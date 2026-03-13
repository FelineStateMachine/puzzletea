package registry

import "testing"

func TestResolveNormalizesSpacingAndAliases(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "word   search", want: "Word Search"},
		{name: "hashi", want: "Hashiwokakero"},
		{name: "polyomino", want: "Fillomino"},
	}

	for _, tt := range tests {
		entry, ok := Resolve(tt.name)
		if !ok {
			t.Fatalf("Resolve(%q) = false, want true", tt.name)
		}
		if got := entry.Definition.Name; got != tt.want {
			t.Fatalf("Resolve(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestModeSeededFlagMatchesSpawnerAvailability(t *testing.T) {
	for _, entry := range Entries() {
		for _, mode := range entry.Modes {
			if got, want := mode.Definition.Seeded, mode.Seeded != nil; got != want {
				t.Fatalf("%s/%s seeded flag = %v, want %v",
					entry.Definition.Name, mode.Definition.Title, got, want)
			}
		}
	}
}

func TestEntriesStayAlignedWithDefinitions(t *testing.T) {
	definitions := Definitions()
	entries := Entries()
	if got, want := len(entries), len(definitions); got != want {
		t.Fatalf("len(Entries()) = %d, want %d", got, want)
	}

	for _, def := range definitions {
		entry, ok := Lookup(def.Name)
		if !ok {
			t.Fatalf("Lookup(%q) = false", def.Name)
		}
		if entry.Import == nil {
			t.Fatalf("%s missing import function", def.Name)
		}
		if got, want := len(entry.Modes), len(def.Modes); got != want {
			t.Fatalf("%s mode count = %d, want %d", def.Name, got, want)
		}
	}
}

func TestPrintAdaptersStayAlignedWithRegistryEntries(t *testing.T) {
	for _, entry := range Entries() {
		if entry.Definition.Name == "Lights Out" {
			if entry.Print != nil {
				t.Fatal("Lights Out should not expose a print adapter")
			}
			continue
		}
		if entry.Print == nil {
			t.Fatalf("%s missing print adapter", entry.Definition.Name)
		}
	}
}
