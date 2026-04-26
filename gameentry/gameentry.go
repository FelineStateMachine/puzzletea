// Package gameentry builds concrete runtime game entries from puzzle definitions
// and runtime mode implementations.
package gameentry

import (
	"fmt"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/export/pdf"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

type ModeEntry struct {
	Definition puzzle.ModeDef
	Spawner    game.Spawner
	Seeded     game.SeededSpawner
	Elo        game.EloSpawner
}

type VariantEntry struct {
	Definition puzzle.VariantDef
	Elo        game.EloSpawner
	Seeded     game.SeededSpawner
}

type Entry struct {
	Definition  puzzle.Definition
	Help        string
	Import      func([]byte) (game.Gamer, error)
	Variants    []VariantEntry
	LegacyModes []puzzle.LegacyModeAlias
	Modes       []ModeEntry
	Print       pdfexport.PrintAdapter
}

type EntrySpec struct {
	Definition puzzle.Definition
	Help       string
	Import     func([]byte) (game.Gamer, error)
	Modes      []game.Mode
	Print      pdfexport.PrintAdapter
}

func BuildModeDefs(modes []game.Mode) []puzzle.ModeDef {
	defs := make([]puzzle.ModeDef, 0, len(modes))
	for _, mode := range modes {
		def := puzzle.NewModeDef(puzzle.ModeSpec{
			Title:       mode.Title(),
			Description: mode.Description(),
		})
		if _, ok := mode.(game.SeededSpawner); ok {
			def.Seeded = true
		}
		defs = append(defs, def)
	}
	return defs
}

func BuildDefaultVariantDef(title, description string, defaultElo difficulty.Elo) puzzle.VariantDef {
	return puzzle.NewVariantDef(puzzle.VariantSpec{
		Title:       title,
		Description: description,
		DefaultElo:  defaultElo,
	})
}

func BuildLegacyModeAliases(modes []puzzle.ModeDef, target puzzle.VariantID) []puzzle.LegacyModeAlias {
	aliases := make([]puzzle.LegacyModeAlias, 0, len(modes))
	for _, mode := range modes {
		if mode.PresetElo == nil {
			continue
		}
		aliases = append(aliases, puzzle.NewLegacyModeAlias(puzzle.LegacyModeAliasSpec{
			ID:              mode.ID,
			Title:           mode.Title,
			Description:     mode.Description,
			TargetVariantID: target,
			PresetElo:       *mode.PresetElo,
		}))
	}
	return aliases
}

func NewEntry(spec EntrySpec) Entry {
	definition := spec.Definition
	if len(definition.Variants) == 0 && hasEloMode(definition.Modes) {
		definition.Variants = []puzzle.VariantDef{
			BuildDefaultVariantDef(definition.Name, definition.Description, defaultEloForModes(definition.Modes)),
		}
	}
	if len(definition.LegacyModes) == 0 && len(definition.Variants) == 1 {
		definition.LegacyModes = BuildLegacyModeAliases(definition.Modes, definition.Variants[0].ID)
	}

	if len(definition.Modes) != len(spec.Modes) {
		panic(fmt.Sprintf(
			"gameentry: definition %q has %d mode definitions but %d runtime modes",
			definition.Name,
			len(definition.Modes),
			len(spec.Modes),
		))
	}

	entries := make([]ModeEntry, 0, len(spec.Modes))
	for i, mode := range spec.Modes {
		spawner, ok := mode.(game.Spawner)
		if !ok {
			panic(fmt.Sprintf(
				"gameentry: definition %q mode %q does not implement game.Spawner",
				definition.Name,
				mode.Title(),
			))
		}

		modeDef := definition.Modes[i]
		if modeDef.Title != mode.Title() || modeDef.Description != mode.Description() {
			panic(fmt.Sprintf(
				"gameentry: definition %q mode %d metadata does not match runtime mode %q",
				definition.Name,
				i,
				mode.Title(),
			))
		}

		entry := ModeEntry{
			Definition: modeDef,
			Spawner:    spawner,
		}
		if seeded, ok := mode.(game.SeededSpawner); ok {
			entry.Seeded = seeded
		}
		if elo, ok := mode.(game.EloSpawner); ok {
			entry.Elo = elo
		}
		if entry.Definition.Seeded != (entry.Seeded != nil) {
			panic(fmt.Sprintf(
				"gameentry: definition %q mode %q seeded flag does not match runtime mode",
				definition.Name,
				modeDef.Title,
			))
		}
		if entry.Definition.PresetElo != nil {
			if err := difficulty.ValidateElo(*entry.Definition.PresetElo); err != nil {
				panic(fmt.Sprintf(
					"gameentry: definition %q mode %q preset Elo is invalid: %v",
					definition.Name,
					modeDef.Title,
					err,
				))
			}
			if entry.Elo == nil {
				panic(fmt.Sprintf(
					"gameentry: definition %q mode %q has preset Elo but runtime mode does not implement game.EloSpawner",
					definition.Name,
					modeDef.Title,
				))
			}
		}
		entries = append(entries, entry)
	}

	variants := buildVariantEntries(definition, entries)

	return Entry{
		Definition:  definition,
		Help:        spec.Help,
		Import:      spec.Import,
		Variants:    variants,
		LegacyModes: append([]puzzle.LegacyModeAlias(nil), definition.LegacyModes...),
		Modes:       entries,
		Print:       spec.Print,
	}
}

func hasEloMode(modes []puzzle.ModeDef) bool {
	for _, mode := range modes {
		if mode.PresetElo != nil {
			return true
		}
	}
	return false
}

func defaultEloForModes(modes []puzzle.ModeDef) difficulty.Elo {
	presets := make([]difficulty.Elo, 0, len(modes))
	for _, mode := range modes {
		if mode.PresetElo != nil {
			presets = append(presets, *mode.PresetElo)
		}
	}
	if len(presets) == 0 {
		return 1200
	}
	return presets[len(presets)/2]
}

func buildVariantEntries(def puzzle.Definition, modes []ModeEntry) []VariantEntry {
	if len(def.Variants) == 0 {
		return nil
	}
	result := make([]VariantEntry, 0, len(def.Variants))
	for _, variant := range def.Variants {
		spawners := make([]ModeEntry, 0, len(modes))
		for _, mode := range modes {
			if mode.Elo != nil {
				spawners = append(spawners, mode)
			}
		}
		if len(spawners) == 0 {
			panic(fmt.Sprintf(
				"gameentry: definition %q variant %q has no Elo-capable legacy modes",
				def.Name,
				variant.Title,
			))
		}
		elo := variantEloSpawner{
			defaultElo: variant.DefaultElo,
			modes:      spawners,
		}
		result = append(result, VariantEntry{
			Definition: variant,
			Elo:        elo,
			Seeded: variantSeededSpawner{
				elo:        elo,
				defaultElo: variant.DefaultElo,
			},
		})
	}
	return result
}

type variantEloSpawner struct {
	defaultElo difficulty.Elo
	modes      []ModeEntry
}

func (s variantEloSpawner) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}
	mode := s.modeForElo(elo)
	return mode.Elo.SpawnElo(seed, elo)
}

func (s variantEloSpawner) modeForElo(elo difficulty.Elo) ModeEntry {
	if len(s.modes) == 1 {
		return s.modes[0]
	}
	band := int(elo) * len(s.modes) / (int(difficulty.SoftCapElo) + 1)
	if band < 0 {
		band = 0
	}
	if band >= len(s.modes) {
		band = len(s.modes) - 1
	}
	return s.modes[band]
}

type variantSeededSpawner struct {
	elo        game.EloSpawner
	defaultElo difficulty.Elo
}

func (s variantSeededSpawner) Spawn() (game.Gamer, error) {
	return s.SpawnSeeded(rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())))
}

func (s variantSeededSpawner) SpawnSeeded(rng *rand.Rand) (game.Gamer, error) {
	if rng == nil {
		return nil, fmt.Errorf("nil RNG")
	}
	seed := fmt.Sprintf("variant-seeded:%016x:%016x", rng.Uint64(), rng.Uint64())
	g, _, err := s.elo.SpawnElo(seed, s.defaultElo)
	return g, err
}
