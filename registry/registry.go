// Package registry is the concrete runtime composition root for built-in games,
// imports, help text, and daily-capable modes.
package registry

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/gameentry"
	"github.com/FelineStateMachine/puzzletea/games/fillomino"
	"github.com/FelineStateMachine/puzzletea/games/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/games/hitori"
	"github.com/FelineStateMachine/puzzletea/games/lightsout"
	"github.com/FelineStateMachine/puzzletea/games/netwalk"
	"github.com/FelineStateMachine/puzzletea/games/nonogram"
	"github.com/FelineStateMachine/puzzletea/games/nurikabe"
	"github.com/FelineStateMachine/puzzletea/games/rippleeffect"
	"github.com/FelineStateMachine/puzzletea/games/shikaku"
	"github.com/FelineStateMachine/puzzletea/games/spellpuzzle"
	"github.com/FelineStateMachine/puzzletea/games/sudoku"
	"github.com/FelineStateMachine/puzzletea/games/sudokurgb"
	"github.com/FelineStateMachine/puzzletea/games/takuzu"
	"github.com/FelineStateMachine/puzzletea/games/takuzuplus"
	"github.com/FelineStateMachine/puzzletea/games/wordsearch"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

type (
	ModeEntry    = gameentry.ModeEntry
	VariantEntry = gameentry.VariantEntry
	Entry        = gameentry.Entry
)

type DailyEntry struct {
	Spawner  game.SeededSpawner
	GameID   puzzle.GameID
	GameType string
	ModeID   puzzle.ModeID
	Mode     string
}

var all = []Entry{
	fillomino.Entry,
	hashiwokakero.Entry,
	hitori.Entry,
	lightsout.Entry,
	netwalk.Entry,
	nonogram.Entry,
	nurikabe.Entry,
	rippleeffect.Entry,
	shikaku.Entry,
	spellpuzzle.Entry,
	sudoku.Entry,
	sudokurgb.Entry,
	takuzu.Entry,
	takuzuplus.Entry,
	wordsearch.Entry,
}

var (
	definitions = buildDefinitions(all)
	index       = catalog.MustBuild(definitions)
	entriesByID = buildEntriesByID(all)
)

func buildDefinitions(entries []Entry) []puzzle.Definition {
	defs := make([]puzzle.Definition, 0, len(entries))
	for _, entry := range entries {
		defs = append(defs, entry.Definition)
	}
	return defs
}

func buildEntriesByID(entries []Entry) map[puzzle.GameID]Entry {
	result := make(map[puzzle.GameID]Entry, len(entries))
	for _, entry := range entries {
		result[entry.Definition.ID] = entry
	}
	return result
}

func Entries() []Entry {
	return append([]Entry(nil), all...)
}

func Definitions() []puzzle.Definition {
	return index.Definitions()
}

func Names() []string {
	return index.Names()
}

func Lookup(name string) (Entry, bool) {
	def, ok := index.Lookup(name)
	if !ok {
		return Entry{}, false
	}
	entry, ok := entriesByID[def.ID]
	return entry, ok
}

func Resolve(name string) (Entry, bool) {
	def, ok := index.Resolve(name)
	if !ok {
		return Entry{}, false
	}
	entry, ok := entriesByID[def.ID]
	return entry, ok
}

func Import(gameType string, data []byte) (game.Gamer, error) {
	entry, ok := Lookup(gameType)
	if !ok {
		return nil, fmt.Errorf("unknown game type %q in save data", gameType)
	}
	g, err := entry.Import(data)
	if err != nil {
		return nil, fmt.Errorf("failed to import game: %w", err)
	}
	return g, nil
}

func DailyEntries() []DailyEntry {
	dailies := index.DailyEntries()
	result := make([]DailyEntry, 0, len(dailies))
	seen := make(map[puzzle.GameID]struct{}, len(dailies))
	for _, daily := range dailies {
		if _, ok := seen[daily.GameID]; ok {
			continue
		}
		entry, ok := entriesByID[daily.GameID]
		if !ok {
			continue
		}
		if len(entry.Variants) > 0 && entry.Variants[0].Seeded != nil {
			variant := entry.Variants[0]
			seen[daily.GameID] = struct{}{}
			result = append(result, DailyEntry{
				Spawner:  variant.Seeded,
				GameID:   daily.GameID,
				GameType: entry.Definition.Name,
				ModeID:   puzzle.ModeID(variant.Definition.ID),
				Mode:     variant.Definition.Title,
			})
		}
	}
	return result
}

func Help(name string) (string, bool) {
	entry, ok := Resolve(name)
	if !ok {
		return "", false
	}
	return entry.Help, true
}
