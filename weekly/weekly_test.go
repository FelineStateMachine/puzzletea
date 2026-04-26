package weekly

import (
	"testing"
	"time"
)

func TestNameAndParseRoundTrip(t *testing.T) {
	name := Name(2026, 1, 1)
	if name != "Week 01-2026 - #01" {
		t.Fatalf("Name() = %q, want %q", name, "Week 01-2026 - #01")
	}

	info, ok := ParseName(name)
	if !ok {
		t.Fatal("ParseName() = false, want true")
	}
	if info.Year != 2026 || info.Week != 1 || info.Index != 1 {
		t.Fatalf("ParseName() = %+v, want year=2026 week=1 index=1", info)
	}
}

func TestParseNameRejectsInvalidValues(t *testing.T) {
	tests := []string{
		"Week 1-2026 - #01",
		"Week 01-2026 - #1",
		"Week 00-2026 - #01",
		"Week 01-2026 - #00",
		"Week 54-2026 - #01",
		"Week 01-2026 - #100",
		"week 01-2026 - #01",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			if _, ok := ParseName(name); ok {
				t.Fatalf("ParseName(%q) = true, want false", name)
			}
		})
	}
}

func TestRNGDeterministic(t *testing.T) {
	rngA := RNG(2026, 12, 7)
	rngB := RNG(2026, 12, 7)

	for i := range 10 {
		if got, want := rngA.Int64(), rngB.Int64(); got != want {
			t.Fatalf("draw %d = %d, want %d", i, got, want)
		}
	}
}

func TestModeDeterministic(t *testing.T) {
	_, gameTypeA, modeA := Mode(2026, 20, 17)
	_, gameTypeB, modeB := Mode(2026, 20, 17)

	if gameTypeA != gameTypeB || modeA != modeB {
		t.Fatalf("Mode() mismatch: (%q,%q) vs (%q,%q)", gameTypeA, modeA, gameTypeB, modeB)
	}
}

func TestISOWeekBoundary(t *testing.T) {
	date := time.Date(2025, time.December, 29, 12, 0, 0, 0, time.UTC)
	info := Current(date)
	if info.Year != 2026 || info.Week != 1 {
		t.Fatalf("Current() = %+v, want year=2026 week=1", info)
	}

	start := StartOfWeek(2026, 1, time.UTC)
	if !start.Equal(time.Date(2025, time.December, 29, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("StartOfWeek() = %v, want 2025-12-29", start)
	}
}

func TestBonusXP(t *testing.T) {
	tests := []struct {
		index int
		want  int
	}{
		{index: 1, want: 0},
		{index: 9, want: 0},
		{index: 10, want: 1},
		{index: 19, want: 1},
		{index: 20, want: 2},
		{index: 99, want: 9},
	}

	for _, tt := range tests {
		t.Run(Name(2026, 1, tt.index), func(t *testing.T) {
			if got := BonusXP(tt.index); got != tt.want {
				t.Fatalf("BonusXP(%d) = %d, want %d", tt.index, got, tt.want)
			}
		})
	}
}

func TestEligibleModesIncludeSudokuRGBVariant(t *testing.T) {
	found := map[string]bool{}
	for _, entry := range eligibleModes {
		if entry.GameType == "Sudoku RGB" {
			found[entry.Mode] = true
		}
	}

	if !found["Sudoku RGB"] {
		t.Fatalf("eligibleModes missing Sudoku RGB variant")
	}
}
