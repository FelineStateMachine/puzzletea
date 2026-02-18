package resolve

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"

	"charm.land/bubbles/v2/list"
)

// Normalize lowercases and replaces hyphens/underscores with spaces for
// fuzzy matching of CLI arguments to game/mode names.
func Normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.TrimSpace(s)
}

// CategoryAliases maps short or alternate names to canonical category names.
var CategoryAliases = map[string]string{
	"hashi":      "hashiwokakero",
	"bridges":    "hashiwokakero",
	"hitori":     "hitori",
	"lights":     "lights out",
	"nonogram":   "nonogram",
	"sudoku":     "sudoku",
	"takuzu":     "takuzu",
	"binairo":    "takuzu",
	"binary":     "takuzu",
	"shikaku":    "shikaku",
	"rectangles": "shikaku",
	"words":      "word search",
	"wordsearch": "word search",
	"ws":         "word search",
}

// Category finds a game category by name (case-insensitive,
// hyphen/underscore-tolerant, with alias support).
func Category(name string, categories []list.Item) (game.Category, error) {
	norm := Normalize(name)

	// Check aliases first.
	if canonical, ok := CategoryAliases[norm]; ok {
		norm = canonical
	}

	for _, item := range categories {
		cat := item.(game.Category)
		if Normalize(cat.Name) == norm {
			return cat, nil
		}
	}

	return game.Category{}, fmt.Errorf("unknown game %q\n\nAvailable games:\n  %s",
		name, strings.Join(CategoryNames(categories), "\n  "))
}

// Mode finds a mode within a category by name. If name is empty,
// returns the first (default) mode.
func Mode(cat game.Category, name string) (game.Spawner, string, error) {
	if len(cat.Modes) == 0 {
		return nil, "", fmt.Errorf("game %q has no available modes", cat.Name)
	}

	if name == "" {
		m := cat.Modes[0].(game.Mode)
		return m.(game.Spawner), m.Title(), nil
	}

	norm := Normalize(name)
	for _, item := range cat.Modes {
		m := item.(game.Mode)
		if Normalize(m.Title()) == norm {
			return m.(game.Spawner), m.Title(), nil
		}
	}

	return nil, "", fmt.Errorf("unknown mode %q for %s\n\nAvailable modes:\n  %s",
		name, cat.Name, strings.Join(ModeNames(cat), "\n  "))
}

// CategoryNames returns the display names of all game categories.
func CategoryNames(categories []list.Item) []string {
	names := make([]string, len(categories))
	for i, item := range categories {
		names[i] = item.(game.Category).Name
	}
	return names
}

// ModeNames returns the display names of all modes in a category.
func ModeNames(cat game.Category) []string {
	names := make([]string, len(cat.Modes))
	for i, item := range cat.Modes {
		names[i] = item.(game.Mode).Title()
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
func SeededMode(seed string, categories []list.Item) (game.SeededSpawner, string, string, error) {
	var best seededEntry
	var bestHash uint64
	found := false
	for _, item := range categories {
		cat := item.(game.Category)
		for _, modeItem := range cat.Modes {
			s, ok := modeItem.(game.SeededSpawner)
			if !ok {
				continue
			}
			h := fnv.New64a()
			h.Write([]byte(seed))
			h.Write([]byte{0})
			h.Write([]byte(cat.Name))
			h.Write([]byte{0})
			h.Write([]byte(modeItem.(game.Mode).Title()))
			score := h.Sum64()
			if !found || score > bestHash {
				bestHash = score
				best = seededEntry{
					spawner:  s,
					gameType: cat.Name,
					mode:     modeItem.(game.Mode).Title(),
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
