package schedule

import (
	"hash/fnv"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/registry"
)

type Entry struct {
	Spawner  game.SeededSpawner
	GameType string
	Mode     string
}

func BuildEligibleModes(entries []registry.DailyEntry) []Entry {
	result := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, Entry{
			Spawner:  entry.Spawner,
			GameType: entry.GameType,
			Mode:     entry.Mode,
		})
	}
	return result
}

func SelectBySeed(seed string, entries []Entry) (Entry, bool) {
	var best Entry
	var bestHash uint64
	found := false
	for _, entry := range entries {
		h := fnv.New64a()
		h.Write([]byte(seed))
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
	return best, found
}
