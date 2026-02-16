package shikaku

import (
	"encoding/json"
	"math/rand/v2"
	"testing"
)

// --- helpers ---

// simplePuzzle builds a 4x4 puzzle with 4 clues:
//
//	[4] .  .  .
//	 .  .  .  .
//	 .  . [2] .
//	 . [6] .  .
//
// Solution: 2x2 rect for clue 0, 1x2 rect for clue 2, 2x3 rect for clue 1, ...
func simplePuzzle() *Puzzle {
	return &Puzzle{
		Width:  4,
		Height: 4,
		Clues: []Clue{
			{ID: 0, X: 0, Y: 0, Value: 4},
			{ID: 1, X: 1, Y: 3, Value: 6},
			{ID: 2, X: 2, Y: 2, Value: 2},
			{ID: 3, X: 3, Y: 0, Value: 4},
		},
	}
}

// tinyPuzzle builds a 2x2 puzzle with 2 clues for minimal testing.
func tinyPuzzle() *Puzzle {
	return &Puzzle{
		Width:  2,
		Height: 2,
		Clues: []Clue{
			{ID: 0, X: 0, Y: 0, Value: 2},
			{ID: 1, X: 1, Y: 1, Value: 2},
		},
	}
}

// --- FindClueAt ---

func TestFindClueAt(t *testing.T) {
	t.Run("clue exists", func(t *testing.T) {
		p := simplePuzzle()
		c := p.FindClueAt(0, 0)
		if c == nil {
			t.Fatal("expected clue at (0,0)")
		}
		if c.ID != 0 || c.Value != 4 {
			t.Errorf("clue = %+v, want ID=0 Value=4", c)
		}
	})

	t.Run("no clue", func(t *testing.T) {
		p := simplePuzzle()
		c := p.FindClueAt(1, 1)
		if c != nil {
			t.Errorf("expected nil at (1,1), got %+v", c)
		}
	})

	t.Run("lazy index build", func(t *testing.T) {
		p := simplePuzzle()
		if p.clueIndex != nil {
			t.Fatal("clueIndex should be nil before first call")
		}
		_ = p.FindClueAt(0, 0)
		if p.clueIndex == nil {
			t.Error("clueIndex should be built after first call")
		}
	})
}

// --- FindClueByID ---

func TestFindClueByID(t *testing.T) {
	t.Run("found by sequential ID", func(t *testing.T) {
		p := simplePuzzle()
		c := p.FindClueByID(2)
		if c == nil {
			t.Fatal("expected clue with ID=2")
		}
		if c.X != 2 || c.Y != 2 {
			t.Errorf("clue 2 at (%d,%d), want (2,2)", c.X, c.Y)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		p := simplePuzzle()
		if c := p.FindClueByID(99); c != nil {
			t.Errorf("expected nil, got %+v", c)
		}
	})
}

// --- Rectangle methods ---

func TestRectangleContains(t *testing.T) {
	r := Rectangle{ClueID: 0, X: 1, Y: 1, W: 2, H: 3}

	if !r.Contains(1, 1) {
		t.Error("should contain top-left corner")
	}
	if !r.Contains(2, 3) {
		t.Error("should contain bottom-right cell")
	}
	if r.Contains(0, 0) {
		t.Error("should not contain (0,0)")
	}
	if r.Contains(3, 1) {
		t.Error("should not contain cell just outside right edge")
	}
}

func TestRectangleArea(t *testing.T) {
	r := Rectangle{W: 3, H: 4}
	if r.Area() != 12 {
		t.Errorf("Area() = %d, want 12", r.Area())
	}
}

// --- CellOwner ---

func TestCellOwner(t *testing.T) {
	p := simplePuzzle()
	p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 2, H: 2})

	t.Run("owned cell", func(t *testing.T) {
		if owner := p.CellOwner(1, 1); owner != 0 {
			t.Errorf("CellOwner(1,1) = %d, want 0", owner)
		}
	})

	t.Run("unowned cell", func(t *testing.T) {
		if owner := p.CellOwner(3, 3); owner != -1 {
			t.Errorf("CellOwner(3,3) = %d, want -1", owner)
		}
	})
}

// --- Overlaps ---

func TestOverlaps(t *testing.T) {
	p := simplePuzzle()
	p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 2, H: 2})

	t.Run("overlapping", func(t *testing.T) {
		r := Rectangle{ClueID: 1, X: 1, Y: 1, W: 2, H: 2}
		if !p.Overlaps(r, 1) {
			t.Error("should overlap with existing rectangle")
		}
	})

	t.Run("non-overlapping", func(t *testing.T) {
		r := Rectangle{ClueID: 1, X: 2, Y: 0, W: 2, H: 2}
		if p.Overlaps(r, 1) {
			t.Error("should not overlap")
		}
	})

	t.Run("exclude self", func(t *testing.T) {
		// Replacing the same clue's rectangle should not count as overlap.
		r := Rectangle{ClueID: 0, X: 0, Y: 0, W: 2, H: 2}
		if p.Overlaps(r, 0) {
			t.Error("should not overlap when excluding self")
		}
	})
}

// --- SetRectangle / RemoveRectangle ---

func TestSetAndRemoveRectangle(t *testing.T) {
	p := simplePuzzle()

	t.Run("set new", func(t *testing.T) {
		p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 2, H: 2})
		if r := p.FindRectangleForClue(0); r == nil {
			t.Fatal("expected rectangle for clue 0")
		}
	})

	t.Run("replace existing", func(t *testing.T) {
		p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 4, H: 1})
		r := p.FindRectangleForClue(0)
		if r == nil {
			t.Fatal("expected rectangle for clue 0")
		}
		if r.W != 4 || r.H != 1 {
			t.Errorf("rectangle = %dx%d, want 4x1", r.W, r.H)
		}
		if len(p.Rectangles) != 1 {
			t.Errorf("len(Rectangles) = %d, want 1", len(p.Rectangles))
		}
	})

	t.Run("remove", func(t *testing.T) {
		p.RemoveRectangle(0)
		if r := p.FindRectangleForClue(0); r != nil {
			t.Error("expected nil after removal")
		}
	})

	t.Run("remove nonexistent", func(t *testing.T) {
		p.RemoveRectangle(99) // must not panic
	})
}

// --- IsSolved ---

func TestIsSolved(t *testing.T) {
	t.Run("solved 2x2", func(t *testing.T) {
		p := tinyPuzzle()
		p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 1, H: 2})
		p.SetRectangle(Rectangle{ClueID: 1, X: 1, Y: 0, W: 1, H: 2})

		if !p.IsSolved() {
			t.Error("puzzle should be solved")
		}
	})

	t.Run("incomplete coverage", func(t *testing.T) {
		p := tinyPuzzle()
		p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 1, H: 2})
		// Only one rectangle placed.

		if p.IsSolved() {
			t.Error("incomplete coverage should not be solved")
		}
	})

	t.Run("wrong area", func(t *testing.T) {
		p := tinyPuzzle()
		// Both clues want area 2 but we give them area 1 rects
		// (which won't cover everything either).
		p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 1, H: 1})
		p.SetRectangle(Rectangle{ClueID: 1, X: 1, Y: 1, W: 1, H: 1})

		if p.IsSolved() {
			t.Error("wrong area should not be solved")
		}
	})

	t.Run("overlapping rectangles", func(t *testing.T) {
		p := tinyPuzzle()
		p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 2, H: 1})
		p.SetRectangle(Rectangle{ClueID: 1, X: 0, Y: 0, W: 2, H: 1}) // same area

		if p.IsSolved() {
			t.Error("overlapping rectangles should not be solved")
		}
	})

	t.Run("no rectangles", func(t *testing.T) {
		p := simplePuzzle()
		if p.IsSolved() {
			t.Error("no rectangles should not be solved")
		}
	})

	t.Run("clue outside its rectangle", func(t *testing.T) {
		p := &Puzzle{
			Width:  2,
			Height: 2,
			Clues: []Clue{
				{ID: 0, X: 0, Y: 0, Value: 2},
				{ID: 1, X: 1, Y: 0, Value: 2},
			},
		}
		// Place clue 0's rect in row 1 (doesn't contain clue 0).
		p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 1, W: 2, H: 1})
		p.SetRectangle(Rectangle{ClueID: 1, X: 0, Y: 0, W: 2, H: 1})

		if p.IsSolved() {
			t.Error("clue outside its rectangle should not be solved")
		}
	})

	t.Run("two clues in same rectangle", func(t *testing.T) {
		p := &Puzzle{
			Width:  4,
			Height: 1,
			Clues: []Clue{
				{ID: 0, X: 0, Y: 0, Value: 4},
				{ID: 1, X: 3, Y: 0, Value: 4},
			},
		}
		// One huge rectangle containing both clues.
		p.SetRectangle(Rectangle{ClueID: 0, X: 0, Y: 0, W: 4, H: 1})
		// No rect for clue 1 — count mismatch.
		if p.IsSolved() {
			t.Error("two clues in same rectangle should not be solved")
		}
	})
}

// --- expansion.rect ---

func TestExpansionRect(t *testing.T) {
	clue := &Clue{ID: 0, X: 3, Y: 3, Value: 6}
	e := expansion{clueID: 0, left: 1, right: 2, up: 0, down: 1}
	r := e.rect(clue)

	if r.X != 2 || r.Y != 3 || r.W != 4 || r.H != 2 {
		t.Errorf("rect = (%d,%d) %dx%d, want (2,3) 4x2", r.X, r.Y, r.W, r.H)
	}
	if r.Area() != 8 {
		t.Errorf("Area() = %d, want 8", r.Area())
	}
}

// --- Save/Load round-trip ---

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Run("full round trip", func(t *testing.T) {
		original := &Model{
			puzzle: Puzzle{
				Width:  5,
				Height: 5,
				Clues: []Clue{
					{ID: 0, X: 0, Y: 0, Value: 4},
					{ID: 1, X: 2, Y: 0, Value: 6},
					{ID: 2, X: 0, Y: 2, Value: 3},
				},
				Rectangles: []Rectangle{
					{ClueID: 0, X: 0, Y: 0, W: 2, H: 2},
					{ClueID: 1, X: 2, Y: 0, W: 3, H: 2},
				},
			},
			keys:      DefaultKeyMap,
			modeTitle: "Test Mode",
		}

		data, err := original.GetSave()
		if err != nil {
			t.Fatal(err)
		}

		got, err := ImportModel(data)
		if err != nil {
			t.Fatal(err)
		}

		if got.puzzle.Width != 5 || got.puzzle.Height != 5 {
			t.Errorf("dimensions = %dx%d, want 5x5", got.puzzle.Width, got.puzzle.Height)
		}
		if len(got.puzzle.Clues) != 3 {
			t.Fatalf("clue count = %d, want 3", len(got.puzzle.Clues))
		}
		if len(got.puzzle.Rectangles) != 2 {
			t.Fatalf("rectangle count = %d, want 2", len(got.puzzle.Rectangles))
		}
		if got.modeTitle != "Test Mode" {
			t.Errorf("modeTitle = %q, want %q", got.modeTitle, "Test Mode")
		}

		for i, c := range original.puzzle.Clues {
			g := got.puzzle.Clues[i]
			if g.ID != c.ID || g.X != c.X || g.Y != c.Y || g.Value != c.Value {
				t.Errorf("clue[%d] = %+v, want %+v", i, g, c)
			}
		}
		for i, r := range original.puzzle.Rectangles {
			g := got.puzzle.Rectangles[i]
			if g.ClueID != r.ClueID || g.X != r.X || g.Y != r.Y || g.W != r.W || g.H != r.H {
				t.Errorf("rectangle[%d] = %+v, want %+v", i, g, r)
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

func TestSaveJSON(t *testing.T) {
	m := Model{
		puzzle: Puzzle{
			Width:  4,
			Height: 4,
			Clues: []Clue{
				{ID: 0, X: 0, Y: 0, Value: 4},
			},
			Rectangles: []Rectangle{
				{ClueID: 0, X: 0, Y: 0, W: 2, H: 2},
			},
		},
		keys:      DefaultKeyMap,
		modeTitle: "Test",
	}

	data, err := m.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"width", "height", "clues", "rectangles", "mode_title"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing key %q in save JSON", key)
		}
	}
}

// --- Generator ---

func TestGeneratePuzzle(t *testing.T) {
	t.Run("produces valid puzzle", func(t *testing.T) {
		puzzle, err := GeneratePuzzle(5, 5, 5)
		if err != nil {
			t.Fatal(err)
		}
		if puzzle.Width != 5 || puzzle.Height != 5 {
			t.Errorf("dimensions = %dx%d, want 5x5", puzzle.Width, puzzle.Height)
		}
		if len(puzzle.Clues) == 0 {
			t.Error("expected at least one clue")
		}
		// Sum of clue values should equal grid area.
		sum := 0
		for _, c := range puzzle.Clues {
			sum += c.Value
		}
		if sum != 25 {
			t.Errorf("sum of clue values = %d, want 25", sum)
		}
	})

	t.Run("no rectangles placed", func(t *testing.T) {
		puzzle, err := GeneratePuzzle(5, 5, 5)
		if err != nil {
			t.Fatal(err)
		}
		if len(puzzle.Rectangles) != 0 {
			t.Errorf("len(Rectangles) = %d, want 0", len(puzzle.Rectangles))
		}
	})

	t.Run("clues within bounds", func(t *testing.T) {
		puzzle, err := GeneratePuzzle(7, 7, 8)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range puzzle.Clues {
			if c.X < 0 || c.X >= 7 || c.Y < 0 || c.Y >= 7 {
				t.Errorf("clue %d at (%d,%d) out of bounds", c.ID, c.X, c.Y)
			}
		}
	})

	t.Run("seeded deterministic", func(t *testing.T) {
		rng1 := rand.New(rand.NewPCG(42, 42))
		rng2 := rand.New(rand.NewPCG(42, 42))

		p1, err := GeneratePuzzleSeeded(5, 5, 5, rng1)
		if err != nil {
			t.Fatal(err)
		}
		p2, err := GeneratePuzzleSeeded(5, 5, 5, rng2)
		if err != nil {
			t.Fatal(err)
		}

		if len(p1.Clues) != len(p2.Clues) {
			t.Fatalf("different clue counts: %d vs %d", len(p1.Clues), len(p2.Clues))
		}
		for i := range p1.Clues {
			if p1.Clues[i] != p2.Clues[i] {
				t.Errorf("clue[%d] differs: %+v vs %+v", i, p1.Clues[i], p2.Clues[i])
			}
		}
	})

	t.Run("larger grid", func(t *testing.T) {
		puzzle, err := GeneratePuzzle(10, 10, 15)
		if err != nil {
			t.Fatal(err)
		}
		sum := 0
		for _, c := range puzzle.Clues {
			sum += c.Value
		}
		if sum != 100 {
			t.Errorf("sum of clue values = %d, want 100", sum)
		}
	})
}

// --- Mode Spawn ---

func TestModeSpawn(t *testing.T) {
	mode := NewMode("Test", "test mode", 5, 5, 5)

	t.Run("Spawn", func(t *testing.T) {
		g, err := mode.Spawn()
		if err != nil {
			t.Fatal(err)
		}
		if g == nil {
			t.Fatal("expected non-nil game")
		}
		if g.IsSolved() {
			t.Error("new game should not be solved")
		}
	})

	t.Run("SpawnSeeded", func(t *testing.T) {
		rng := rand.New(rand.NewPCG(123, 456))
		g, err := mode.SpawnSeeded(rng)
		if err != nil {
			t.Fatal(err)
		}
		if g == nil {
			t.Fatal("expected non-nil game")
		}
	})
}
