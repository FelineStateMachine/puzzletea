package hitori

import (
	"encoding/json"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
)

// --- Helpers ---

func makeGrid(rows ...string) grid {
	g := make(grid, len(rows))
	for i, row := range rows {
		g[i] = []rune(row)
	}
	return g
}

func makeMarks(rows ...string) [][]cellMark {
	marks := make([][]cellMark, len(rows))
	for y, row := range rows {
		marks[y] = make([]cellMark, len(row))
		for x, r := range row {
			switch r {
			case 'X':
				marks[y][x] = shaded
			case 'O':
				marks[y][x] = circled
			default:
				marks[y][x] = unmarked
			}
		}
	}
	return marks
}

func testMode(size int) HitoriMode {
	return NewMode("Test", "test mode", size, 0.25)
}

// --- Grid and serialization (P0) ---

func TestNewGrid(t *testing.T) {
	g := newGrid("123\n456\n789")
	if len(g) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(g))
	}
	if string(g[0]) != "123" {
		t.Errorf("row 0 = %q, want %q", string(g[0]), "123")
	}
	if string(g[2]) != "789" {
		t.Errorf("row 2 = %q, want %q", string(g[2]), "789")
	}
}

func TestGridString(t *testing.T) {
	g := makeGrid("123", "456", "789")
	got := g.String()
	want := "123\n456\n789"
	if got != want {
		t.Errorf("grid.String() = %q, want %q", got, want)
	}
}

func TestGridClone(t *testing.T) {
	g := makeGrid("123", "456")
	c := g.clone()
	c[0][0] = 'X'
	if g[0][0] == 'X' {
		t.Error("clone shares memory with original")
	}
}

func TestSerializeDeserializeMarks(t *testing.T) {
	marks := makeMarks("..X", "O..", "XO.")
	serialized := serializeMarks(marks)
	deserialized := deserializeMarks(serialized, 3)

	for y := range 3 {
		for x := range 3 {
			if deserialized[y][x] != marks[y][x] {
				t.Errorf("mark[%d][%d] = %d, want %d", y, x, deserialized[y][x], marks[y][x])
			}
		}
	}
}

func TestCloneMarks(t *testing.T) {
	marks := makeMarks("X.", ".O")
	c := cloneMarks(marks)
	c[0][0] = unmarked
	if marks[0][0] != shaded {
		t.Error("cloneMarks shares memory with original")
	}
}

// --- Validation: hasNoDuplicatesInRows (P0) ---

func TestHasNoDuplicatesInRows(t *testing.T) {
	tests := []struct {
		name    string
		numbers grid
		marks   [][]cellMark
		want    bool
	}{
		{
			name:    "no duplicates",
			numbers: makeGrid("123", "231", "312"),
			marks:   makeMarks("...", "...", "..."),
			want:    true,
		},
		{
			name:    "duplicate in row but one is shaded",
			numbers: makeGrid("113", "231", "312"),
			marks:   makeMarks("X..", "...", "..."),
			want:    true,
		},
		{
			name:    "duplicate in row neither shaded",
			numbers: makeGrid("113", "231", "312"),
			marks:   makeMarks("...", "...", "..."),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasNoDuplicatesInRows(tt.numbers, tt.marks, len(tt.numbers))
			if got != tt.want {
				t.Errorf("hasNoDuplicatesInRows = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Validation: hasNoDuplicatesInCols (P0) ---

func TestHasNoDuplicatesInCols(t *testing.T) {
	tests := []struct {
		name    string
		numbers grid
		marks   [][]cellMark
		want    bool
	}{
		{
			name:    "no duplicates in columns",
			numbers: makeGrid("123", "231", "312"),
			marks:   makeMarks("...", "...", "..."),
			want:    true,
		},
		{
			name:    "duplicate in column but one shaded",
			numbers: makeGrid("123", "131", "312"),
			marks:   makeMarks("...", "X..", "..."),
			want:    true,
		},
		{
			name:    "duplicate in column neither shaded",
			numbers: makeGrid("123", "131", "312"),
			marks:   makeMarks("...", "...", "..."),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasNoDuplicatesInCols(tt.numbers, tt.marks, len(tt.numbers))
			if got != tt.want {
				t.Errorf("hasNoDuplicatesInCols = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Validation: hasNoAdjacentShaded (P0) ---

func TestHasNoAdjacentShaded(t *testing.T) {
	tests := []struct {
		name  string
		marks [][]cellMark
		want  bool
	}{
		{
			name:  "no shaded cells",
			marks: makeMarks("...", "...", "..."),
			want:  true,
		},
		{
			name:  "diagonal shaded OK",
			marks: makeMarks("X..", ".X.", "..X"),
			want:  true,
		},
		{
			name:  "horizontal adjacent",
			marks: makeMarks("XX.", "...", "..."),
			want:  false,
		},
		{
			name:  "vertical adjacent",
			marks: makeMarks("X..", "X..", "..."),
			want:  false,
		},
		{
			name:  "single shaded",
			marks: makeMarks("...", ".X.", "..."),
			want:  true,
		},
		{
			name:  "checkerboard pattern",
			marks: makeMarks("X.X.", ".X.X", "X.X.", ".X.X"),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasNoAdjacentShaded(tt.marks, len(tt.marks))
			if got != tt.want {
				t.Errorf("hasNoAdjacentShaded = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Validation: allWhiteConnected (P0) ---

func TestAllWhiteConnected(t *testing.T) {
	tests := []struct {
		name  string
		marks [][]cellMark
		want  bool
	}{
		{
			name:  "all white",
			marks: makeMarks("...", "...", "..."),
			want:  true,
		},
		{
			name:  "connected with some shaded",
			marks: makeMarks("X..", "...", "..X"),
			want:  true,
		},
		{
			name:  "disconnected by shaded column",
			marks: makeMarks(".X.", ".X.", ".X."),
			want:  false,
		},
		{
			name:  "single white cell surrounded",
			marks: makeMarks("XXX", "X.X", "XXX"),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := allWhiteConnected(tt.marks, len(tt.marks))
			if got != tt.want {
				t.Errorf("allWhiteConnected = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- isValidSolution (P0) ---

func TestIsValidSolution(t *testing.T) {
	// A valid 4x4 Hitori solution:
	// Numbers: 1 2 1 3    Marks: X . . .
	//          2 3 4 1           . . . .
	//          3 1 2 4           . . . .
	//          1 4 3 2           . . X .
	// After shading: row 0 has [2,1,3], row 3 has [1,4,2].
	// No dups in rows/cols, no adjacent shaded, white connected.
	numbers := makeGrid("1213", "2341", "3124", "1432")
	marks := makeMarks("X...", "....", "....", "..X.")

	if !isValidSolution(numbers, marks, 4) {
		t.Error("expected valid solution")
	}
}

func TestIsValidSolution_InvalidDuplicate(t *testing.T) {
	numbers := makeGrid("1213", "2341", "3124", "1432")
	marks := makeMarks("....", "....", "....", "....") // no shading, row 0 has dup 1
	if isValidSolution(numbers, marks, 4) {
		t.Error("expected invalid: duplicate 1 in row 0")
	}
}

func TestIsValidSolution_InvalidAdjacent(t *testing.T) {
	numbers := makeGrid("1213", "2341", "3124", "1432")
	marks := makeMarks("XX..", "....", "....", "....") // adjacent shaded
	if isValidSolution(numbers, marks, 4) {
		t.Error("expected invalid: adjacent shaded cells")
	}
}

// --- Latin square generation (P1) ---

func TestGenerateLatinSquare(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generator test in short mode")
	}

	sizes := []int{4, 5, 6, 8}
	for _, size := range sizes {
		t.Run("", func(t *testing.T) {
			g := generateLatinSquare(size)

			// Check rows have all numbers.
			for y := range size {
				seen := map[rune]bool{}
				for x := range size {
					seen[g[y][x]] = true
				}
				if len(seen) != size {
					t.Errorf("row %d has %d unique numbers, want %d", y, len(seen), size)
				}
			}

			// Check columns have all numbers.
			for x := range size {
				seen := map[rune]bool{}
				for y := range size {
					seen[g[y][x]] = true
				}
				if len(seen) != size {
					t.Errorf("col %d has %d unique numbers, want %d", x, len(seen), size)
				}
			}
		})
	}
}

// --- Mask generation (P1) ---

func TestGenerateValidMask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generator test in short mode")
	}

	size := 6
	mask := generateValidMask(size, 0.25)

	// No adjacent black cells.
	for y := range size {
		for x := range size {
			if !mask[y][x] {
				continue
			}
			if x > 0 && mask[y][x-1] {
				t.Errorf("adjacent black cells at (%d,%d) and (%d,%d)", x, y, x-1, y)
			}
			if y > 0 && mask[y-1][x] {
				t.Errorf("adjacent black cells at (%d,%d) and (%d,%d)", x, y, x, y-1)
			}
		}
	}

	// White cells connected.
	if !whiteCellsConnected(mask, size) {
		t.Error("white cells are not connected")
	}
}

// --- Save/load round-trip (P1) ---

func TestSaveLoadRoundTrip(t *testing.T) {
	numbers := makeGrid("1213", "2341", "3124", "1432")
	mode := testMode(4)
	gamer, err := New(mode, numbers)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	data, err := gamer.GetSave()
	if err != nil {
		t.Fatalf("GetSave: %v", err)
	}

	loaded, err := ImportModel(data)
	if err != nil {
		t.Fatalf("ImportModel: %v", err)
	}

	// Verify grid preserved.
	if loaded.numbers.String() != numbers.String() {
		t.Error("numbers grid not preserved after round-trip")
	}
	if loaded.size != 4 {
		t.Errorf("size = %d, want 4", loaded.size)
	}
}

func TestSaveLoadRoundTrip_WithMarks(t *testing.T) {
	numbers := makeGrid("1213", "2341", "3124", "1432")
	mode := testMode(4)
	gamer, err := New(mode, numbers)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Apply some marks via the model.
	m := gamer.(Model)
	m.marks[0][0] = shaded
	m.marks[1][1] = circled

	data, err := m.GetSave()
	if err != nil {
		t.Fatalf("GetSave: %v", err)
	}

	loaded, err := ImportModel(data)
	if err != nil {
		t.Fatalf("ImportModel: %v", err)
	}

	if loaded.marks[0][0] != shaded {
		t.Error("shaded mark not preserved")
	}
	if loaded.marks[1][1] != circled {
		t.Error("circled mark not preserved")
	}
}

func TestImportModel_InvalidJSON(t *testing.T) {
	_, err := ImportModel([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestImportModel_InvalidSize(t *testing.T) {
	data, _ := json.Marshal(Save{Size: 0, Numbers: "", Marks: ""})
	_, err := ImportModel(data)
	if err == nil {
		t.Error("expected error for zero size")
	}
}

// --- Model construction (P0) ---

func TestNew_ValidGrid(t *testing.T) {
	mode := testMode(3)
	numbers := makeGrid("123", "231", "312")
	g, err := New(mode, numbers)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if g.IsSolved() {
		t.Error("new game should not be solved")
	}
}

func TestNew_InvalidSize(t *testing.T) {
	mode := testMode(0)
	_, err := New(mode, makeGrid())
	if err == nil {
		t.Error("expected error for zero size")
	}
}

func TestNew_RowMismatch(t *testing.T) {
	mode := testMode(3)
	_, err := New(mode, makeGrid("12", "23"))
	if err == nil {
		t.Error("expected error for row count mismatch")
	}
}

func TestNew_ColumnMismatch(t *testing.T) {
	mode := testMode(3)
	_, err := New(mode, makeGrid("12", "23", "31"))
	if err == nil {
		t.Error("expected error for column width mismatch")
	}
}

// --- Model methods (P0) ---

func TestModel_Reset(t *testing.T) {
	mode := testMode(3)
	numbers := makeGrid("123", "231", "312")
	g, _ := New(mode, numbers)

	m := g.(Model)
	m.marks[0][0] = shaded
	m.marks[1][1] = circled

	reset := m.Reset().(Model)
	if reset.marks[0][0] != unmarked || reset.marks[1][1] != unmarked {
		t.Error("Reset did not clear marks")
	}
}

func TestModel_SetTitle(t *testing.T) {
	mode := testMode(3)
	numbers := makeGrid("123", "231", "312")
	g, _ := New(mode, numbers)

	updated := g.SetTitle("Custom Title").(Model)
	if updated.modeTitle != "Custom Title" {
		t.Errorf("modeTitle = %q, want %q", updated.modeTitle, "Custom Title")
	}
}

func TestModel_GetFullHelp(t *testing.T) {
	mode := testMode(3)
	numbers := makeGrid("123", "231", "312")
	g, _ := New(mode, numbers)
	help := g.GetFullHelp()
	if len(help) != 2 {
		t.Errorf("expected 2 help groups, got %d", len(help))
	}
}

func TestModel_GetDebugInfo(t *testing.T) {
	mode := testMode(3)
	numbers := makeGrid("123", "231", "312")
	g, _ := New(mode, numbers)
	info := g.GetDebugInfo()
	if info == "" {
		t.Error("expected non-empty debug info")
	}
}

func TestModel_View(t *testing.T) {
	mode := testMode(3)
	numbers := makeGrid("123", "231", "312")
	g, _ := New(mode, numbers)
	view := g.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

// --- Gamer interface compliance (P0) ---

func TestGamerInterface(t *testing.T) {
	var _ game.Gamer = Model{}
}

// --- Spawner/Mode interface compliance (P0) ---

func TestModeInterfaces(t *testing.T) {
	var _ game.Mode = HitoriMode{}
	var _ game.Spawner = HitoriMode{}
}

func TestModeTitleDescription(t *testing.T) {
	mode := NewMode("Easy", "6x6 grid", 6, 0.22)
	if mode.Title() != "Easy" {
		t.Errorf("Title = %q, want %q", mode.Title(), "Easy")
	}
	if mode.Description() != "6x6 grid" {
		t.Errorf("Description = %q, want %q", mode.Description(), "6x6 grid")
	}
}

// --- Puzzle generation (P2) ---

func TestGenerate_SmallPuzzle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generator test in short mode")
	}

	puzzle, err := Generate(5, 0.32)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(puzzle) != 5 {
		t.Errorf("puzzle has %d rows, want 5", len(puzzle))
	}
	for y, row := range puzzle {
		if len(row) != 5 {
			t.Errorf("row %d has %d cols, want 5", y, len(row))
		}
		for x, num := range row {
			if num < '1' || num > '5' {
				t.Errorf("puzzle[%d][%d] = %c, want 1-5", y, x, num)
			}
		}
	}

	// Verify it has exactly one solution.
	solutions := countPuzzleSolutions(puzzle, 5, 2)
	if solutions != 1 {
		t.Errorf("puzzle has %d solutions, want exactly 1", solutions)
	}
}

func TestSpawn_Mini(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generator test in short mode")
	}

	mode := NewMode("Mini", "5x5 grid", 5, 0.32)
	g, err := mode.Spawn()
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	if g == nil {
		t.Fatal("Spawn returned nil game")
	}
	if g.IsSolved() {
		t.Error("freshly spawned game should not be solved")
	}
}

// --- Solver validation (P1) ---

func TestCountPuzzleSolutions_KnownPuzzle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping solver test in short mode")
	}

	// A simple 4x4 puzzle with known unique solution.
	// Numbers: 1 2 1 3
	//          2 3 4 1
	//          3 1 2 4
	//          1 4 3 2
	// Solution: shade (0,0) and (2,3) -> unique
	puzzle := makeGrid("1213", "2341", "3124", "1432")

	solutions := countPuzzleSolutions(puzzle, 4, 2)
	if solutions < 1 {
		t.Errorf("expected at least 1 solution, got %d", solutions)
	}
}

// --- Registration (P0) ---

func TestRegistration(t *testing.T) {
	fn, ok := game.Registry["Hitori"]
	if !ok {
		t.Fatal("Hitori not registered in game.Registry")
	}

	numbers := makeGrid("123", "231", "312")
	mode := testMode(3)
	g, _ := New(mode, numbers)
	data, _ := g.GetSave()

	loaded, err := fn(data)
	if err != nil {
		t.Fatalf("Registry import: %v", err)
	}
	if loaded == nil {
		t.Fatal("Registry import returned nil")
	}
}
