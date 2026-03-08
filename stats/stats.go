package stats

import (
	"math"
	"slices"
	"time"

	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/store"
)

// ModeKey identifies a specific mode within a game category.
type ModeKey struct {
	GameType string
	Mode     string
}

type Weights map[ModeKey]int

func WeightsFromDefinitions(definitions []puzzle.Definition) Weights {
	weights := make(Weights, 64)
	for _, def := range definitions {
		count := len(def.Modes)
		for i, mode := range def.Modes {
			xp := int(math.Round(float64(i) / float64(count) * 10))
			xp = max(xp, 1)
			weights[ModeKey{GameType: def.Name, Mode: mode.Title}] = xp
		}
	}
	return weights
}

// LevelFromXP returns the level for the given total XP.
// Uses the inverse of XPForLevel: level = floor((xp / 5) ^ (1/1.6))
func LevelFromXP(xp int) int {
	if xp <= 0 {
		return 0
	}
	return int(math.Floor(math.Pow(float64(xp)/5.0, 1.0/1.6)))
}

// XPForLevel returns the total XP required to reach the given level.
// Formula: ceil(5 * level^1.6)
func XPForLevel(level int) int {
	if level <= 0 {
		return 0
	}
	return int(math.Ceil(5.0 * math.Pow(float64(level), 1.6)))
}

// ComputeCategoryXP calculates total XP for a game category by summing
// XP across all modes. Daily victories earn 2x XP.
func ComputeCategoryXP(weights Weights, gameType string, modeStats []store.ModeStat) int {
	total := 0
	for _, ms := range modeStats {
		if ms.GameType != gameType {
			continue
		}
		baseXP := weights[ModeKey{ms.GameType, ms.Mode}]
		if baseXP == 0 {
			baseXP = 1
		}
		normalVictories := ms.Victories - ms.DailyVictories
		total += normalVictories * baseXP
		total += ms.DailyVictories * baseXP * 2
		total += ms.WeeklyBonusXP
	}
	return total
}

// ComputeDailyStreak calculates the length of the current daily completion
// streak. The streak is "alive" if the most recent completion is today or
// yesterday (giving the player until end of day to extend it).
func ComputeDailyStreak(dates []time.Time, now time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	today := TruncateToDate(now)
	yesterday := today.AddDate(0, 0, -1)

	most := TruncateToDate(dates[0])
	var start time.Time
	switch {
	case most.Equal(today):
		start = today
	case most.Equal(yesterday):
		start = yesterday
	default:
		return 0
	}

	streak := 0
	expected := start
	for _, d := range dates {
		dt := TruncateToDate(d)
		if dt.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else if dt.Before(expected) {
			break
		}
	}
	return streak
}

// TruncateToDate strips the time component, returning midnight local time.
func TruncateToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// --- Stats card ---

// Card holds the computed stats for a single game category.
type Card struct {
	GameType      string
	Level         int
	PreferredMode string
	Victories     int
	Attempts      int
	DailyPlayed   int
	CurrentXP     int
	NextLevelXP   int
}

// BuildCards constructs a Card for each category the player has
// interacted with. Categories with zero attempts are excluded.
func BuildCards(weights Weights, catStats []store.CategoryStat, modeStats []store.ModeStat) []Card {
	var cards []Card
	for _, cs := range catStats {
		if cs.TotalAttempts == 0 {
			continue
		}
		totalXP := ComputeCategoryXP(weights, cs.GameType, modeStats)
		lvl := LevelFromXP(totalXP)
		nextXP := XPForLevel(lvl + 1)

		preferred := cs.PreferredMode
		if preferred == "" {
			preferred = "\u2014" // em dash
		}

		cards = append(cards, Card{
			GameType:      cs.GameType,
			Level:         lvl,
			PreferredMode: preferred,
			Victories:     cs.TotalVictories,
			Attempts:      cs.TotalAttempts,
			DailyPlayed:   cs.TimesDaily,
			CurrentXP:     totalXP,
			NextLevelXP:   nextXP,
		})
	}

	// Sort by level descending, then by XP descending for ties.
	slices.SortStableFunc(cards, func(a, b Card) int {
		if a.Level != b.Level {
			return b.Level - a.Level
		}
		return b.CurrentXP - a.CurrentXP
	})

	return cards
}

// --- Profile banner ---

// ProfileBanner holds the summary data shown above the card grid.
type ProfileBanner struct {
	ProfileLevel         int
	DailyStreak          int
	TotalDailies         int
	CurrentDaily         bool
	WeekliesCompleted    int
	ThisWeekHighestIndex int
}

// BuildProfileBanner constructs the summary banner shown above the card grid.
func BuildProfileBanner(
	catStats []store.CategoryStat,
	modeStats []store.ModeStat,
	weights Weights,
	streakDates []time.Time,
	currentDaily bool,
	weekliesCompleted int,
	thisWeekHighestIndex int,
) ProfileBanner {
	profileLevel := 0
	totalDailies := 0
	for _, cs := range catStats {
		xp := ComputeCategoryXP(weights, cs.GameType, modeStats)
		profileLevel += LevelFromXP(xp)
		totalDailies += cs.TimesDaily
	}

	streak := ComputeDailyStreak(streakDates, time.Now())

	return ProfileBanner{
		ProfileLevel:         profileLevel,
		DailyStreak:          streak,
		TotalDailies:         totalDailies,
		CurrentDaily:         currentDaily,
		WeekliesCompleted:    weekliesCompleted,
		ThisWeekHighestIndex: thisWeekHighestIndex,
	}
}
