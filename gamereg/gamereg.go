package gamereg

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

type ModeEntry struct {
	Definition puzzle.ModeDef
	Spawner    game.Spawner
	Seeded     game.SeededSpawner
}

type Entry struct {
	Definition puzzle.Definition
	Help       string
	Import     func([]byte) (game.Gamer, error)
	Modes      []ModeEntry
}

type EntrySpec struct {
	Definition puzzle.Definition
	Help       string
	Import     func([]byte) (game.Gamer, error)
	Modes      []game.Mode
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

func NewEntry(spec EntrySpec) Entry {
	if len(spec.Definition.Modes) != len(spec.Modes) {
		panic(fmt.Sprintf(
			"gamereg: definition %q has %d mode definitions but %d runtime modes",
			spec.Definition.Name,
			len(spec.Definition.Modes),
			len(spec.Modes),
		))
	}

	entries := make([]ModeEntry, 0, len(spec.Modes))
	for i, mode := range spec.Modes {
		spawner, ok := mode.(game.Spawner)
		if !ok {
			panic(fmt.Sprintf(
				"gamereg: definition %q mode %q does not implement game.Spawner",
				spec.Definition.Name,
				mode.Title(),
			))
		}

		modeDef := spec.Definition.Modes[i]
		if modeDef.Title != mode.Title() || modeDef.Description != mode.Description() {
			panic(fmt.Sprintf(
				"gamereg: definition %q mode %d metadata does not match runtime mode %q",
				spec.Definition.Name,
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
		if entry.Definition.Seeded != (entry.Seeded != nil) {
			panic(fmt.Sprintf(
				"gamereg: definition %q mode %q seeded flag does not match runtime mode",
				spec.Definition.Name,
				modeDef.Title,
			))
		}
		entries = append(entries, entry)
	}

	return Entry{
		Definition: spec.Definition,
		Help:       spec.Help,
		Import:     spec.Import,
		Modes:      entries,
	}
}
