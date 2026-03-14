package netwalk

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

const visualFixtureModeTitle = "Visual Fixture"

type visualFixtureCell struct {
	x    int
	y    int
	mask directionMask
	kind cellKind
	lock bool
}

type visualFixtureCase struct {
	name   string
	puzzle Puzzle
}

var visualFixtureCases = []visualFixtureCase{
	{
		name: "cursor-root-horizontal",
		puzzle: buildVisualFixturePuzzle(4, point{X: 0, Y: 0},
			serverTile(0, 0, east),
			nodeTile(1, 0, west),
			nodeTile(3, 2, south),
			nodeTile(3, 3, north),
		),
	},
	{
		name: "leaf-gallery",
		puzzle: buildVisualFixturePuzzle(5, point{X: 0, Y: 0},
			serverTile(0, 0, east),
			nodeTile(1, 0, west),
			nodeTile(4, 0, west),
			nodeTile(4, 1, north),
			nodeTile(3, 4, east),
			nodeTile(4, 4, north),
			nodeTile(0, 3, south),
			nodeTile(0, 4, east),
		),
	},
	{
		name: "straight-and-corner-gallery",
		puzzle: buildVisualFixturePuzzle(5, point{X: 0, Y: 0},
			serverTile(0, 0, east),
			nodeTile(1, 0, west),
			nodeTile(2, 0, east|west),
			nodeTile(3, 0, west|south),
			nodeTile(3, 1, north|south),
			nodeTile(3, 2, north|east),
			nodeTile(4, 2, west|south),
			nodeTile(4, 3, north|west),
			nodeTile(2, 2, east|south),
			nodeTile(2, 3, north),
		),
	},
	{
		name: "tee-and-cross-gallery",
		puzzle: buildVisualFixturePuzzle(5, point{X: 0, Y: 0},
			serverTile(0, 0, east),
			nodeTile(1, 0, west),
			nodeTile(3, 1, north|east|south),
			nodeTile(3, 0, south),
			nodeTile(4, 1, west),
			nodeTile(3, 2, north|east|south|west),
			nodeTile(2, 2, east),
			nodeTile(4, 2, west),
			nodeTile(3, 3, north),
		),
	},
	{
		name: "connected-horizontal-bridge",
		puzzle: buildVisualFixturePuzzle(5, point{X: 0, Y: 0},
			serverTile(0, 0, east),
			nodeTile(1, 0, east|west),
			nodeTile(2, 0, west),
			nodeTile(4, 3, south),
			nodeTile(4, 4, north),
		),
	},
	{
		name: "connected-vertical-bridge",
		puzzle: buildVisualFixturePuzzle(5, point{X: 0, Y: 0},
			serverTile(0, 0, south),
			nodeTile(0, 1, north|south),
			nodeTile(0, 2, north),
			nodeTile(4, 3, south),
			nodeTile(4, 4, north),
		),
	},
	{
		name: "disconnected-default-foreground",
		puzzle: buildVisualFixturePuzzle(5, point{X: 0, Y: 0},
			serverTile(0, 0, east),
			nodeTile(1, 0, west),
			nodeTile(3, 1, east|south),
			nodeTile(4, 1, south|west),
			nodeTile(3, 2, north|east),
			nodeTile(4, 2, north|west),
		),
	},
	{
		name: "dangling-error-state",
		puzzle: buildVisualFixturePuzzle(4, point{X: 0, Y: 0},
			serverTile(0, 0, east),
			nodeTile(1, 0, west),
			nodeTile(2, 2, east),
		),
	},
	{
		name: "locked-root-cursor",
		puzzle: buildVisualFixturePuzzle(4, point{X: 0, Y: 0},
			serverLockedTile(0, 0, east),
			nodeTile(1, 0, west),
			nodeTile(3, 2, south),
			nodeTile(3, 3, north),
		),
	},
	{
		name: "solved-with-empty-cells",
		puzzle: buildVisualFixturePuzzle(4, point{X: 1, Y: 1},
			serverTile(1, 1, east|south),
			nodeTile(2, 1, west),
			nodeTile(1, 2, north),
		),
	},
}

func VisualFixtureJSONL() ([]byte, error) {
	records, err := visualFixtureRecords()
	if err != nil {
		return nil, err
	}

	var b strings.Builder
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return nil, fmt.Errorf("marshal visual fixture record: %w", err)
		}
		b.Write(data)
		b.WriteByte('\n')
	}

	return []byte(b.String()), nil
}

func visualFixtureRecords() ([]pdfexport.JSONLRecord, error) {
	records := make([]pdfexport.JSONLRecord, 0, len(visualFixtureCases))
	for i, fixture := range visualFixtureCases {
		save, err := visualFixtureSave(fixture.puzzle)
		if err != nil {
			return nil, fmt.Errorf("fixture %q: %w", fixture.name, err)
		}

		records = append(records, pdfexport.JSONLRecord{
			Schema: pdfexport.ExportSchemaV1,
			Pack: pdfexport.JSONLPackMeta{
				Generated:     "2026-03-14T00:00:00Z",
				Version:       "visual-fixture",
				Category:      "Netwalk",
				ModeSelection: visualFixtureModeTitle,
				Count:         len(visualFixtureCases),
			},
			Puzzle: pdfexport.JSONLPuzzle{
				Index: i + 1,
				Name:  fixture.name,
				Game:  "Netwalk",
				Mode:  visualFixtureModeTitle,
				Save:  save,
			},
		})
	}

	return records, nil
}

func visualFixtureSave(puzzle Puzzle) (json.RawMessage, error) {
	cursor := puzzle.firstActive()
	m := Model{
		puzzle:    puzzle,
		keys:      DefaultKeyMap,
		modeTitle: visualFixtureModeTitle,
		cursor:    game.Cursor{X: cursor.X, Y: cursor.Y},
	}
	m.recompute()

	save, err := m.GetSave()
	if err != nil {
		return nil, fmt.Errorf("encode save: %w", err)
	}

	return json.RawMessage(save), nil
}

func buildVisualFixturePuzzle(size int, root point, cells ...visualFixtureCell) Puzzle {
	puzzle := newPuzzle(size)
	puzzle.Root = root
	for _, cell := range cells {
		puzzle.Tiles[cell.y][cell.x] = tile{
			BaseMask:        cell.mask,
			Rotation:        0,
			InitialRotation: 0,
			Locked:          cell.lock,
			Kind:            cell.kind,
		}
	}
	return puzzle
}

func serverTile(x, y int, mask directionMask) visualFixtureCell {
	return visualFixtureCell{x: x, y: y, mask: mask, kind: serverCell}
}

func serverLockedTile(x, y int, mask directionMask) visualFixtureCell {
	return visualFixtureCell{x: x, y: y, mask: mask, kind: serverCell, lock: true}
}

func nodeTile(x, y int, mask directionMask) visualFixtureCell {
	return visualFixtureCell{x: x, y: y, mask: mask, kind: nodeCell}
}
