package takuzuplus

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math/rand/v2"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/charmbracelet/x/ansi"
)

func makeGrid4(rows ...string) grid {
	if len(rows) != 4 {
		panic("makeGrid4 requires exactly 4 rows")
	}
	g := make(grid, 4)
	for i, row := range rows {
		g[i] = []rune(row)
	}
	return g
}

func validGrid4() grid {
	return makeGrid4(
		"0011",
		"1100",
		"1010",
		"0101",
	)
}

func TestCanPlaceWithRelations(t *testing.T) {
	rels := newRelations(4)
	rels.horizontal[0][0] = relationSame
	rels.horizontal[0][1] = relationDiff

	g := makeGrid4(
		"0...",
		"....",
		"....",
		"....",
	)

	if canPlaceWithRelations(g, 4, 1, 0, oneCell, rels) {
		t.Fatal("expected same relation to reject opposite value")
	}
	if !canPlaceWithRelations(g, 4, 1, 0, zeroCell, rels) {
		t.Fatal("expected same relation to allow equal value")
	}

	g[0][1] = zeroCell
	if canPlaceWithRelations(g, 4, 2, 0, zeroCell, rels) {
		t.Fatal("expected opposite relation to reject equal value")
	}
	if !canPlaceWithRelations(g, 4, 2, 0, oneCell, rels) {
		t.Fatal("expected opposite relation to allow different value")
	}
}

func TestCountSolutionsWithRelations(t *testing.T) {
	g := makeGrid4(
		"001.",
		"1100",
		"1010",
		"0101",
	)
	rels := newRelations(4)
	rels.horizontal[0][2] = relationSame

	if got := countSolutions(g.clone(), 4, 2, rels); got != 1 {
		t.Fatalf("countSolutions = %d, want 1", got)
	}
}

func TestGeneratePuzzle(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	complete := generateCompleteSeeded(6, rng)
	mode := Modes[1].(TakuzuPlusMode)
	puzzle, provided, rels := generatePuzzleSeeded(complete, 6, 0.40, mode.profile, rng)

	if got := countSolutions(puzzle.clone(), 6, 2, rels); got != 1 {
		t.Fatalf("generated puzzle has %d solutions, want 1", got)
	}

	relationCount := 0
	for y, row := range rels.horizontal {
		for x, rel := range row {
			if rel == relationNone {
				continue
			}
			relationCount++
			if want := relationForValues(complete[y][x], complete[y][x+1]); rel != want {
				t.Fatalf("horizontal relation (%d,%d) = %q, want %q", x, y, rel, want)
			}
		}
	}
	for y, row := range rels.vertical {
		for x, rel := range row {
			if rel == relationNone {
				continue
			}
			relationCount++
			if want := relationForValues(complete[y][x], complete[y+1][x]); rel != want {
				t.Fatalf("vertical relation (%d,%d) = %q, want %q", x, y, rel, want)
			}
		}
	}
	if relationCount == 0 {
		t.Fatal("expected seeded Takuzu+ generation to emit at least one relation clue")
	}

	for y := range puzzle {
		for x := range puzzle[y] {
			isEmpty := puzzle[y][x] == emptyCell
			if isEmpty == provided[y][x] {
				t.Fatalf("provided mismatch at (%d,%d): empty=%v provided=%v", x, y, isEmpty, provided[y][x])
			}
		}
	}
}

func TestRelationChoiceScorePrefersMixedOverTwoProvided(t *testing.T) {
	profile := modeProfiles[0]
	complete := makeGrid4(
		"0011",
		"1100",
		"1010",
		"0101",
	)
	provided := [][]bool{
		{true, true, true, true},
		{true, true, false, true},
		{true, true, true, true},
		{true, true, true, true},
	}
	rels := newRelations(4)

	choices := additiveRelationChoices(complete, provided, rels, 4, profile)
	if len(choices) == 0 {
		t.Fatal("expected additive relation choices")
	}
	if choices[0].endpointClass != endpointOneProvided {
		t.Fatalf("top additive choice endpoint class = %d, want mixed(1)", choices[0].endpointClass)
	}
}

func TestRelationChoiceScorePrefersUnderrepresentedSymbolAndOrientation(t *testing.T) {
	profile := modeProfiles[3]
	provided := [][]bool{
		{true, true, true, true},
		{true, true, true, true},
		{true, true, true, true},
		{true, true, true, true},
	}
	rels := newRelations(4)
	rels.horizontal[0][0] = relationSame
	rels.horizontal[1][0] = relationSame
	metrics := analyzeRelations(rels, provided, 4)

	diffVertical := relationChoice{
		candidate:     relationCandidate{horizontal: false, x: 2, y: 0},
		endpointClass: endpointTwoProvided,
		symbol:        relationDiff,
	}
	diffVertical.score = relationChoiceScore(diffVertical, profile, metrics, 4)

	sameHorizontal := relationChoice{
		candidate:     relationCandidate{horizontal: true, x: 2, y: 2},
		endpointClass: endpointTwoProvided,
		symbol:        relationSame,
	}
	sameHorizontal.score = relationChoiceScore(sameHorizontal, profile, metrics, 4)

	if diffVertical.score <= sameHorizontal.score {
		t.Fatalf("expected underrepresented diff/vertical clue score %d to exceed same/horizontal score %d", diffVertical.score, sameHorizontal.score)
	}
}

func TestRelationChoiceScoreRespectsSpacingPenalty(t *testing.T) {
	profile := modeProfiles[0]
	provided := [][]bool{
		{true, true, true, true},
		{true, true, true, true},
		{true, true, true, true},
		{true, true, true, true},
	}
	rels := newRelations(4)
	rels.horizontal[0][0] = relationSame
	metrics := analyzeRelations(rels, provided, 4)

	closeChoice := relationChoice{
		candidate:     relationCandidate{horizontal: true, x: 1, y: 0},
		endpointClass: endpointTwoProvided,
		symbol:        relationDiff,
	}
	closeChoice.score = relationChoiceScore(closeChoice, profile, metrics, 4)

	farChoice := relationChoice{
		candidate:     relationCandidate{horizontal: true, x: 2, y: 3},
		endpointClass: endpointTwoProvided,
		symbol:        relationDiff,
	}
	farChoice.score = relationChoiceScore(farChoice, profile, metrics, 4)

	if farChoice.score <= closeChoice.score {
		t.Fatalf("expected far clue score %d to exceed clustered clue score %d", farChoice.score, closeChoice.score)
	}
}

func TestGeneratePuzzleSeeded_ModeRelationBands(t *testing.T) {
	for i, item := range Modes {
		mode := item.(TakuzuPlusMode)
		t.Run(mode.Title(), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(101+i), uint64(701+i)))
			complete := generateCompleteSeeded(mode.Size, rng)
			_, _, rels := generatePuzzleSeeded(complete, mode.Size, mode.Prefilled, mode.profile, rng)

			if count := countRelations(rels); count < mode.profile.MinRelations || count > mode.profile.MaxRelations {
				t.Fatalf("relation count = %d, want %d..%d", count, mode.profile.MinRelations, mode.profile.MaxRelations)
			}
		})
	}
}

func TestSampledModeDistributionStaysNearProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sampled distribution test in short mode")
	}

	sampleCounts := []int{4, 4, 4, 3, 2, 1, 0}
	for i, item := range Modes {
		mode := item.(TakuzuPlusMode)
		samples := sampleCounts[i]
		if samples == 0 {
			continue
		}

		t.Run(fmt.Sprintf("%s-%d", mode.Title(), samples), func(t *testing.T) {
			totalRelations := 0
			totalHorizontal := 0
			totalSame := 0

			for sample := range samples {
				rng := rand.New(rand.NewPCG(uint64(1000+i*37+sample), uint64(2000+i*41+sample)))
				complete := generateCompleteSeeded(mode.Size, rng)
				_, provided, rels := generatePuzzleSeeded(complete, mode.Size, mode.Prefilled, mode.profile, rng)
				metrics := analyzeRelations(rels, provided, mode.Size)
				totalRelations += metrics.Total
				totalHorizontal += metrics.HorizontalCount
				totalSame += metrics.SameCount
			}

			avgRelations := float64(totalRelations) / float64(samples)
			targetMid := float64(mode.profile.MinRelations+mode.profile.MaxRelations) / 2
			if diff := avgRelations - targetMid; diff < -2.0 || diff > 2.0 {
				t.Fatalf("average relations %.2f too far from target midpoint %.2f", avgRelations, targetMid)
			}

			if totalRelations >= 5 {
				horizontalRatio := float64(totalHorizontal) / float64(totalRelations)
				sameRatio := float64(totalSame) / float64(totalRelations)
				if horizontalRatio < 0.30 || horizontalRatio > 0.70 {
					t.Fatalf("horizontal ratio %.2f too skewed", horizontalRatio)
				}
				if sameRatio < 0.30 || sameRatio > 0.70 {
					t.Fatalf("same-symbol ratio %.2f too skewed", sameRatio)
				}
			}
		})
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	rels := newRelations(4)
	rels.horizontal[0][0] = relationSame
	rels.vertical[1][2] = relationDiff

	puzzle := makeGrid4(
		"0.11",
		"1.0.",
		".0.1",
		"0101",
	)
	provided := [][]bool{
		{true, false, true, true},
		{true, false, true, false},
		{false, true, false, true},
		{true, true, true, true},
	}

	model := Model{
		size:        4,
		grid:        puzzle,
		initialGrid: puzzle.clone(),
		provided:    provided,
		relations:   rels,
		keys:        DefaultKeyMap,
		modeTitle:   "Medium",
	}

	data, err := model.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.relations.horizontal[0][0] != relationSame {
		t.Fatal("horizontal relation not preserved")
	}
	if loaded.relations.vertical[1][2] != relationDiff {
		t.Fatal("vertical relation not preserved")
	}
	if loaded.provided[0][1] {
		t.Fatal("provided mask should preserve editable cell")
	}
}

func TestGridViewRendersRelationClues(t *testing.T) {
	rels := newRelations(4)
	rels.horizontal[0][0] = relationDiff
	rels.vertical[0][1] = relationSame

	m := Model{
		size: 4,
		grid: validGrid4(),
		provided: [][]bool{
			{true, true, true, true},
			{true, true, true, true},
			{true, true, true, true},
			{true, true, true, true},
		},
		relations: rels,
		keys:      DefaultKeyMap,
	}

	view := ansi.Strip(gridView(m))
	if !strings.Contains(view, "x") {
		t.Fatal("expected vertical relation clue in view")
	}
	if !strings.Contains(view, "=") {
		t.Fatal("expected horizontal relation clue in view")
	}
}

func TestSaveJSONFields(t *testing.T) {
	m := Model{
		size:        4,
		grid:        validGrid4(),
		initialGrid: validGrid4(),
		provided: [][]bool{
			{true, true, true, true},
			{true, true, true, true},
			{true, true, true, true},
			{true, true, true, true},
		},
		relations: newRelations(4),
		keys:      DefaultKeyMap,
	}

	data, err := m.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	var save Save
	if err := json.Unmarshal(data, &save); err != nil {
		t.Fatal(err)
	}
	if save.HorizontalRelations == "" {
		t.Fatal("expected horizontal_relations to be serialized")
	}
	if save.VerticalRelations == "" {
		t.Fatal("expected vertical_relations to be serialized")
	}
}

func TestMouseClickMovesCursor(t *testing.T) {
	m := testMouseModel()

	x, y := takuzuPlusCellScreenCoords(&m, 2, 1)
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	got := next.(Model)

	if got.cursor.X != 2 || got.cursor.Y != 1 {
		t.Fatalf("cursor = (%d,%d), want (2,1)", got.cursor.X, got.cursor.Y)
	}
	if got.grid[1][2] != emptyCell {
		t.Fatalf("clicked destination cell changed to %q, want unchanged empty", got.grid[1][2])
	}
}

func TestMouseClickSameCellCyclesEditableCell(t *testing.T) {
	m := testMouseModel()
	m.cursor.X, m.cursor.Y = 1, 1

	x, y := takuzuPlusCellScreenCoords(&m, 1, 1)

	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	got := next.(Model)
	if got.grid[1][1] != zeroCell {
		t.Fatalf("first click = %q, want %q", got.grid[1][1], zeroCell)
	}

	next, _ = got.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	got = next.(Model)
	if got.grid[1][1] != oneCell {
		t.Fatalf("second click = %q, want %q", got.grid[1][1], oneCell)
	}

	next, _ = got.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	got = next.(Model)
	if got.grid[1][1] != emptyCell {
		t.Fatalf("third click = %q, want %q", got.grid[1][1], emptyCell)
	}
}

func TestMouseClickSameCellDoesNotCycleProvidedCell(t *testing.T) {
	m := testMouseModel()
	m.cursor.X, m.cursor.Y = 0, 0

	x, y := takuzuPlusCellScreenCoords(&m, 0, 0)
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	got := next.(Model)

	if got.grid[0][0] != zeroCell {
		t.Fatalf("provided cell changed to %q, want %q", got.grid[0][0], zeroCell)
	}
}

func TestRelationBridgeBackgroundSatisfied(t *testing.T) {
	m := testMouseModel()
	m.grid[1][1] = zeroCell
	m.grid[1][2] = zeroCell
	m.relations.horizontal[1][1] = relationSame

	got := relationBridgeBackground(m, game.DynamicGridBridge{
		Kind:  game.DynamicGridBridgeVertical,
		X:     2,
		Y:     1,
		Count: 2,
	})

	if !sameColor(got, theme.Current().SuccessBG) {
		t.Fatal("expected satisfied relation clue to use success background")
	}
}

func TestRelationBridgeBackgroundConflict(t *testing.T) {
	m := testMouseModel()
	m.grid[0][0] = zeroCell
	m.grid[1][0] = zeroCell
	m.relations.vertical[0][0] = relationDiff

	got := relationBridgeBackground(m, game.DynamicGridBridge{
		Kind:  game.DynamicGridBridgeHorizontal,
		X:     0,
		Y:     1,
		Count: 2,
	})

	if !sameColor(got, game.ConflictBG()) {
		t.Fatal("expected conflicting relation clue to use conflict background")
	}
}

func TestRelationBridgeBackgroundIncompleteIsNeutral(t *testing.T) {
	m := testMouseModel()
	m.grid[1][1] = emptyCell
	m.grid[1][2] = oneCell
	m.relations.horizontal[1][1] = relationDiff

	got := relationBridgeBackground(m, game.DynamicGridBridge{
		Kind:  game.DynamicGridBridgeVertical,
		X:     2,
		Y:     1,
		Count: 2,
	})

	if got != nil {
		t.Fatal("expected incomplete relation clue to stay neutral")
	}
}

func testMouseModel() Model {
	return Model{
		size: 4,
		grid: makeGrid4(
			"0..1",
			"....",
			"....",
			"1..0",
		),
		initialGrid: makeGrid4(
			"0..1",
			"....",
			"....",
			"1..0",
		),
		provided: [][]bool{
			{true, false, false, true},
			{false, false, false, false},
			{false, false, false, false},
			{true, false, false, true},
		},
		relations:   newRelations(4),
		keys:        DefaultKeyMap,
		modeTitle:   "Test",
		termWidth:   120,
		termHeight:  40,
		originValid: false,
	}
}

func takuzuPlusCellScreenCoords(m *Model, col, row int) (int, int) {
	ox, oy := m.gridOrigin()
	return ox + col*(cellWidth+1), oy + row*2
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}
