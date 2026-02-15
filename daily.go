package main

import (
	"hash/fnv"
	"math/rand/v2"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/wordsearch"

	"github.com/charmbracelet/bubbles/list"
)

// dailyEntry pairs a SeededSpawner with metadata for the eligible daily pool.
type dailyEntry struct {
	spawner  game.SeededSpawner
	gameType string
	mode     string
}

// dailyPool maps game type names to their DailyModes exports.
// Each package owns which of its modes are eligible for daily rotation.
var dailyPool = []struct {
	gameType string
	modes    []list.Item
}{
	{"Nonogram", nonogram.DailyModes},
	{"Sudoku", sudoku.DailyModes},
	{"Takuzu", takuzu.DailyModes},
	{"Hashiwokakero", hashiwokakero.DailyModes},
	{"Hitori", hitori.DailyModes},
	{"Lights Out", lightsout.DailyModes},
	{"Word Search", wordsearch.DailyModes},
}

// eligibleDailyModes is the flattened pool built from each package's DailyModes.
var eligibleDailyModes = buildEligibleDailyModes()

func buildEligibleDailyModes() []dailyEntry {
	var entries []dailyEntry
	for _, p := range dailyPool {
		for _, item := range p.modes {
			s := item.(game.SeededSpawner)
			entries = append(entries, dailyEntry{
				spawner:  s,
				gameType: p.gameType,
				mode:     item.(game.Mode).Title(),
			})
		}
	}
	return entries
}

// dailySeed returns a deterministic int64 seed derived from the date.
func dailySeed(date time.Time) uint64 {
	dateStr := date.Format("2006-01-02")
	h := fnv.New64a()
	h.Write([]byte(dateStr))
	return h.Sum64()
}

// dailyRNG creates a deterministic RNG seeded from the given date.
func dailyRNG(date time.Time) *rand.Rand {
	seed := dailySeed(date)
	return rand.New(rand.NewPCG(seed, ^seed))
}

// dailyName generates the daily puzzle name in the format:
// "Daily Feb 14 26 - amber-falcon"
//
// NOTE: this must be called before dailyMode on the same RNG — the number
// of draws here determines which mode is selected. Changing this function
// will shift daily puzzles for all future dates.
func dailyName(date time.Time, rng *rand.Rand) string {
	return "Daily " + date.Format("Jan _2 06") + " - " + namegen.GenerateSeeded(rng)
}

// dailyMode selects the daily mode from the eligible pool using the seeded RNG.
// Returns the spawner, game type name, and mode title.
func dailyMode(rng *rand.Rand) (game.SeededSpawner, string, string) {
	entry := eligibleDailyModes[rng.IntN(len(eligibleDailyModes))]
	return entry.spawner, entry.gameType, entry.mode
}
