package weekly

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"regexp"
	"strconv"
	"time"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/game"
)

// Entry pairs a SeededSpawner with metadata for the eligible weekly pool.
type Entry struct {
	Spawner  game.SeededSpawner
	GameType string
	Mode     string
}

// Info identifies a single weekly gauntlet puzzle.
type Info struct {
	Year  int
	Week  int
	Index int
}

var (
	eligibleModes      = buildEligibleModes()
	weeklyNamePattern  = regexp.MustCompile(`^Week (\d{2})-(\d{4}) - #(\d{2})$`)
	currentWeekEntries = 99
)

func buildEligibleModes() []Entry {
	catalogEntries := catalog.DailyEntries()
	entries := make([]Entry, 0, len(catalogEntries))
	for _, entry := range catalogEntries {
		entries = append(entries, Entry{
			Spawner:  entry.Spawner,
			GameType: entry.GameType,
			Mode:     entry.Mode,
		})
	}
	return entries
}

// Name returns the canonical persisted name for a weekly puzzle.
func Name(year, week, index int) string {
	return fmt.Sprintf("Week %02d-%04d - #%02d", week, year, index)
}

// Prefix returns the exact query prefix for all rows in a weekly gauntlet.
func Prefix(year, week int) string {
	return fmt.Sprintf("Week %02d-%04d - #", week, year)
}

// NameForDate returns the weekly puzzle name for the date's ISO week.
func NameForDate(date time.Time, index int) string {
	year, week := date.ISOWeek()
	return Name(year, week, index)
}

// Current returns the current ISO week identity for the provided date.
func Current(date time.Time) Info {
	year, week := date.ISOWeek()
	return Info{Year: year, Week: week}
}

// StartOfWeek returns the Monday for the given ISO week.
func StartOfWeek(year, week int, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, loc)
	weekdayOffset := (int(jan4.Weekday()) + 6) % 7
	weekOneMonday := jan4.AddDate(0, 0, -weekdayOffset)
	return weekOneMonday.AddDate(0, 0, (week-1)*7)
}

// AddWeeks moves to the next or previous ISO week.
func AddWeeks(weekStart time.Time, delta int) time.Time {
	return weekStart.AddDate(0, 0, delta*7)
}

// ParseName parses a canonical weekly name.
func ParseName(name string) (Info, bool) {
	matches := weeklyNamePattern.FindStringSubmatch(name)
	if len(matches) != 4 {
		return Info{}, false
	}

	week, err := strconv.Atoi(matches[1])
	if err != nil {
		return Info{}, false
	}
	year, err := strconv.Atoi(matches[2])
	if err != nil {
		return Info{}, false
	}
	index, err := strconv.Atoi(matches[3])
	if err != nil {
		return Info{}, false
	}
	if week < 1 || week > 53 || index < 1 || index > currentWeekEntries {
		return Info{}, false
	}

	info := Info{Year: year, Week: week, Index: index}
	if Name(info.Year, info.Week, info.Index) != name {
		return Info{}, false
	}
	return info, true
}

// Seed returns a deterministic uint64 seed derived from the weekly identity.
func Seed(year, week, index int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(Name(year, week, index)))
	return h.Sum64()
}

// RNG creates a deterministic RNG seeded from the weekly identity.
func RNG(year, week, index int) *rand.Rand {
	seed := Seed(year, week, index)
	return rand.New(rand.NewPCG(seed, ^seed))
}

// Mode selects the weekly mode from the daily-eligible pool using rendezvous
// hashing on the canonical weekly name.
func Mode(year, week, index int) (game.SeededSpawner, string, string) {
	if len(eligibleModes) == 0 {
		return nil, "", ""
	}

	seedName := Name(year, week, index)
	var best Entry
	var bestHash uint64
	found := false
	for _, entry := range eligibleModes {
		h := fnv.New64a()
		h.Write([]byte(seedName))
		h.Write([]byte{0})
		h.Write([]byte(entry.GameType))
		h.Write([]byte{0})
		h.Write([]byte(entry.Mode))
		score := h.Sum64()
		if !found || score > bestHash {
			bestHash = score
			best = entry
			found = true
		}
	}
	if !found {
		return nil, "", ""
	}
	return best.Spawner, best.GameType, best.Mode
}

// BonusXP returns the slot-based bonus XP for a completed weekly puzzle.
func BonusXP(index int) int {
	if index < 10 {
		return 0
	}
	return index / 10
}
