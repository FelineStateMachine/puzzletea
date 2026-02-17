package stats

import (
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/store"
)

// --- ModeXP map (P0) ---

func TestModeXPMap(t *testing.T) {
	// InitModeXP is called via TestMain or init in the main package.
	// For the stats package tests, we set up a known map directly.
	origXP := ModeXP
	defer func() { ModeXP = origXP }()

	ModeXP = map[ModeKey]int{
		{"Sudoku", "Easy"}:   1,
		{"Sudoku", "Medium"}: 3,
		{"Sudoku", "Hard"}:   5,
	}

	t.Run("all modes have positive XP", func(t *testing.T) {
		if len(ModeXP) == 0 {
			t.Fatal("ModeXP map is empty")
		}
		for k, v := range ModeXP {
			if v <= 0 {
				t.Errorf("ModeXP[%q/%q] = %d, want > 0", k.GameType, k.Mode, v)
			}
		}
	})
}

// --- LevelFromXP (P0) ---

func TestLevelFromXP(t *testing.T) {
	tests := []struct {
		name string
		xp   int
		want int
	}{
		{name: "zero XP", xp: 0, want: 0},
		{name: "negative XP", xp: -5, want: 0},
		{name: "1 XP", xp: 1, want: 0},
		{name: "4 XP", xp: 4, want: 0},
		{name: "5 XP = level 1", xp: 5, want: 1},
		{name: "15 XP still level 1", xp: 15, want: 1},
		{name: "16 XP = level 2", xp: 16, want: 2},
		{name: "28 XP still level 2", xp: 28, want: 2},
		{name: "29 XP = level 3", xp: 29, want: 3},
		{name: "45 XP still level 3", xp: 45, want: 3},
		{name: "46 XP = level 4", xp: 46, want: 4},
		{name: "65 XP still level 4", xp: 65, want: 4},
		{name: "66 XP = level 5", xp: 66, want: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LevelFromXP(tt.xp)
			if got != tt.want {
				t.Errorf("LevelFromXP(%d) = %d, want %d", tt.xp, got, tt.want)
			}
		})
	}
}

// --- XPForLevel (P0) ---

func TestXPForLevel(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  int
	}{
		{name: "level 0", level: 0, want: 0},
		{name: "level 1", level: 1, want: 5},
		{name: "level 2", level: 2, want: 16},
		{name: "level 3", level: 3, want: 29},
		{name: "level 4", level: 4, want: 46},
		{name: "level 5", level: 5, want: 66},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := XPForLevel(tt.level)
			if got != tt.want {
				t.Errorf("XPForLevel(%d) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}

// --- Level/XP round-trip consistency (P0) ---

func TestLevelXPRoundTrip(t *testing.T) {
	// For each level, XPForLevel should be the threshold that LevelFromXP
	// returns that level. One XP less should return level-1.
	for lvl := 1; lvl <= 30; lvl++ {
		threshold := XPForLevel(lvl)
		if LevelFromXP(threshold) != lvl {
			t.Errorf("LevelFromXP(XPForLevel(%d)=%d) = %d, want %d",
				lvl, threshold, LevelFromXP(threshold), lvl)
		}
		if threshold > 0 && LevelFromXP(threshold-1) != lvl-1 {
			t.Errorf("LevelFromXP(%d-1=%d) = %d, want %d",
				threshold, threshold-1, LevelFromXP(threshold-1), lvl-1)
		}
	}
}

// --- ComputeDailyStreak (P0) ---

func TestComputeDailyStreak(t *testing.T) {
	// Helper: date at midnight local time.
	d := func(year int, month time.Month, day int) time.Time {
		return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	}
	now := time.Date(2026, time.February, 16, 15, 30, 0, 0, time.Local)

	tests := []struct {
		name  string
		now   time.Time
		dates []time.Time
		want  int
	}{
		{
			name:  "empty dates",
			now:   now,
			dates: nil,
			want:  0,
		},
		{
			name:  "streak of 1 today",
			now:   now,
			dates: []time.Time{d(2026, 2, 16)},
			want:  1,
		},
		{
			name:  "streak of 1 yesterday",
			now:   now,
			dates: []time.Time{d(2026, 2, 15)},
			want:  1,
		},
		{
			name:  "streak broken two days ago",
			now:   now,
			dates: []time.Time{d(2026, 2, 14)},
			want:  0,
		},
		{
			name: "3-day streak ending today",
			now:  now,
			dates: []time.Time{
				d(2026, 2, 16),
				d(2026, 2, 15),
				d(2026, 2, 14),
			},
			want: 3,
		},
		{
			name: "3-day streak ending yesterday",
			now:  now,
			dates: []time.Time{
				d(2026, 2, 15),
				d(2026, 2, 14),
				d(2026, 2, 13),
			},
			want: 3,
		},
		{
			name: "gap breaks streak",
			now:  now,
			dates: []time.Time{
				d(2026, 2, 16),
				d(2026, 2, 15),
				// gap: Feb 14 missing
				d(2026, 2, 13),
				d(2026, 2, 12),
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeDailyStreak(tt.dates, tt.now)
			if got != tt.want {
				t.Errorf("ComputeDailyStreak() = %d, want %d", got, tt.want)
			}
		})
	}
}

// --- ComputeCategoryXP (P1) ---

func TestComputeCategoryXP(t *testing.T) {
	// Override ModeXP for test.
	origXP := ModeXP
	defer func() { ModeXP = origXP }()

	ModeXP = map[ModeKey]int{
		{"Sudoku", "Easy"}:   1,
		{"Sudoku", "Medium"}: 3,
		{"Sudoku", "Hard"}:   5,
	}

	tests := []struct {
		name      string
		gameType  string
		modeStats []store.ModeStat
		want      int
	}{
		{
			name:      "no stats",
			gameType:  "Sudoku",
			modeStats: nil,
			want:      0,
		},
		{
			name:     "normal victories only",
			gameType: "Sudoku",
			modeStats: []store.ModeStat{
				{GameType: "Sudoku", Mode: "Easy", Victories: 5, DailyVictories: 0},
				{GameType: "Sudoku", Mode: "Hard", Victories: 2, DailyVictories: 0},
			},
			want: 5*1 + 2*5, // 15
		},
		{
			name:     "daily 2x multiplier",
			gameType: "Sudoku",
			modeStats: []store.ModeStat{
				{GameType: "Sudoku", Mode: "Easy", Victories: 3, DailyVictories: 1},
			},
			// 2 normal * 1 + 1 daily * 1 * 2 = 4
			want: 4,
		},
		{
			name:     "ignores other game types",
			gameType: "Sudoku",
			modeStats: []store.ModeStat{
				{GameType: "Sudoku", Mode: "Easy", Victories: 1, DailyVictories: 0},
				{GameType: "Nonogram", Mode: "Mini", Victories: 10, DailyVictories: 0},
			},
			want: 1,
		},
		{
			name:     "unknown mode defaults to 1 XP",
			gameType: "Sudoku",
			modeStats: []store.ModeStat{
				{GameType: "Sudoku", Mode: "Unknown", Victories: 3, DailyVictories: 0},
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeCategoryXP(tt.gameType, tt.modeStats)
			if got != tt.want {
				t.Errorf("ComputeCategoryXP(%q) = %d, want %d", tt.gameType, got, tt.want)
			}
		})
	}
}

// --- BuildCards (P1) ---

func TestBuildCards(t *testing.T) {
	origXP := ModeXP
	defer func() { ModeXP = origXP }()

	ModeXP = map[ModeKey]int{
		{"Sudoku", "Easy"}: 2,
		{"Sudoku", "Hard"}: 8,
	}

	catStats := []store.CategoryStat{
		{
			GameType:       "Sudoku",
			TotalAttempts:  10,
			TotalVictories: 7,
			TotalAbandoned: 1,
			TimesDaily:     3,
			DailyVictories: 2,
			PreferredMode:  "Easy",
		},
	}
	modeStats := []store.ModeStat{
		{GameType: "Sudoku", Mode: "Easy", Victories: 5, DailyVictories: 2},
		{GameType: "Sudoku", Mode: "Hard", Victories: 2, DailyVictories: 0},
	}

	cards := BuildCards(catStats, modeStats)
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}

	c := cards[0]
	if c.GameType != "Sudoku" {
		t.Errorf("GameType = %q, want %q", c.GameType, "Sudoku")
	}
	if c.Victories != 7 {
		t.Errorf("Victories = %d, want 7", c.Victories)
	}
	if c.Attempts != 10 {
		t.Errorf("Attempts = %d, want 10", c.Attempts)
	}
	if c.DailyPlayed != 3 {
		t.Errorf("DailyPlayed = %d, want 3", c.DailyPlayed)
	}
	if c.PreferredMode != "Easy" {
		t.Errorf("PreferredMode = %q, want %q", c.PreferredMode, "Easy")
	}

	// XP: Easy: (5-2)*2 + 2*2*2 = 6+8 = 14; Hard: 2*8 = 16; total = 30
	if c.CurrentXP != 30 {
		t.Errorf("CurrentXP = %d, want 30", c.CurrentXP)
	}
}

// --- BuildCards empty state (P2) ---

func TestBuildCardsEmpty(t *testing.T) {
	cards := BuildCards(nil, nil)
	if len(cards) != 0 {
		t.Errorf("expected 0 cards for nil stats, got %d", len(cards))
	}
}

// --- RenderView empty state (P2) ---

func TestRenderViewEmpty(t *testing.T) {
	result := RenderView(ProfileBanner{}, nil, 80)
	if result == "" {
		t.Error("expected non-empty output for empty state")
	}
}
