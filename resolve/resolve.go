package resolve

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"

	"github.com/charmbracelet/bubbles/list"
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
