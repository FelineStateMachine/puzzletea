package takuzu

import (
	"encoding/json"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
)

// --- Helpers ---

func makeGrid6(rows ...string) grid {
	if len(rows) != 6 {
		panic("makeGrid6 requires exactly 6 rows")
	}
	g := make(grid, 6)
	for i, row := range rows {
		g[i] = []rune(row)
	}
	return g
}

func validGrid6() grid {
	// A valid 6x6 Takuzu grid (3 zeros, 3 ones per line, no triples, unique lines)
	return makeGrid6(
		"001011",
		"110100",
		"010101",
		"101010",
		"100110",
		"011001",
	)
}

// --- canPlace (P0) ---

func TestCanPlace_Horizontal_Triples(t *testing.T) {
	tests := []struct {
		name string
		row  string
		x    int
		val  rune
		want bool
	}{
		{"00. placing 0 at x=2", "00....", 2, zeroCell, false},
		{"00. placing 1 at x=2", "00....", 2, oneCell, true},
		{".00 placing 0 at x=0", ".00...", 0, zeroCell, false},
		{".00 placing 1 at x=0", ".00...", 0, oneCell, true},
		{"0.0 placing 0 at x=1", "0.0...", 1, zeroCell, false},
		{"0.0 placing 1 at x=1", "0.0...", 1, oneCell, true},
		{"01. placing 0 at x=2", "01....", 2, zeroCell, true},
		{"11. placing 1 at x=2", "11....", 2, oneCell, false},
		{".11 placing 1 at x=0", ".11...", 0, oneCell, false},
		{"1.1 placing 1 at x=1", "1.1...", 1, oneCell, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := makeGrid6(tt.row, "......", "......", "......", "......", "......")
			if got := canPlace(g, 6, tt.x, 0, tt.val); got != tt.want {
				t.Errorf("canPlace = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanPlace_Vertical_Triples(t *testing.T) {
	tests := []struct {
		name string
		col  string // column 0 values by row index; '.' = leave empty
		y    int
		val  rune
		want bool
	}{
		{"y-2,y-1 match: 00. placing 0 at y=2", "00....", 2, zeroCell, false},
		{"y-2,y-1 match: 00. placing 1 at y=2", "00....", 2, oneCell, true},
		{"y+1,y+2 match: .00 placing 0 at y=0", ".00...", 0, zeroCell, false},
		{"y-1,y+1 match: 0.0 placing 0 at y=1", "0.0...", 1, zeroCell, false},
		{"y-1,y+1 match: 0.0 placing 1 at y=1", "0.0...", 1, oneCell, true},
		{"no triple: 01. placing 0 at y=2", "01....", 2, zeroCell, true},
		{"y-2,y-1 match ones: 11. placing 1 at y=2", "11....", 2, oneCell, false},
		{"y+1,y+2 match ones: .11 placing 1 at y=0", ".11...", 0, oneCell, false},
		{"y-1,y+1 match ones: 1.1 placing 1 at y=1", "1.1...", 1, oneCell, false},
		{"bottom edge: ....00 placing 0 at y=3", "....00", 3, zeroCell, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := make(grid, 6)
			for y := range 6 {
				g[y] = []rune("......")
				if y < len(tt.col) && tt.col[y] != '.' {
					g[y][0] = rune(tt.col[y])
				}
			}
			if got := canPlace(g, 6, 0, tt.y, tt.val); got != tt.want {
				t.Errorf("canPlace = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanPlace_EqualCounts(t *testing.T) {
	t.Run("row at limit for zeros", func(t *testing.T) {
		g := makeGrid6("000...", "......", "......", "......", "......", "......")
		// Row already has 3 zeros, can't place another.
		if got := canPlace(g, 6, 3, 0, zeroCell); got != false {
			t.Error("canPlace should reject when row has size/2 zeros")
		}
	})

	t.Run("row at limit for ones", func(t *testing.T) {
		g := makeGrid6("111...", "......", "......", "......", "......", "......")
		if got := canPlace(g, 6, 3, 0, oneCell); got != false {
			t.Error("canPlace should reject when row has size/2 ones")
		}
	})

	t.Run("row has room for both", func(t *testing.T) {
		g := makeGrid6("01....", "......", "......", "......", "......", "......")
		// Only 1 zero and 1 one, room for both.
		if got := canPlace(g, 6, 2, 0, zeroCell); got != true {
			t.Error("canPlace should allow when row has room")
		}
		if got := canPlace(g, 6, 2, 0, oneCell); got != true {
			t.Error("canPlace should allow when row has room")
		}
	})

	t.Run("column at limit", func(t *testing.T) {
		g := make(grid, 6)
		for y := range 6 {
			g[y] = []rune("......")
		}
		// Put 3 zeros in column 0.
		g[0][0] = zeroCell
		g[1][0] = zeroCell
		g[2][0] = zeroCell

		if got := canPlace(g, 6, 0, 3, zeroCell); got != false {
			t.Error("canPlace should reject when column has size/2 zeros")
		}
	})
}

func TestCanPlace_UniqueLines(t *testing.T) {
	t.Run("completing row creates duplicate", func(t *testing.T) {
		g := makeGrid6(
			"00101.", // Will complete to 001011
			"001011", // Already exists
			"......",
			"......",
			"......",
			"......",
		)
		// Placing '1' at (5, 0) would make row 0 identical to row 1.
		if got := canPlace(g, 6, 5, 0, oneCell); got != false {
			t.Error("canPlace should reject when completing row creates duplicate")
		}
	})

	t.Run("completing row is unique", func(t *testing.T) {
		g := makeGrid6(
			"01101.", // Will complete to 011010 (3 zeros, 3 ones)
			"001011", // Different
			"......",
			"......",
			"......",
			"......",
		)
		// Placing '0' at (5, 0) makes row 0 unique and has valid counts.
		if got := canPlace(g, 6, 5, 0, zeroCell); got != true {
			t.Error("canPlace should allow unique row completion")
		}
	})

	t.Run("completing column creates duplicate", func(t *testing.T) {
		g := make(grid, 6)
		for y := range 6 {
			g[y] = []rune("......")
		}
		// Column 0: 00101.  Column 1: 001011 (complete)
		g[0][0], g[1][0], g[2][0], g[3][0], g[4][0] = '0', '0', '1', '0', '1'
		g[0][1], g[1][1], g[2][1], g[3][1], g[4][1], g[5][1] = '0', '0', '1', '0', '1', '1'

		// Placing '1' at (0, 5) would duplicate column 1.
		if got := canPlace(g, 6, 0, 5, oneCell); got != false {
			t.Error("canPlace should reject when completing column creates duplicate")
		}
	})

	t.Run("row not yet complete no uniqueness check", func(t *testing.T) {
		g := makeGrid6(
			"0010..", // Not complete
			"001011",
			"......",
			"......",
			"......",
			"......",
		)
		// Placing at (4, 0) doesn't trigger uniqueness check (row still incomplete after).
		if got := canPlace(g, 6, 4, 0, oneCell); got != true {
			t.Error("canPlace should allow when row not fully complete")
		}
	})
}

// --- rowEqual / colEqual (P0) ---

func TestRowEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b []rune
		want bool
	}{
		{"identical", []rune{'0', '1', '0', '1'}, []rune{'0', '1', '0', '1'}, true},
		{"different", []rune{'0', '1', '0', '1'}, []rune{'0', '1', '1', '0'}, false},
		{"different lengths", []rune{'0', '1'}, []rune{'0', '1', '0'}, false},
		{"both empty", []rune{}, []rune{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rowEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("rowEqual = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColEqual(t *testing.T) {
	g := makeGrid6(
		"001011",
		"010101",
		"101010",
		"110100",
		"001011",
		"010001",
	)

	t.Run("identical columns", func(t *testing.T) {
		// Columns 2 and 4 are both [1,0,1,0,1,0].
		if !colEqual(g, 6, 2, 4) {
			t.Error("colEqual should return true for identical columns")
		}
	})

	t.Run("different columns", func(t *testing.T) {
		if colEqual(g, 6, 0, 1) {
			t.Error("colEqual should return false for different columns")
		}
	})
}

// --- rowEqualWith / colEqualWith (P0) ---

func TestRowEqualWith(t *testing.T) {
	a := []rune{'0', '0', '1', '0', '1', '1'}
	b := []rune{'0', '0', '1', '0', '1', '1'}

	t.Run("override makes no difference when already equal", func(t *testing.T) {
		// a[3]='0', override with '0', should still match b[3]='0'
		if !rowEqualWith(a, b, 3, '0') {
			t.Error("rowEqualWith should return true")
		}
	})

	t.Run("override fixes difference", func(t *testing.T) {
		b2 := []rune{'0', '0', '1', '1', '1', '1'} // Differs at index 3
		if !rowEqualWith(a, b2, 3, '1') {
			t.Error("rowEqualWith should return true when override fixes it")
		}
	})

	t.Run("override doesn't help", func(t *testing.T) {
		b2 := []rune{'1', '0', '1', '0', '1', '1'} // Differs at index 0
		if rowEqualWith(a, b2, 3, '1') {
			t.Error("rowEqualWith should return false when override doesn't match")
		}
	})

	t.Run("different lengths", func(t *testing.T) {
		b2 := []rune{'0', '0', '1'}
		if rowEqualWith(a, b2, 0, '0') {
			t.Error("rowEqualWith should return false for different lengths")
		}
	})
}

func TestColEqualWith(t *testing.T) {
	g := makeGrid6(
		"001011",
		"001011",
		"101010",
		"110100",
		"001011",
		"001001",
	)

	t.Run("override fixes difference", func(t *testing.T) {
		// Column 0 and 1 differ only at row 2 (g[2][0]='1', g[2][1]='0')
		// If we treat g[2][0] as '0', they'd be equal.
		if !colEqualWith(g, 6, 0, 1, 2, '0') {
			t.Error("colEqualWith should return true when override matches")
		}
	})

	t.Run("override doesn't help", func(t *testing.T) {
		if colEqualWith(g, 6, 0, 2, 1, '0') {
			t.Error("colEqualWith should return false when columns differ elsewhere")
		}
	})
}

// --- rowFilled / colFilled / rowFilledExcept / colFilledExcept (P1) ---

func TestRowFilled(t *testing.T) {
	g := makeGrid6(
		"001011",
		"01....",
		"......",
		"101010",
		"......",
		"......",
	)

	tests := []struct {
		name string
		y    int
		want bool
	}{
		{"all filled", 0, true},
		{"has empty", 1, false},
		{"all empty", 2, false},
		{"all filled again", 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rowFilled(g, tt.y, 6); got != tt.want {
				t.Errorf("rowFilled(row %d) = %v, want %v", tt.y, got, tt.want)
			}
		})
	}
}

func TestColFilled(t *testing.T) {
	g := makeGrid6(
		"0.....",
		"0.....",
		"1.....",
		"0.....",
		"1.....",
		"1.....",
	)

	t.Run("all filled", func(t *testing.T) {
		if !colFilled(g, 0, 6) {
			t.Error("colFilled should return true for filled column")
		}
	})

	t.Run("has empty", func(t *testing.T) {
		if colFilled(g, 1, 6) {
			t.Error("colFilled should return false for empty column")
		}
	})
}

func TestRowFilledExcept(t *testing.T) {
	g := makeGrid6(
		"00101.",
		"0.1...",
		"......",
		"......",
		"......",
		"......",
	)

	t.Run("skip is only empty", func(t *testing.T) {
		if !rowFilledExcept(g, 0, 5, 6) {
			t.Error("rowFilledExcept should return true when skip is only empty cell")
		}
	})

	t.Run("another empty exists", func(t *testing.T) {
		if rowFilledExcept(g, 1, 1, 6) {
			t.Error("rowFilledExcept should return false when other empties exist")
		}
	})

	t.Run("all empty except skip", func(t *testing.T) {
		if rowFilledExcept(g, 2, 0, 6) {
			t.Error("rowFilledExcept should return false for all-empty row")
		}
	})
}

func TestColFilledExcept(t *testing.T) {
	g := makeGrid6(
		"0.....",
		"0.....",
		"......",
		"1.....",
		"0.....",
		"1.....",
	)

	t.Run("skip is only empty", func(t *testing.T) {
		if !colFilledExcept(g, 0, 2, 6) {
			t.Error("colFilledExcept should return true when skip is only empty")
		}
	})

	t.Run("another empty exists", func(t *testing.T) {
		if colFilledExcept(g, 1, 0, 6) {
			t.Error("colFilledExcept should return false when column has other empties")
		}
	})
}

// --- checkConstraints (P0) ---

func TestCheckConstraints(t *testing.T) {
	t.Run("valid grid", func(t *testing.T) {
		g := validGrid6()
		if !checkConstraints(g, 6) {
			t.Error("checkConstraints should return true for valid grid")
		}
	})

	t.Run("triple in row", func(t *testing.T) {
		g := makeGrid6(
			"000111", // Triple 0s and triple 1s
			"010101",
			"101010",
			"110100",
			"100110",
			"011001",
		)
		if checkConstraints(g, 6) {
			t.Error("checkConstraints should reject grid with triple in row")
		}
	})

	t.Run("triple in column", func(t *testing.T) {
		g := makeGrid6(
			"001011",
			"001011",
			"001011", // Column 0 has triple 0s
			"110100",
			"100110",
			"011001",
		)
		if checkConstraints(g, 6) {
			t.Error("checkConstraints should reject grid with triple in column")
		}
	})

	t.Run("unequal row count", func(t *testing.T) {
		g := makeGrid6(
			"000011", // Valid count
			"010101",
			"101010",
			"111000", // Valid count
			"100110",
			"011001",
		)
		// Modify to break count: row 0 now has 4 zeros, 2 ones.
		g[0] = []rune("000001")
		if checkConstraints(g, 6) {
			t.Error("checkConstraints should reject unequal row count")
		}
	})

	t.Run("unequal column count", func(t *testing.T) {
		g := validGrid6()
		// Break column 0 count.
		g[0][0] = oneCell
		g[1][0] = oneCell
		g[2][0] = oneCell
		g[3][0] = oneCell // Now column has 4 ones
		if checkConstraints(g, 6) {
			t.Error("checkConstraints should reject unequal column count")
		}
	})
}

// --- hasUniqueLines (P0) ---

func TestHasUniqueLines(t *testing.T) {
	t.Run("all unique", func(t *testing.T) {
		g := validGrid6()
		if !hasUniqueLines(g, 6) {
			t.Error("hasUniqueLines should return true for unique grid")
		}
	})

	t.Run("duplicate rows", func(t *testing.T) {
		g := makeGrid6(
			"001011",
			"010101",
			"001011", // Duplicate of row 0
			"110100",
			"100110",
			"011001",
		)
		if hasUniqueLines(g, 6) {
			t.Error("hasUniqueLines should return false for duplicate rows")
		}
	})

	t.Run("duplicate columns", func(t *testing.T) {
		g := makeGrid6(
			"001011",
			"001101",
			"101010",
			"101100",
			"001011",
			"001001",
		)
		// Columns 0 and 3 are both [0,0,1,1,0,0].
		if hasUniqueLines(g, 6) {
			t.Error("hasUniqueLines should return false for duplicate columns")
		}
	})

	t.Run("partial duplicate still unique", func(t *testing.T) {
		g := makeGrid6(
			"001011",
			"001101", // Shares prefix with row 0 but differs at end
			"101010",
			"110100",
			"100110",
			"011001",
		)
		if !hasUniqueLines(g, 6) {
			t.Error("hasUniqueLines should return true when rows only partially match")
		}
	})
}

// --- Grid serialization (P1) ---

func TestGridSerialization(t *testing.T) {
	t.Run("6x6 round-trip", func(t *testing.T) {
		g := makeGrid6("001011", "010101", "101010", "110100", "100110", "011001")
		s := g.String()
		g2 := newGrid(state(s))

		if len(g2) != len(g) {
			t.Fatalf("row count = %d, want %d", len(g2), len(g))
		}
		for y := range g {
			for x := range g[y] {
				if g2[y][x] != g[y][x] {
					t.Errorf("cell (%d,%d) = %c, want %c", x, y, g2[y][x], g[y][x])
				}
			}
		}
	})

	t.Run("8x8 round-trip", func(t *testing.T) {
		s := createEmptyState(8)
		g := newGrid(s)
		s2 := g.String()
		if string(s) != s2 {
			t.Error("8x8 round-trip failed")
		}
	})

	t.Run("all empty", func(t *testing.T) {
		s := createEmptyState(6)
		g := newGrid(s)
		for y := range g {
			for x := range g[y] {
				if g[y][x] != emptyCell {
					t.Errorf("cell (%d,%d) = %c, want emptyCell", x, y, g[y][x])
				}
			}
		}
	})
}

func TestGridClone(t *testing.T) {
	original := makeGrid6("001011", "010101", "101010", "110100", "100110", "011001")
	cloned := original.clone()

	t.Run("content match", func(t *testing.T) {
		for y := range original {
			for x := range original[y] {
				if cloned[y][x] != original[y][x] {
					t.Errorf("cloned[%d][%d] = %c, want %c", y, x, cloned[y][x], original[y][x])
				}
			}
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		cloned[0][0] = 'X'
		if original[0][0] == 'X' {
			t.Error("clone modified original grid")
		}
	})
}

func TestCreateEmptyState(t *testing.T) {
	t.Run("size 6", func(t *testing.T) {
		s := createEmptyState(6)
		g := newGrid(s)
		if len(g) != 6 {
			t.Fatalf("row count = %d, want 6", len(g))
		}
		for y := range g {
			if len(g[y]) != 6 {
				t.Fatalf("col count at row %d = %d, want 6", y, len(g[y]))
			}
			for x := range g[y] {
				if g[y][x] != emptyCell {
					t.Errorf("cell (%d,%d) = %c, want emptyCell", x, y, g[y][x])
				}
			}
		}
	})

	t.Run("size 0", func(t *testing.T) {
		s := createEmptyState(0)
		if s != "" {
			t.Errorf("createEmptyState(0) = %q, want empty", s)
		}
	})
}

// --- serializeProvided / deserializeProvided (P1) ---

func TestProvidedSerialization(t *testing.T) {
	t.Run("round-trip", func(t *testing.T) {
		p := [][]bool{
			{true, false, true, false, true, false},
			{false, true, false, true, false, true},
			{true, true, false, false, true, true},
			{false, false, true, true, false, false},
			{true, false, true, false, true, false},
			{false, true, false, true, false, true},
		}
		s := serializeProvided(p)
		p2 := deserializeProvided(s, 6)

		if len(p2) != 6 {
			t.Fatalf("row count = %d, want 6", len(p2))
		}
		for y := range p {
			for x := range p[y] {
				if p2[y][x] != p[y][x] {
					t.Errorf("provided[%d][%d] = %v, want %v", y, x, p2[y][x], p[y][x])
				}
			}
		}
	})

	t.Run("all provided", func(t *testing.T) {
		p := make([][]bool, 3)
		for y := range p {
			p[y] = []bool{true, true, true}
		}
		s := serializeProvided(p)
		if s != "###\n###\n###" {
			t.Errorf("serializeProvided = %q, want ###\\n###\\n###", s)
		}
	})

	t.Run("none provided", func(t *testing.T) {
		p := make([][]bool, 3)
		for y := range p {
			p[y] = []bool{false, false, false}
		}
		s := serializeProvided(p)
		if s != "...\n...\n..." {
			t.Errorf("serializeProvided = %q, want ...\\n...\\n...", s)
		}
	})

	t.Run("checkerboard", func(t *testing.T) {
		p := [][]bool{
			{true, false},
			{false, true},
		}
		s := serializeProvided(p)
		expected := "#.\n.#"
		if s != expected {
			t.Errorf("serializeProvided = %q, want %q", s, expected)
		}
		p2 := deserializeProvided(s, 2)
		for y := range p {
			for x := range p[y] {
				if p2[y][x] != p[y][x] {
					t.Errorf("deserialized[%d][%d] = %v, want %v", y, x, p2[y][x], p[y][x])
				}
			}
		}
	})
}

// --- fillGrid (P2) ---

func TestFillGrid(t *testing.T) {
	t.Run("fills 6x6", func(t *testing.T) {
		g := newGrid(createEmptyState(6))
		ok := fillGrid(g, 6)
		if !ok {
			t.Fatal("fillGrid returned false")
		}

		// Check no empty cells.
		for y := range g {
			for x := range g[y] {
				if g[y][x] == emptyCell {
					t.Errorf("cell (%d,%d) is still empty", x, y)
				}
			}
		}
	})

	t.Run("result valid", func(t *testing.T) {
		g := newGrid(createEmptyState(6))
		fillGrid(g, 6)
		if !checkConstraints(g, 6) {
			t.Error("fillGrid produced grid that fails checkConstraints")
		}
	})

	t.Run("result has unique lines", func(t *testing.T) {
		g := newGrid(createEmptyState(6))
		fillGrid(g, 6)
		if !hasUniqueLines(g, 6) {
			t.Error("fillGrid produced grid with duplicate lines")
		}
	})
}

// --- countSolutions (P2) ---

func TestCountSolutions(t *testing.T) {
	t.Run("complete valid grid", func(t *testing.T) {
		g := validGrid6()
		count := countSolutions(g, 6, 10)
		if count != 1 {
			t.Errorf("countSolutions = %d, want 1", count)
		}
	})

	t.Run("one empty cell unique", func(t *testing.T) {
		g := validGrid6()
		// Remove one cell.
		saved := g[0][0]
		g[0][0] = emptyCell
		count := countSolutions(g, 6, 10)
		if count != 1 {
			t.Errorf("countSolutions = %d, want 1", count)
		}
		g[0][0] = saved
	})

	t.Run("multiple solutions with limit", func(t *testing.T) {
		g := newGrid(createEmptyState(6))
		// Empty grid has many solutions, should hit limit.
		count := countSolutions(g, 6, 2)
		if count < 2 {
			t.Errorf("countSolutions = %d, want >= 2", count)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		g := newGrid(createEmptyState(6))
		count := countSolutions(g, 6, 3)
		if count > 3 {
			t.Errorf("countSolutions = %d, should not exceed limit 3", count)
		}
	})
}

// --- generatePuzzle (P2) ---

func TestGeneratePuzzle(t *testing.T) {
	complete := generateComplete(6)

	t.Run("unique solution", func(t *testing.T) {
		puzzle, _ := generatePuzzle(complete, 6, 0.5)
		count := countSolutions(puzzle, 6, 2)
		if count != 1 {
			t.Errorf("generated puzzle has %d solutions, want 1", count)
		}
	})

	t.Run("provided mask consistent with empties", func(t *testing.T) {
		puzzle, provided := generatePuzzle(complete, 6, 0.5)
		for y := range puzzle {
			for x := range puzzle[y] {
				isEmpty := puzzle[y][x] == emptyCell
				isProvided := provided[y][x]
				if isEmpty == isProvided {
					t.Errorf("cell (%d,%d): isEmpty=%v but isProvided=%v", x, y, isEmpty, isProvided)
				}
			}
		}
	})

	t.Run("approximate prefill", func(t *testing.T) {
		_, provided := generatePuzzle(complete, 6, 0.5)
		count := 0
		for y := range provided {
			for x := range provided[y] {
				if provided[y][x] {
					count++
				}
			}
		}
		target := 18 // 50% of 36
		if count < target-5 || count > target+5 {
			t.Errorf("provided cell count = %d, want ~%d (±5)", count, target)
		}
	})
}

// --- Save/Load round-trip (P1) ---

func TestSaveLoadRoundTrip(t *testing.T) {
	puzzle := makeGrid6("00101.", "01....", "......", "......", "......", "......")
	provided := make([][]bool, 6)
	for y := range 6 {
		provided[y] = make([]bool, 6)
		for x := range 6 {
			provided[y][x] = puzzle[y][x] != emptyCell
		}
	}

	m := Model{
		size:      6,
		grid:      puzzle,
		provided:  provided,
		cursor:    game.Cursor{X: 2, Y: 3},
		keys:      DefaultKeyMap,
		modeTitle: "Medium",
		solved:    false,
	}

	data, err := m.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}

	// Size.
	if got.size != 6 {
		t.Errorf("size = %d, want 6", got.size)
	}

	// Grid content.
	if got.grid.String() != m.grid.String() {
		t.Error("grid mismatch")
	}

	// Provided matrix.
	for y := range provided {
		for x := range provided[y] {
			if got.provided[y][x] != provided[y][x] {
				t.Errorf("provided[%d][%d] = %v, want %v", y, x, got.provided[y][x], provided[y][x])
			}
		}
	}

	// Mode title.
	if got.modeTitle != "Medium" {
		t.Errorf("modeTitle = %q, want %q", got.modeTitle, "Medium")
	}
}

func TestImportModel_InvalidJSON(t *testing.T) {
	_, err := ImportModel([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestImportModel_InvalidSize(t *testing.T) {
	data, _ := json.Marshal(Save{Size: 0})
	_, err := ImportModel(data)
	if err == nil {
		t.Fatal("expected error for invalid size")
	}
}

func TestImportModel_SolvedRecalculated(t *testing.T) {
	g := validGrid6()
	provided := make([][]bool, 6)
	for y := range 6 {
		provided[y] = make([]bool, 6)
	}

	data, _ := json.Marshal(Save{
		Size:     6,
		State:    g.String(),
		Provided: serializeProvided(provided),
	})

	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}

	if !got.solved {
		t.Error("ImportModel should set solved=true for complete valid grid")
	}
}

// --- Uniqueness across all modes (P0) ---

func TestGeneratePuzzle_AllModes_Unique(t *testing.T) {
	modes := []struct {
		name      string
		size      int
		prefilled float64
		runs      int
	}{
		{"Beginner 6x6", 6, 0.50, 3},
		{"Easy 6x6", 6, 0.40, 3},
		{"Medium 8x8", 8, 0.40, 3},
		{"Tricky 10x10", 10, 0.38, 2},
		{"Hard 10x10", 10, 0.32, 2},
		{"Very Hard 12x12", 12, 0.30, 1},
		{"Extreme 14x14", 14, 0.28, 1},
	}

	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			for range m.runs {
				complete := generateComplete(m.size)
				puzzle, _ := generatePuzzle(complete, m.size, m.prefilled)
				count := countSolutions(puzzle, m.size, 2)
				if count != 1 {
					t.Errorf("puzzle has %d solutions, want 1", count)
					return
				}
			}
		})
	}
}
