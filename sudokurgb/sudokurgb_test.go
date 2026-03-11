package sudokurgb

import (
	"math/rand/v2"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/x/ansi"
)

func validCompleteGrid() grid {
	rows := [gridSize][gridSize]int{
		{1, 2, 3, 1, 2, 3, 1, 2, 3},
		{2, 3, 1, 2, 3, 1, 2, 3, 1},
		{3, 1, 2, 3, 1, 2, 3, 1, 2},
		{1, 2, 3, 1, 2, 3, 1, 2, 3},
		{2, 3, 1, 2, 3, 1, 2, 3, 1},
		{3, 1, 2, 3, 1, 2, 3, 1, 2},
		{1, 2, 3, 1, 2, 3, 1, 2, 3},
		{2, 3, 1, 2, 3, 1, 2, 3, 1},
		{3, 1, 2, 3, 1, 2, 3, 1, 2},
	}

	var g grid
	for y := range gridSize {
		for x := range gridSize {
			g[y][x] = cell{x: x, y: y, v: rows[y][x]}
		}
	}
	return g
}

func TestFillGridSeededProducesValidBoard(t *testing.T) {
	g := newGrid(nil)
	if !fillGridSeeded(&g, rand.New(rand.NewPCG(7, 11))) {
		t.Fatal("fillGridSeeded returned false")
	}
	if analysis := analyzeGrid(g); !isSolvedWith(g, analysis) {
		t.Fatal("generated board is not solved")
	}
}

func TestCountSolutionsRejectsOverQuotaState(t *testing.T) {
	g := validCompleteGrid()
	g[0][1].v = 1
	if got := countSolutions(&g, 2); got != 0 {
		t.Fatalf("countSolutions(overquota) = %d, want 0", got)
	}
}

func TestGenerateProvidedCellsSeededMatchesModeTargets(t *testing.T) {
	for i, item := range Modes {
		mode := item.(SudokuRGBMode)
		t.Run(mode.Title(), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(i+1), uint64(i+101)))
			cells := GenerateProvidedCellsSeeded(mode, rng)
			if got, want := len(cells), mode.ProvidedCount; got != want {
				t.Fatalf("provided count = %d, want %d", got, want)
			}

			g := newGrid(cells)
			if got := countSolutions(&g, 2); got != 1 {
				t.Fatalf("countSolutions() = %d, want 1", got)
			}
		})
	}
}

func TestModelUpdateRestrictsInputsAndRespectsProvidedCells(t *testing.T) {
	model := Model{
		grid:         newGrid([]cell{{x: 0, y: 0, v: 1}}),
		provided:     []cell{{x: 0, y: 0, v: 1}},
		providedGrid: buildProvidedGrid([]cell{{x: 0, y: 0, v: 1}}),
		analysis:     analyzeGrid(newGrid([]cell{{x: 0, y: 0, v: 1}})),
		keys:         DefaultKeyMap,
		cursor:       game.Cursor{X: 1, Y: 1},
	}

	next, _ := model.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	got := next.(Model)
	if got.grid[1][1].v != 1 {
		t.Fatalf("value after 1 = %d, want 1", got.grid[1][1].v)
	}

	next, _ = got.Update(tea.KeyPressMsg{Code: '4', Text: "4"})
	got = next.(Model)
	if got.grid[1][1].v != 1 {
		t.Fatalf("value after ignored 4 = %d, want 1", got.grid[1][1].v)
	}

	next, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	got = next.(Model)
	if got.grid[1][1].v != 0 {
		t.Fatalf("value after backspace = %d, want 0", got.grid[1][1].v)
	}

	got.cursor.X, got.cursor.Y = 0, 0
	next, _ = got.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	got = next.(Model)
	if got.grid[0][0].v != 1 {
		t.Fatalf("provided cell mutated to %d, want 1", got.grid[0][0].v)
	}
}

func TestCellViewUsesUnicodeSymbolsWithCursorMarkers(t *testing.T) {
	g := newGrid(nil)
	g[0][0].v = 3
	m := Model{
		grid:     g,
		keys:     DefaultKeyMap,
		cursor:   game.Cursor{X: 0, Y: 0},
		analysis: analyzeGrid(g),
	}

	got := ansi.Strip(cellView(m, 0, 0, false))
	if !strings.Contains(got, game.CursorLeft+"●"+game.CursorRight) {
		t.Fatalf("filled cursor cell = %q, want marker-wrapped unicode symbol", got)
	}

	g[0][0].v = 0
	m.grid = g
	got = ansi.Strip(cellView(m, 0, 0, false))
	if !strings.Contains(got, game.CursorLeft+"·"+game.CursorRight) {
		t.Fatalf("empty cursor cell = %q, want marker-wrapped dot", got)
	}
}

func TestScreenToGridUsesStandardCellWidth(t *testing.T) {
	m := Model{
		grid:     newGrid(nil),
		analysis: analyzeGrid(newGrid(nil)),
		keys:     DefaultKeyMap,
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	got := next.(Model)
	x, y := got.cachedGridOrigin()

	if col, row, ok := got.screenToGrid(x, y); !ok || col != 0 || row != 0 {
		t.Fatalf("screenToGrid(origin) = (%d,%d,%v), want (0,0,true)", col, row, ok)
	}
	if _, _, ok := got.screenToGrid(x+cellWidth, y); ok {
		t.Fatal("screenToGrid(separator) = ok, want false")
	}
	if _, _, ok := got.screenToGrid(x-rowHintWidth, y); ok {
		t.Fatal("screenToGrid(row hint band) = ok, want false")
	}
	if _, _, ok := got.screenToGrid(x, y-colHintHeight); ok {
		t.Fatal("screenToGrid(column hint band) = ok, want false")
	}
}

func TestSaveRoundTrip(t *testing.T) {
	g := newGrid([]cell{{x: 0, y: 0, v: 1}})
	g[1][1].v = 3

	saveModel := Model{
		grid:         g,
		provided:     []cell{{x: 0, y: 0, v: 1}},
		providedGrid: buildProvidedGrid([]cell{{x: 0, y: 0, v: 1}}),
		analysis:     analyzeGrid(g),
		keys:         DefaultKeyMap,
		modeTitle:    "Easy",
	}

	save, err := saveModel.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := ImportModel(save)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.grid[0][0].v != 1 || loaded.grid[1][1].v != 3 {
		t.Fatalf("loaded grid values = (%d,%d), want (1,3)", loaded.grid[0][0].v, loaded.grid[1][1].v)
	}
	if loaded.modeTitle != "Easy" {
		t.Fatalf("mode title = %q, want Easy", loaded.modeTitle)
	}
}

func TestAnalyzeGridTracksRowColumnAndBoxState(t *testing.T) {
	t.Run("counts solved board", func(t *testing.T) {
		analysis := analyzeGrid(validCompleteGrid())
		for row := range gridSize {
			if got := analysis.rowCountsString(row); got != "3/3/3" {
				t.Fatalf("rowCountsString(%d) = %q, want 3/3/3", row, got)
			}
		}
		for col := range gridSize {
			if got := analysis.colCountsString(col); got != "3/3/3" {
				t.Fatalf("colCountsString(%d) = %q, want 3/3/3", col, got)
			}
		}
		if analysis.hasConflicts() {
			t.Fatal("solved board unexpectedly has conflicts")
		}
	})

	t.Run("row over-quota stays off cells", func(t *testing.T) {
		g := newGrid(nil)
		for x := 0; x < 4; x++ {
			g[0][x].v = 1
		}

		analysis := analyzeGrid(g)
		if !analysis.rowOverQuota[0][1] {
			t.Fatal("expected row over-quota for red")
		}
		if analysis.colOverQuota[0][1] {
			t.Fatal("unexpected column over-quota")
		}
		for y := range gridSize {
			for x := range gridSize {
				if analysis.boxConflictCells[y][x] {
					t.Fatalf("unexpected box conflict at (%d,%d)", x, y)
				}
			}
		}
	})

	t.Run("column over-quota stays off cells", func(t *testing.T) {
		g := newGrid(nil)
		for _, y := range []int{0, 3, 6, 8} {
			g[y][0].v = 2
		}

		analysis := analyzeGrid(g)
		if !analysis.colOverQuota[0][2] {
			t.Fatal("expected column over-quota for green")
		}
		if analysis.rowOverQuota[0][2] {
			t.Fatal("unexpected row over-quota")
		}
		for y := range gridSize {
			for x := range gridSize {
				if analysis.boxConflictCells[y][x] {
					t.Fatalf("unexpected box conflict at (%d,%d)", x, y)
				}
			}
		}
	})

	t.Run("box over-quota marks only box cells", func(t *testing.T) {
		g := newGrid(nil)
		for _, cell := range []cell{
			{x: 0, y: 0, v: 3},
			{x: 1, y: 0, v: 3},
			{x: 2, y: 1, v: 3},
			{x: 0, y: 2, v: 3},
		} {
			g[cell.y][cell.x] = cell
		}

		analysis := analyzeGrid(g)
		if analysis.rowOverQuota[0][3] || analysis.colOverQuota[0][3] {
			t.Fatal("box-only conflict leaked into row or column state")
		}
		for _, cell := range []cell{
			{x: 0, y: 0},
			{x: 1, y: 0},
			{x: 2, y: 1},
			{x: 0, y: 2},
		} {
			if !analysis.boxConflictCells[cell.y][cell.x] {
				t.Fatalf("expected box conflict at (%d,%d)", cell.x, cell.y)
			}
		}
		if analysis.boxConflictCells[3][3] {
			t.Fatal("unexpected box conflict outside offending box")
		}
	})
}
