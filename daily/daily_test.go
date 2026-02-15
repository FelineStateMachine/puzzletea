package daily

import (
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

// --- Seed (P0) ---

func TestSeed(t *testing.T) {
	tests := []struct {
		name     string
		dateA    time.Time
		dateB    time.Time
		wantSame bool
	}{
		{
			name:     "same date same seed",
			dateA:    time.Date(2026, 2, 14, 10, 30, 0, 0, time.UTC),
			dateB:    time.Date(2026, 2, 14, 23, 59, 0, 0, time.UTC),
			wantSame: true,
		},
		{
			name:     "different dates different seeds",
			dateA:    time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC),
			dateB:    time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
			wantSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seedA := Seed(tt.dateA)
			seedB := Seed(tt.dateB)
			if (seedA == seedB) != tt.wantSame {
				t.Errorf("Seed(%v)=%d, Seed(%v)=%d, wantSame=%v",
					tt.dateA, seedA, tt.dateB, seedB, tt.wantSame)
			}
		})
	}
}

func TestSeedDeterministic(t *testing.T) {
	date := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	s1 := Seed(date)
	s2 := Seed(date)
	if s1 != s2 {
		t.Errorf("Seed is not deterministic: %d != %d", s1, s2)
	}
}

// --- RNG (P0) ---

func TestRNG(t *testing.T) {
	date := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	rng1 := RNG(date)
	rng2 := RNG(date)

	// Same date should produce the same sequence of random numbers.
	for i := range 10 {
		a := rng1.Int64()
		b := rng2.Int64()
		if a != b {
			t.Fatalf("RNG diverged at draw %d: %d != %d", i, a, b)
		}
	}
}

func TestRNGDifferentDates(t *testing.T) {
	rng1 := RNG(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	rng2 := RNG(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))

	// Different dates should produce different sequences.
	same := true
	for range 5 {
		if rng1.Int64() != rng2.Int64() {
			same = false
			break
		}
	}
	if same {
		t.Error("RNG produced identical sequences for different dates")
	}
}

// --- Name (P0) ---

func TestName(t *testing.T) {
	date := time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC)
	rng := RNG(date)
	name := Name(date, rng)

	if !strings.HasPrefix(name, "Daily Feb 14 26 - ") {
		t.Errorf("Name = %q, want prefix %q", name, "Daily Feb 14 26 - ")
	}
}

func TestNameDeterministic(t *testing.T) {
	date := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	name1 := Name(date, RNG(date))
	name2 := Name(date, RNG(date))
	if name1 != name2 {
		t.Errorf("Name not deterministic: %q != %q", name1, name2)
	}
}

// --- Mode (P1) ---

func TestMode(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 99))
	spawner, gameType, mode := Mode(rng)

	if spawner == nil {
		t.Fatal("Mode returned nil spawner")
	}
	if gameType == "" {
		t.Error("Mode returned empty gameType")
	}
	if mode == "" {
		t.Error("Mode returned empty mode title")
	}
}

func TestModeDeterministic(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(100, 200))
	rng2 := rand.New(rand.NewPCG(100, 200))

	_, gt1, m1 := Mode(rng1)
	_, gt2, m2 := Mode(rng2)

	if gt1 != gt2 || m1 != m2 {
		t.Errorf("Mode not deterministic: (%q,%q) vs (%q,%q)", gt1, m1, gt2, m2)
	}
}

// --- EligibleModes pool (P1) ---

func TestEligibleModesNotEmpty(t *testing.T) {
	if len(eligibleModes) == 0 {
		t.Fatal("eligibleModes is empty; no daily modes registered")
	}
}

func TestEligibleModesHaveMetadata(t *testing.T) {
	for i, e := range eligibleModes {
		if e.Spawner == nil {
			t.Errorf("eligibleModes[%d].Spawner is nil", i)
		}
		if e.GameType == "" {
			t.Errorf("eligibleModes[%d].GameType is empty", i)
		}
		if e.Mode == "" {
			t.Errorf("eligibleModes[%d].Mode is empty", i)
		}
	}
}
