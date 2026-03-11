package registry

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/fillomino"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/nurikabe"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/rippleeffect"
	"github.com/FelineStateMachine/puzzletea/shikaku"
	"github.com/FelineStateMachine/puzzletea/spellpuzzle"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/sudokurgb"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/takuzuplus"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
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

type DailyEntry struct {
	Spawner  game.SeededSpawner
	GameID   puzzle.GameID
	GameType string
	ModeID   puzzle.ModeID
	Mode     string
}

var all = []Entry{
	adaptLegacy(fillomino.Definition),
	adaptLegacy(hashiwokakero.Definition),
	adaptLegacy(hitori.Definition),
	adaptLegacy(lightsout.Definition),
	adaptLegacy(nonogram.Definition),
	adaptLegacy(nurikabe.Definition),
	adaptLegacy(rippleeffect.Definition),
	adaptLegacy(shikaku.Definition),
	adaptLegacy(spellpuzzle.Definition),
	adaptLegacy(sudoku.Definition),
	adaptLegacy(sudokurgb.Definition),
	adaptLegacy(takuzu.Definition),
	adaptLegacy(takuzuplus.Definition),
	adaptLegacy(wordsearch.Definition),
}

var (
	definitions = buildDefinitions(all)
	index       = catalog.MustBuild(definitions)
	entriesByID = buildEntriesByID(all)
)

func adaptLegacy(def game.Definition) Entry {
	gameID := puzzle.CanonicalGameID(def.Name)
	modes := make([]ModeEntry, 0, len(def.Modes))
	for _, item := range def.Modes {
		mode, ok := item.(game.Mode)
		if !ok {
			continue
		}
		spawner, ok := item.(game.Spawner)
		if !ok {
			continue
		}
		modeID := puzzle.CanonicalModeID(mode.Title())
		modeDef := puzzle.ModeDef{
			ID:          modeID,
			Title:       mode.Title(),
			Description: mode.Description(),
		}
		var seeded game.SeededSpawner
		if s, ok := item.(game.SeededSpawner); ok {
			modeDef.Seeded = true
			seeded = s
		}
		modes = append(modes, ModeEntry{
			Definition: modeDef,
			Spawner:    spawner,
			Seeded:     seeded,
		})
	}

	dailyIDs := make([]puzzle.ModeID, 0, len(def.DailyModes))
	for _, item := range def.DailyModes {
		mode, ok := item.(game.Mode)
		if !ok {
			continue
		}
		dailyIDs = append(dailyIDs, puzzle.CanonicalModeID(mode.Title()))
	}

	return Entry{
		Definition: puzzle.Definition{
			ID:           gameID,
			Name:         def.Name,
			Description:  def.Description,
			Aliases:      append([]string(nil), def.Aliases...),
			Modes:        extractModeDefs(modes),
			DailyModeIDs: dailyIDs,
		},
		Help:   def.Help,
		Import: def.Import,
		Modes:  modes,
	}
}

func extractModeDefs(modes []ModeEntry) []puzzle.ModeDef {
	defs := make([]puzzle.ModeDef, 0, len(modes))
	for _, mode := range modes {
		defs = append(defs, mode.Definition)
	}
	return defs
}

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
	for _, daily := range dailies {
		entry, ok := entriesByID[daily.GameID]
		if !ok {
			continue
		}
		for _, mode := range entry.Modes {
			if mode.Definition.ID != daily.ModeID || mode.Seeded == nil {
				continue
			}
			result = append(result, DailyEntry{
				Spawner:  mode.Seeded,
				GameID:   daily.GameID,
				GameType: entry.Definition.Name,
				ModeID:   daily.ModeID,
				Mode:     mode.Definition.Title,
			})
			break
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
