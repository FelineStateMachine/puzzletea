package registry

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/fillomino"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/gamereg"
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

type (
	ModeEntry = gamereg.ModeEntry
	Entry     = gamereg.Entry
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
