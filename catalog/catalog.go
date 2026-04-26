// Package catalog indexes puzzle definitions, aliases, and daily metadata
// without importing concrete game implementations.
package catalog

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

type DailyEntry struct {
	GameID   puzzle.GameID
	GameName string
	ModeID   puzzle.ModeID
	Mode     string
}

type Index struct {
	definitions []puzzle.Definition
	names       []string
	dailies     []DailyEntry
	byName      map[string]puzzle.Definition
	byAlias     map[string]string
}

func MustBuild(definitions []puzzle.Definition) Index {
	idx, err := Build(definitions)
	if err != nil {
		panic(err)
	}
	return idx
}

func Build(definitions []puzzle.Definition) (Index, error) {
	if err := Validate(definitions); err != nil {
		return Index{}, err
	}

	idx := Index{
		definitions: append([]puzzle.Definition(nil), definitions...),
		names:       make([]string, 0, len(definitions)),
		dailies:     make([]DailyEntry, 0, len(definitions)*2),
		byName:      make(map[string]puzzle.Definition, len(definitions)),
		byAlias:     make(map[string]string),
	}

	for _, def := range definitions {
		idx.names = append(idx.names, def.Name)
		idx.byName[puzzle.NormalizeName(def.Name)] = def
		for _, alias := range def.Aliases {
			idx.byAlias[puzzle.NormalizeName(alias)] = def.Name
		}
		byModeID := make(map[puzzle.ModeID]puzzle.ModeDef, len(def.Modes))
		for _, mode := range def.Modes {
			byModeID[mode.ID] = mode
		}
		for _, dailyID := range def.DailyModeIDs {
			mode, ok := byModeID[dailyID]
			if !ok {
				continue
			}
			idx.dailies = append(idx.dailies, DailyEntry{
				GameID:   def.ID,
				GameName: def.Name,
				ModeID:   mode.ID,
				Mode:     mode.Title,
			})
		}
	}

	return idx, nil
}

func Validate(definitions []puzzle.Definition) error {
	names := make(map[string]string, len(definitions))
	aliases := make(map[string]string)

	for _, def := range definitions {
		normName := puzzle.NormalizeName(def.Name)
		if normName == "" {
			return fmt.Errorf("game %q has empty canonical name", def.Name)
		}
		if def.ID == "" {
			return fmt.Errorf("game %q has empty id", def.Name)
		}
		if prior, exists := names[normName]; exists {
			return fmt.Errorf("duplicate game name %q conflicts with %q", def.Name, prior)
		}
		names[normName] = def.Name

		modeTitles := make(map[puzzle.ModeID]struct{}, len(def.Modes))
		for _, mode := range def.Modes {
			if mode.ID == "" {
				return fmt.Errorf("game %q has mode %q with empty id", def.Name, mode.Title)
			}
			modeTitles[mode.ID] = struct{}{}
		}
		variantIDs := make(map[puzzle.VariantID]struct{}, len(def.Variants))
		for _, variant := range def.Variants {
			if variant.ID == "" {
				return fmt.Errorf("game %q has variant %q with empty id", def.Name, variant.Title)
			}
			if err := difficulty.ValidateElo(variant.DefaultElo); err != nil {
				return fmt.Errorf("game %q variant %q default Elo is invalid: %w", def.Name, variant.Title, err)
			}
			if _, exists := variantIDs[variant.ID]; exists {
				return fmt.Errorf("game %q has duplicate variant %q", def.Name, variant.ID)
			}
			variantIDs[variant.ID] = struct{}{}
		}
		for _, alias := range def.LegacyModes {
			if alias.ID == "" {
				return fmt.Errorf("game %q has legacy mode %q with empty id", def.Name, alias.Title)
			}
			if _, ok := variantIDs[alias.TargetVariantID]; !ok {
				return fmt.Errorf("game %q legacy mode %q targets missing variant %q", def.Name, alias.Title, alias.TargetVariantID)
			}
			if err := difficulty.ValidateElo(alias.PresetElo); err != nil {
				return fmt.Errorf("game %q legacy mode %q preset Elo is invalid: %w", def.Name, alias.Title, err)
			}
			if alias.XPWeight < 1 {
				return fmt.Errorf("game %q legacy mode %q XP weight must be positive", def.Name, alias.Title)
			}
		}

		for _, dailyID := range def.DailyModeIDs {
			if _, ok := modeTitles[dailyID]; !ok {
				return fmt.Errorf("game %q daily mode %q is not present in Modes", def.Name, dailyID)
			}
		}

		for _, alias := range def.Aliases {
			normAlias := puzzle.NormalizeName(alias)
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

func (i Index) Names() []string {
	return append([]string(nil), i.names...)
}

func (i Index) Definitions() []puzzle.Definition {
	return append([]puzzle.Definition(nil), i.definitions...)
}

func (i Index) DailyEntries() []DailyEntry {
	return append([]DailyEntry(nil), i.dailies...)
}

func (i Index) Lookup(name string) (puzzle.Definition, bool) {
	def, ok := i.byName[puzzle.NormalizeName(name)]
	return def, ok
}

func (i Index) Resolve(name string) (puzzle.Definition, bool) {
	norm := puzzle.NormalizeName(name)
	if def, ok := i.byName[norm]; ok {
		return def, true
	}
	canonical, ok := i.byAlias[norm]
	if !ok {
		return puzzle.Definition{}, false
	}
	return i.Lookup(canonical)
}
