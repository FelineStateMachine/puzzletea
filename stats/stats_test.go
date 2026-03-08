package stats

import (
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/store"
)

func TestWeightsFromDefinitions(t *testing.T) {
	weights := WeightsFromDefinitions([]puzzle.Definition{{
		Name: "Sudoku",
		Modes: []puzzle.ModeDef{
			{Title: "Easy"},
			{Title: "Medium"},
			{Title: "Hard"},
		},
	}})
	if len(weights) != 3 {
		t.Fatalf("len(weights) = %d, want 3", len(weights))
	}
}

func TestLevelFromXPAndXPForLevel(t *testing.T) {
	if got := LevelFromXP(0); got != 0 {
		t.Fatalf("LevelFromXP(0) = %d, want 0", got)
	}
	if got := XPForLevel(3); got != 29 {
		t.Fatalf("XPForLevel(3) = %d, want 29", got)
	}
}

func TestComputeDailyStreak(t *testing.T) {
	now := time.Date(2026, time.February, 16, 15, 30, 0, 0, time.Local)
	dates := []time.Time{
		time.Date(2026, time.February, 16, 0, 0, 0, 0, time.Local),
		time.Date(2026, time.February, 15, 0, 0, 0, 0, time.Local),
		time.Date(2026, time.February, 14, 0, 0, 0, 0, time.Local),
	}
	if got := ComputeDailyStreak(dates, now); got != 3 {
		t.Fatalf("ComputeDailyStreak() = %d, want 3", got)
	}
}

func TestComputeCategoryXP(t *testing.T) {
	weights := Weights{
		{"Sudoku", "Easy"}: 2,
		{"Sudoku", "Hard"}: 8,
	}
	modeStats := []store.ModeStat{
		{GameType: "Sudoku", Mode: "Easy", Victories: 5, DailyVictories: 2},
		{GameType: "Sudoku", Mode: "Hard", Victories: 2},
	}
	if got := ComputeCategoryXP(weights, "Sudoku", modeStats); got != 30 {
		t.Fatalf("ComputeCategoryXP() = %d, want 30", got)
	}
}

func TestBuildCardsAndProfile(t *testing.T) {
	weights := Weights{
		{"Sudoku", "Easy"}: 2,
		{"Sudoku", "Hard"}: 8,
	}
	catStats := []store.CategoryStat{{
		GameType:       "Sudoku",
		TotalAttempts:  10,
		TotalVictories: 7,
		TimesDaily:     3,
		PreferredMode:  "Easy",
	}}
	modeStats := []store.ModeStat{
		{GameType: "Sudoku", Mode: "Easy", Victories: 5, DailyVictories: 2},
		{GameType: "Sudoku", Mode: "Hard", Victories: 2},
	}
	cards := BuildCards(weights, catStats, modeStats)
	if len(cards) != 1 || cards[0].CurrentXP != 30 {
		t.Fatalf("BuildCards() = %#v", cards)
	}
	profile := BuildProfileBanner(catStats, modeStats, weights, nil, false, 4, 17)
	if profile.WeekliesCompleted != 4 || profile.ThisWeekHighestIndex != 17 {
		t.Fatalf("BuildProfileBanner() = %#v", profile)
	}
}
