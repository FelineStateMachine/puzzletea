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
