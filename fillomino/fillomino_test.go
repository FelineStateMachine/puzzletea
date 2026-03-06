package fillomino

import (
	"image/color"
	"math/rand/v2"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestValidateGridState(t *testing.T) {
	tests := []struct {
		name   string
		grid   grid
		solved bool
	}{
		{
			name: "connected equal-number region is valid",
			grid: grid{
				{2, 2, 1},
				{3, 3, 3},
			},
			solved: true,
		},
		{
			name: "disconnected equal-number cells can coexist",
			grid: grid{
				{2, 2, 1},
				{3, 0, 3},
			},
			solved: false,
		},
		{
			name: "region size mismatch is invalid",
			grid: grid{
				{2, 2, 1},
				{3, 3, 1},
			},
			solved: false,
		},
		{
			name: "incomplete grid is not solved",
			grid: grid{
				{1, 0},
				{0, 1},
			},
			solved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateGridState(tt.grid)
			if got.solved != tt.solved {
				t.Fatalf("solved = %v, want %v", got.solved, tt.solved)
			}
		})
	}
}

func TestCountSolutions(t *testing.T) {
	tests := []struct {
		name     string
		givens   grid
		maxValue int
		want     int
	}{
		{
			name: "unique solution puzzle",
			givens: grid{
				{2, 2, 1},
				{3, 3, 3},
				{2, 2, 1},
			},
			maxValue: 3,
			want:     1,
		},
		{
			name: "ambiguous puzzle",
			givens: grid{
				{0, 0, 0},
				{0, 0, 0},
			},
			maxValue: 3,
			want:     2,
		},
		{
			name: "invalid fixed clues",
			givens: grid{
				{1, 1, 0},
				{0, 0, 0},
			},
			maxValue: 3,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countSolutions(tt.givens, tt.maxValue, 2)
			if got != tt.want {
				t.Fatalf("countSolutions() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGeneratePuzzleSeeded(t *testing.T) {
	mode := Modes[0].(Mode)
	rngA := rand.New(rand.NewPCG(11, 22))
	rngB := rand.New(rand.NewPCG(11, 22))

	a, err := GeneratePuzzleSeeded(mode.Size, mode.Size, mode.MaxRegion, mode.GivenRatio, rngA)
	if err != nil {
		t.Fatal(err)
	}
	b, err := GeneratePuzzleSeeded(mode.Size, mode.Size, mode.MaxRegion, mode.GivenRatio, rngB)
	if err != nil {
		t.Fatal(err)
	}

	if encodeGrid(a.Givens) != encodeGrid(b.Givens) {
		t.Fatalf("seeded givens differ:\n%s\n!=\n%s", encodeGrid(a.Givens), encodeGrid(b.Givens))
	}
	if validateGridState(a.Solution).solved != true {
		t.Fatal("generated solution is not solved")
	}
	if got := countSolutions(a.Givens, mode.MaxRegion, 2); got != 1 {
		t.Fatalf("generated puzzle solutions = %d, want 1", got)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	puzzle := Puzzle{
		Width:  3,
		Height: 3,
		Givens: grid{{2, 0, 1}, {0, 3, 0}, {2, 0, 1}},
	}
	g, err := New(NewMode("Test", "desc", 3, 3, 1), puzzle)
	if err != nil {
		t.Fatal(err)
	}
	model := g.(Model)
	model.cursor = pointCursor(1, 0)
	model.setCell(2)

	save, err := model.GetSave()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := ImportModel(save)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := loaded.grid[0][1], 2; got != want {
		t.Fatalf("loaded grid[0][1] = %d, want %d", got, want)
	}
	if !loaded.provided[0][0] || !loaded.provided[0][2] || !loaded.provided[1][1] || !loaded.provided[2][2] {
		t.Fatal("expected provided mask to persist")
	}
	if got, want := loaded.maxCellValue, 3; got != want {
		t.Fatalf("loaded maxCellValue = %d, want %d", got, want)
	}
}

func TestImportLegacySaveRestoresDigitCapFromState(t *testing.T) {
	loaded, err := ImportModel([]byte(`{"width":3,"height":3,"state":"2 . 1\n. 3 .\n2 . 1","provided":"#.#\n.#.\n#.#","mode_title":"Test"}`))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := loaded.maxCellValue, 3; got != want {
		t.Fatalf("loaded maxCellValue = %d, want %d", got, want)
	}
}

func TestCellViewConflictedCursorIsDistinct(t *testing.T) {
	conflict := cellView(3, false, false, false, false, false, false, true, nil)
	cursorConflict := cellView(3, false, true, false, false, false, false, true, nil)

	if conflict == cursorConflict {
		t.Fatal("conflicted cursor should render differently from plain conflicted cell")
	}
	if !strings.Contains(cursorConflict, game.CursorLeft+"3"+game.CursorRight) {
		t.Fatal("conflicted cursor should include cursor markers")
	}
}

func TestDynamicGridEdges(t *testing.T) {
	g := grid{
		{2, 2, 1},
		{3, 3, 1},
	}
	h := grid{
		{4, 1},
		{4, 2},
	}

	if hasVerticalEdge(g, 1, 0) {
		t.Fatal("expected open interior edge inside a region")
	}
	if !hasVerticalEdge(g, 2, 0) {
		t.Fatal("expected border between different regions")
	}
	if horizontalEdge(h, 0, 1) {
		t.Fatal("expected open horizontal edge inside a region")
	}
	if !horizontalEdge(g, 0, 0) {
		t.Fatal("expected top outer border")
	}
}

func TestCompletedRegionGapBackgrounds(t *testing.T) {
	conflicts := validateGridState(grid{{4, 4}, {4, 4}}).conflicts
	m := Model{
		width:     2,
		height:    2,
		grid:      grid{{4, 4}, {4, 4}},
		conflicts: conflicts,
	}
	completed := completedRegionBackgrounds(m.grid, m.conflicts)
	regionSet := map[point]struct{}{}

	if got := verticalGapBackground(m, regionSet, completed, 1, 0); got == nil {
		t.Fatal("expected completed region to color vertical interior gap")
	}
	if got := horizontalGapBackground(m, regionSet, completed, 0, 1); got == nil {
		t.Fatal("expected completed region to color horizontal interior gap")
	}
	if got := junctionGapBackground(m, regionSet, completed, 1, 1); got == nil {
		t.Fatal("expected completed region to color interior junction")
	}
}

func TestConflictGapBackgroundUsesConflictColor(t *testing.T) {
	conflicts := validateGridState(grid{{4, 4}, {4, 4}}).conflicts
	m := Model{
		width:     2,
		height:    2,
		grid:      grid{{4, 4}, {4, 4}},
		conflicts: conflicts,
	}
	m.conflicts[0][0] = true

	if got := verticalGapBackground(m, nil, nil, 1, 0); !sameColor(got, game.ConflictBG()) {
		t.Fatal("expected conflict color to bridge vertical interior gap")
	}
	if got := junctionGapBackground(m, nil, nil, 1, 1); !sameColor(got, game.ConflictBG()) {
		t.Fatal("expected conflict color to bridge interior junction")
	}
}

func TestGridViewUsesUniformRows(t *testing.T) {
	conflicts := validateGridState(grid{{2, 2, 1}, {3, 3, 3}, {1, 2, 2}}).conflicts
	m := Model{
		width:     3,
		height:    3,
		grid:      grid{{2, 2, 1}, {3, 3, 3}, {1, 2, 2}},
		provided:  newProvidedMask(grid{{2, 2, 1}, {3, 3, 3}, {1, 2, 2}}),
		conflicts: conflicts,
		cursor:    pointCursor(1, 1),
	}

	view := gridView(m)
	lines := strings.Split(view, "\n")
	if got, want := len(lines), m.height*2+1; got != want {
		t.Fatalf("grid line count = %d, want %d", got, want)
	}

	wantWidth := lipgloss.Width(lines[0])
	for i, line := range lines {
		if width := lipgloss.Width(line); width != wantWidth {
			t.Fatalf("line %d width = %d, want %d", i, width, wantWidth)
		}
	}
}

func TestGridViewOmitsBorderBetweenAdjacentEmptyCells(t *testing.T) {
	m := Model{
		width:     2,
		height:    1,
		grid:      grid{{0, 0}},
		provided:  newProvidedMask(grid{{0, 0}}),
		conflicts: validateGridState(grid{{0, 0}}).conflicts,
	}

	view := ansi.Strip(gridView(m))
	lines := strings.Split(view, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}

	content := []rune(lines[1])
	if got := content[4]; got != ' ' {
		t.Fatalf("separator between adjacent empty cells = %q, want space", got)
	}
}

func TestSolvedGapBackgroundUsesSolvedColor(t *testing.T) {
	conflicts := validateGridState(grid{{2, 2}}).conflicts
	m := Model{
		width:     2,
		height:    1,
		grid:      grid{{2, 2}},
		conflicts: conflicts,
		solved:    true,
	}
	got := verticalGapBackground(m, nil, nil, 1, 0)
	if got == nil {
		t.Fatal("expected solved gap background")
	}
	if !sameColor(got, theme.Current().SuccessBG) {
		t.Fatal("expected solved gap background to use solved color")
	}
}

func TestScreenToGrid(t *testing.T) {
	m := Model{
		width:      3,
		height:     3,
		grid:       grid{{2, 2, 1}, {3, 3, 3}, {1, 2, 2}},
		provided:   newProvidedMask(grid{{2, 2, 1}, {3, 3, 3}, {1, 2, 2}}),
		conflicts:  validateGridState(grid{{2, 2, 1}, {3, 3, 3}, {1, 2, 2}}).conflicts,
		keys:       DefaultKeyMap,
		modeTitle:  "Test",
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
		wantOK  bool
	}{
		{"origin cell", ox, oy, 0, 0, true},
		{"cell (1,0)", ox + cellWidth + 1, oy, 1, 0, true},
		{"cell (0,1)", ox, oy + 2, 0, 1, true},
		{"vertical border miss", ox + cellWidth, oy, 0, 0, false},
		{"horizontal border miss", ox, oy + 1, 0, 0, false},
		{"outside left", ox - 1, oy, 0, 0, false},
		{"outside bottom", ox, oy + 6, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, row, ok := m.screenToGrid(tt.screenX, tt.screenY)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && (col != tt.wantCol || row != tt.wantRow) {
				t.Fatalf("screenToGrid(%d, %d) = (%d, %d), want (%d, %d)", tt.screenX, tt.screenY, col, row, tt.wantCol, tt.wantRow)
			}
		})
	}
}

func TestMouseClickMovesCursor(t *testing.T) {
	puzzle := Puzzle{
		Width:  3,
		Height: 3,
		Givens: grid{{2, 0, 1}, {0, 3, 0}, {2, 0, 1}},
	}
	g, err := New(NewMode("Test", "desc", 3, 3, 3), puzzle)
	if err != nil {
		t.Fatal(err)
	}
	m := g.(Model)
	m.termWidth = 120
	m.termHeight = 40

	x, y := fillominoCellScreenCoords(&m, 2, 1)
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	got := next.(Model)

	if got.cursor.X != 2 || got.cursor.Y != 1 {
		t.Fatalf("cursor = (%d,%d), want (2,1)", got.cursor.X, got.cursor.Y)
	}
}

func TestMouseClickOnBorderDoesNotMoveCursor(t *testing.T) {
	puzzle := Puzzle{
		Width:  3,
		Height: 3,
		Givens: grid{{2, 0, 1}, {0, 3, 0}, {2, 0, 1}},
	}
	g, err := New(NewMode("Test", "desc", 3, 3, 3), puzzle)
	if err != nil {
		t.Fatal(err)
	}
	m := g.(Model)
	m.termWidth = 120
	m.termHeight = 40
	m.cursor = pointCursor(1, 1)

	x, y := fillominoCellScreenCoords(&m, 1, 1)
	next, _ := m.Update(tea.MouseClickMsg{X: x + cellWidth, Y: y, Button: tea.MouseLeft})
	got := next.(Model)

	if got.cursor.X != 1 || got.cursor.Y != 1 {
		t.Fatalf("cursor moved to (%d,%d), want unchanged (1,1)", got.cursor.X, got.cursor.Y)
	}
}

func pointCursor(x, y int) game.Cursor {
	return game.Cursor{X: x, Y: y}
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func fillominoCellScreenCoords(m *Model, col, row int) (int, int) {
	ox, oy := m.gridOrigin()
	return ox + col*(cellWidth+1), oy + row*2
}
