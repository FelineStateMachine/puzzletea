package sudoku

import (
	"image"
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderGridUsesBoxBoundaries(t *testing.T) {
	m := Model{
		grid:      emptyGrid(),
		conflicts: computeConflicts(emptyGrid()),
		keys:      DefaultKeyMap,
	}

	lines := strings.Split(ansi.Strip(renderGrid(m, false)), "\n")
	if got, want := len(lines), 19; got != want {
		t.Fatalf("line count = %d, want %d", got, want)
	}

	contentRow := []rune(lines[1])
	if got := contentRow[4]; got != ' ' {
		t.Fatalf("interior bridge rune = %q, want %q", got, ' ')
	}
	if got := contentRow[12]; got != '│' {
		t.Fatalf("box boundary rune = %q, want %q", got, '│')
	}

	innerBoundary := []rune(lines[2])
	if got := innerBoundary[1]; got != ' ' {
		t.Fatalf("interior row separator = %q, want %q", got, ' ')
	}

	boxBoundary := []rune(lines[6])
	if got := boxBoundary[1]; got != '─' {
		t.Fatalf("box row separator = %q, want %q", got, '─')
	}
}

func TestBridgeFillUsesAxisSpillOnly(t *testing.T) {
	m := Model{cursor: game.Cursor{X: 1, Y: 1}}
	surface := theme.Current().Surface

	tests := []struct {
		name   string
		bridge game.DynamicGridBridge
		want   color.Color
	}{
		{
			name: "vertical bridge spills left right on cursor row",
			bridge: game.DynamicGridBridge{
				Kind: game.DynamicGridBridgeVertical,
				X:    3,
				Y:    1,
				Cells: [4]image.Point{
					image.Pt(2, 1),
					image.Pt(3, 1),
				},
				Count: 2,
			},
			want: surface,
		},
		{
			name: "horizontal bridge spills up down on cursor column",
			bridge: game.DynamicGridBridge{
				Kind: game.DynamicGridBridgeHorizontal,
				X:    1,
				Y:    3,
				Cells: [4]image.Point{
					image.Pt(1, 2),
					image.Pt(1, 3),
				},
				Count: 2,
			},
			want: surface,
		},
		{
			name: "off axis bridge stays empty",
			bridge: game.DynamicGridBridge{
				Kind: game.DynamicGridBridgeVertical,
				X:    3,
				Y:    4,
				Cells: [4]image.Point{
					image.Pt(2, 4),
					image.Pt(3, 4),
				},
				Count: 2,
			},
			want: nil,
		},
		{
			name: "junction stays empty",
			bridge: game.DynamicGridBridge{
				Kind:    game.DynamicGridBridgeJunction,
				X:       2,
				Y:       1,
				Cells:   [4]image.Point{image.Pt(1, 0), image.Pt(1, 1), image.Pt(2, 0), image.Pt(2, 1)},
				Count:   4,
				Zone:    0,
				Uniform: true,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bridgeFill(m, false, tt.bridge); !sameColor(got, tt.want) {
				t.Fatalf("bridgeFill() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZoneFillHighlightsOnlyActiveBox(t *testing.T) {
	m := Model{cursor: game.Cursor{X: 1, Y: 1}}
	surface := theme.Current().Surface

	inBoxZone := 0
	if got := activeBoxZoneFill(m.cursor, false, inBoxZone); !sameColor(got, surface) {
		t.Fatalf("in-box bridge background = %v, want %v", got, surface)
	}

	if got := activeBoxZoneFill(m.cursor, false, inBoxZone); !sameColor(got, surface) {
		t.Fatalf("in-box junction background = %v, want %v", got, surface)
	}

	outOfBoxZone := 1
	if got := activeBoxZoneFill(m.cursor, false, outOfBoxZone); got != nil {
		t.Fatalf("out-of-box bridge background = %v, want nil", got)
	}

	if got := activeBoxZoneFill(m.cursor, true, inBoxZone); got != nil {
		t.Fatalf("solved in-box bridge background = %v, want nil", got)
	}
}

func TestCellViewSameNumberOverridesProvidedCrosshair(t *testing.T) {
	g := validCompleteGrid()
	m := Model{
		cursor: game.Cursor{X: 0, Y: 0},
		grid:   g,
		keys:   DefaultKeyMap,
	}
	m.providedGrid[4][4] = true

	got := cellView(m, 4, 4, false)
	want := sameNumberStyle(5).Width(cellWidth).AlignHorizontal(lipgloss.Center).Render("5")
	if got != want {
		t.Fatalf("cellView() = %q, want %q", got, want)
	}
}

func TestCellViewUsesSemanticDigitColors(t *testing.T) {
	p := theme.Current()
	g := emptyGrid()
	g[0][0].v = 5
	g[0][1].v = 7
	m := Model{
		cursor: game.Cursor{X: 8, Y: 8},
		grid:   g,
		keys:   DefaultKeyMap,
	}
	m.providedGrid[0][1] = true

	gotUser := cellView(m, 0, 0, false)
	wantUser := lipgloss.NewStyle().
		Foreground(digitColor(5)).
		Background(p.BG).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("5")
	if gotUser != wantUser {
		t.Fatalf("user cellView() = %q, want %q", gotUser, wantUser)
	}

	gotProvided := cellView(m, 1, 0, false)
	wantProvided := lipgloss.NewStyle().
		Bold(true).
		Foreground(digitColor(7)).
		Background(p.BG).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("7")
	if gotProvided != wantProvided {
		t.Fatalf("provided cellView() = %q, want %q", gotProvided, wantProvided)
	}
}

func TestCellViewUsesDigitColorForFilledCursor(t *testing.T) {
	g := emptyGrid()
	g[1][1].v = 6
	m := Model{
		cursor: game.Cursor{X: 1, Y: 1},
		grid:   g,
		keys:   DefaultKeyMap,
	}

	got := cellView(m, 1, 1, false)
	want := digitCursorStyle(6).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render(game.CursorLeft + "6" + game.CursorRight)
	if got != want {
		t.Fatalf("filled cursor cellView() = %q, want %q", got, want)
	}
}

func TestCellViewKeepsDefaultCursorStyleForEmptyCell(t *testing.T) {
	m := Model{
		cursor: game.Cursor{X: 1, Y: 1},
		grid:   emptyGrid(),
		keys:   DefaultKeyMap,
	}

	got := cellView(m, 1, 1, false)
	want := game.CursorStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render(game.CursorLeft + "·" + game.CursorRight)
	if got != want {
		t.Fatalf("empty cursor cellView() = %q, want %q", got, want)
	}
}

func TestScreenToGrid(t *testing.T) {
	m := testDynamicBoardModel()
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
		{"next cell", ox + cellWidth + 1, oy, 1, 0, true},
		{"next row", ox, oy + 2, 0, 1, true},
		{"vertical separator", ox + cellWidth, oy, 0, 0, false},
		{"horizontal separator", ox, oy + 1, 0, 0, false},
		{"outside left", ox - 1, oy, 0, 0, false},
		{"outside bottom", ox, oy + 18, 0, 0, false},
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
	m := testDynamicBoardModel()

	x, y := sudokuCellScreenCoords(&m, 2, 1)
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	got := next.(Model)

	if got.cursor.X != 2 || got.cursor.Y != 1 {
		t.Fatalf("cursor = (%d,%d), want (2,1)", got.cursor.X, got.cursor.Y)
	}
}

func TestMouseClickOnSeparatorDoesNotMoveCursor(t *testing.T) {
	m := testDynamicBoardModel()
	m.cursor = game.Cursor{X: 1, Y: 1}

	x, y := sudokuCellScreenCoords(&m, 1, 1)
	next, _ := m.Update(tea.MouseClickMsg{X: x + cellWidth, Y: y, Button: tea.MouseLeft})
	got := next.(Model)

	if got.cursor.X != 1 || got.cursor.Y != 1 {
		t.Fatalf("cursor = (%d,%d), want unchanged (1,1)", got.cursor.X, got.cursor.Y)
	}
}

func TestMouseClickSameCellDoesNotMutateValue(t *testing.T) {
	m := testDynamicBoardModel()
	m.cursor = game.Cursor{X: 1, Y: 1}
	m.grid[1][1].v = 4

	x, y := sudokuCellScreenCoords(&m, 1, 1)
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	got := next.(Model)

	if got.grid[1][1].v != 4 {
		t.Fatalf("cell value = %d, want unchanged 4", got.grid[1][1].v)
	}
}

func TestOriginInvalidation(t *testing.T) {
	t.Run("help toggle", func(t *testing.T) {
		m := testDynamicBoardModel()
		m.originValid = true

		next, _ := m.Update(game.HelpToggleMsg{Show: true})
		got := next.(Model)
		if got.originValid {
			t.Fatal("expected help toggle to invalidate origin cache")
		}
	})

	t.Run("window size", func(t *testing.T) {
		m := testDynamicBoardModel()
		m.originValid = true

		next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		got := next.(Model)
		if got.originValid {
			t.Fatal("expected window resize to invalidate origin cache")
		}
	})

	t.Run("reset", func(t *testing.T) {
		m := testDynamicBoardModel()
		m.originValid = true

		got := m.Reset().(Model)
		if got.originValid {
			t.Fatal("expected reset to invalidate origin cache")
		}
	})

	t.Run("solved transition", func(t *testing.T) {
		g := validCompleteGrid()
		g[0][0].v = 0

		m := Model{
			grid:        g,
			conflicts:   computeConflicts(g),
			keys:        DefaultKeyMap,
			originValid: true,
		}

		m.updateCell(5)
		if m.originValid {
			t.Fatal("expected solving move to invalidate origin cache")
		}
		if !m.isSolved() {
			t.Fatal("expected board to be solved after filling final cell")
		}
	})
}

func testDynamicBoardModel() Model {
	g := emptyGrid()
	return Model{
		grid:       g,
		conflicts:  computeConflicts(g),
		keys:       DefaultKeyMap,
		modeTitle:  "Test",
		termWidth:  120,
		termHeight: 40,
	}
}

func sudokuCellScreenCoords(m *Model, col, row int) (int, int) {
	ox, oy := m.gridOrigin()
	return ox + col*(cellWidth+1), oy + row*2
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}

	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}
