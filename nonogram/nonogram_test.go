package nonogram

import (
	"reflect"
	"strings"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
)

// --- generateTomography (P0) ---

func TestGenerateTomography(t *testing.T) {
	tests := []struct {
		name     string
		grid     grid
		wantRows TomographyDefinition
		wantCols TomographyDefinition
	}{
		{
			name:     "all filled row",
			grid:     grid{[]rune{filledTile, filledTile, filledTile}},
			wantRows: TomographyDefinition{{3}},
			wantCols: TomographyDefinition{{1}, {1}, {1}},
		},
		{
			name:     "all empty row",
			grid:     grid{[]rune{emptyTile, emptyTile, emptyTile}},
			wantRows: TomographyDefinition{{0}},
			wantCols: TomographyDefinition{{0}, {0}, {0}},
		},
		{
			name:     "single block",
			grid:     grid{[]rune{filledTile, filledTile, emptyTile, emptyTile}},
			wantRows: TomographyDefinition{{2}},
			wantCols: TomographyDefinition{{1}, {1}, {0}, {0}},
		},
		{
			name:     "two blocks",
			grid:     grid{[]rune{filledTile, filledTile, emptyTile, filledTile}},
			wantRows: TomographyDefinition{{2, 1}},
			wantCols: TomographyDefinition{{1}, {1}, {0}, {1}},
		},
		{
			name:     "alternating",
			grid:     grid{[]rune{filledTile, emptyTile, filledTile, emptyTile, filledTile}},
			wantRows: TomographyDefinition{{1, 1, 1}},
			wantCols: TomographyDefinition{{1}, {0}, {1}, {0}, {1}},
		},
		{
			name: "column hints vertical pattern",
			grid: grid{
				[]rune{filledTile, emptyTile},
				[]rune{filledTile, emptyTile},
				[]rune{emptyTile, filledTile},
			},
			wantRows: TomographyDefinition{{1}, {1}, {1}},
			wantCols: TomographyDefinition{{2}, {1}},
		},
		{
			name:     "1x1 filled",
			grid:     grid{[]rune{filledTile}},
			wantRows: TomographyDefinition{{1}},
			wantCols: TomographyDefinition{{1}},
		},
		{
			name:     "1x1 empty",
			grid:     grid{[]rune{emptyTile}},
			wantRows: TomographyDefinition{{0}},
			wantCols: TomographyDefinition{{0}},
		},
		{
			name:     "empty grid",
			grid:     grid{},
			wantRows: TomographyDefinition{},
			wantCols: TomographyDefinition{},
		},
		{
			name: "rectangular grid 3x5",
			grid: grid{
				[]rune{filledTile, emptyTile, filledTile, emptyTile, filledTile},
				[]rune{emptyTile, emptyTile, emptyTile, emptyTile, emptyTile},
				[]rune{filledTile, filledTile, filledTile, filledTile, filledTile},
			},
			wantRows: TomographyDefinition{{1, 1, 1}, {0}, {5}},
			wantCols: TomographyDefinition{{1, 1}, {1}, {1, 1}, {1}, {1, 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := generateTomography(tt.grid)
			if !reflect.DeepEqual([][]int(hints.rows), [][]int(tt.wantRows)) {
				t.Errorf("rows = %v, want %v", hints.rows, tt.wantRows)
			}
			if !reflect.DeepEqual([][]int(hints.cols), [][]int(tt.wantCols)) {
				t.Errorf("cols = %v, want %v", hints.cols, tt.wantCols)
			}
		})
	}
}

// --- TomographyDefinition.equal (P0) ---

func TestTomographyDefinitionEqual(t *testing.T) {
	tests := []struct {
		name string
		a    TomographyDefinition
		b    TomographyDefinition
		want bool
	}{
		{
			name: "identical single hints",
			a:    TomographyDefinition{{1, 2}, {3}},
			b:    TomographyDefinition{{1, 2}, {3}},
			want: true,
		},
		{
			name: "different values",
			a:    TomographyDefinition{{1, 2}},
			b:    TomographyDefinition{{1, 3}},
			want: false,
		},
		{
			name: "different inner lengths",
			a:    TomographyDefinition{{1, 2}},
			b:    TomographyDefinition{{1, 2, 3}},
			want: false,
		},
		{
			name: "different outer lengths",
			a:    TomographyDefinition{{1}, {2}},
			b:    TomographyDefinition{{1}},
			want: false,
		},
		{
			name: "both empty",
			a:    TomographyDefinition{},
			b:    TomographyDefinition{},
			want: true,
		},
		{
			name: "nil vs empty",
			a:    nil,
			b:    TomographyDefinition{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.equal(tt.b); got != tt.want {
				t.Errorf("equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- TomographyDefinition.RequiredLen (P1) ---

func TestTomographyDefinitionRequiredLen(t *testing.T) {
	tests := []struct {
		name string
		td   TomographyDefinition
		want int
	}{
		{"varying lengths", TomographyDefinition{{1}, {2, 3}, {4}}, 2},
		{"single element each", TomographyDefinition{{5}, {3}}, 1},
		{"empty definition", TomographyDefinition{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.td.RequiredLen(); got != tt.want {
				t.Errorf("RequiredLen() = %d, want %d", got, tt.want)
			}
		})
	}
}

// --- Grid serialization round-trip (P1) ---

func TestGridRoundTrip(t *testing.T) {
	t.Run("5x5 round-trip", func(t *testing.T) {
		g := grid{
			[]rune{filledTile, emptyTile, filledTile, emptyTile, filledTile},
			[]rune{emptyTile, filledTile, emptyTile, filledTile, emptyTile},
			[]rune{filledTile, filledTile, filledTile, filledTile, filledTile},
			[]rune{emptyTile, emptyTile, emptyTile, emptyTile, emptyTile},
			[]rune{filledTile, emptyTile, emptyTile, emptyTile, filledTile},
		}
		s := g.String()
		g2 := newGrid(state(s))
		s2 := g2.String()
		if s != s2 {
			t.Errorf("round-trip mismatch:\n  got:  %q\n  want: %q", s2, s)
		}
	})

	t.Run("preserves all tile types", func(t *testing.T) {
		g := grid{
			[]rune{filledTile, markedTile, emptyTile},
		}
		s := g.String()
		g2 := newGrid(state(s))
		s2 := g2.String()
		if s != s2 {
			t.Errorf("tile types not preserved:\n  got:  %q\n  want: %q", s2, s)
		}
	})

	t.Run("single row", func(t *testing.T) {
		g := grid{
			[]rune{filledTile, emptyTile, filledTile},
		}
		s := g.String()
		g2 := newGrid(state(s))
		if len(g2) != 1 {
			t.Fatalf("expected 1 row, got %d", len(g2))
		}
		if g2.String() != s {
			t.Error("single row round-trip mismatch")
		}
	})

	t.Run("single column", func(t *testing.T) {
		g := grid{
			[]rune{filledTile},
			[]rune{emptyTile},
			[]rune{filledTile},
		}
		s := g.String()
		g2 := newGrid(state(s))
		if len(g2) != 3 {
			t.Fatalf("expected 3 rows, got %d", len(g2))
		}
		if g2.String() != s {
			t.Error("single column round-trip mismatch")
		}
	})
}

// --- createEmptyState (P1) ---

func TestCreateEmptyState(t *testing.T) {
	t.Run("5x5", func(t *testing.T) {
		s := createEmptyState(5, 5)
		rows := strings.Split(string(s), "\n")
		if len(rows) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(rows))
		}
		for i, row := range rows {
			if len([]rune(row)) != 5 {
				t.Errorf("row %d: expected 5 runes, got %d", i, len([]rune(row)))
			}
			for _, r := range row {
				if r != emptyTile {
					t.Errorf("row %d: expected emptyTile, got %q", i, r)
				}
			}
		}
	})

	t.Run("1x1", func(t *testing.T) {
		s := createEmptyState(1, 1)
		if string(s) != string(emptyTile) {
			t.Errorf("expected single emptyTile, got %q", s)
		}
		if strings.Contains(string(s), "\n") {
			t.Error("1x1 should not contain newline")
		}
	})

	t.Run("0 dimension", func(t *testing.T) {
		if s := createEmptyState(0, 5); s != "" {
			t.Errorf("expected empty string for h=0, got %q", s)
		}
		if s := createEmptyState(5, 0); s != "" {
			t.Errorf("expected empty string for w=0, got %q", s)
		}
	})

	t.Run("negative dimension", func(t *testing.T) {
		if s := createEmptyState(-1, 5); s != "" {
			t.Errorf("expected empty string for h=-1, got %q", s)
		}
	})
}

// --- generateRandomState (P2) ---

func TestGenerateRandomState(t *testing.T) {
	t.Run("density clamping low", func(t *testing.T) {
		// Use a larger grid and multiple runs to avoid flakiness from per-row jitter.
		totalFilled := 0
		for range 10 {
			s := generateRandomState(10, 10, 0.0)
			g := newGrid(s)
			for _, row := range g {
				for _, r := range row {
					if r == filledTile {
						totalFilled++
					}
				}
			}
		}
		if totalFilled == 0 {
			t.Error("expected some filled cells with density clamped from 0.0")
		}
	})

	t.Run("density clamping high", func(t *testing.T) {
		totalEmpty := 0
		for range 10 {
			s := generateRandomState(10, 10, 1.0)
			g := newGrid(s)
			for _, row := range g {
				for _, r := range row {
					if r == emptyTile {
						totalEmpty++
					}
				}
			}
		}
		if totalEmpty == 0 {
			t.Error("expected some empty cells with density clamped from 1.0")
		}
	})

	t.Run("correct dimensions", func(t *testing.T) {
		s := generateRandomState(5, 10, 0.5)
		g := newGrid(s)
		if len(g) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(g))
		}
		for i, row := range g {
			if len(row) != 10 {
				t.Errorf("row %d: expected 10 cols, got %d", i, len(row))
			}
		}
	})

	t.Run("0 dimensions", func(t *testing.T) {
		if s := generateRandomState(0, 5, 0.5); s != "" {
			t.Errorf("expected empty string for h=0, got %q", s)
		}
		if s := generateRandomState(5, 0, 0.5); s != "" {
			t.Errorf("expected empty string for w=0, got %q", s)
		}
	})

	t.Run("approximate density", func(t *testing.T) {
		totalFilled := 0
		totalCells := 0
		runs := 20
		for range runs {
			s := generateRandomState(10, 10, 0.5)
			g := newGrid(s)
			for _, row := range g {
				for _, r := range row {
					totalCells++
					if r == filledTile {
						totalFilled++
					}
				}
			}
		}
		ratio := float64(totalFilled) / float64(totalCells)
		if ratio < 0.2 || ratio > 0.8 {
			t.Errorf("average density = %.2f, expected roughly 0.5 (within 0.2–0.8)", ratio)
		}
	})
}

// --- isValidPuzzle (P0) ---

func TestIsValidPuzzle(t *testing.T) {
	tests := []struct {
		name   string
		hints  Hints
		height int
		width  int
		want   bool
	}{
		{
			name: "all-zero row",
			hints: Hints{
				rows: TomographyDefinition{{0}, {1, 2}},
				cols: TomographyDefinition{{1}, {2}},
			},
			height: 2, width: 2,
			want: false,
		},
		{
			name: "full-width row",
			hints: Hints{
				rows: TomographyDefinition{{5}, {1, 2}},
				cols: TomographyDefinition{{1}, {1}, {1}, {1}, {1}},
			},
			height: 2, width: 5,
			want: false,
		},
		{
			name: "all-zero column",
			hints: Hints{
				rows: TomographyDefinition{{1}, {1}},
				cols: TomographyDefinition{{0}, {2}},
			},
			height: 2, width: 2,
			want: false,
		},
		{
			name: "full-height column",
			hints: Hints{
				rows: TomographyDefinition{{1}, {1}, {1}, {1}, {1}},
				cols: TomographyDefinition{{5}, {1, 2}},
			},
			height: 5, width: 2,
			want: false,
		},
		{
			name: "valid mixed hints",
			hints: Hints{
				rows: TomographyDefinition{{1, 1}, {2}},
				cols: TomographyDefinition{{1}, {1, 1}, {1}},
			},
			height: 2, width: 3,
			want: true,
		},
		{
			name: "multi-segment rows valid",
			hints: Hints{
				rows: TomographyDefinition{{2, 1}, {1, 1}},
				cols: TomographyDefinition{{1}, {1, 1}, {1}, {1}, {1}},
			},
			height: 2, width: 5,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidPuzzle(tt.hints, tt.height, tt.width); got != tt.want {
				t.Errorf("isValidPuzzle() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- GenerateRandomTomography (P2) ---

func TestGenerateRandomTomography(t *testing.T) {
	mode := NewMode("Test", "test mode", 5, 5, 0.5)

	t.Run("produces valid puzzle", func(t *testing.T) {
		hints := GenerateRandomTomography(mode)
		if !isValidPuzzle(hints, mode.Height, mode.Width) {
			t.Error("GenerateRandomTomography produced invalid puzzle")
		}
	})

	t.Run("hints match dimensions", func(t *testing.T) {
		hints := GenerateRandomTomography(mode)
		if len(hints.rows) != mode.Height {
			t.Errorf("len(rows) = %d, want %d", len(hints.rows), mode.Height)
		}
		if len(hints.cols) != mode.Width {
			t.Errorf("len(cols) = %d, want %d", len(hints.cols), mode.Width)
		}
	})
}

// --- Save/Load round-trip (P1) ---

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		g := grid{
			[]rune{filledTile, emptyTile, markedTile, filledTile, emptyTile},
			[]rune{emptyTile, filledTile, filledTile, emptyTile, filledTile},
			[]rune{filledTile, filledTile, emptyTile, emptyTile, filledTile},
			[]rune{emptyTile, emptyTile, filledTile, filledTile, emptyTile},
			[]rune{filledTile, emptyTile, filledTile, emptyTile, filledTile},
		}
		hints := Hints{
			rows: TomographyDefinition{{2, 1}, {1, 2}, {2, 1}, {2}, {1, 1, 1}},
			cols: TomographyDefinition{{1, 1, 1}, {1, 1}, {1, 1, 1}, {1, 1}, {1, 1, 1}},
		}
		curr := generateTomography(g)

		m := Model{
			width:        5,
			height:       5,
			rowHints:     hints.rows,
			colHints:     hints.cols,
			grid:         g,
			keys:         DefaultKeyMap,
			currentHints: curr,
			solved:       curr.rows.equal(hints.rows) && curr.cols.equal(hints.cols),
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
		if got.grid.String() != m.grid.String() {
			t.Errorf("grid mismatch:\n  got:  %q\n  want: %q", got.grid.String(), m.grid.String())
		}

		// Dimensions.
		if got.width != m.width {
			t.Errorf("width = %d, want %d", got.width, m.width)
		}
		if got.height != m.height {
			t.Errorf("height = %d, want %d", got.height, m.height)
		}

		// Hints.
		if !got.rowHints.equal(m.rowHints) {
			t.Errorf("rowHints mismatch: got %v, want %v", got.rowHints, m.rowHints)
		}
		if !got.colHints.equal(m.colHints) {
			t.Errorf("colHints mismatch: got %v, want %v", got.colHints, m.colHints)
		}

		// Solved state.
		if got.solved != m.solved {
			t.Errorf("solved = %v, want %v", got.solved, m.solved)
		}
	})

	t.Run("preserves solved state", func(t *testing.T) {
		// Create a grid that matches its own tomography (solved).
		g := grid{
			[]rune{filledTile, emptyTile},
			[]rune{emptyTile, filledTile},
		}
		hints := generateTomography(g)

		m := Model{
			width:        2,
			height:       2,
			rowHints:     hints.rows,
			colHints:     hints.cols,
			grid:         g,
			keys:         DefaultKeyMap,
			currentHints: hints,
			solved:       true,
		}

		data, err := m.GetSave()
		if err != nil {
			t.Fatal(err)
		}

		got, err := ImportModel(data)
		if err != nil {
			t.Fatal(err)
		}

		if !got.solved {
			t.Error("expected imported model to be solved")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := ImportModel([]byte("not json"))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

// Ensure the game import is used (it's needed for the init() registration side effect
// and the NewMode constructor uses game.NewBaseMode internally).
var _ game.Gamer = Model{}
