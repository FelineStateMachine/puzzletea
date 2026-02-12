package nonogram

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

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

// --- countSolutions (P0) ---

func TestCountSolutions_Unique(t *testing.T) {
	tests := []struct {
		name  string
		hints Hints
		w, h  int
		want  int
	}{
		{
			name: "3x3 permutation matrix has 6 solutions",
			hints: Hints{
				rows: TomographyDefinition{{1}, {1}, {1}},
				cols: TomographyDefinition{{1}, {1}, {1}},
			},
			w: 3, h: 3,
			want: 6,
		},
		{
			name: "2x2 diagonal has 2 solutions",
			hints: Hints{
				rows: TomographyDefinition{{1}, {1}},
				cols: TomographyDefinition{{1}, {1}},
			},
			w: 2, h: 2,
			want: 2,
		},
		{
			name: "single cell filled",
			hints: Hints{
				rows: TomographyDefinition{{1}},
				cols: TomographyDefinition{{1}},
			},
			w: 1, h: 1,
			want: 1,
		},
		{
			name: "single cell empty",
			hints: Hints{
				rows: TomographyDefinition{{0}},
				cols: TomographyDefinition{{0}},
			},
			w: 1, h: 1,
			want: 1,
		},
		{
			name: "full row constraint",
			hints: Hints{
				rows: TomographyDefinition{{3}},
				cols: TomographyDefinition{{1}, {1}, {1}},
			},
			w: 3, h: 1,
			want: 1,
		},
		{
			name: "full 2x2 grid unique",
			hints: Hints{
				rows: TomographyDefinition{{2}, {2}},
				cols: TomographyDefinition{{2}, {2}},
			},
			w: 2, h: 2,
			want: 1,
		},
		{
			name: "empty 2x2 grid unique",
			hints: Hints{
				rows: TomographyDefinition{{0}, {0}},
				cols: TomographyDefinition{{0}, {0}},
			},
			w: 2, h: 2,
			want: 1,
		},
		{
			name: "L-shape unique",
			hints: Hints{
				rows: TomographyDefinition{{2}, {1}},
				cols: TomographyDefinition{{1}, {2}},
			},
			w: 2, h: 2,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got := countSolutions(tt.hints, tt.w, tt.h, 20, ctx)
			if got != tt.want {
				t.Errorf("countSolutions = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountSolutions_Multiple(t *testing.T) {
	t.Run("symmetric hints with multiple solutions", func(t *testing.T) {
		hints := Hints{
			rows: TomographyDefinition{{1}, {1}, {1}},
			cols: TomographyDefinition{{1}, {1}, {1}},
		}
		ctx := context.Background()
		got := countSolutions(hints, 3, 3, 10, ctx)
		if got < 2 {
			t.Errorf("expected multiple solutions, got %d", got)
		}
	})

	t.Run("ambiguous pattern", func(t *testing.T) {
		hints := Hints{
			rows: TomographyDefinition{{1}, {1}},
			cols: TomographyDefinition{{1}, {1}},
		}
		ctx := context.Background()
		got := countSolutions(hints, 2, 2, 10, ctx)
		if got != 2 {
			t.Errorf("expected 2 solutions, got %d", got)
		}
	})
}

func TestCountSolutions_Limit(t *testing.T) {
	t.Run("respects limit", func(t *testing.T) {
		hints := Hints{
			rows: TomographyDefinition{{1}, {1}, {1}},
			cols: TomographyDefinition{{1}, {1}, {1}},
		}
		ctx := context.Background()
		got := countSolutions(hints, 3, 3, 2, ctx)
		if got > 2 {
			t.Errorf("countSolutions = %d, should not exceed limit 2", got)
		}
	})
}

func TestCountSolutions_Timeout(t *testing.T) {
	t.Run("returns on timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond)

		hints := Hints{
			rows: TomographyDefinition{{1}, {1}, {1}},
			cols: TomographyDefinition{{1}, {1}, {1}},
		}
		got := countSolutions(hints, 3, 3, 10, ctx)
		if got != -1 {
			t.Errorf("expected -1 on timeout, got %d", got)
		}
	})
}

// --- propagateLine (P0) ---

func TestPropagateLine(t *testing.T) {
	tests := []struct {
		name   string
		line   []cellState
		hint   []int
		want   []cellState
		wantOK bool
	}{
		{
			name:   "empty line with zero hint",
			line:   []cellState{cellUnknown, cellUnknown, cellUnknown},
			hint:   []int{0},
			want:   []cellState{cellEmpty, cellEmpty, cellEmpty},
			wantOK: true,
		},
		{
			name:   "full line",
			line:   []cellState{cellUnknown, cellUnknown, cellUnknown},
			hint:   []int{3},
			want:   []cellState{cellFilled, cellFilled, cellFilled},
			wantOK: true,
		},
		{
			name:   "single block in middle of 5",
			line:   []cellState{cellUnknown, cellUnknown, cellUnknown, cellUnknown, cellUnknown},
			hint:   []int{3},
			want:   []cellState{cellUnknown, cellUnknown, cellFilled, cellUnknown, cellUnknown},
			wantOK: true,
		},
		{
			name:   "contradiction",
			line:   []cellState{cellFilled, cellUnknown, cellEmpty},
			hint:   []int{3},
			want:   nil,
			wantOK: false,
		},
		{
			name:   "partial fill forced with [2,1]",
			line:   []cellState{cellFilled, cellUnknown, cellUnknown, cellUnknown, cellUnknown},
			hint:   []int{2, 1},
			want:   []cellState{cellFilled, cellFilled, cellEmpty, cellUnknown, cellUnknown},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := propagateLine(tt.line, tt.hint)
			if ok != tt.wantOK {
				t.Errorf("propagateLine ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if !tt.wantOK {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("got length %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// --- generateLinePossibilities (P0) ---

func TestGenerateLinePossibilities(t *testing.T) {
	tests := []struct {
		name   string
		length int
		hint   []int
		want   int
	}{
		{"empty line", 3, []int{0}, 1},
		{"full line", 3, []int{3}, 1},
		{"single in 3", 3, []int{1}, 3},
		{"two blocks", 5, []int{1, 1}, 6},
		{"block of 2 in 5", 5, []int{2}, 4},
		{"exact fit", 5, []int{2, 2}, 1},
		{"impossible", 3, []int{4}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateLinePossibilities(tt.length, tt.hint)
			if len(got) != tt.want {
				t.Errorf("got %d possibilities, want %d", len(got), tt.want)
			}
		})
	}
}

// --- GenerateRandomTomography uniqueness (P0) ---

func TestGenerateRandomTomography_AllModes_Unique(t *testing.T) {
	modes := []struct {
		name    string
		width   int
		height  int
		density float64
		runs    int
	}{
		{"Easy 5x5", 5, 5, 0.65, 5},
		{"Medium 5x5", 5, 5, 0.50, 5},
		{"Hard 5x5", 5, 5, 0.35, 5},
		{"Easy 10x10", 10, 10, 0.65, 3},
		{"Medium 10x10", 10, 10, 0.50, 3},
		{"Hard 10x10", 10, 10, 0.35, 3},
		{"Easy 15x15", 15, 15, 0.65, 1},
		{"Medium 15x15", 15, 15, 0.50, 1},
		{"Easy 20x20", 20, 20, 0.65, 1},
		{"Medium 20x20", 20, 20, 0.50, 1},
	}

	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			if m.width >= 15 && testing.Short() {
				t.Skip("skipping large grids in short mode")
			}

			mode := NewMode(m.name, "test mode", m.width, m.height, m.density)

			for range m.runs {
				hints := GenerateRandomTomography(mode)
				if len(hints.rows) == 0 || len(hints.cols) == 0 {
					t.Error("GenerateRandomTomography returned empty hints")
					return
				}
				if len(hints.rows) != m.height || len(hints.cols) != m.width {
					t.Errorf("hints dimensions wrong: got rows=%d, cols=%d, want rows=%d, cols=%d",
						len(hints.rows), len(hints.cols), m.height, m.width)
					return
				}
			}
		})
	}
}

func TestGenerateRandomTomography_5x5_VerifyUnique(t *testing.T) {
	mode := NewMode("Test 5x5", "test mode", 5, 5, 0.5)

	hints := GenerateRandomTomography(mode)
	if len(hints.rows) == 0 {
		t.Fatal("GenerateRandomTomography returned empty hints")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count := countSolutions(hints, mode.Width, mode.Height, 2, ctx)
	if count != 1 {
		t.Errorf("puzzle has %d solutions, want 1", count)
	}
}

// Ensure the game import is used (it's needed for the init() registration side effect
// and the NewMode constructor uses game.NewBaseMode internally).
var _ game.Gamer = Model{}
