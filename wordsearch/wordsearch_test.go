package wordsearch

import (
	"encoding/json"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
)

// --- Direction.Vector (P0) ---

func TestDirectionVector(t *testing.T) {
	tests := []struct {
		name   string
		dir    Direction
		wantDX int
		wantDY int
	}{
		{"Right", Right, 1, 0},
		{"Down", Down, 0, 1},
		{"DownRight", DownRight, 1, 1},
		{"DownLeft", DownLeft, -1, 1},
		{"Left", Left, -1, 0},
		{"Up", Up, 0, -1},
		{"UpRight", UpRight, 1, -1},
		{"UpLeft", UpLeft, -1, -1},
		{"invalid", Direction(99), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy := tt.dir.Vector()
			if dx != tt.wantDX || dy != tt.wantDY {
				t.Errorf("Direction(%d).Vector() = (%d,%d), want (%d,%d)", tt.dir, dx, dy, tt.wantDX, tt.wantDY)
			}
		})
	}
}

// --- Word.Contains (P0) ---

func TestWordContains(t *testing.T) {
	hello := Word{Text: "HELLO", Start: Position{0, 0}, End: Position{4, 0}, Direction: Right}
	cat := Word{Text: "CAT", Start: Position{1, 1}, End: Position{3, 3}, Direction: DownRight}
	single := Word{Text: "A", Start: Position{3, 3}, End: Position{3, 3}, Direction: Right}

	tests := []struct {
		name string
		word *Word
		pos  Position
		want bool
	}{
		{"start position", &hello, Position{0, 0}, true},
		{"end position", &hello, Position{4, 0}, true},
		{"middle position", &hello, Position{2, 0}, true},
		{"off-path position", &hello, Position{0, 1}, false},
		{"past-end position", &hello, Position{5, 0}, false},
		{"diagonal word hit", &cat, Position{2, 2}, true},
		{"diagonal start", &cat, Position{1, 1}, true},
		{"diagonal off-path", &cat, Position{1, 2}, false},
		{"single-letter hit", &single, Position{3, 3}, true},
		{"single-letter miss", &single, Position{3, 4}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.word.Contains(tt.pos); got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.pos, got, tt.want)
			}
		})
	}
}

// --- Word.Positions (P0) ---

func TestWordPositions(t *testing.T) {
	tests := []struct {
		name string
		word Word
		want []Position
	}{
		{
			"horizontal 5-letter",
			Word{Text: "HELLO", Start: Position{0, 0}, Direction: Right},
			[]Position{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}},
		},
		{
			"vertical 3-letter",
			Word{Text: "CAT", Start: Position{2, 0}, Direction: Down},
			[]Position{{2, 0}, {2, 1}, {2, 2}},
		},
		{
			"diagonal 2-letter",
			Word{Text: "AB", Start: Position{0, 0}, Direction: DownRight},
			[]Position{{0, 0}, {1, 1}},
		},
		{
			"single letter",
			Word{Text: "A", Start: Position{3, 3}, Direction: Right},
			[]Position{{3, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.word.Positions()
			if len(got) != len(tt.want) {
				t.Fatalf("len(Positions()) = %d, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("Positions()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}

	t.Run("count matches text length", func(t *testing.T) {
		w := Word{Text: "TESTING", Start: Position{0, 0}, Direction: Right}
		if got := len(w.Positions()); got != len(w.Text) {
			t.Errorf("len(Positions()) = %d, want %d", got, len(w.Text))
		}
	})
}

// --- Grid operations (P1) ---

func TestGridGetSet(t *testing.T) {
	g := createEmptyGrid(5, 5)

	t.Run("Set then Get", func(t *testing.T) {
		g.Set(2, 3, 'A')
		if got := g.Get(2, 3); got != 'A' {
			t.Errorf("Get(2,3) = %c, want A", got)
		}
	})

	t.Run("Get out of bounds negative x", func(t *testing.T) {
		if got := g.Get(-1, 0); got != 0 {
			t.Errorf("Get(-1,0) = %d, want 0", got)
		}
	})

	t.Run("Get out of bounds negative y", func(t *testing.T) {
		if got := g.Get(0, -1); got != 0 {
			t.Errorf("Get(0,-1) = %d, want 0", got)
		}
	})

	t.Run("Get out of bounds over x", func(t *testing.T) {
		if got := g.Get(5, 0); got != 0 {
			t.Errorf("Get(5,0) = %d, want 0", got)
		}
	})

	t.Run("Get out of bounds over y", func(t *testing.T) {
		if got := g.Get(0, 5); got != 0 {
			t.Errorf("Get(0,5) = %d, want 0", got)
		}
	})

	t.Run("Set out of bounds no panic", func(t *testing.T) {
		g.Set(-1, 0, 'X') // must not panic
		g.Set(5, 0, 'X')  // must not panic
	})

	t.Run("overwrite existing", func(t *testing.T) {
		g.Set(0, 0, 'A')
		g.Set(0, 0, 'B')
		if got := g.Get(0, 0); got != 'B' {
			t.Errorf("Get(0,0) = %c, want B", got)
		}
	})
}

func TestGridSerialization(t *testing.T) {
	t.Run("round-trip", func(t *testing.T) {
		original := newGrid(state("ABC\nDEF\nGHI"))
		roundTripped := newGrid(state(original.String()))

		if len(roundTripped) != len(original) {
			t.Fatalf("row count = %d, want %d", len(roundTripped), len(original))
		}
		for y := range original {
			for x := range original[y] {
				if roundTripped.Get(x, y) != original.Get(x, y) {
					t.Errorf("cell (%d,%d) = %c, want %c", x, y, roundTripped.Get(x, y), original.Get(x, y))
				}
			}
		}
	})

	t.Run("empty state", func(t *testing.T) {
		g := newGrid(state(""))
		if len(g) != 0 {
			t.Errorf("len(newGrid empty) = %d, want 0", len(g))
		}
	})

	t.Run("createEmptyGrid dimensions", func(t *testing.T) {
		g := createEmptyGrid(5, 5)
		if len(g) != 5 {
			t.Fatalf("height = %d, want 5", len(g))
		}
		for y := range g {
			if len(g[y]) != 5 {
				t.Fatalf("width of row %d = %d, want 5", y, len(g[y]))
			}
			for x := range g[y] {
				if g[y][x] != ' ' {
					t.Errorf("cell (%d,%d) = %c, want space", x, y, g[y][x])
				}
			}
		}
	})

	t.Run("single row", func(t *testing.T) {
		g := newGrid(state("HELLO"))
		if len(g) != 1 {
			t.Fatalf("row count = %d, want 1", len(g))
		}
		if len(g[0]) != 5 {
			t.Errorf("col count = %d, want 5", len(g[0]))
		}
	})

	t.Run("single column", func(t *testing.T) {
		g := newGrid(state("A\nB\nC"))
		if len(g) != 3 {
			t.Fatalf("row count = %d, want 3", len(g))
		}
		for _, row := range g {
			if len(row) != 1 {
				t.Errorf("col count = %d, want 1", len(row))
			}
		}
	})
}

// --- lineDirection (P0) ---

func TestLineDirection(t *testing.T) {
	tests := []struct {
		name      string
		start     game.Cursor
		end       game.Cursor
		wantDX    int
		wantDY    int
		wantValid bool
	}{
		{"horizontal right", game.Cursor{X: 0, Y: 0}, game.Cursor{X: 4, Y: 0}, 1, 0, true},
		{"horizontal left", game.Cursor{X: 4, Y: 0}, game.Cursor{X: 0, Y: 0}, -1, 0, true},
		{"vertical down", game.Cursor{X: 0, Y: 0}, game.Cursor{X: 0, Y: 4}, 0, 1, true},
		{"vertical up", game.Cursor{X: 0, Y: 4}, game.Cursor{X: 0, Y: 0}, 0, -1, true},
		{"diagonal down-right", game.Cursor{X: 0, Y: 0}, game.Cursor{X: 3, Y: 3}, 1, 1, true},
		{"diagonal up-left", game.Cursor{X: 3, Y: 3}, game.Cursor{X: 0, Y: 0}, -1, -1, true},
		{"diagonal down-left", game.Cursor{X: 3, Y: 0}, game.Cursor{X: 0, Y: 3}, -1, 1, true},
		{"diagonal up-right", game.Cursor{X: 0, Y: 3}, game.Cursor{X: 3, Y: 0}, 1, -1, true},
		{"invalid slope", game.Cursor{X: 0, Y: 0}, game.Cursor{X: 2, Y: 3}, 0, 0, false},
		{"invalid steep", game.Cursor{X: 0, Y: 0}, game.Cursor{X: 1, Y: 3}, 0, 0, false},
		{"same point", game.Cursor{X: 2, Y: 2}, game.Cursor{X: 2, Y: 2}, 0, 0, false},
		{"adjacent horizontal", game.Cursor{X: 0, Y: 0}, game.Cursor{X: 1, Y: 0}, 1, 0, true},
		{"adjacent diagonal", game.Cursor{X: 0, Y: 0}, game.Cursor{X: 1, Y: 1}, 1, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy, valid := lineDirection(tt.start, tt.end)
			if valid != tt.wantValid {
				t.Fatalf("valid = %v, want %v", valid, tt.wantValid)
			}
			if dx != tt.wantDX || dy != tt.wantDY {
				t.Errorf("dx,dy = (%d,%d), want (%d,%d)", dx, dy, tt.wantDX, tt.wantDY)
			}
		})
	}
}

// --- walkLine (P0) ---

func TestWalkLine(t *testing.T) {
	tests := []struct {
		name    string
		start   game.Cursor
		end     game.Cursor
		wantOK  bool
		wantPos []Position
	}{
		{
			"horizontal line",
			game.Cursor{X: 0, Y: 0},
			game.Cursor{X: 3, Y: 0},
			true,
			[]Position{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
		},
		{
			"vertical line",
			game.Cursor{X: 2, Y: 0},
			game.Cursor{X: 2, Y: 3},
			true,
			[]Position{{2, 0}, {2, 1}, {2, 2}, {2, 3}},
		},
		{
			"diagonal line",
			game.Cursor{X: 0, Y: 0},
			game.Cursor{X: 2, Y: 2},
			true,
			[]Position{{0, 0}, {1, 1}, {2, 2}},
		},
		{
			"reverse diagonal",
			game.Cursor{X: 2, Y: 2},
			game.Cursor{X: 0, Y: 0},
			true,
			[]Position{{2, 2}, {1, 1}, {0, 0}},
		},
		{
			"adjacent cells",
			game.Cursor{X: 0, Y: 0},
			game.Cursor{X: 1, Y: 0},
			true,
			[]Position{{0, 0}, {1, 0}},
		},
		{
			"invalid line",
			game.Cursor{X: 0, Y: 0},
			game.Cursor{X: 1, Y: 2},
			false,
			nil,
		},
		{
			"same point",
			game.Cursor{X: 1, Y: 1},
			game.Cursor{X: 1, Y: 1},
			false,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var visited []Position
			ok := walkLine(tt.start, tt.end, func(x, y int) {
				visited = append(visited, Position{x, y})
			})

			if ok != tt.wantOK {
				t.Fatalf("walkLine returned %v, want %v", ok, tt.wantOK)
			}

			if !tt.wantOK {
				if len(visited) != 0 {
					t.Errorf("fn was called %d times on invalid line", len(visited))
				}
				return
			}

			if len(visited) != len(tt.wantPos) {
				t.Fatalf("visited %d positions, want %d", len(visited), len(tt.wantPos))
			}
			for i := range tt.wantPos {
				if visited[i] != tt.wantPos[i] {
					t.Errorf("visited[%d] = %v, want %v", i, visited[i], tt.wantPos[i])
				}
			}
		})
	}
}

// --- reverseString (P1) ---

func TestReverseString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal", "HELLO", "OLLEH"},
		{"single char", "A", "A"},
		{"empty", "", ""},
		{"palindrome", "ABBA", "ABBA"},
		{"two chars", "AB", "BA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverseString(tt.input); got != tt.want {
				t.Errorf("reverseString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- tryPlaceWord (P2) ---

func TestTryPlaceWord(t *testing.T) {
	t.Run("places word in empty grid", func(t *testing.T) {
		g := createEmptyGrid(10, 10)
		w := tryPlaceWord(g, "HELLO", []Direction{Right, Down, DownRight}, 1000)
		if w == nil {
			t.Fatal("expected word to be placed")
		}

		// Verify letters are in the grid at word positions.
		for i, pos := range w.Positions() {
			got := g.Get(pos.X, pos.Y)
			want := rune(w.Text[i])
			if got != want {
				t.Errorf("grid(%d,%d) = %c, want %c", pos.X, pos.Y, got, want)
			}
		}
	})

	t.Run("respects allowed directions", func(t *testing.T) {
		g := createEmptyGrid(10, 10)
		w := tryPlaceWord(g, "HELLO", []Direction{Right}, 1000)
		if w == nil {
			t.Fatal("expected word to be placed")
		}
		if w.Direction != Right {
			t.Errorf("direction = %d, want Right(%d)", w.Direction, Right)
		}
	})

	t.Run("allows letter overlap", func(t *testing.T) {
		g := createEmptyGrid(1, 5)
		g.Set(0, 0, 'H')
		w := tryPlaceWord(g, "HELLO", []Direction{Right}, 1000)
		if w == nil {
			t.Fatal("expected word to be placed with matching overlap")
		}
	})

	t.Run("rejects conflicting overlap", func(t *testing.T) {
		g := createEmptyGrid(1, 5)
		for x := range 5 {
			g.Set(x, 0, 'X')
		}
		w := tryPlaceWord(g, "HELLO", []Direction{Right}, 100)
		if w != nil {
			t.Error("expected nil for conflicting overlap")
		}
	})

	t.Run("returns nil when impossible", func(t *testing.T) {
		g := createEmptyGrid(3, 3)
		w := tryPlaceWord(g, "TOOLONGWORD", []Direction{Right, Down}, 100)
		if w != nil {
			t.Error("expected nil for word that cannot fit")
		}
	})

	t.Run("uppercases text", func(t *testing.T) {
		g := createEmptyGrid(10, 10)
		w := tryPlaceWord(g, "hello", []Direction{Right}, 1000)
		if w == nil {
			t.Fatal("expected word to be placed")
		}
		if w.Text != "HELLO" {
			t.Errorf("Text = %q, want %q", w.Text, "HELLO")
		}
	})
}

// --- selectWords (P2) ---

func TestSelectWords(t *testing.T) {
	t.Run("respects count", func(t *testing.T) {
		words := selectWords(5, 4, 6)
		if len(words) > 5 {
			t.Errorf("len(selectWords) = %d, want <= 5", len(words))
		}
	})

	t.Run("respects length filter", func(t *testing.T) {
		words := selectWords(20, 4, 6)
		for _, w := range words {
			if len(w) < 4 || len(w) > 6 {
				t.Errorf("word %q has length %d, want 4-6", w, len(w))
			}
		}
	})

	t.Run("clamps when too few", func(t *testing.T) {
		words := selectWords(99999, 4, 4)
		// Should return all 4-letter words, not 99999.
		if len(words) > 99999 {
			t.Error("returned more than requested")
		}
		if len(words) == 0 {
			t.Error("expected some words")
		}
	})
}

// --- fillEmptyCells (P2) ---

func TestFillEmptyCells(t *testing.T) {
	t.Run("fills all spaces", func(t *testing.T) {
		g := createEmptyGrid(5, 5)
		fillEmptyCells(g)
		for y := range g {
			for x := range g[y] {
				if g[y][x] == ' ' {
					t.Errorf("cell (%d,%d) is still a space", x, y)
				}
			}
		}
	})

	t.Run("preserves existing letters", func(t *testing.T) {
		g := createEmptyGrid(5, 5)
		g.Set(0, 0, 'Z')
		fillEmptyCells(g)
		if got := g.Get(0, 0); got != 'Z' {
			t.Errorf("cell (0,0) = %c, want Z", got)
		}
	})

	t.Run("only uppercase letters", func(t *testing.T) {
		g := createEmptyGrid(5, 5)
		fillEmptyCells(g)
		for y := range g {
			for x := range g[y] {
				c := g[y][x]
				if c < 'A' || c > 'Z' {
					t.Errorf("cell (%d,%d) = %c, not in A-Z", x, y, c)
				}
			}
		}
	})
}

// --- GenerateWordSearch (P2) ---

func TestGenerateWordSearch(t *testing.T) {
	t.Run("returns populated grid", func(t *testing.T) {
		g, _ := GenerateWordSearch(15, 15, 10, 4, 7, []Direction{Right, Down, DownRight})
		for y := range g {
			for x := range g[y] {
				if g[y][x] == ' ' {
					t.Errorf("cell (%d,%d) is empty", x, y)
				}
			}
		}
	})

	t.Run("placed words match grid", func(t *testing.T) {
		g, words := GenerateWordSearch(15, 15, 10, 4, 7, []Direction{Right, Down, DownRight})
		for _, w := range words {
			for i, pos := range w.Positions() {
				got := g.Get(pos.X, pos.Y)
				want := rune(w.Text[i])
				if got != want {
					t.Errorf("word %q pos %d: grid(%d,%d) = %c, want %c", w.Text, i, pos.X, pos.Y, got, want)
				}
			}
		}
	})

	t.Run("word count within requested", func(t *testing.T) {
		_, words := GenerateWordSearch(15, 15, 10, 4, 7, []Direction{Right, Down, DownRight})
		if len(words) > 10 {
			t.Errorf("len(words) = %d, want <= 10", len(words))
		}
	})
}

// --- Save/Load round-trip (P1) ---

func TestSaveLoadRoundTrip(t *testing.T) {
	words := []Word{
		{Text: "HELLO", Start: Position{0, 0}, End: Position{4, 0}, Direction: Right, Found: true},
		{Text: "WORLD", Start: Position{0, 1}, End: Position{4, 1}, Direction: Right, Found: false},
	}

	m := Model{
		width:          5,
		height:         5,
		grid:           newGrid(state("HELLO\nWORLD\nABCDE\nFGHIJ\nKLMNO")),
		words:          words,
		cursor:         game.Cursor{X: 3, Y: 2},
		selection:      startSelected,
		selectionStart: game.Cursor{X: 1, Y: 1},
		keys:           DefaultKeyMap,
		solved:         false,
		foundCells:     buildFoundCells(5, 5, words),
	}

	data, err := m.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}

	// Dimensions.
	if got.width != m.width || got.height != m.height {
		t.Errorf("dimensions = %dx%d, want %dx%d", got.width, got.height, m.width, m.height)
	}

	// Grid content.
	if got.grid.String() != m.grid.String() {
		t.Errorf("grid mismatch")
	}

	// Cursor.
	if got.cursor.X != 3 || got.cursor.Y != 2 {
		t.Errorf("cursor = (%d,%d), want (3,2)", got.cursor.X, got.cursor.Y)
	}

	// Selection state.
	if got.selection != startSelected {
		t.Errorf("selection = %d, want %d", got.selection, startSelected)
	}
	if got.selectionStart.X != 1 || got.selectionStart.Y != 1 {
		t.Errorf("selectionStart = (%d,%d), want (1,1)", got.selectionStart.X, got.selectionStart.Y)
	}

	// Solved.
	if got.solved != false {
		t.Errorf("solved = %v, want false", got.solved)
	}

	// Words and found status.
	if len(got.words) != 2 {
		t.Fatalf("word count = %d, want 2", len(got.words))
	}
	if !got.words[0].Found {
		t.Error("words[0].Found = false, want true")
	}
	if got.words[1].Found {
		t.Error("words[1].Found = true, want false")
	}
}

func TestImportModel_InvalidJSON(t *testing.T) {
	_, err := ImportModel([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestImportModel_EmptyData(t *testing.T) {
	data, _ := json.Marshal(Save{})
	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}
	// Should produce a model with zero dimensions and empty grid.
	if got.width != 0 || got.height != 0 {
		t.Errorf("dimensions = %dx%d, want 0x0", got.width, got.height)
	}
}
