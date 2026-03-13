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
