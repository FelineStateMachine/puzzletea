package resolve

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
)

// Normalize lowercases and replaces hyphens/underscores with spaces for
// fuzzy matching of CLI arguments to mode names.
func Normalize(s string) string {
	return puzzle.NormalizeName(s)
}

// Mode finds a mode within a game entry by name. If name is empty,
// returns the first (default) mode.
func Mode(entry registry.Entry, name string) (game.Spawner, string, error) {
	mode, err := ModeEntry(entry, name)
	if err != nil {
		return nil, "", err
	}
	return mode.Spawner, mode.Definition.Title, nil
}

func ModeEntry(entry registry.Entry, name string) (registry.ModeEntry, error) {
	if len(entry.Modes) == 0 {
		return registry.ModeEntry{}, fmt.Errorf("game %q has no available modes", entry.Definition.Name)
	}

	if name == "" {
		mode := entry.Modes[0]
		return mode, nil
	}

	norm := Normalize(name)
	for _, mode := range entry.Modes {
		if Normalize(mode.Definition.Title) == norm {
			return mode, nil
		}
	}

	return registry.ModeEntry{}, fmt.Errorf("unknown mode %q for %s\n\nAvailable modes:\n  %s",
		name, entry.Definition.Name, strings.Join(ModeNames(entry), "\n  "))
}

// ModeNames returns the display names of all modes in an entry.
func ModeNames(entry registry.Entry) []string {
	names := make([]string, 0, len(entry.Modes))
	for _, mode := range entry.Modes {
		names = append(names, mode.Definition.Title)
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

func seededModeForDefinition(seed string, entry registry.Entry) (seededEntry, bool) {
	var best seededEntry
	var bestHash uint64
	found := false

	for _, mode := range entry.Modes {
		if mode.Seeded == nil {
			continue
		}
		h := fnv.New64a()
		h.Write([]byte(seed))
		h.Write([]byte{0})
		h.Write([]byte(entry.Definition.Name))
		h.Write([]byte{0})
		h.Write([]byte(mode.Definition.Title))
		score := h.Sum64()
		if !found || score > bestHash {
			bestHash = score
			best = seededEntry{
				spawner:  mode.Seeded,
				gameType: entry.Definition.Name,
				mode:     mode.Definition.Title,
			}
			found = true
		}
	}

	return best, found
}

// SeededMode selects a game type and mode from all available modes across
// all categories using rendezvous hashing (highest random weight). Each
// eligible mode is scored by hashing the seed string together with its
// (gameType, modeTitle) pair; the highest score wins.
//
// This is resilient to changes in the category/mode lists: adding or
// removing a mode only affects seeds where the changed entry would have
// been the winner.
func SeededMode(seed string, entries []registry.Entry) (game.SeededSpawner, string, string, error) {
	var best seededEntry
	var bestHash uint64
	found := false
	for _, entry := range entries {
		selected, ok := seededModeForDefinition(seed, entry)
		if !ok {
			continue
		}
		h := fnv.New64a()
		h.Write([]byte(seed))
		h.Write([]byte{0})
		h.Write([]byte(selected.gameType))
		h.Write([]byte{0})
		h.Write([]byte(selected.mode))
		score := h.Sum64()
		if !found || score > bestHash {
			bestHash = score
			best = selected
			found = true
		}
	}
	if !found {
		return nil, "", "", errors.New("no seeded modes available")
	}
	return best.spawner, best.gameType, best.mode, nil
}

// SeededModeForGame deterministically selects a seeded mode within a single
// game definition, so the same seed and game name always produce the same
// puzzle for that game.
func SeededModeForGame(seed, gameType string, entries []registry.Entry) (game.SeededSpawner, string, string, error) {
	norm := Normalize(gameType)
	for _, entry := range entries {
		if Normalize(entry.Definition.Name) != norm {
			continue
		}

		selected, ok := seededModeForDefinition(seed, entry)
		if !ok {
			return nil, "", "", fmt.Errorf("game %q has no seeded modes", entry.Definition.Name)
		}
		return selected.spawner, selected.gameType, selected.mode, nil
	}

	return nil, "", "", fmt.Errorf("unknown game %q", gameType)
}
