package daily

import (
	"hash/fnv"
	"math/rand/v2"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/game/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/game/hitori"
	"github.com/FelineStateMachine/puzzletea/game/lightsout"
	"github.com/FelineStateMachine/puzzletea/game/nonogram"
	"github.com/FelineStateMachine/puzzletea/game/shikaku"
	"github.com/FelineStateMachine/puzzletea/game/sudoku"
	"github.com/FelineStateMachine/puzzletea/game/takuzu"
	"github.com/FelineStateMachine/puzzletea/game/wordsearch"
	"github.com/FelineStateMachine/puzzletea/namegen"

	"charm.land/bubbles/v2/list"
)

// Entry pairs a SeededSpawner with metadata for the eligible daily pool.
type Entry struct {
	Spawner  game.SeededSpawner
	GameType string
	Mode     string
}

// pool maps game type names to their DailyModes exports.
// Each package owns which of its modes are eligible for daily rotation.
var pool = []struct {
	gameType string
	modes    []list.Item
}{
	{"Nonogram", nonogram.DailyModes},
	{"Sudoku", sudoku.DailyModes},
	{"Takuzu", takuzu.DailyModes},
	{"Hashiwokakero", hashiwokakero.DailyModes},
	{"Hitori", hitori.DailyModes},
	{"Lights Out", lightsout.DailyModes},
	{"Shikaku", shikaku.DailyModes},
	{"Word Search", wordsearch.DailyModes},
}

// eligibleModes is the flattened pool built from each package's DailyModes.
var eligibleModes = buildEligibleModes()

func buildEligibleModes() []Entry {
	var entries []Entry
	for _, p := range pool {
		for _, item := range p.modes {
			s := item.(game.SeededSpawner)
			entries = append(entries, Entry{
				Spawner:  s,
				GameType: p.gameType,
				Mode:     item.(game.Mode).Title(),
			})
		}
	}
	return entries
}

// Seed returns a deterministic uint64 seed derived from the date.
func Seed(date time.Time) uint64 {
	dateStr := date.Format("2006-01-02")
	h := fnv.New64a()
	h.Write([]byte(dateStr))
	return h.Sum64()
}

// RNG creates a deterministic RNG seeded from the given date.
func RNG(date time.Time) *rand.Rand {
	seed := Seed(date)
	return rand.New(rand.NewPCG(seed, ^seed))
}

// Name generates the daily puzzle name in the format:
// "Daily Feb 14 26 - amber-falcon"
//
// NOTE: this must be called before Mode on the same RNG — the number
// of draws here determines which mode is selected. Changing this function
// will shift daily puzzles for all future dates.
func Name(date time.Time, rng *rand.Rand) string {
	return "Daily " + date.Format("Jan _2 06") + " - " + namegen.GenerateSeeded(rng)
}

// Mode selects the daily mode from the eligible pool using the seeded RNG.
// Returns the spawner, game type name, and mode title.
func Mode(rng *rand.Rand) (game.SeededSpawner, string, string) {
	entry := eligibleModes[rng.IntN(len(eligibleModes))]
	return entry.Spawner, entry.GameType, entry.Mode
}
