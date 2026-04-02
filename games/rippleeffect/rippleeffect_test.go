package rippleeffect

import (
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/charmbracelet/x/ansi"
)

func sampleCages() []Cage {
	return []Cage{
		{ID: 0, Size: 3, Cells: []Cell{{0, 0}, {1, 0}, {2, 0}}},
		{ID: 1, Size: 3, Cells: []Cell{{0, 1}, {1, 1}, {2, 1}}},
		{ID: 2, Size: 3, Cells: []Cell{{0, 2}, {1, 2}, {2, 2}}},
	}
}

func sampleSolution() grid {
	return grid{
		{1, 2, 3},
		{2, 3, 1},
		{3, 1, 2},
	}
}

func TestBuildGeometryBoundaries(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	if !geo.boundaries[0][0].has(boundaryTop) || !geo.boundaries[0][0].has(boundaryLeft) {
		t.Fatal("expected top-left cell to have outer top/left boundaries")
	}
	if geo.boundaries[0][0].has(boundaryRight) {
		t.Fatal("expected shared cage edge on the right to be open")
	}
	if !geo.boundaries[1][1].has(boundaryTop) {
		t.Fatal("expected cage boundary between first and second rows")
	}
	if !geo.boundaries[2][2].has(boundaryBottom) || !geo.boundaries[2][2].has(boundaryRight) {
		t.Fatal("expected bottom-right outer boundaries")
	}
}

func TestValidateGridState(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		grid   grid
		solved bool
	}{
		{
			name:   "valid solved board",
			grid:   sampleSolution(),
			solved: true,
		},
		{
			name: "duplicate within cage",
			grid: grid{
				{1, 1, 3},
				{2, 3, 1},
				{3, 1, 2},
			},
			solved: false,
		},
		{
			name: "ripple conflict",
			grid: grid{
				{2, 1, 3},
				{2, 3, 1},
				{3, 1, 2},
			},
			solved: false,
		},
		{
			name: "incomplete board",
			grid: grid{
				{1, 2, 3},
				{2, 0, 1},
				{3, 1, 2},
			},
			solved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateGridState(tt.grid, geo)
			if result.solved != tt.solved {
				t.Fatalf("solved = %v, want %v", result.solved, tt.solved)
			}
		})
	}
}

func TestCountSolutions(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		givens grid
		want   int
	}{
		{
			name: "unique solution",
			givens: grid{
				{1, 2, 3},
				{2, 3, 1},
				{3, 1, 2},
			},
			want: 1,
		},
		{
			name: "ambiguous puzzle",
			givens: grid{
				{0, 0, 0},
				{0, 0, 0},
				{0, 0, 0},
			},
			want: 2,
		},
		{
			name: "impossible puzzle",
			givens: grid{
				{2, 2, 3},
				{0, 0, 0},
				{0, 0, 0},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countSolutions(geo, tt.givens, 2)
			if got != tt.want {
				t.Fatalf("countSolutions() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGeneratePuzzleSeeded(t *testing.T) {
	mode := Modes[0].(Mode)
	rngA := rand.New(rand.NewPCG(11, 22))
	rngB := rand.New(rand.NewPCG(11, 22))

	a, err := GeneratePuzzleSeeded(mode.Size, mode.Size, mode.MaxCage, mode.GivenRatio, rngA)
	if err != nil {
		t.Fatal(err)
	}
	b, err := GeneratePuzzleSeeded(mode.Size, mode.Size, mode.MaxCage, mode.GivenRatio, rngB)
	if err != nil {
		t.Fatal(err)
	}

	if encodeGrid(a.Givens) != encodeGrid(b.Givens) {
		t.Fatalf("seeded givens differ:\n%s\n!=\n%s", encodeGrid(a.Givens), encodeGrid(b.Givens))
	}
	if len(a.Cages) != len(b.Cages) {
		t.Fatalf("seeded cage count mismatch: %d != %d", len(a.Cages), len(b.Cages))
	}
	geo, err := buildGeometry(a.Width, a.Height, a.Cages)
	if err != nil {
		t.Fatal(err)
	}
	if !validateGridState(a.Solution, geo).solved {
		t.Fatal("generated solution is not solved")
	}
	if got := countSolutions(geo, a.Givens, 2); got != 1 {
		t.Fatalf("generated puzzle solutions = %d, want 1", got)
	}
}

func TestMinGivensForCageAllowsZero(t *testing.T) {
	profile := generationProfile{minGivensByCage: []int{0, 0, 0, 0}}

	if got := minGivensForCage(profile, 1); got != 0 {
		t.Fatalf("minGivensForCage(..., 1) = %d, want 0", got)
	}
	if got := minGivensForCage(profile, 3); got != 0 {
		t.Fatalf("minGivensForCage(..., 3) = %d, want 0", got)
	}
}

func TestNormalizeSingletonCagesCapsSingletons(t *testing.T) {
	rng := rand.New(rand.NewPCG(77, 99))
	cages := []Cage{
		{ID: 0, Size: 1, Cells: []Cell{{0, 0}}},
		{ID: 1, Size: 1, Cells: []Cell{{1, 0}}},
		{ID: 2, Size: 2, Cells: []Cell{{2, 0}, {2, 1}}},
		{ID: 3, Size: 2, Cells: []Cell{{0, 1}, {0, 2}}},
		{ID: 4, Size: 3, Cells: []Cell{{1, 1}, {1, 2}, {2, 2}}},
	}

	normalized := normalizeSingletonCages(3, 3, 3, cages, generationProfile{maxSingletonCages: 1}, rng)
	if got := countSingletonCages(normalized); got > 1 {
		t.Fatalf("singleton cages = %d, want <= 1", got)
	}

	geo, err := buildGeometry(3, 3, normalized)
	if err != nil {
		t.Fatalf("buildGeometry() after normalization: %v", err)
	}
	if got := len(geo.cages); got != len(normalized) {
		t.Fatalf("geometry cages = %d, want %d", got, len(normalized))
	}
}

func TestSparseModesGenerateUniquePuzzlesWithLimitedSingletons(t *testing.T) {
	modeIndices := []int{2, 3, 4}
	sampleCount := 6
	sawZeroGivenCage := false

	for _, idx := range modeIndices {
		mode := Modes[idx].(Mode)

		for sample := range sampleCount {
			rng := rand.New(rand.NewPCG(uint64(7000+idx*31+sample), uint64(8000+idx*37+sample)))
			puzzle, err := mode.generatePuzzleSeeded(rng)
			if err != nil {
				t.Fatalf("%s generatePuzzleSeeded failed: %v", mode.Title(), err)
			}

			geo, err := buildGeometry(puzzle.Width, puzzle.Height, puzzle.Cages)
			if err != nil {
				t.Fatalf("%s buildGeometry failed: %v", mode.Title(), err)
			}
			if got := countSolutions(geo, puzzle.Givens, 2); got != 1 {
				t.Fatalf("%s solutions = %d, want 1", mode.Title(), got)
			}

			if got := countSingletonCages(puzzle.Cages); got > 1 {
				t.Fatalf("%s singleton cages = %d, want <= 1", mode.Title(), got)
			}

			if countZeroGivenCages(puzzle) > 0 {
				sawZeroGivenCage = true
			}
		}
	}

	if !sawZeroGivenCage {
		t.Fatal("expected sampled medium+ puzzles to include at least one zero-given cage")
	}
}

func TestSampledGivenValueDistributionMatchesSolutions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sampled ripple given-value distribution test in short mode")
	}

	sampleCounts := []int{8, 6, 4, 2, 1}
	shareDriftTolerances := []float64{0.12, 0.10, 0.10, 0.12, 0.16}

	for i, item := range Modes {
		mode := item.(Mode)
		samples := sampleCounts[i]
		tolerance := shareDriftTolerances[i]

		t.Run(fmt.Sprintf("%s-%d", mode.Title(), samples), func(t *testing.T) {
			solutionCounts := make([]int, mode.MaxCage+1)
			givenCounts := make([]int, mode.MaxCage+1)
			totalSolutions := 0
			totalGivens := 0

			for sample := range samples {
				rng := rand.New(rand.NewPCG(uint64(3000+i*37+sample), uint64(4000+i*41+sample)))
				puzzle, err := mode.generatePuzzleSeeded(rng)
				if err != nil {
					t.Fatalf("generatePuzzleSeeded failed: %v", err)
				}

				totalSolutions += accumulateValueCounts(solutionCounts, puzzle.Solution)
				totalGivens += accumulateValueCounts(givenCounts, puzzle.Givens)
			}

			if totalSolutions == 0 {
				t.Fatal("expected sampled solutions to contain values")
			}
			if totalGivens == 0 {
				t.Fatal("expected sampled puzzles to contain givens")
			}

			t.Logf("sampled solution values: %s", formatValueDistribution(solutionCounts))
			t.Logf("sampled given values: %s", formatValueDistribution(givenCounts))

			for value := 1; value < len(solutionCounts); value++ {
				if solutionCounts[value] == 0 {
					continue
				}

				solutionShare := float64(solutionCounts[value]) / float64(totalSolutions)
				givenShare := float64(givenCounts[value]) / float64(totalGivens)
				if diff := givenShare - solutionShare; diff < -tolerance || diff > tolerance {
					t.Fatalf(
						"value %d share %.2f outside %.2f +/- %.2f",
						value,
						givenShare,
						solutionShare,
						tolerance,
					)
				}
			}
		})
	}
}

func TestSaveRoundTrip(t *testing.T) {
	puzzle := Puzzle{
		Width:    3,
		Height:   3,
		Cages:    sampleCages(),
		Givens:   grid{{1, 0, 3}, {0, 3, 0}, {3, 0, 2}},
		Solution: sampleSolution(),
	}
	g, err := New(NewMode("Test", "desc", 3, 3, 1), puzzle)
	if err != nil {
		t.Fatal(err)
	}
	model := g.(Model)
	model.cursor = pointCursor(1, 0)
	model.setCell(2)

	save, err := model.GetSave()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := ImportModel(save)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := loaded.grid[0][1], 2; got != want {
		t.Fatalf("loaded grid[0][1] = %d, want %d", got, want)
	}
	if got, want := loaded.geo.cageGrid[2][2], 2; got != want {
		t.Fatalf("loaded cage grid[2][2] = %d, want %d", got, want)
	}
}

func TestChooseVisualKindPriority(t *testing.T) {
	if got := chooseVisualKind(true, false, true, true, true); got != visualConflictCursor {
		t.Fatalf("conflicted cursor should win, got %v", got)
	}
	if got := chooseVisualKind(true, true, false, true, true); got != visualSolved {
		t.Fatalf("solved should beat cursor, got %v", got)
	}
	if got := chooseVisualKind(true, false, false, true, true); got != visualCursor {
		t.Fatalf("cursor should beat cage highlight, got %v", got)
	}
	if got := chooseVisualKind(false, false, false, true, true); got != visualCompleted {
		t.Fatalf("completed should beat transient cage highlight, got %v", got)
	}
}

func TestCursorText(t *testing.T) {
	if got, want := cursorText(0), game.CursorLeft+"·"+game.CursorRight; got != want {
		t.Fatalf("cursorText(0) = %q, want %q", got, want)
	}
	if got, want := cursorText(3), game.CursorLeft+"3"+game.CursorRight; got != want {
		t.Fatalf("cursorText(3) = %q, want %q", got, want)
	}
}

func TestCellViewConflictedCursorIsDistinct(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	m := Model{
		width:     3,
		height:    3,
		grid:      sampleSolution(),
		givens:    grid{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		conflicts: newConflictGrid(3, 3),
		cursor:    pointCursor(1, 1),
		geo:       geo,
	}
	m.conflicts[1][1] = true

	conflict := cellView(m, 0, 0, 0, nil)
	cursorConflict := cellView(m, 1, 1, 0, nil)

	if conflict == cursorConflict {
		t.Fatal("conflicted cursor should render differently from plain conflicted cell")
	}
	if !strings.Contains(cursorConflict, game.CursorLeft+"3"+game.CursorRight) {
		t.Fatal("conflicted cursor should include cursor markers")
	}
}

func TestCellViewUsesGivenTintForImmutableGivens(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	m := Model{
		width:     3,
		height:    3,
		grid:      sampleSolution(),
		givens:    grid{{1, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		conflicts: newConflictGrid(3, 3),
		cursor:    pointCursor(2, 2),
		geo:       geo,
	}

	got := cellView(m, 0, 0, -1, nil)
	want := lipgloss.NewStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Bold(true).
		Foreground(theme.Current().FG).
		Background(theme.GivenTint(theme.Current().BG)).
		Render(" 1 ")
	if got != want {
		t.Fatalf("given cellView() = %q, want %q", got, want)
	}
}

func TestCellViewDoesNotHighlightCrosshairAxis(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	m := Model{
		width:     3,
		height:    3,
		grid:      newGrid(3, 3),
		givens:    newGrid(3, 3),
		conflicts: newConflictGrid(3, 3),
		cursor:    pointCursor(0, 0),
		geo:       geo,
	}

	got := cellView(m, 0, 1, geo.cageGrid[m.cursor.Y][m.cursor.X], nil)
	want := lipgloss.NewStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Foreground(theme.Current().FG).
		Background(theme.Current().BG).
		Render(" · ")
	if got != want {
		t.Fatalf("crosshair-axis cellView() = %q, want %q", got, want)
	}
}

func TestGridViewUsesUniformRowsForBronzeBirchState(t *testing.T) {
	model, err := ImportModel([]byte(`{"width":5,"height":5,"state":". 2 1 . 3\n2 1 1 1 .\n1 1 3 . .\n1 . 1 3 2\n2 3 2 1 1","givens":". 2 1 . 3\n2 1 1 1 .\n1 1 3 . .\n1 . 1 3 2\n2 3 2 1 1","cages":[{"id":0,"size":3,"cells":[{"x":0,"y":0},{"x":1,"y":0},{"x":1,"y":1}]},{"id":1,"size":2,"cells":[{"x":2,"y":0},{"x":3,"y":0}]},{"id":2,"size":3,"cells":[{"x":4,"y":0},{"x":4,"y":1},{"x":4,"y":2}]},{"id":3,"size":2,"cells":[{"x":0,"y":1},{"x":0,"y":2}]},{"id":4,"size":1,"cells":[{"x":2,"y":1}]},{"id":5,"size":3,"cells":[{"x":3,"y":1},{"x":3,"y":2},{"x":3,"y":3}]},{"id":6,"size":3,"cells":[{"x":1,"y":2},{"x":1,"y":3},{"x":2,"y":2}]},{"id":7,"size":2,"cells":[{"x":0,"y":3},{"x":0,"y":4}]},{"id":8,"size":3,"cells":[{"x":2,"y":3},{"x":2,"y":4},{"x":1,"y":4}]},{"id":9,"size":2,"cells":[{"x":4,"y":3},{"x":4,"y":4}]},{"id":10,"size":1,"cells":[{"x":3,"y":4}]}],"mode_title":"bronze-birch"}`))
	if err != nil {
		t.Fatal(err)
	}

	view := gridView(*model)
	lines := strings.Split(view, "\n")
	if got, want := len(lines), model.height*2+1; got != want {
		t.Fatalf("grid line count = %d, want %d", got, want)
	}

	wantWidth := lipgloss.Width(lines[0])
	for i, line := range lines {
		if width := lipgloss.Width(line); width != wantWidth {
			t.Fatalf("line %d width = %d, want %d", i, width, wantWidth)
		}
	}
}

func TestCompletedCageBackgrounds(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	m := Model{
		width:     3,
		height:    3,
		geo:       geo,
		grid:      sampleSolution(),
		givens:    newGrid(3, 3),
		conflicts: newConflictGrid(3, 3),
	}

	completed := completedCageBackgrounds(m)
	if len(completed) != len(sampleCages()) {
		t.Fatalf("completed cages = %d, want %d", len(completed), len(sampleCages()))
	}

	m.conflicts[0][0] = true
	completed = completedCageBackgrounds(m)
	if _, ok := completed[0]; ok {
		t.Fatal("conflicted cage should not get completed background")
	}
}

func TestSolvedBridgeBackgroundsUseSolvedColor(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	m := Model{
		width:  3,
		height: 3,
		geo:    geo,
		solved: true,
	}

	backgrounds := solvedBridgeBackgrounds(m)
	if len(backgrounds) != len(sampleCages()) {
		t.Fatalf("solved bridge backgrounds = %d, want %d", len(backgrounds), len(sampleCages()))
	}

	want := theme.Current().SuccessBG
	for cageIdx, bg := range backgrounds {
		if !sameColor(bg, want) {
			t.Fatalf("cage %d bridge background does not use solved color", cageIdx)
		}
	}
}

func TestGridViewUsesCageBoundaries(t *testing.T) {
	geo, err := buildGeometry(3, 3, sampleCages())
	if err != nil {
		t.Fatal(err)
	}

	m := Model{
		width:     3,
		height:    3,
		grid:      newGrid(3, 3),
		givens:    newGrid(3, 3),
		conflicts: newConflictGrid(3, 3),
		cursor:    pointCursor(0, 0),
		geo:       geo,
	}

	lines := strings.Split(ansi.Strip(gridView(m)), "\n")
	content := []rune(lines[1])
	if got := content[4]; got != ' ' {
		t.Fatalf("interior separator inside cage = %q, want space", got)
	}
	boundary := []rune(lines[2])
	if got := boundary[1]; got != '─' {
		t.Fatalf("separator between cages = %q, want horizontal wall", got)
	}
}

func TestBridgeFillDefersToUniformZoneBackground(t *testing.T) {
	m := Model{}
	bridgeBG := map[int]color.Color{7: theme.Current().SelectionBG}

	if got := bridgeFill(m, bridgeBG, game.DynamicGridBridge{
		Kind:    game.DynamicGridBridgeVertical,
		Zone:    7,
		Uniform: true,
		Count:   2,
		Cells: [4]image.Point{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
		},
	}); got != nil {
		t.Fatal("expected uniform bridge to defer to zone background")
	}
}

func TestBridgeFillLeavesOpenAnchoredBridgeUnfilled(t *testing.T) {
	geo, err := buildGeometry(3, 2, []Cage{
		{ID: 0, Size: 2, Cells: []Cell{{0, 0}, {0, 1}}},
		{ID: 1, Size: 2, Cells: []Cell{{1, 0}, {2, 0}}},
		{ID: 2, Size: 1, Cells: []Cell{{1, 1}}},
		{ID: 3, Size: 1, Cells: []Cell{{2, 1}}},
	})
	if err != nil {
		t.Fatal(err)
	}

	m := Model{
		width:     3,
		height:    2,
		grid:      newGrid(3, 2),
		givens:    newGrid(3, 2),
		conflicts: newConflictGrid(3, 2),
		cursor:    pointCursor(0, 0),
		geo:       geo,
	}
	bridgeBG := bridgeBackgrounds(m, nil)

	got := bridgeFill(m, bridgeBG, game.DynamicGridBridge{
		Kind:    game.DynamicGridBridgeVertical,
		X:       2,
		Y:       0,
		Count:   2,
		Zone:    geo.cageGrid[0][1],
		Uniform: true,
		Cells: [4]image.Point{
			{X: 1, Y: 0},
			{X: 2, Y: 0},
		},
	})
	if got != nil {
		t.Fatal("expected anchored open bridge to remain unfilled")
	}
}

func TestBridgeFillDoesNotExpandIntoOpenInterior(t *testing.T) {
	geo, err := buildGeometry(4, 3, []Cage{
		{ID: 0, Size: 1, Cells: []Cell{{0, 1}}},
		{ID: 1, Size: 9, Cells: []Cell{{1, 0}, {2, 0}, {3, 0}, {1, 1}, {2, 1}, {3, 1}, {1, 2}, {2, 2}, {3, 2}}},
		{ID: 2, Size: 1, Cells: []Cell{{0, 0}}},
		{ID: 3, Size: 1, Cells: []Cell{{0, 2}}},
	})
	if err != nil {
		t.Fatal(err)
	}

	m := Model{
		width:     4,
		height:    3,
		grid:      newGrid(4, 3),
		givens:    newGrid(4, 3),
		conflicts: newConflictGrid(4, 3),
		cursor:    pointCursor(0, 1),
		geo:       geo,
	}
	bridgeBG := bridgeBackgrounds(m, nil)

	if got := bridgeFill(m, bridgeBG, game.DynamicGridBridge{
		Kind:    game.DynamicGridBridgeVertical,
		X:       2,
		Y:       1,
		Count:   2,
		Zone:    geo.cageGrid[1][1],
		Uniform: true,
		Cells: [4]image.Point{
			{X: 1, Y: 1},
			{X: 2, Y: 1},
		},
	}); got != nil {
		t.Fatal("expected fully open interior bridge to remain unfilled")
	}
}

func TestBridgeFillVerticalSeparatorStaysUnfilled(t *testing.T) {
	m := Model{cursor: pointCursor(0, 1)}

	if got := bridgeFill(m, nil, game.DynamicGridBridge{
		Kind: game.DynamicGridBridgeVertical,
		X:    1,
		Y:    0,
	}); got != nil {
		t.Fatal("expected vertical separator to remain unfilled")
	}
}

func TestBridgeFillHorizontalSeparatorStaysUnfilled(t *testing.T) {
	m := Model{cursor: pointCursor(1, 0)}

	if got := bridgeFill(m, nil, game.DynamicGridBridge{
		Kind: game.DynamicGridBridgeHorizontal,
		X:    0,
		Y:    1,
	}); got != nil {
		t.Fatal("expected horizontal separator to remain unfilled")
	}
}

func TestBridgeFillClosedVerticalSeparatorStaysUnfilled(t *testing.T) {
	m := Model{cursor: pointCursor(0, 0)}

	got := bridgeFill(m, nil, game.DynamicGridBridge{
		Kind: game.DynamicGridBridgeVertical,
		X:    1,
		Y:    0,
	})
	if got != nil {
		t.Fatal("expected closed vertical separator to remain unfilled")
	}
}

func TestBridgeFillClosedHorizontalSeparatorStaysUnfilled(t *testing.T) {
	m := Model{cursor: pointCursor(0, 0)}

	got := bridgeFill(m, nil, game.DynamicGridBridge{
		Kind: game.DynamicGridBridgeHorizontal,
		X:    0,
		Y:    1,
	})
	if got != nil {
		t.Fatal("expected closed horizontal separator to remain unfilled")
	}
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func TestImportRejectsInvalidGivens(t *testing.T) {
	_, err := ImportModel([]byte(`{"width":3,"height":3,"state":"1 2 3\n2 3 1\n3 1 2","givens":"1 2 3\n2 4 1\n3 1 2","cages":[{"id":0,"size":3,"cells":[{"x":0,"y":0},{"x":1,"y":0},{"x":2,"y":0}]},{"id":1,"size":3,"cells":[{"x":0,"y":1},{"x":1,"y":1},{"x":2,"y":1}]},{"id":2,"size":3,"cells":[{"x":0,"y":2},{"x":1,"y":2},{"x":2,"y":2}]}],"mode_title":"Test"}`))
	if err == nil {
		t.Fatal("expected import error")
	}
}

func pointCursor(x, y int) game.Cursor {
	return game.Cursor{X: x, Y: y}
}

func accumulateValueCounts(counts []int, state grid) int {
	total := 0
	for y := range len(state) {
		for x := range len(state[y]) {
			value := state[y][x]
			if value <= 0 {
				continue
			}
			counts[value]++
			total++
		}
	}
	return total
}

func formatValueDistribution(counts []int) string {
	total := 0
	for value := 1; value < len(counts); value++ {
		total += counts[value]
	}
	if total == 0 {
		return "none"
	}

	parts := make([]string, 0, len(counts)-1)
	for value := 1; value < len(counts); value++ {
		if counts[value] == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d=%d(%.1f%%)", value, counts[value], 100*float64(counts[value])/float64(total)))
	}
	return strings.Join(parts, ", ")
}

func countZeroGivenCages(puzzle Puzzle) int {
	count := 0
	for _, cage := range puzzle.Cages {
		givens := 0
		for _, cell := range cage.Cells {
			if puzzle.Givens[cell.Y][cell.X] != 0 {
				givens++
			}
		}
		if givens == 0 {
			count++
		}
	}
	return count
}
