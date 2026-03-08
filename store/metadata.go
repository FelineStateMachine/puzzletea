package store

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/puzzle"
)

var weeklyNamePattern = regexp.MustCompile(`^Week (\d{2})-(\d{4}) - #(\d{2})$`)

func RunKindForName(name string) RunKind {
	switch {
	case RunDateForName(name) != nil:
		return RunKindDaily
	case isWeeklyName(name):
		return RunKindWeekly
	case SeedTextForName(name) != "":
		return RunKindSeeded
	default:
		return RunKindNormal
	}
}

func RunDateForName(name string) *time.Time {
	if !strings.HasPrefix(name, "Daily ") {
		return nil
	}
	label, _, found := strings.Cut(strings.TrimPrefix(name, "Daily "), " - ")
	if !found {
		return nil
	}
	t, err := time.ParseInLocation("Jan _2 06", label, time.Local)
	if err != nil {
		return nil
	}
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return &day
}

func WeeklyIdentityForName(name string) (year, week, index int, ok bool) {
	matches := weeklyNamePattern.FindStringSubmatch(name)
	if len(matches) != 4 {
		return 0, 0, 0, false
	}

	week, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, false
	}
	year, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, false
	}
	index, err = strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, false
	}
	return year, week, index, true
}

func SeedTextForName(name string) string {
	if strings.HasPrefix(name, "Daily ") || isWeeklyName(name) {
		return ""
	}
	seedPart, _, found := strings.Cut(name, " - ")
	if !found {
		return ""
	}
	seedPart = strings.TrimSpace(seedPart)
	if seedPart == "" {
		return ""
	}
	if before, _, hasGame := strings.Cut(seedPart, " ["); hasGame {
		return strings.TrimSpace(before)
	}
	return seedPart
}

func CanonicalGameID(gameType string) string {
	return string(puzzle.CanonicalGameID(gameType))
}

func CanonicalModeID(mode string) string {
	return string(puzzle.CanonicalModeID(mode))
}

func isWeeklyName(name string) bool {
	_, _, _, ok := WeeklyIdentityForName(name)
	return ok
}
