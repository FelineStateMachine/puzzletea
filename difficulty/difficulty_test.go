package difficulty

import "testing"

func TestValidateElo(t *testing.T) {
	for _, elo := range []Elo{MinElo, 600, 1200, 1800, 2400, SoftCapElo} {
		if err := ValidateElo(elo); err != nil {
			t.Fatalf("ValidateElo(%d) returned error: %v", elo, err)
		}
	}

	for _, elo := range []Elo{-1, SoftCapElo + 1} {
		if err := ValidateElo(elo); err == nil {
			t.Fatalf("ValidateElo(%d) returned nil error", elo)
		}
	}
}

func TestClampElo(t *testing.T) {
	tests := []struct {
		input Elo
		want  Elo
	}{
		{input: -100, want: MinElo},
		{input: MinElo, want: MinElo},
		{input: 1500, want: 1500},
		{input: SoftCapElo, want: SoftCapElo},
		{input: 3200, want: SoftCapElo},
	}

	for _, tt := range tests {
		if got := ClampElo(tt.input); got != tt.want {
			t.Fatalf("ClampElo(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestScore01(t *testing.T) {
	tests := []struct {
		input Elo
		want  float64
	}{
		{input: -100, want: 0},
		{input: MinElo, want: 0},
		{input: 1500, want: 0.5},
		{input: SoftCapElo, want: 1},
		{input: 3200, want: 1},
	}

	for _, tt := range tests {
		if got := Score01(tt.input); got != tt.want {
			t.Fatalf("Score01(%d) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
