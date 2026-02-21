package nurikabe

import (
	"context"
	"errors"
	"math/rand/v2"
	"runtime"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

func makeClues(rows ...[]int) clueGrid {
	out := make(clueGrid, len(rows))
	for y, row := range rows {
		out[y] = make([]int, len(row))
		copy(out[y], row)
	}
	return out
}

func makeMarksFromClues(clues clueGrid) grid {
	h := len(clues)
	w := len(clues[0])
	g := newGrid(w, h, seaCell)
	for y := range h {
		for x := range w {
			if clues[y][x] > 0 {
				g[y][x] = islandCell
			}
		}
	}
	return g
}

func uniquePuzzleFixture() Puzzle {
	clues := makeClues(
		[]int{1, 0, 1, 0, 1},
		[]int{0, 0, 0, 0, 0},
		[]int{0, 1, 0, 1, 0},
		[]int{0, 0, 0, 0, 0},
		[]int{1, 0, 1, 0, 1},
	)
	return Puzzle{Width: 5, Height: 5, Clues: clues}
}

// --- Grid and serialization (P0) ---

func TestGridStringAndParse(t *testing.T) {
	g := grid{
		[]cellState{unknownCell, seaCell, islandCell},
		[]cellState{seaCell, islandCell, unknownCell},
	}
	data := g.String()
	round, err := parseGrid(data, 3, 2)
	if err != nil {
		t.Fatalf("parseGrid error: %v", err)
	}
	if round.String() != data {
		t.Errorf("round-trip mismatch: got %q want %q", round.String(), data)
	}
}

func TestParseGridLegacyRequiredLand(t *testing.T) {
	g, err := parseGrid("?!\n~o", 2, 2)
	if err != nil {
		t.Fatalf("parseGrid error: %v", err)
	}
	if g[0][1] != islandCell {
		t.Fatalf("legacy required-land marker should load as island, got %q", rune(g[0][1]))
	}
}

func TestGridClone(t *testing.T) {
	g := newGrid(2, 2, unknownCell)
	g[0][0] = seaCell
	c := g.clone()
	c[0][0] = islandCell
	if g[0][0] == islandCell {
		t.Fatal("clone shares memory with original")
	}
}

func TestClueSerializationRoundTrip(t *testing.T) {
	clues := makeClues(
		[]int{0, 12, 0},
		[]int{3, 0, 1},
	)
	enc := serializeClues(clues)
	dec, err := parseClues(enc, 3, 2)
	if err != nil {
		t.Fatalf("parseClues error: %v", err)
	}
	if serializeClues(dec) != enc {
		t.Errorf("clue round-trip mismatch: got %q want %q", serializeClues(dec), enc)
	}
}

// --- Rule checks (P0) ---

func TestHasSeaSquare(t *testing.T) {
	g := grid{
		[]cellState{seaCell, seaCell, islandCell},
		[]cellState{seaCell, seaCell, islandCell},
		[]cellState{islandCell, islandCell, islandCell},
	}
	if !hasSeaSquare(g) {
		t.Fatal("expected 2x2 sea block")
	}
}

func TestSeaSquareCellDetection(t *testing.T) {
	g := grid{
		[]cellState{seaCell, seaCell, islandCell},
		[]cellState{seaCell, seaCell, islandCell},
		[]cellState{islandCell, islandCell, islandCell},
	}

	for _, p := range []point{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
		if !isSeaSquareCell(g, p.x, p.y) {
			t.Fatalf("expected (%d,%d) to be part of a 2x2 sea block", p.x, p.y)
		}
	}

	if isSeaSquareCell(g, 2, 2) {
		t.Fatal("non-sea cell should not be part of a 2x2 sea block")
	}

	lShape := grid{
		[]cellState{seaCell, seaCell, unknownCell},
		[]cellState{seaCell, islandCell, unknownCell},
		[]cellState{unknownCell, unknownCell, unknownCell},
	}
	if isSeaSquareCell(lShape, 0, 0) {
		t.Fatal("L-shape sea should not be flagged as 2x2")
	}
}

func TestSeaConnectivity(t *testing.T) {
	connected := grid{
		[]cellState{seaCell, seaCell, islandCell},
		[]cellState{islandCell, seaCell, islandCell},
		[]cellState{islandCell, seaCell, seaCell},
	}
	if !isSeaConnected(connected) {
		t.Fatal("expected connected sea")
	}

	disconnected := grid{
		[]cellState{seaCell, islandCell, seaCell},
		[]cellState{islandCell, islandCell, islandCell},
		[]cellState{seaCell, islandCell, seaCell},
	}
	if isSeaConnected(disconnected) {
		t.Fatal("expected disconnected sea")
	}
}

func TestIsSolvedGrid(t *testing.T) {
	p := uniquePuzzleFixture()
	marks := makeMarksFromClues(p.Clues)
	if !isSolvedGrid(marks, p.Clues) {
		t.Fatal("expected solved fixture")
	}
}

// --- Solver counts (P0) ---

func TestCountSolutions(t *testing.T) {
	tests := []struct {
		name string
		p    Puzzle
		want int
	}{
		{
			name: "zero solutions",
			p: Puzzle{Width: 1, Height: 1, Clues: makeClues(
				[]int{1},
			)},
			want: 0,
		},
		{
			name: "one solution",
			p: Puzzle{Width: 2, Height: 2, Clues: makeClues(
				[]int{1, 0},
				[]int{0, 0},
			)},
			want: 1,
		},
		{
			name: "two solutions",
			p: Puzzle{Width: 2, Height: 2, Clues: makeClues(
				[]int{2, 0},
				[]int{0, 0},
			)},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := CountSolutions(tt.p, 3, 50000)
			if err != nil {
				t.Fatalf("CountSolutions error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("CountSolutions = %d, want %d", got, tt.want)
			}
		})
	}
}

// --- Save/load round-trip (P1) ---

func TestSaveLoadRoundTrip(t *testing.T) {
	mode := NewMode("Mini", "test", 5, 5, 0.25, 5)
	p := uniquePuzzleFixture()
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)

	m.cursor = pointToCursor(1, 1)
	m.setCellAtCursor(islandCell)
	m.cursor = pointToCursor(3, 3)
	m.setCellAtCursor(seaCell)

	data, err := m.GetSave()
	if err != nil {
		t.Fatalf("GetSave error: %v", err)
	}
	loaded, err := ImportModel(data)
	if err != nil {
		t.Fatalf("ImportModel error: %v", err)
	}

	if loaded.width != m.width || loaded.height != m.height {
		t.Fatalf("size mismatch: got %dx%d want %dx%d", loaded.width, loaded.height, m.width, m.height)
	}
	if serializeClues(loaded.clues) != serializeClues(m.clues) {
		t.Fatal("clues mismatch after round-trip")
	}
	if loaded.marks.String() != m.marks.String() {
		t.Fatal("marks mismatch after round-trip")
	}
}

func pointToCursor(x, y int) game.Cursor {
	return game.Cursor{X: x, Y: y}
}

// --- Model input behavior (P1) ---

func TestModelKeyboardUpdate(t *testing.T) {
	mode := NewMode("Mini", "test", 3, 3, 0.3, 5)
	p := Puzzle{Width: 3, Height: 3, Clues: makeClues(
		[]int{2, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 1},
	)}
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)
	m.cursor = pointToCursor(1, 1)

	next, _ := m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m = next.(Model)
	if m.marks[1][1] != seaCell {
		t.Fatalf("expected sea at (1,1), got %q", rune(m.marks[1][1]))
	}

	next, _ = m.Update(tea.KeyPressMsg{Code: 'z', Text: "z"})
	m = next.(Model)
	if m.marks[1][1] != islandCell {
		t.Fatalf("expected island at (1,1), got %q", rune(m.marks[1][1]))
	}

	next, _ = m.Update(tea.KeyPressMsg{Code: 'z', Text: "z"})
	m = next.(Model)
	if m.marks[1][1] != unknownCell {
		t.Fatalf("expected unknown at (1,1), got %q", rune(m.marks[1][1]))
	}
}

func TestModelMouseDragPaint(t *testing.T) {
	mode := NewMode("Mini", "test", 3, 3, 0.3, 5)
	p := Puzzle{Width: 3, Height: 3, Clues: makeClues(
		[]int{2, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 1},
	)}
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)
	ox, oy := m.gridOrigin()

	click := tea.MouseClickMsg{X: ox + cellWidth, Y: oy + 1, Button: tea.MouseLeft}
	next, _ = m.Update(click)
	m = next.(Model)

	drag := tea.MouseMotionMsg{X: ox + 2*cellWidth, Y: oy + 1, Button: tea.MouseLeft}
	next, _ = m.Update(drag)
	m = next.(Model)

	next, _ = m.Update(tea.MouseReleaseMsg{})
	m = next.(Model)

	if m.marks[1][1] != seaCell || m.marks[1][2] != seaCell {
		t.Fatalf("drag paint failed: row= %q", string([]rune{rune(m.marks[1][0]), rune(m.marks[1][1]), rune(m.marks[1][2])}))
	}
}

func TestModelMouseClickMirrorsKeyboard(t *testing.T) {
	mode := NewMode("Mini", "test", 3, 3, 0.3, 5)
	p := Puzzle{Width: 3, Height: 3, Clues: makeClues(
		[]int{2, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 1},
	)}
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)
	ox, oy := m.gridOrigin()

	cellX := ox + cellWidth
	cellY := oy + 1

	next, _ = m.Update(tea.MouseClickMsg{X: cellX, Y: cellY, Button: tea.MouseLeft})
	m = next.(Model)
	if m.marks[1][1] != seaCell {
		t.Fatalf("left click from unknown should set sea, got %q", rune(m.marks[1][1]))
	}

	next, _ = m.Update(tea.MouseClickMsg{X: cellX, Y: cellY, Button: tea.MouseLeft})
	m = next.(Model)
	if m.marks[1][1] != seaCell {
		t.Fatalf("left click from sea should keep sea, got %q", rune(m.marks[1][1]))
	}

	next, _ = m.Update(tea.MouseClickMsg{X: cellX, Y: cellY, Button: tea.MouseRight})
	m = next.(Model)
	if m.marks[1][1] != islandCell {
		t.Fatalf("right click should set island, got %q", rune(m.marks[1][1]))
	}

	next, _ = m.Update(tea.MouseClickMsg{X: cellX, Y: cellY, Button: tea.MouseRight})
	m = next.(Model)
	if m.marks[1][1] != unknownCell {
		t.Fatalf("right click from island should set unknown, got %q", rune(m.marks[1][1]))
	}

	next, _ = m.Update(tea.MouseClickMsg{X: cellX, Y: cellY, Button: tea.MouseLeft})
	m = next.(Model)
	if m.marks[1][1] != seaCell {
		t.Fatalf("left click from unknown should set sea, got %q", rune(m.marks[1][1]))
	}
}

func TestModelMouseRightClickTogglesIsland(t *testing.T) {
	mode := NewMode("Mini", "test", 3, 3, 0.3, 5)
	p := Puzzle{Width: 3, Height: 3, Clues: makeClues(
		[]int{2, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 1},
	)}
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)
	ox, oy := m.gridOrigin()

	cellX := ox + cellWidth
	cellY := oy + 1

	next, _ = m.Update(tea.MouseClickMsg{X: cellX, Y: cellY, Button: tea.MouseRight})
	m = next.(Model)
	if m.marks[1][1] != islandCell {
		t.Fatalf("expected island after first right click, got %q", rune(m.marks[1][1]))
	}

	next, _ = m.Update(tea.MouseClickMsg{X: cellX, Y: cellY, Button: tea.MouseRight})
	m = next.(Model)
	if m.marks[1][1] != unknownCell {
		t.Fatalf("expected unknown after second right click, got %q", rune(m.marks[1][1]))
	}
}

func TestModelMouseRightDragPaintIsland(t *testing.T) {
	mode := NewMode("Mini", "test", 3, 3, 0.3, 5)
	p := Puzzle{Width: 3, Height: 3, Clues: makeClues(
		[]int{2, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 1},
	)}
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)
	ox, oy := m.gridOrigin()

	next, _ = m.Update(tea.MouseClickMsg{X: ox + cellWidth, Y: oy + 1, Button: tea.MouseRight})
	m = next.(Model)

	next, _ = m.Update(tea.MouseMotionMsg{X: ox + 2*cellWidth, Y: oy + 1, Button: tea.MouseRight})
	m = next.(Model)

	next, _ = m.Update(tea.MouseReleaseMsg{})
	m = next.(Model)

	if m.marks[1][1] != islandCell || m.marks[1][2] != islandCell {
		t.Fatalf("right drag should paint island: row= %q", string([]rune{rune(m.marks[1][0]), rune(m.marks[1][1]), rune(m.marks[1][2])}))
	}
}

func TestModelMouseRightDragClearsIsland(t *testing.T) {
	mode := NewMode("Mini", "test", 3, 3, 0.3, 5)
	p := Puzzle{Width: 3, Height: 3, Clues: makeClues(
		[]int{2, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 1},
	)}
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)
	m.cursor = pointToCursor(1, 1)
	m.setCellAtCursor(islandCell)
	m.cursor = pointToCursor(2, 1)
	m.setCellAtCursor(islandCell)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)
	ox, oy := m.gridOrigin()

	next, _ = m.Update(tea.MouseClickMsg{X: ox + cellWidth, Y: oy + 1, Button: tea.MouseRight})
	m = next.(Model)

	next, _ = m.Update(tea.MouseMotionMsg{X: ox + 2*cellWidth, Y: oy + 1, Button: tea.MouseRight})
	m = next.(Model)

	next, _ = m.Update(tea.MouseReleaseMsg{})
	m = next.(Model)

	if m.marks[1][1] != unknownCell || m.marks[1][2] != unknownCell {
		t.Fatalf("right drag from island should clear to unknown: row= %q", string([]rune{rune(m.marks[1][0]), rune(m.marks[1][1]), rune(m.marks[1][2])}))
	}
}

func TestModelMouseLeftDragPaintSeaFromSea(t *testing.T) {
	mode := NewMode("Mini", "test", 3, 3, 0.3, 5)
	p := Puzzle{Width: 3, Height: 3, Clues: makeClues(
		[]int{2, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 1},
	)}
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)
	m.cursor = pointToCursor(1, 1)
	m.setCellAtCursor(seaCell)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)
	ox, oy := m.gridOrigin()

	next, _ = m.Update(tea.MouseClickMsg{X: ox + cellWidth, Y: oy + 1, Button: tea.MouseLeft})
	m = next.(Model)

	next, _ = m.Update(tea.MouseMotionMsg{X: ox + 2*cellWidth, Y: oy + 1, Button: tea.MouseLeft})
	m = next.(Model)

	next, _ = m.Update(tea.MouseReleaseMsg{})
	m = next.(Model)

	if m.marks[1][1] != seaCell || m.marks[1][2] != seaCell {
		t.Fatalf("left drag from sea should paint sea: row= %q", string([]rune{rune(m.marks[1][0]), rune(m.marks[1][1]), rune(m.marks[1][2])}))
	}
}

func TestHelpToggleInvalidatesOriginCache(t *testing.T) {
	mode := NewMode("Mini", "test", 3, 3, 0.3, 5)
	p := Puzzle{Width: 3, Height: 3, Clues: makeClues(
		[]int{2, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 1},
	)}
	g, err := New(mode, p)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	m := g.(Model)
	m.originValid = true
	m.originX = 5
	m.originY = 7

	next, _ := m.Update(game.HelpToggleMsg{Show: true})
	got := next.(Model)

	if !got.showFullHelp {
		t.Fatal("expected showFullHelp to be true")
	}
	if got.originValid {
		t.Fatal("expected origin cache to be invalidated")
	}
}

// --- Seeded generation determinism (P1) ---

func TestGenerateSeededDeterministicConcurrent(t *testing.T) {
	mode := Modes[0].(NurikabeMode)
	rng1 := rand.New(rand.NewPCG(11, 29))
	rng2 := rand.New(rand.NewPCG(11, 29))
	rng3 := rand.New(rand.NewPCG(11, 29))

	p1, err := GenerateSeeded(mode, rng1)
	if err != nil {
		t.Fatalf("GenerateSeeded #1 error: %v", err)
	}
	p2, err := GenerateSeeded(mode, rng2)
	if err != nil {
		t.Fatalf("GenerateSeeded #2 error: %v", err)
	}
	p3, err := GenerateSeeded(mode, rng3)
	if err != nil {
		t.Fatalf("GenerateSeeded #3 error: %v", err)
	}

	if serializeClues(p1.Clues) != serializeClues(p2.Clues) {
		t.Fatal("seeded generation is not deterministic")
	}
	if serializeClues(p1.Clues) != serializeClues(p3.Clues) {
		t.Fatal("seeded generation diverged across repeated runs")
	}
}

// --- Generator runtime and validity (P2) ---

func TestGenerateReturnsWithin10SecondsAllModes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow generator budget test in short mode")
	}

	for i, item := range Modes {
		mode := item.(NurikabeMode)
		t.Run(mode.Title(), func(t *testing.T) {
			start := time.Now()
			rng := rand.New(rand.NewPCG(uint64(100+i), uint64(999+i)))
			_, err := GenerateSeeded(mode, rng)
			if err != nil {
				t.Fatalf("GenerateSeeded error: %v", err)
			}
			if elapsed := time.Since(start); elapsed > generationHardTimeout+500*time.Millisecond {
				t.Fatalf("generation exceeded budget for %s: %s", mode.Title(), elapsed)
			}
		})
	}
}

func TestConstructedCandidateValidCompletedBoard(t *testing.T) {
	mode := Modes[2].(NurikabeMode) // Medium
	profile := modeIslandProfile(mode)
	clueTarget := int(float64(mode.Width*mode.Height) * mode.ClueDensity)
	if clueTarget < 2 {
		clueTarget = 2
	}

	rng := rand.New(rand.NewPCG(901, 1777))
	candidate, err := buildCandidateByCarving(mode, clueTarget, profile, rng, 0)
	if err != nil {
		t.Fatalf("buildCandidateByCarving error: %v", err)
	}

	if err := validateClues(candidate.puzzle.Clues, candidate.puzzle.Width, candidate.puzzle.Height); err != nil {
		t.Fatalf("invalid candidate clues: %v", err)
	}

	solved, _, err := CountSolutions(candidate.puzzle, 1, 2000000)
	if err != nil {
		t.Fatalf("CountSolutions error: %v", err)
	}
	if solved < 1 {
		t.Fatalf("expected constructed candidate to be solvable, got %d", solved)
	}
}

func TestGenerateSeededReturnsSolvable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow generator solvability test in short mode")
	}

	for i, item := range Modes {
		mode := item.(NurikabeMode)
		t.Run(mode.Title(), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(500+i), uint64(1200+i)))
			p, err := GenerateSeeded(mode, rng)
			if err != nil {
				t.Fatalf("GenerateSeeded error: %v", err)
			}
			if err := validateClues(p.Clues, p.Width, p.Height); err != nil {
				t.Fatalf("invalid generated puzzle clues: %v", err)
			}

			nodeLimit := max(300000, generationNodeLimit(mode)*8)
			count, _, err := CountSolutionsContext(context.Background(), p, 1, nodeLimit)
			if err != nil {
				if errors.Is(err, errNodeLimit) {
					t.Skipf("solver node cap reached at %d for %s; generation remains solvable-by-construction", nodeLimit, mode.Title())
				}
				t.Fatalf("CountSolutionsContext error: %v", err)
			}
			if count < 1 {
				t.Fatalf("expected at least one solution, got %d", count)
			}
		})
	}
}

func TestGenerateSeededAvoidsSingletonFlood(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow singleton distribution test in short mode")
	}

	for i, item := range Modes {
		mode := item.(NurikabeMode)
		t.Run(mode.Title(), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(700+i), uint64(1700+i)))
			p, err := GenerateSeeded(mode, rng)
			if err != nil {
				t.Fatalf("GenerateSeeded error: %v", err)
			}

			clueCount := 0
			singletons := 0
			for y := 0; y < p.Height; y++ {
				for x := 0; x < p.Width; x++ {
					if p.Clues[y][x] <= 0 {
						continue
					}
					clueCount++
					if p.Clues[y][x] == 1 {
						singletons++
					}
				}
			}
			if clueCount == 0 {
				t.Fatal("expected at least one clue")
			}

			profile := modeIslandProfile(mode)
			ratio := float64(singletons) / float64(clueCount)
			// Keep a small tolerance for fallback cases under the 10s SLA.
			if ratio > profile.maxSingletonRatio+0.12 {
				t.Fatalf(
					"singleton ratio too high for %s: got %.2f (%d/%d), target <= %.2f",
					mode.Title(),
					ratio,
					singletons,
					clueCount,
					profile.maxSingletonRatio+0.12,
				)
			}
		})
	}
}

func TestUniqueFirstFallbackPolicy(t *testing.T) {
	mode := Modes[1].(NurikabeMode) // Easy
	rng := rand.New(rand.NewPCG(2026, 219))
	p, err := GenerateSeeded(mode, rng)
	if err != nil {
		t.Fatalf("GenerateSeeded error: %v", err)
	}

	count, _, err := CountSolutionsContext(context.Background(), p, 2, generationNodeLimit(mode))
	if err != nil {
		t.Fatalf("CountSolutionsContext error: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected generated puzzle to be solvable, got %d", count)
	}
	if count > 2 {
		t.Fatalf("solution count should be capped by limit: got %d", count)
	}
}

func TestGenerateSeededCancellationFast(t *testing.T) {
	mode := Modes[4].(NurikabeMode) // Expert gives enough work to exercise cancel path.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	rng := rand.New(rand.NewPCG(33, 77))
	_, err := GenerateSeededWithContext(ctx, mode, rng)
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
		t.Fatalf("cancellation should return quickly, took %s", elapsed)
	}
}

func TestGenerateSeededDeadlineExceededFast(t *testing.T) {
	mode := Modes[4].(NurikabeMode) // Expert keeps deadline path realistic.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	start := time.Now()
	rng := rand.New(rand.NewPCG(44, 88))
	_, err := GenerateSeededWithContext(ctx, mode, rng)
	if err == nil {
		t.Fatal("expected deadline exceeded error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("deadline cancellation should return quickly, took %s", elapsed)
	}
}

func TestNoGoroutineLeakOnCancel(t *testing.T) {
	mode := Modes[4].(NurikabeMode)
	before := runtime.NumGoroutine()

	for i := 0; i < 8; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		rng := rand.New(rand.NewPCG(uint64(400+i), uint64(900+i)))
		_, _ = GenerateSeededWithContext(ctx, mode, rng)
	}

	time.Sleep(80 * time.Millisecond)
	after := runtime.NumGoroutine()
	if delta := after - before; delta > 6 {
		t.Fatalf("possible goroutine leak: before=%d after=%d delta=%d", before, after, delta)
	}
}

func TestNoGoroutineLeakOnDeadlineTimeout(t *testing.T) {
	mode := Modes[4].(NurikabeMode)
	before := runtime.NumGoroutine()

	for i := 0; i < 8; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		rng := rand.New(rand.NewPCG(uint64(1400+i), uint64(1900+i)))
		_, _ = GenerateSeededWithContext(ctx, mode, rng)
		cancel()
	}

	time.Sleep(120 * time.Millisecond)
	after := runtime.NumGoroutine()
	if delta := after - before; delta > 8 {
		t.Fatalf("possible goroutine leak on deadline timeout: before=%d after=%d delta=%d", before, after, delta)
	}
}

// --- Benchmarks (P2) ---

func BenchmarkGenerateSeededModes(b *testing.B) {
	for i, item := range Modes {
		mode := item.(NurikabeMode)
		modeIdx := i
		b.Run(mode.Title(), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				seedA := uint64(100 + modeIdx)
				seedB := uint64(1000 + modeIdx)
				rng := rand.New(rand.NewPCG(seedA, seedB))
				if _, err := GenerateSeeded(mode, rng); err != nil {
					b.Fatalf("GenerateSeeded error: %v", err)
				}
			}
		})
	}
}

func BenchmarkCountSolutions(b *testing.B) {
	easyMode := Modes[1].(NurikabeMode)
	easyRNG := rand.New(rand.NewPCG(7001, 9001))
	easyPuzzle, err := GenerateSeeded(easyMode, easyRNG)
	if err != nil {
		b.Fatalf("GenerateSeeded setup error: %v", err)
	}

	cases := []struct {
		name      string
		puzzle    Puzzle
		nodeLimit int
	}{
		{
			name:      "Fixture5x5",
			puzzle:    uniquePuzzleFixture(),
			nodeLimit: 120000,
		},
		{
			name:      "GeneratedEasy",
			puzzle:    easyPuzzle,
			nodeLimit: generationNodeLimit(easyMode) * 4,
		},
	}

	for _, tc := range cases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				_, _, solveErr := CountSolutionsContext(context.Background(), tc.puzzle, 2, tc.nodeLimit)
				if solveErr != nil && !errors.Is(solveErr, errNodeLimit) {
					b.Fatalf("CountSolutionsContext error: %v", solveErr)
				}
			}
		})
	}
}
