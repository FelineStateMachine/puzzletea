package resolve

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
)

// Normalize lowercases and replaces hyphens/underscores with spaces for
// fuzzy matching of CLI arguments to mode names.
func Normalize(s string) string {
	return game.NormalizeName(s)
}

// Mode finds a mode within a category by name. If name is empty,
// returns the first (default) mode.
func Mode(cat game.Category, name string) (game.Spawner, string, error) {
	if len(cat.Modes) == 0 {
		return nil, "", fmt.Errorf("game %q has no available modes", cat.Name)
	}

	if name == "" {
		for _, item := range cat.Modes {
			m, ok := item.(game.Mode)
			if !ok {
				continue
			}
			s, ok := item.(game.Spawner)
			if !ok {
				continue
			}
			return s, m.Title(), nil
		}
		return nil, "", fmt.Errorf("game %q has no spawnable modes", cat.Name)
	}

	norm := Normalize(name)
	for _, item := range cat.Modes {
		m, ok := item.(game.Mode)
		if !ok {
			continue
		}
		if Normalize(m.Title()) == norm {
			s, ok := item.(game.Spawner)
			if !ok {
				return nil, "", fmt.Errorf("mode %q for %s is not spawnable", m.Title(), cat.Name)
			}
			return s, m.Title(), nil
		}
	}

	return nil, "", fmt.Errorf("unknown mode %q for %s\n\nAvailable modes:\n  %s",
		name, cat.Name, strings.Join(ModeNames(cat), "\n  "))
}

// ModeNames returns the display names of all modes in a category.
func ModeNames(cat game.Category) []string {
	names := make([]string, 0, len(cat.Modes))
	for _, item := range cat.Modes {
		mode, ok := item.(game.Mode)
		if !ok {
			continue
		}
		names = append(names, mode.Title())
	}
	return names
}

// RNGFromString creates a deterministic RNG seeded from an arbitrary string.
// The string is hashed with FNV-64a to produce a uint64 seed.
func RNGFromString(seed string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(seed))
	s := h.Sum64()
	return rand.New(rand.NewPCG(s, ^s))
}

// seededEntry pairs a SeededSpawner with its category and mode metadata.
type seededEntry struct {
	spawner  game.SeededSpawner
	gameType string
	mode     string
}

// SeededMode selects a game type and mode from all available modes across
// all categories using rendezvous hashing (highest random weight). Each
// eligible mode is scored by hashing the seed string together with its
// (gameType, modeTitle) pair; the highest score wins.
//
// This is resilient to changes in the category/mode lists: adding or
// removing a mode only affects seeds where the changed entry would have
// been the winner.
func SeededMode(seed string, definitions []game.Definition) (game.SeededSpawner, string, string, error) {
	var best seededEntry
	var bestHash uint64
	found := false
	for _, def := range definitions {
		for _, modeItem := range def.Modes {
			mode, ok := modeItem.(game.Mode)
			if !ok {
				continue
			}
			s, ok := modeItem.(game.SeededSpawner)
			if !ok {
				continue
			}
			h := fnv.New64a()
			h.Write([]byte(seed))
			h.Write([]byte{0})
			h.Write([]byte(def.Name))
			h.Write([]byte{0})
			h.Write([]byte(mode.Title()))
			score := h.Sum64()
			if !found || score > bestHash {
				bestHash = score
				best = seededEntry{
					spawner:  s,
					gameType: def.Name,
					mode:     mode.Title(),
				}
				found = true
			}
		}
	}
	if !found {
		return nil, "", "", errors.New("no seeded modes available")
	}
	return best.spawner, best.gameType, best.mode, nil
}
