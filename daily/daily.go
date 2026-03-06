package daily

import (
	"hash/fnv"
	"math/rand/v2"
	"time"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/namegen"
)

// Entry pairs a SeededSpawner with metadata for the eligible daily pool.
type Entry struct {
	Spawner  game.SeededSpawner
	GameType string
	Mode     string
}

// eligibleModes is the flattened pool built from each package's DailyModes.
var eligibleModes = buildEligibleModes()

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
// Name uses its own sub-RNG derived from the date so that changes to
// the namegen word lists cannot affect mode selection or puzzle generation.
func Name(date time.Time) string {
	h := fnv.New64a()
	h.Write([]byte("name:"))
	h.Write([]byte(date.Format("2006-01-02")))
	seed := h.Sum64()
	nameRNG := rand.New(rand.NewPCG(seed, ^seed))
	return "Daily " + date.Format("Jan _2 06") + " - " + namegen.GenerateSeeded(nameRNG)
}

// Mode selects the daily mode from the eligible pool using rendezvous
// hashing (highest random weight). Each entry is scored by hashing the
// date together with its (GameType, Mode) pair; the highest score wins.
//
// This is resilient to changes in the pool: adding or removing an entry
// only affects dates where the changed entry would have been the winner.
func Mode(date time.Time) (game.SeededSpawner, string, string) {
	if len(eligibleModes) == 0 {
		return nil, "", ""
	}

	dateStr := date.Format("2006-01-02")

	var best Entry
	var bestHash uint64
	found := false
	for _, entry := range eligibleModes {
		h := fnv.New64a()
		h.Write([]byte(dateStr))
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
