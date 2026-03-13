package puzzle

import "testing"

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Word Search", want: "word search"},
		{input: "lights-out", want: "lights out"},
		{input: "word_search", want: "word search"},
		{input: "  Sudoku  ", want: "sudoku"},
		{input: "Sudoku   RGB", want: "sudoku rgb"},
	}

	for _, tt := range tests {
		if got := NormalizeName(tt.input); got != tt.want {
			t.Fatalf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCanonicalIDsUseNormalizedNames(t *testing.T) {
	if got, want := CanonicalGameID("  Sudoku   RGB "), GameID("sudoku rgb"); got != want {
		t.Fatalf("CanonicalGameID() = %q, want %q", got, want)
	}
	if got, want := CanonicalModeID("Mini-Hard"), ModeID("mini hard"); got != want {
		t.Fatalf("CanonicalModeID() = %q, want %q", got, want)
	}
}

func TestNewModeDefAndDefinitionCloneInput(t *testing.T) {
	mode := NewModeDef(ModeSpec{
		Title:       "Medium",
		Description: "Balanced board",
		Seeded:      true,
	})
	if got, want := mode.ID, ModeID("medium"); got != want {
		t.Fatalf("ModeDef.ID = %q, want %q", got, want)
	}

	spec := DefinitionSpec{
		Name:         "Sudoku",
		Description:  "Classic grid logic",
		Aliases:      []string{"classic"},
		Modes:        []ModeDef{mode},
		DailyModeIDs: []ModeID{mode.ID},
	}
	def := NewDefinition(spec)

	spec.Aliases[0] = "changed"
	spec.Modes[0].Title = "Changed"
	spec.DailyModeIDs[0] = "changed"

	if got, want := def.ID, GameID("sudoku"); got != want {
		t.Fatalf("Definition.ID = %q, want %q", got, want)
	}
	if got, want := def.Aliases[0], "classic"; got != want {
		t.Fatalf("Aliases[0] = %q, want %q", got, want)
	}
	if got, want := def.Modes[0].Title, "Medium"; got != want {
		t.Fatalf("Modes[0].Title = %q, want %q", got, want)
	}
	if got, want := def.DailyModeIDs[0], ModeID("medium"); got != want {
		t.Fatalf("DailyModeIDs[0] = %q, want %q", got, want)
	}
}

func TestSelectModeIDsByIndex(t *testing.T) {
	modes := []ModeDef{
		NewModeDef(ModeSpec{Title: "Beginner"}),
		NewModeDef(ModeSpec{Title: "Medium"}),
		NewModeDef(ModeSpec{Title: "Expert"}),
	}

	got := SelectModeIDsByIndex(modes, 1, 2, 99, -1)
	if len(got) != 2 {
		t.Fatalf("len(SelectModeIDsByIndex) = %d, want 2", len(got))
	}
	if got[0] != ModeID("medium") || got[1] != ModeID("expert") {
		t.Fatalf("SelectModeIDsByIndex = %v, want [medium expert]", got)
	}
}
