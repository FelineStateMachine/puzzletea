package resolve

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"strings"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
)

// Normalize lowercases and replaces hyphens/underscores with spaces for
// fuzzy matching of CLI arguments to mode names.
func Normalize(s string) string {
	return puzzle.NormalizeName(s)
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

type VariantSelection struct {
	Variant      registry.VariantEntry
	ExplicitElo  *difficulty.Elo
	LegacyAlias  *puzzle.LegacyModeAlias
	DisplayTitle string
}

func VariantEntry(entry registry.Entry, name string) (VariantSelection, error) {
	if len(entry.Variants) == 0 {
		return VariantSelection{}, fmt.Errorf("game %q has no available variants", entry.Definition.Name)
	}

	if name == "" {
		variant := entry.Variants[0]
		return VariantSelection{
			Variant:      variant,
			DisplayTitle: variant.Definition.Title,
		}, nil
	}

	norm := Normalize(name)
	for _, variant := range entry.Variants {
		if Normalize(variant.Definition.Title) == norm || Normalize(string(variant.Definition.ID)) == norm {
			return VariantSelection{
				Variant:      variant,
				DisplayTitle: variant.Definition.Title,
			}, nil
		}
	}

	for _, alias := range entry.LegacyModes {
		if legacyAliasMatches(alias, norm) {
			variant, ok := variantByID(entry, alias.TargetVariantID)
			if !ok {
				return VariantSelection{}, fmt.Errorf("legacy mode %q targets missing variant %q", alias.Title, alias.TargetVariantID)
			}
			elo := alias.PresetElo
			aliasCopy := alias
			return VariantSelection{
				Variant:      variant,
				ExplicitElo:  &elo,
				LegacyAlias:  &aliasCopy,
				DisplayTitle: variant.Definition.Title,
			}, nil
		}
	}

	return VariantSelection{}, fmt.Errorf("unknown variant or legacy mode %q for %s\n\nAvailable variants:\n  %s",
		name, entry.Definition.Name, strings.Join(VariantNames(entry), "\n  "))
}

func legacyAliasMatches(alias puzzle.LegacyModeAlias, norm string) bool {
	if Normalize(alias.Title) == norm || Normalize(string(alias.ID)) == norm {
		return true
	}
	for _, cliAlias := range alias.CLIAliases {
		if Normalize(cliAlias) == norm {
			return true
		}
	}
	return false
}

func variantByID(entry registry.Entry, id puzzle.VariantID) (registry.VariantEntry, bool) {
	for _, variant := range entry.Variants {
		if variant.Definition.ID == id {
			return variant, true
		}
	}
	return registry.VariantEntry{}, false
}

// ModeNames returns the display names of all modes in an entry.
func ModeNames(entry registry.Entry) []string {
	names := make([]string, 0, len(entry.Modes))
	for _, mode := range entry.Modes {
		names = append(names, mode.Definition.Title)
	}
	return names
}

func VariantNames(entry registry.Entry) []string {
	names := make([]string, 0, len(entry.Variants))
	for _, variant := range entry.Variants {
		names = append(names, variant.Definition.Title)
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

	for _, variant := range entry.Variants {
		if variant.Seeded == nil {
			continue
		}
		h := fnv.New64a()
		h.Write([]byte(seed))
		h.Write([]byte{0})
		h.Write([]byte(entry.Definition.Name))
		h.Write([]byte{0})
		h.Write([]byte(variant.Definition.Title))
		score := h.Sum64()
		if !found || score > bestHash {
			bestHash = score
			best = seededEntry{
				spawner:  variant.Seeded,
				gameType: entry.Definition.Name,
				mode:     variant.Definition.Title,
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
