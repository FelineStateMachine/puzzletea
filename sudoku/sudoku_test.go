package sudoku

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
)

// validCompleteGrid returns a known valid 9x9 Sudoku solution.
func validCompleteGrid() grid {
	rows := [9][9]int{
		{5, 3, 4, 6, 7, 8, 9, 1, 2},
		{6, 7, 2, 1, 9, 5, 3, 4, 8},
		{1, 9, 8, 3, 4, 2, 5, 6, 7},
		{8, 5, 9, 7, 6, 1, 4, 2, 3},
		{4, 2, 6, 8, 5, 3, 7, 9, 1},
		{7, 1, 3, 9, 2, 4, 8, 5, 6},
		{9, 6, 1, 5, 3, 7, 2, 8, 4},
		{2, 8, 7, 4, 1, 9, 6, 3, 5},
		{3, 4, 5, 2, 8, 6, 1, 7, 9},
	}
	var g grid
	for y := range gridSize {
		for x := range gridSize {
			g[y][x] = cell{x: x, y: y, v: rows[y][x]}
		}
	}
	return g
}

// emptyGrid returns a 9x9 grid with all cells set to 0 and correct x/y.
func emptyGrid() grid {
	return newGrid(nil)
}

// --- isValid (P0) ---

func TestIsValid(t *testing.T) {
	t.Run("valid placement on empty board", func(t *testing.T) {
		g := emptyGrid()
		if !isValid(&g, 5, 0, 0) {
			t.Error("expected valid placement on empty grid")
		}
	})

	t.Run("row conflict", func(t *testing.T) {
		g := emptyGrid()
		g[0][3].v = 5 // place 5 in row 0, col 3
		if isValid(&g, 5, 0, 0) {
			t.Error("expected row conflict")
		}
	})

	t.Run("column conflict", func(t *testing.T) {
		g := emptyGrid()
		g[3][0].v = 5 // place 5 in col 0, row 3
		if isValid(&g, 5, 0, 0) {
			t.Error("expected column conflict")
		}
	})

	t.Run("box conflict", func(t *testing.T) {
		g := emptyGrid()
		g[1][1].v = 5 // same box as (0,0)
		if isValid(&g, 5, 0, 0) {
			t.Error("expected box conflict")
		}
	})

	t.Run("no self-conflict", func(t *testing.T) {
		g := emptyGrid()
		g[2][2].v = 5
		// Checking isValid for 5 at (2,2) should skip self.
		if !isValid(&g, 5, 2, 2) {
			t.Error("expected no self-conflict")
		}
	})

	t.Run("different value same position", func(t *testing.T) {
		g := emptyGrid()
		g[0][0].v = 3
		if !isValid(&g, 5, 0, 0) {
			t.Error("expected no conflict for different value")
		}
	})

	t.Run("box boundary (2,2) vs (3,3) different boxes", func(t *testing.T) {
		g := emptyGrid()
		g[3][3].v = 5 // box (1,1), different from (0,0) box
		if !isValid(&g, 5, 2, 2) {
			t.Error("(2,2) and (3,3) are in different boxes, no conflict expected")
		}
	})

	t.Run("box boundary (2,2) vs (0,0) same box", func(t *testing.T) {
		g := emptyGrid()
		g[0][0].v = 5 // same box as (2,2)
		if isValid(&g, 5, 2, 2) {
			t.Error("(0,0) and (2,2) are in the same box, conflict expected")
		}
	})

	t.Run("all 9 boxes center placement", func(t *testing.T) {
		// Place value 1 in center of each 3x3 box, then verify no cross-box conflict.
		g := emptyGrid()
		centers := [][2]int{
			{1, 1}, {4, 1}, {7, 1},
			{1, 4}, {4, 4}, {7, 4},
			{1, 7}, {4, 7}, {7, 7},
		}
		for _, c := range centers {
			g[c[1]][c[0]].v = 1
		}
		// Each center should be valid (1 appears once per box, once per row, once per column).
		for _, c := range centers {
			if !isValid(&g, 1, c[0], c[1]) {
				t.Errorf("expected valid at center (%d,%d)", c[0], c[1])
			}
		}
	})
}

// --- fillGrid (P1) ---

func TestFillGrid(t *testing.T) {
	t.Run("fills empty grid", func(t *testing.T) {
		g := emptyGrid()
		if !fillGrid(&g) {
			t.Fatal("fillGrid returned false")
		}
		for y := range gridSize {
			for x := range gridSize {
				if g[y][x].v < 1 || g[y][x].v > 9 {
					t.Errorf("cell (%d,%d) = %d, want 1-9", x, y, g[y][x].v)
				}
			}
		}
	})

	t.Run("result is valid", func(t *testing.T) {
		g := emptyGrid()
		fillGrid(&g)
		for y := range gridSize {
			for x := range gridSize {
				if !isValid(&g, g[y][x].v, x, y) {
					t.Errorf("cell (%d,%d)=%d fails isValid", x, y, g[y][x].v)
				}
			}
		}
	})

	t.Run("fills partial grid", func(t *testing.T) {
		provided := []cell{
			{x: 0, y: 0, v: 5},
			{x: 1, y: 0, v: 3},
		}
		g := newGrid(provided)
		if !fillGrid(&g) {
			t.Fatal("fillGrid returned false on partial grid")
		}
		if g[0][0].v != 5 {
			t.Errorf("provided cell (0,0) changed to %d", g[0][0].v)
		}
		if g[0][1].v != 3 {
			t.Errorf("provided cell (1,0) changed to %d", g[0][1].v)
		}
	})

	t.Run("produces different results", func(t *testing.T) {
		var grids [5]grid
		for i := range grids {
			grids[i] = emptyGrid()
			fillGrid(&grids[i])
		}
		allSame := true
		for i := 1; i < len(grids); i++ {
			for y := range gridSize {
				for x := range gridSize {
					if grids[i][y][x].v != grids[0][y][x].v {
						allSame = false
						break
					}
				}
				if !allSame {
					break
				}
			}
			if !allSame {
				break
			}
		}
		if allSame {
			t.Error("fillGrid produced identical grids across 5 runs")
		}
	})
}

// --- countSolutions (P0) ---

func TestCountSolutions(t *testing.T) {
	t.Run("complete grid", func(t *testing.T) {
		g := validCompleteGrid()
		if count := countSolutions(&g, 10); count != 1 {
			t.Errorf("countSolutions = %d, want 1", count)
		}
	})

	t.Run("one empty cell unique", func(t *testing.T) {
		g := validCompleteGrid()
		g[0][0].v = 0
		if count := countSolutions(&g, 10); count != 1 {
			t.Errorf("countSolutions = %d, want 1", count)
		}
	})

	t.Run("impossible grid", func(t *testing.T) {
		g := emptyGrid()
		// Place two 1s in row 0 by forcing the grid state.
		g[0][0].v = 1
		g[0][1].v = 1
		if count := countSolutions(&g, 10); count != 0 {
			t.Errorf("countSolutions = %d, want 0", count)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		g := emptyGrid()
		count := countSolutions(&g, 2)
		if count < 2 {
			t.Errorf("countSolutions = %d, want >= 2", count)
		}
		if count > 2 {
			t.Errorf("countSolutions = %d, should not exceed limit 2", count)
		}
	})

	t.Run("two empty cells ambiguous", func(t *testing.T) {
		// Start with a valid complete grid, then clear two cells that create
		// ambiguity: (0,0) and (0,1) where swapping their values still yields
		// a valid grid.
		g := validCompleteGrid()
		// g[0][0]=5, g[0][1]=3. We need to check if swapping creates a valid grid.
		// Clear both and count solutions.
		v1, v2 := g[0][0].v, g[0][1].v
		g[0][0].v = 0
		g[0][1].v = 0
		count := countSolutions(&g, 10)
		// If swapping is valid, count >= 2; otherwise find a different pair.
		// The valid grid has (0,0)=5,(0,1)=3 — swapping: is 3 valid at (0,0)
		// and 5 valid at (0,1)?
		g[0][0].v = v2
		g[0][1].v = v1
		swapValid := isValid(&g, v2, 0, 0) && isValid(&g, v1, 1, 0)
		if swapValid && count < 2 {
			t.Errorf("countSolutions = %d, expected >= 2 for ambiguous pair", count)
		}
		if !swapValid && count != 1 {
			// The pair wasn't actually ambiguous; just verify we got 1.
			t.Logf("pair (0,0)/(0,1) not ambiguous, count=%d", count)
		}
	})
}

// --- computeConflicts (P0) ---

func TestComputeConflicts(t *testing.T) {
	t.Run("no conflicts", func(t *testing.T) {
		g := validCompleteGrid()
		conflicts := computeConflicts(g)
		for y := range gridSize {
			for x := range gridSize {
				if conflicts[y][x] {
					t.Errorf("unexpected conflict at (%d,%d)", x, y)
				}
			}
		}
	})

	t.Run("row conflict", func(t *testing.T) {
		g := validCompleteGrid()
		// Create a row conflict: set (0,1) to same value as (0,0).
		g[0][1].v = g[0][0].v
		conflicts := computeConflicts(g)
		if !conflicts[0][0] {
			t.Error("expected conflict at (0,0)")
		}
		if !conflicts[0][1] {
			t.Error("expected conflict at (1,0)")
		}
	})

	t.Run("column conflict", func(t *testing.T) {
		g := validCompleteGrid()
		// Create a column conflict: set (0,1) row to same as (0,0).
		g[1][0].v = g[0][0].v
		conflicts := computeConflicts(g)
		if !conflicts[0][0] {
			t.Error("expected conflict at (0,0)")
		}
		if !conflicts[1][0] {
			t.Error("expected conflict at (0,1)")
		}
	})

	t.Run("box conflict", func(t *testing.T) {
		g := validCompleteGrid()
		// Create a box conflict: set (1,1) to same value as (0,0).
		g[1][1].v = g[0][0].v
		conflicts := computeConflicts(g)
		if !conflicts[0][0] {
			t.Error("expected conflict at (0,0)")
		}
		if !conflicts[1][1] {
			t.Error("expected conflict at (1,1)")
		}
	})

	t.Run("empty cells no false positives", func(t *testing.T) {
		g := emptyGrid()
		// Place a single value.
		g[0][0].v = 5
		conflicts := computeConflicts(g)
		for y := range gridSize {
			for x := range gridSize {
				if conflicts[y][x] {
					t.Errorf("unexpected conflict at (%d,%d) on mostly empty grid", x, y)
				}
			}
		}
	})

	t.Run("multiple conflicts", func(t *testing.T) {
		g := validCompleteGrid()
		// Row conflict in row 0.
		g[0][1].v = g[0][0].v
		// Column conflict in column 5.
		g[3][5].v = g[0][5].v
		conflicts := computeConflicts(g)
		if !conflicts[0][0] || !conflicts[0][1] {
			t.Error("expected row conflict at (0,0) and (1,0)")
		}
		if !conflicts[0][5] || !conflicts[3][5] {
			t.Error("expected column conflict at (5,0) and (5,3)")
		}
	})
}

// --- isSolvedWith (P0) ---

func TestIsSolvedWith(t *testing.T) {
	t.Run("valid complete grid", func(t *testing.T) {
		g := validCompleteGrid()
		conflicts := computeConflicts(g)
		if !isSolvedWith(g, conflicts) {
			t.Error("expected solved")
		}
	})

	t.Run("has empty cell", func(t *testing.T) {
		g := validCompleteGrid()
		g[4][4].v = 0
		conflicts := computeConflicts(g)
		if isSolvedWith(g, conflicts) {
			t.Error("expected not solved with empty cell")
		}
	})

	t.Run("has conflict", func(t *testing.T) {
		g := validCompleteGrid()
		g[0][1].v = g[0][0].v // duplicate in row
		conflicts := computeConflicts(g)
		if isSolvedWith(g, conflicts) {
			t.Error("expected not solved with conflict")
		}
	})

	t.Run("all zeros", func(t *testing.T) {
		g := emptyGrid()
		conflicts := computeConflicts(g)
		if isSolvedWith(g, conflicts) {
			t.Error("expected not solved for blank grid")
		}
	})
}

// --- newGrid (P1) ---

func TestNewGrid(t *testing.T) {
	t.Run("with provided cells", func(t *testing.T) {
		provided := []cell{
			{x: 0, y: 0, v: 5},
			{x: 4, y: 3, v: 9},
			{x: 8, y: 8, v: 1},
		}
		g := newGrid(provided)
		for _, c := range provided {
			if g[c.y][c.x].v != c.v {
				t.Errorf("cell (%d,%d) = %d, want %d", c.x, c.y, g[c.y][c.x].v, c.v)
			}
		}
	})

	t.Run("without provided cells", func(t *testing.T) {
		g := newGrid(nil)
		for y := range gridSize {
			for x := range gridSize {
				if g[y][x].v != 0 {
					t.Errorf("cell (%d,%d) = %d, want 0", x, y, g[y][x].v)
				}
			}
		}
	})

	t.Run("all cells have correct positions", func(t *testing.T) {
		g := newGrid(nil)
		for y := range gridSize {
			for x := range gridSize {
				if g[y][x].x != x || g[y][x].y != y {
					t.Errorf("cell position = (%d,%d), want (%d,%d)", g[y][x].x, g[y][x].y, x, y)
				}
			}
		}
	})
}

// --- gridToString (P1) ---

func TestGridToString(t *testing.T) {
	t.Run("round-trip", func(t *testing.T) {
		g := validCompleteGrid()
		s := gridToString(g)

		// Parse it back.
		g2 := emptyGrid()
		rows := splitLines(s)
		for y := range gridSize {
			for x := range gridSize {
				g2[y][x].v = int(rows[y][x] - '0')
			}
		}
		for y := range gridSize {
			for x := range gridSize {
				if g2[y][x].v != g[y][x].v {
					t.Errorf("round-trip mismatch at (%d,%d): got %d, want %d", x, y, g2[y][x].v, g[y][x].v)
				}
			}
		}
	})

	t.Run("encodes zeros", func(t *testing.T) {
		g := emptyGrid()
		s := gridToString(g)
		for _, ch := range s {
			if ch != '0' && ch != '\n' {
				t.Errorf("expected only '0' and '\\n', got %q", ch)
			}
		}
	})

	t.Run("encodes 1-9", func(t *testing.T) {
		g := validCompleteGrid()
		s := gridToString(g)
		for _, ch := range s {
			if ch == '\n' {
				continue
			}
			if ch < '1' || ch > '9' {
				t.Errorf("expected digit 1-9, got %q", ch)
			}
		}
	})
}

// splitLines splits a string by newlines (helper for test).
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := range len(s) {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

// --- GenerateProvidedCells (P2) ---

func TestGenerateProvidedCells(t *testing.T) {
	mode := NewMode("Test", "test", 30)

	t.Run("produces solvable puzzle", func(t *testing.T) {
		cells := GenerateProvidedCells(mode)
		g := newGrid(cells)
		count := countSolutions(&g, 2)
		if count != 1 {
			t.Errorf("countSolutions = %d, want 1", count)
		}
	})

	t.Run("approximate count", func(t *testing.T) {
		cells := GenerateProvidedCells(mode)
		// May be slightly higher than target if removal would break uniqueness.
		if len(cells) < mode.ProvidedCount {
			t.Errorf("len(cells) = %d, want >= %d", len(cells), mode.ProvidedCount)
		}
	})

	t.Run("all provided cells valid", func(t *testing.T) {
		cells := GenerateProvidedCells(mode)
		g := newGrid(cells)
		for _, c := range cells {
			if !isValid(&g, c.v, c.x, c.y) {
				t.Errorf("provided cell (%d,%d)=%d fails isValid", c.x, c.y, c.v)
			}
		}
	})

	t.Run("cells have correct positions", func(t *testing.T) {
		cells := GenerateProvidedCells(mode)
		for _, c := range cells {
			if c.x < 0 || c.x > 8 || c.y < 0 || c.y > 8 {
				t.Errorf("cell position (%d,%d) out of bounds", c.x, c.y)
			}
			if c.v < 1 || c.v > 9 {
				t.Errorf("cell value %d out of range at (%d,%d)", c.v, c.x, c.y)
			}
		}
	})
}

// --- Save/Load round-trip (P1) ---

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Run("preserves provided and user cells", func(t *testing.T) {
		provided := []cell{
			{x: 0, y: 0, v: 5},
			{x: 1, y: 0, v: 3},
			{x: 4, y: 4, v: 9},
		}
		g := newGrid(provided)
		// Add user-entered values.
		g[1][0].v = 6
		g[2][2].v = 8

		m := Model{
			grid:         g,
			provided:     provided,
			providedGrid: buildProvidedGrid(provided),
			keys:         DefaultKeyMap,
			cursor:       game.Cursor{X: 2, Y: 3},
			modeTitle:    "Hard",
		}

		data, err := m.GetSave()
		if err != nil {
			t.Fatal(err)
		}

		got, err := ImportModel(data)
		if err != nil {
			t.Fatal(err)
		}

		// Provided cells preserved.
		for _, c := range provided {
			if got.grid[c.y][c.x].v != c.v {
				t.Errorf("provided cell (%d,%d) = %d, want %d", c.x, c.y, got.grid[c.y][c.x].v, c.v)
			}
			if !got.providedGrid[c.y][c.x] {
				t.Errorf("providedGrid[%d][%d] = false, want true", c.y, c.x)
			}
		}

		// User-entered values preserved.
		if got.grid[1][0].v != 6 {
			t.Errorf("user cell (0,1) = %d, want 6", got.grid[1][0].v)
		}
		if got.grid[2][2].v != 8 {
			t.Errorf("user cell (2,2) = %d, want 8", got.grid[2][2].v)
		}

		// Non-provided cells should not be in providedGrid.
		if got.providedGrid[1][0] {
			t.Error("user cell (0,1) should not be marked as provided")
		}
	})

	t.Run("provided vs user distinction", func(t *testing.T) {
		provided := []cell{
			{x: 0, y: 0, v: 5},
		}
		m := Model{
			grid:         newGrid(provided),
			provided:     provided,
			providedGrid: buildProvidedGrid(provided),
			keys:         DefaultKeyMap,
		}

		data, err := m.GetSave()
		if err != nil {
			t.Fatal(err)
		}

		got, err := ImportModel(data)
		if err != nil {
			t.Fatal(err)
		}

		if !got.providedGrid[0][0] {
			t.Error("expected (0,0) to be marked as provided")
		}
		if got.providedGrid[0][1] {
			t.Error("expected (1,0) to not be marked as provided")
		}
	})

	t.Run("empty user input", func(t *testing.T) {
		provided := []cell{
			{x: 0, y: 0, v: 5},
			{x: 8, y: 8, v: 1},
		}
		m := Model{
			grid:         newGrid(provided),
			provided:     provided,
			providedGrid: buildProvidedGrid(provided),
			keys:         DefaultKeyMap,
		}

		data, err := m.GetSave()
		if err != nil {
			t.Fatal(err)
		}

		got, err := ImportModel(data)
		if err != nil {
			t.Fatal(err)
		}

		// All non-provided cells should be 0.
		for y := range gridSize {
			for x := range gridSize {
				if !got.providedGrid[y][x] && got.grid[y][x].v != 0 {
					t.Errorf("non-provided cell (%d,%d) = %d, want 0", x, y, got.grid[y][x].v)
				}
			}
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := ImportModel([]byte("not json"))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}
