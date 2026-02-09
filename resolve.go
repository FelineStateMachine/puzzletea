package main

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
)

// normalize lowercases and replaces hyphens/underscores with spaces for
// fuzzy matching of CLI arguments to game/mode names.
func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.TrimSpace(s)
}

// categoryAliases maps short or alternate names to canonical category names.
var categoryAliases = map[string]string{
	"hashi":    "hashiwokakero",
	"bridges":  "hashiwokakero",
	"lights":   "lights out",
	"nonogram": "nonogram",
	"sudoku":   "sudoku",
	"words":    "word search",
	"ws":       "word search",
}

// resolveCategory finds a game category by name (case-insensitive,
// hyphen/underscore-tolerant, with alias support).
func resolveCategory(name string) (game.Category, error) {
	norm := normalize(name)

	// Check aliases first.
	if canonical, ok := categoryAliases[norm]; ok {
		norm = canonical
	}

	for _, item := range GameCategories {
		cat := item.(game.Category)
		if normalize(cat.Name) == norm {
			return cat, nil
		}
	}

	return game.Category{}, fmt.Errorf("unknown game %q\n\nAvailable games:\n  %s",
		name, strings.Join(listCategoryNames(), "\n  "))
}

// resolveMode finds a mode within a category by name. If name is empty,
// returns the first (default) mode.
func resolveMode(cat game.Category, name string) (game.Spawner, string, error) {
	if len(cat.Modes) == 0 {
		return nil, "", fmt.Errorf("game %q has no available modes", cat.Name)
	}

	if name == "" {
		m := cat.Modes[0].(game.Mode)
		return m.(game.Spawner), m.Title(), nil
	}

	norm := normalize(name)
	for _, item := range cat.Modes {
		m := item.(game.Mode)
		if normalize(m.Title()) == norm {
			return m.(game.Spawner), m.Title(), nil
		}
	}

	return nil, "", fmt.Errorf("unknown mode %q for %s\n\nAvailable modes:\n  %s",
		name, cat.Name, strings.Join(listModeNames(cat), "\n  "))
}

// listCategoryNames returns the display names of all game categories.
func listCategoryNames() []string {
	names := make([]string, len(GameCategories))
	for i, item := range GameCategories {
		names[i] = item.(game.Category).Name
	}
	return names
}

// listModeNames returns the display names of all modes in a category.
func listModeNames(cat game.Category) []string {
	names := make([]string, len(cat.Modes))
	for i, item := range cat.Modes {
		names[i] = item.(game.Mode).Title()
	}
	return names
}
