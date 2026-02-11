package lightsout

import (
	"encoding/json"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
)

// --- helpers ---

func copyGrid(g [][]bool) [][]bool {
	c := make([][]bool, len(g))
	for i, row := range g {
		c[i] = make([]bool, len(row))
		copy(c[i], row)
	}
	return c
}

func countDiffs(a, b [][]bool) int {
	n := 0
	for y := range a {
		for x := range a[y] {
			if a[y][x] != b[y][x] {
				n++
			}
		}
	}
	return n
}

func makeGrid(h, w int, val bool) [][]bool {
	g := make([][]bool, h)
	for y := range g {
		g[y] = make([]bool, w)
		if val {
			for x := range g[y] {
				g[y][x] = true
			}
		}
	}
	return g
}

// --- Toggle (P0) ---

func TestToggle_Center(t *testing.T) {
	t.Run("all false to true", func(t *testing.T) {
		g := makeGrid(3, 3, false)
		Toggle(g, 1, 1)

		// Center + 4 neighbors should be true.
		for _, pos := range [][2]int{{1, 1}, {0, 1}, {2, 1}, {1, 0}, {1, 2}} {
			if !g[pos[1]][pos[0]] {
				t.Errorf("expected (%d,%d) to be true", pos[0], pos[1])
			}
		}
		// Corners should remain false.
		for _, pos := range [][2]int{{0, 0}, {2, 0}, {0, 2}, {2, 2}} {
			if g[pos[1]][pos[0]] {
				t.Errorf("expected corner (%d,%d) to be false", pos[0], pos[1])
			}
		}
	})

	t.Run("all true to false", func(t *testing.T) {
		g := makeGrid(3, 3, true)
		Toggle(g, 1, 1)

		for _, pos := range [][2]int{{1, 1}, {0, 1}, {2, 1}, {1, 0}, {1, 2}} {
			if g[pos[1]][pos[0]] {
				t.Errorf("expected (%d,%d) to be false", pos[0], pos[1])
			}
		}
		for _, pos := range [][2]int{{0, 0}, {2, 0}, {0, 2}, {2, 2}} {
			if !g[pos[1]][pos[0]] {
				t.Errorf("expected corner (%d,%d) to be true", pos[0], pos[1])
			}
		}
	})

	t.Run("mixed state flips exactly 5", func(t *testing.T) {
		// Checkerboard pattern.
		g := [][]bool{
			{true, false, true},
			{false, true, false},
			{true, false, true},
		}
		before := copyGrid(g)
		Toggle(g, 1, 1)

		if d := countDiffs(before, g); d != 5 {
			t.Errorf("countDiffs = %d, want 5", d)
		}
	})
}

func TestToggle_Edges(t *testing.T) {
	tests := []struct {
		name      string
		x, y      int
		wantDiffs int
	}{
		{"top-left corner", 0, 0, 3},
		{"top-right corner", 4, 0, 3},
		{"bottom-left corner", 0, 4, 3},
		{"bottom-right corner", 4, 4, 3},
		{"top edge", 2, 0, 4},
		{"left edge", 0, 2, 4},
		{"bottom edge", 2, 4, 4},
		{"right edge", 4, 2, 4},
		{"center", 2, 2, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := makeGrid(5, 5, false)
			before := copyGrid(g)
			Toggle(g, tt.x, tt.y)

			if d := countDiffs(before, g); d != tt.wantDiffs {
				t.Errorf("countDiffs = %d, want %d", d, tt.wantDiffs)
			}
		})
	}
}

func TestToggle_DoubleToggleCancels(t *testing.T) {
	g := makeGrid(3, 3, false)
	g[0][0] = true
	before := copyGrid(g)

	Toggle(g, 1, 1)
	Toggle(g, 1, 1)

	if d := countDiffs(before, g); d != 0 {
		t.Errorf("double toggle should restore grid, got %d diffs", d)
	}
}

func TestToggle_EmptyGrid(t *testing.T) {
	g := [][]bool{}
	Toggle(g, 0, 0) // must not panic
}

func TestToggle_1x1(t *testing.T) {
	g := [][]bool{{false}}
	Toggle(g, 0, 0)
	if !g[0][0] {
		t.Error("expected single cell to be true after toggle")
	}
}

func TestToggle_Rectangular(t *testing.T) {
	g := makeGrid(5, 3, false) // 5 rows, 3 cols
	before := copyGrid(g)
	Toggle(g, 1, 2) // center of grid, all neighbors in bounds

	if d := countDiffs(before, g); d != 5 {
		t.Errorf("countDiffs = %d, want 5", d)
	}
}

// --- IsSolved (P0) ---

func TestIsSolved(t *testing.T) {
	tests := []struct {
		name string
		grid [][]bool
		want bool
	}{
		{"all false 3x3", makeGrid(3, 3, false), true},
		{"one true", func() [][]bool {
			g := makeGrid(3, 3, false)
			g[1][1] = true
			return g
		}(), false},
		{"all true", makeGrid(3, 3, true), false},
		{"empty grid", [][]bool{}, true},
		{"nil grid", nil, true},
		{"single cell false", [][]bool{{false}}, true},
		{"single cell true", [][]bool{{true}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSolved(tt.grid); got != tt.want {
				t.Errorf("IsSolved = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Generate (P1) ---

func TestGenerate_NotSolved(t *testing.T) {
	for range 10 {
		g := Generate(5, 5)
		if IsSolved(g) {
			t.Fatal("Generate(5,5) returned a solved grid")
		}
	}
}

func TestGenerate_Dimensions(t *testing.T) {
	g := Generate(3, 4)
	if len(g) != 4 {
		t.Fatalf("len(grid) = %d, want 4", len(g))
	}
	for y, row := range g {
		if len(row) != 3 {
			t.Errorf("len(grid[%d]) = %d, want 3", y, len(row))
		}
	}
}

func TestGenerate_HasLightsOn(t *testing.T) {
	for range 10 {
		g := Generate(5, 5)
		found := false
		for _, row := range g {
			for _, cell := range row {
				if cell {
					found = true
				}
			}
		}
		if !found {
			t.Fatal("Generate produced a grid with no lights on")
		}
	}
}

// --- New (P1) ---

func TestNew(t *testing.T) {
	t.Run("returns valid model", func(t *testing.T) {
		m, err := New(5, 5)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.grid) != 5 {
			t.Errorf("grid height = %d, want 5", len(m.grid))
		}
		for y, row := range m.grid {
			if len(row) != 5 {
				t.Errorf("grid[%d] width = %d, want 5", y, len(row))
			}
		}
	})

	t.Run("cursor centered odd", func(t *testing.T) {
		m, _ := New(5, 5)
		if m.cursor.X != 2 || m.cursor.Y != 2 {
			t.Errorf("cursor = (%d,%d), want (2,2)", m.cursor.X, m.cursor.Y)
		}
	})

	t.Run("cursor centered even", func(t *testing.T) {
		m, _ := New(4, 4)
		if m.cursor.X != 2 || m.cursor.Y != 2 {
			t.Errorf("cursor = (%d,%d), want (2,2)", m.cursor.X, m.cursor.Y)
		}
	})

	t.Run("grid not solved", func(t *testing.T) {
		m, _ := New(3, 3)
		if m.IsSolved() {
			t.Error("new game should not be solved")
		}
	})
}

// --- Mode.Spawn (P1) ---

func TestModeSpawn(t *testing.T) {
	mode := NewMode("Easy", "3x3 grid", 3, 3)
	g, err := mode.Spawn()
	if err != nil {
		t.Fatal(err)
	}
	if g == nil {
		t.Fatal("Spawn returned nil")
	}
}

// --- Save/Load round-trip (P1) ---

func TestSaveLoadRoundTrip(t *testing.T) {
	m := Model{
		grid:      [][]bool{{true, false, true}, {false, true, false}},
		width:     3,
		height:    2,
		cursor:    game.Cursor{X: 2, Y: 1},
		modeTitle: "Medium",
		keys:      DefaultKeyMap,
	}

	data, err := m.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}

	// Grid content.
	if len(got.grid) != len(m.grid) {
		t.Fatalf("grid height = %d, want %d", len(got.grid), len(m.grid))
	}
	for y := range m.grid {
		for x := range m.grid[y] {
			if got.grid[y][x] != m.grid[y][x] {
				t.Errorf("grid[%d][%d] = %v, want %v", y, x, got.grid[y][x], m.grid[y][x])
			}
		}
	}

	// Dimensions.
	if got.width != 3 {
		t.Errorf("width = %d, want 3", got.width)
	}
	if got.height != 2 {
		t.Errorf("height = %d, want 2", got.height)
	}

	// Cursor.
	if got.cursor.X != 2 || got.cursor.Y != 1 {
		t.Errorf("cursor = (%d,%d), want (2,1)", got.cursor.X, got.cursor.Y)
	}

	// ModeTitle.
	if got.modeTitle != "Medium" {
		t.Errorf("modeTitle = %q, want %q", got.modeTitle, "Medium")
	}
}

func TestImportModel_EmptyGrid(t *testing.T) {
	data, _ := json.Marshal(Save{Grid: [][]bool{}})
	_, err := ImportModel(data)
	if err == nil {
		t.Fatal("expected error for empty grid")
	}
}

func TestImportModel_InvalidJSON(t *testing.T) {
	_, err := ImportModel([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestImportModel_DimensionsDerived(t *testing.T) {
	// 4 rows, 5 cols → height=4, width=5.
	grid := make([][]bool, 4)
	for i := range grid {
		grid[i] = make([]bool, 5)
	}

	data, _ := json.Marshal(Save{Grid: grid})
	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.width != 5 {
		t.Errorf("width = %d, want 5", got.width)
	}
	if got.height != 4 {
		t.Errorf("height = %d, want 4", got.height)
	}
}
