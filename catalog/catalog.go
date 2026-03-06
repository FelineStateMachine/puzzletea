package catalog

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/nurikabe"
	"github.com/FelineStateMachine/puzzletea/shikaku"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
)

// All is the explicit ordered catalog of every puzzle game.
var All = []game.Definition{
	hashiwokakero.Definition,
	hitori.Definition,
	lightsout.Definition,
	nonogram.Definition,
	nurikabe.Definition,
	shikaku.Definition,
	sudoku.Definition,
	takuzu.Definition,
	wordsearch.Definition,
}

// DailyEntry is a flattened seeded daily candidate with metadata.
type DailyEntry struct {
	Spawner  game.SeededSpawner
	GameType string
	Mode     string
}

type index struct {
	categories []game.Category
	items      []list.Item
	names      []string
	dailies    []DailyEntry
	byName     map[string]game.Definition
	byAlias    map[string]string
}

var built = mustBuildIndex(All)

func mustBuildIndex(definitions []game.Definition) index {
	idx, err := buildIndex(definitions)
	if err != nil {
		panic(err)
	}
	return idx
}

func buildIndex(definitions []game.Definition) (index, error) {
	if err := Validate(definitions); err != nil {
		return index{}, err
	}

	idx := index{
		categories: make([]game.Category, 0, len(definitions)),
		items:      make([]list.Item, 0, len(definitions)),
		names:      make([]string, 0, len(definitions)),
		dailies:    make([]DailyEntry, 0, len(definitions)*2),
		byName:     make(map[string]game.Definition, len(definitions)),
		byAlias:    make(map[string]string),
	}

	for _, def := range definitions {
		category := def.Category()
		idx.categories = append(idx.categories, category)
		idx.items = append(idx.items, category)
		idx.names = append(idx.names, def.Name)
		idx.byName[game.NormalizeName(def.Name)] = def
		for _, alias := range def.Aliases {
			idx.byAlias[game.NormalizeName(alias)] = def.Name
		}
		for _, item := range def.DailyModes {
			mode, ok := item.(game.Mode)
			if !ok {
				continue
			}
			spawner, ok := item.(game.SeededSpawner)
			if !ok {
				continue
			}
			idx.dailies = append(idx.dailies, DailyEntry{
				Spawner:  spawner,
				GameType: def.Name,
				Mode:     mode.Title(),
			})
		}
	}

	return idx, nil
}

// Validate checks that the catalog is internally consistent.
func Validate(definitions []game.Definition) error {
	names := make(map[string]string, len(definitions))
	aliases := make(map[string]string)

	for _, def := range definitions {
		if def.Import == nil {
			return fmt.Errorf("game %q has nil importer", def.Name)
		}

		normName := game.NormalizeName(def.Name)
		if normName == "" {
			return fmt.Errorf("game %q has empty canonical name", def.Name)
		}
		if prior, exists := names[normName]; exists {
			return fmt.Errorf("duplicate game name %q conflicts with %q", def.Name, prior)
		}
		names[normName] = def.Name

		modeTitles := make(map[string]struct{}, len(def.Modes))
		for _, item := range def.Modes {
			mode, ok := item.(game.Mode)
			if !ok {
				continue
			}
			modeTitles[mode.Title()] = struct{}{}
		}

		for _, item := range def.DailyModes {
			mode, ok := item.(game.Mode)
			if !ok {
				return fmt.Errorf("game %q has daily mode without game.Mode implementation", def.Name)
			}
			if _, ok := item.(game.SeededSpawner); !ok {
				return fmt.Errorf("game %q daily mode %q is not seeded", def.Name, mode.Title())
			}
			if _, ok := modeTitles[mode.Title()]; !ok {
				return fmt.Errorf("game %q daily mode %q is not present in Modes", def.Name, mode.Title())
			}
		}

		for _, alias := range def.Aliases {
			normAlias := game.NormalizeName(alias)
			if normAlias == "" {
				return fmt.Errorf("game %q has empty alias", def.Name)
			}
			if normAlias == normName {
				return fmt.Errorf("game %q alias %q duplicates the canonical name", def.Name, alias)
			}
			if prior, exists := names[normAlias]; exists {
				return fmt.Errorf("game %q alias %q conflicts with canonical game %q", def.Name, alias, prior)
			}
			if prior, exists := aliases[normAlias]; exists {
				return fmt.Errorf("game %q alias %q conflicts with alias owned by %q", def.Name, alias, prior)
			}
			aliases[normAlias] = def.Name
		}
	}

	return nil
}

// Names returns the canonical game names in catalog order.
func Names() []string {
	return append([]string(nil), built.names...)
}

// Categories returns the category view of the catalog in menu order.
func Categories() []game.Category {
	return append([]game.Category(nil), built.categories...)
}

// CategoryItems returns the category view as list items for Bubble Tea lists.
func CategoryItems() []list.Item {
	return append([]list.Item(nil), built.items...)
}

// DailyEntries returns the flattened daily-eligible seeded modes.
func DailyEntries() []DailyEntry {
	return append([]DailyEntry(nil), built.dailies...)
}

// Lookup returns a definition by canonical name.
func Lookup(name string) (game.Definition, bool) {
	def, ok := built.byName[game.NormalizeName(name)]
	return def, ok
}

// Resolve returns a definition by canonical name or alias.
func Resolve(name string) (game.Definition, bool) {
	norm := game.NormalizeName(name)
	if def, ok := built.byName[norm]; ok {
		return def, true
	}
	canonical, ok := built.byAlias[norm]
	if !ok {
		return game.Definition{}, false
	}
	return Lookup(canonical)
}

// Import reconstructs a saved game for the given canonical game type.
func Import(gameType string, data []byte) (game.Gamer, error) {
	def, ok := Lookup(gameType)
	if !ok {
		return nil, fmt.Errorf("unknown game type %q in save data", gameType)
	}
	g, err := def.Import(data)
	if err != nil {
		return nil, fmt.Errorf("failed to import game: %w", err)
	}
	return g, nil
}
