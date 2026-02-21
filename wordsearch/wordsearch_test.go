package wordsearch

import (
	"encoding/json"
	"math/rand/v2"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
)

func testRNG(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed+1))
}

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

func TestTryPlaceWordSeededWithFallback(t *testing.T) {
	g := createEmptyGrid(1, 5)
	rng := testRNG(123)

	word, stats := tryPlaceWordSeededWithFallback(g, "HELLO", []Direction{Right}, 0, rng)
	if word == nil {
		t.Fatal("expected fallback to place word")
	}
	if !stats.UsedFallback {
		t.Fatal("expected UsedFallback to be true")
	}
	if stats.RandomAttempts != 0 {
		t.Fatalf("RandomAttempts = %d, want 0", stats.RandomAttempts)
	}
	if stats.FallbackAttempts == 0 {
		t.Fatal("expected fallback attempts to be > 0")
	}
	if word.Start != (Position{X: 0, Y: 0}) || word.End != (Position{X: 4, Y: 0}) {
		t.Fatalf("unexpected placement: start=%v end=%v", word.Start, word.End)
	}
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

	t.Run("no duplicates", func(t *testing.T) {
		rng := testRNG(42)
		words := selectWordsSeeded(500, 3, 10, rng)
		seen := make(map[string]struct{}, len(words))

		for _, w := range words {
			if _, ok := seen[w]; ok {
				t.Fatalf("duplicate word selected: %q", w)
			}
			seen[w] = struct{}{}
		}
	})
}

func TestOrderWordsForPlacement(t *testing.T) {
	words := []string{"AA", "BBBB", "CCC", "DDDD", "E"}
	orderWordsForPlacement(words)

	want := []string{"BBBB", "DDDD", "CCC", "AA", "E"}
	for i := range want {
		if words[i] != want[i] {
			t.Fatalf("words[%d]=%q, want %q", i, words[i], want[i])
		}
	}
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

// --- generation integrity regression (P0) ---

func TestGenerateWordSearchSeeded_WordCountRegression(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		wordCount   int
		minWordLen  int
		maxWordLen  int
		allowedDirs []Direction
		minPlaced   int
	}{
		{
			name:        "easy_10x10",
			width:       10,
			height:      10,
			wordCount:   6,
			minWordLen:  3,
			maxWordLen:  5,
			allowedDirs: []Direction{Right, Down, DownRight},
			minPlaced:   6,
		},
		{
			name:        "medium_15x15",
			width:       15,
			height:      15,
			wordCount:   10,
			minWordLen:  4,
			maxWordLen:  7,
			allowedDirs: []Direction{Right, Down, DownRight, DownLeft, Left, Up},
			minPlaced:   10,
		},
		{
			name:        "hard_20x20",
			width:       20,
			height:      20,
			wordCount:   15,
			minWordLen:  5,
			maxWordLen:  10,
			allowedDirs: []Direction{Right, Down, DownRight, DownLeft, Left, Up, UpRight, UpLeft},
			minPlaced:   13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for seed := range 128 {
				rng := testRNG(uint64(seed + 1))
				_, words := GenerateWordSearchSeeded(
					tt.width,
					tt.height,
					tt.wordCount,
					tt.minWordLen,
					tt.maxWordLen,
					tt.allowedDirs,
					rng,
				)

				if len(words) > tt.wordCount {
					t.Fatalf("seed %d: len(words)=%d, want <= %d", seed, len(words), tt.wordCount)
				}
				if len(words) < tt.minPlaced {
					t.Fatalf("seed %d: len(words)=%d, want >= %d", seed, len(words), tt.minPlaced)
				}
			}
		})
	}
}

func TestGenerateWordSearchSeeded_PlacementValidityRegression(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		wordCount   int
		minWordLen  int
		maxWordLen  int
		allowedDirs []Direction
	}{
		{
			name:        "easy_10x10",
			width:       10,
			height:      10,
			wordCount:   6,
			minWordLen:  3,
			maxWordLen:  5,
			allowedDirs: []Direction{Right, Down, DownRight},
		},
		{
			name:        "medium_15x15",
			width:       15,
			height:      15,
			wordCount:   10,
			minWordLen:  4,
			maxWordLen:  7,
			allowedDirs: []Direction{Right, Down, DownRight, DownLeft, Left, Up},
		},
		{
			name:        "hard_20x20",
			width:       20,
			height:      20,
			wordCount:   15,
			minWordLen:  5,
			maxWordLen:  10,
			allowedDirs: []Direction{Right, Down, DownRight, DownLeft, Left, Up, UpRight, UpLeft},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for seed := range 64 {
				rng := testRNG(uint64(seed + 1))
				g, words := GenerateWordSearchSeeded(
					tt.width,
					tt.height,
					tt.wordCount,
					tt.minWordLen,
					tt.maxWordLen,
					tt.allowedDirs,
					rng,
				)
				assertWordPlacementsValid(t, g, words)
			}
		})
	}
}

func assertWordPlacementsValid(t *testing.T, g grid, words []Word) {
	t.Helper()

	type cell struct {
		x int
		y int
	}

	occupied := map[cell]rune{}

	for wi, w := range words {
		positions := w.Positions()
		if len(positions) != len(w.Text) {
			t.Fatalf("word[%d]=%q positions=%d, want %d", wi, w.Text, len(positions), len(w.Text))
		}
		if positions[0] != w.Start {
			t.Fatalf("word[%d]=%q start=%v, want %v", wi, w.Text, w.Start, positions[0])
		}
		if positions[len(positions)-1] != w.End {
			t.Fatalf("word[%d]=%q end=%v, want %v", wi, w.Text, w.End, positions[len(positions)-1])
		}

		for i, pos := range positions {
			if pos.X < 0 || pos.X >= len(g[0]) || pos.Y < 0 || pos.Y >= len(g) {
				t.Fatalf("word[%d]=%q pos[%d]=(%d,%d) out of bounds", wi, w.Text, i, pos.X, pos.Y)
			}

			want := rune(w.Text[i])
			if got := g.Get(pos.X, pos.Y); got != want {
				t.Fatalf(
					"word[%d]=%q pos[%d]=(%d,%d) grid=%q, want %q",
					wi,
					w.Text,
					i,
					pos.X,
					pos.Y,
					got,
					want,
				)
			}

			key := cell{x: pos.X, y: pos.Y}
			if prev, ok := occupied[key]; ok && prev != want {
				t.Fatalf(
					"conflicting overlap at (%d,%d): had %q, word[%d]=%q wants %q",
					pos.X,
					pos.Y,
					prev,
					wi,
					w.Text,
					want,
				)
			}
			occupied[key] = want
		}
	}
}

func TestGenerateWordSearchSeeded_PlacementStats(t *testing.T) {
	rng := testRNG(99)
	_, words, stats := generateWordSearchSeededWithStats(
		20,
		20,
		15,
		5,
		10,
		[]Direction{Right, Down, DownRight, DownLeft, Left, Up, UpRight, UpLeft},
		rng,
	)

	if stats.TargetWords != 15 {
		t.Fatalf("TargetWords = %d, want 15", stats.TargetWords)
	}
	if stats.PlacedWords != len(words) {
		t.Fatalf("PlacedWords = %d, want %d", stats.PlacedWords, len(words))
	}
	if stats.PlacedWords+stats.FailedWords != stats.TargetWords {
		t.Fatalf(
			"placed+failed = %d, want %d",
			stats.PlacedWords+stats.FailedWords,
			stats.TargetWords,
		)
	}
	if stats.RandomAttempts < stats.TargetWords {
		t.Fatalf("RandomAttempts = %d, want >= %d", stats.RandomAttempts, stats.TargetWords)
	}
	if stats.FallbackPlaced > stats.FallbackUsed {
		t.Fatalf("FallbackPlaced = %d, want <= FallbackUsed = %d", stats.FallbackPlaced, stats.FallbackUsed)
	}
	if got := stats.SuccessRate(); got <= 0 || got > 1 {
		t.Fatalf("SuccessRate = %.4f, want (0,1]", got)
	}
	if got := stats.AttemptsPerWord(); got <= 0 {
		t.Fatalf("AttemptsPerWord = %.4f, want > 0", got)
	}
}

// --- screenToGrid (P1) ---

func TestScreenToGrid(t *testing.T) {
	m := &Model{
		width:      10,
		height:     10,
		grid:       createEmptyGrid(10, 10),
		words:      nil,
		keys:       DefaultKeyMap,
		modeTitle:  "Test",
		foundCells: buildFoundCells(10, 10, nil),
		termWidth:  120,
		termHeight: 40,
	}

	ox, oy := m.gridOrigin()

	tests := []struct {
		name    string
		screenX int
		screenY int
		wantCol int
		wantRow int
		wantOk  bool
	}{
		{"origin cell", ox, oy, 0, 0, true},
		{"cell (1,0)", ox + cellWidth, oy, 1, 0, true},
		{"cell (0,1)", ox, oy + 1, 0, 1, true},
		{"cell (2,3)", ox + 2*cellWidth, oy + 3, 2, 3, true},
		{"cell (9,9) last", ox + 9*cellWidth, oy + 9, 9, 9, true},
		{"outside left", ox - 1, oy, 0, 0, false},
		{"outside top", ox, oy - 1, 0, 0, false},
		{"outside right", ox + 10*cellWidth, oy, 0, 0, false},
		{"outside bottom", ox, oy + 10, 0, 0, false},
		{"far outside", 0, 0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, row, ok := m.screenToGrid(tt.screenX, tt.screenY)
			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && (col != tt.wantCol || row != tt.wantRow) {
				t.Errorf("screenToGrid(%d, %d) = (%d, %d), want (%d, %d)",
					tt.screenX, tt.screenY, col, row, tt.wantCol, tt.wantRow)
			}
		})
	}
}

func TestHelpToggleInvalidatesOriginCache(t *testing.T) {
	m := Model{
		width:       10,
		height:      10,
		grid:        createEmptyGrid(10, 10),
		keys:        DefaultKeyMap,
		modeTitle:   "Test",
		foundCells:  buildFoundCells(10, 10, nil),
		originX:     9,
		originY:     13,
		originValid: true,
	}

	next, _ := m.Update(game.HelpToggleMsg{Show: true})
	got := next.(Model)

	if !got.showFullHelp {
		t.Fatal("expected showFullHelp to be true")
	}
	if got.originValid {
		t.Fatal("expected origin cache to be invalidated")
	}
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
