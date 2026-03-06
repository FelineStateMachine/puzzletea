package game

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderDynamicGridCenterJunctionRune(t *testing.T) {
	tests := []struct {
		name     string
		zones    [][]int
		wantRune rune
	}{
		{
			name:     "open interior",
			zones:    [][]int{{0, 0}, {0, 0}},
			wantRune: ' ',
		},
		{
			name:     "vertical wall",
			zones:    [][]int{{0, 1}, {0, 1}},
			wantRune: '│',
		},
		{
			name:     "cross",
			zones:    [][]int{{0, 1}, {2, 3}},
			wantRune: '┼',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := RenderDynamicGrid(DynamicGridSpec{
				Width:  2,
				Height: 2,
				Cell: func(_, _ int) string {
					return "   "
				},
				ZoneAt: func(x, y int) int {
					return tt.zones[y][x]
				},
			})

			lines := strings.Split(ansi.Strip(view), "\n")
			got := []rune(lines[2])[4]
			if got != tt.wantRune {
				t.Fatalf("center junction = %q, want %q", got, tt.wantRune)
			}
		})
	}
}

func TestRenderDynamicGridUsesUniformRows(t *testing.T) {
	view := RenderDynamicGrid(DynamicGridSpec{
		Width:  3,
		Height: 3,
		Cell: func(_, _ int) string {
			return "   "
		},
		ZoneAt: func(x, y int) int {
			return y*10 + x
		},
	})

	lines := strings.Split(view, "\n")
	if got, want := len(lines), 7; got != want {
		t.Fatalf("line count = %d, want %d", got, want)
	}

	wantWidth := lipgloss.Width(lines[0])
	for i, line := range lines {
		if width := lipgloss.Width(line); width != wantWidth {
			t.Fatalf("line %d width = %d, want %d", i, width, wantWidth)
		}
	}
}

func TestDynamicGridRenderBorderCharSolvedDefaults(t *testing.T) {
	got := dynamicGridRenderBorderChar('│', DefaultBorderColors(), true, nil)
	want := lipgloss.NewStyle().
		Foreground(theme.Current().SuccessBorder).
		Background(theme.Current().SuccessBG).
		Render("│")
	if got != want {
		t.Fatal("expected solved border char to use solved colors")
	}
}

func TestDynamicGridScreenToCell(t *testing.T) {
	metrics := DynamicGridMetrics{Width: 3, Height: 3, CellWidth: DynamicGridCellWidth}
	originX, originY := 10, 20

	tests := []struct {
		name              string
		screenX           int
		screenY           int
		includeSeparators bool
		wantCol           int
		wantRow           int
		wantOK            bool
	}{
		{"cell origin", originX, originY, false, 0, 0, true},
		{"next cell", originX + DynamicGridCellWidth + 1, originY, false, 1, 0, true},
		{"separator rejected", originX + DynamicGridCellWidth, originY, false, 0, 0, false},
		{"separator snaps during drag", originX + DynamicGridCellWidth, originY, true, 1, 0, true},
		{"outside right", originX + 3*(DynamicGridCellWidth+1), originY, false, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, row, ok := DynamicGridScreenToCell(
				metrics,
				originX,
				originY,
				tt.screenX,
				tt.screenY,
				tt.includeSeparators,
			)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && (col != tt.wantCol || row != tt.wantRow) {
				t.Fatalf("cell = (%d,%d), want (%d,%d)", col, row, tt.wantCol, tt.wantRow)
			}
		})
	}
}

func TestDynamicGridOriginStableAcrossSecondaryRows(t *testing.T) {
	grid := RenderDynamicGrid(DynamicGridSpec{
		Width:  3,
		Height: 2,
		Cell: func(_, _ int) string {
			return "   "
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
	})
	title := TitleBarView("Test", "Mode", false)
	shortView := ComposeGameViewRows(title, grid, StaticRow("1-9: place"))
	longView := ComposeGameViewRows(title, grid, StableRow("1-9: place", strings.Repeat("long help ", 6)))

	shortX, shortY := DynamicGridOrigin(120, 40, shortView, title, grid)
	longX, longY := DynamicGridOrigin(120, 40, longView, title, grid)
	if shortX != longX || shortY != longY {
		t.Fatalf("origin changed from (%d,%d) to (%d,%d)", shortX, shortY, longX, longY)
	}
}
