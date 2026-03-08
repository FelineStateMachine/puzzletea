package game

import (
	"image"
	"image/color"
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

func TestRenderDynamicGridSupportsCustomCellWidth(t *testing.T) {
	view := RenderDynamicGrid(DynamicGridSpec{
		Width:     2,
		Height:    1,
		CellWidth: 5,
		Cell: func(_, _ int) string {
			return "     "
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
		VerticalBridgeText: func(x, y int) string {
			if x == 1 && y == 0 {
				return "x"
			}
			return ""
		},
	})

	lines := strings.Split(ansi.Strip(view), "\n")
	if got := len([]rune(lines[0])); got != 13 {
		t.Fatalf("top row width = %d, want %d", got, 13)
	}
	if !strings.Contains(lines[1], "x") {
		t.Fatalf("content row = %q, want custom-width bridge text", lines[1])
	}
}

func TestRenderDynamicGridCollapsesSparseBridges(t *testing.T) {
	view := RenderDynamicGrid(DynamicGridSpec{
		Width:  2,
		Height: 2,
		Cell: func(_, _ int) string {
			return "   "
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
		HasVerticalEdge: func(x, _ int) bool {
			return x <= 0 || x >= 2
		},
		HasHorizontalEdge: func(_, y int) bool {
			return y <= 0 || y >= 2
		},
		VerticalBridgeWidth: func(x int) int {
			if x <= 0 || x >= 2 {
				return 1
			}
			return 0
		},
		HorizontalBridgeHeight: func(y int) int {
			if y <= 0 || y >= 2 {
				return 1
			}
			return 0
		},
	})

	lines := strings.Split(ansi.Strip(view), "\n")
	if got, want := len(lines), 4; got != want {
		t.Fatalf("line count = %d, want %d", got, want)
	}
	if got, want := lines[0], "┌──────┐"; got != want {
		t.Fatalf("top row = %q, want %q", got, want)
	}
	if got := strings.Count(lines[1], "│"); got != 2 {
		t.Fatalf("content row border count = %d, want 2", got)
	}
}

func TestRenderDynamicGridEdgeOverridesSuppressInteriorBorders(t *testing.T) {
	view := RenderDynamicGrid(DynamicGridSpec{
		Width:  2,
		Height: 2,
		Cell: func(_, _ int) string {
			return "   "
		},
		ZoneAt: func(x, y int) int {
			return y*10 + x
		},
		HasVerticalEdge: func(x, _ int) bool {
			return x <= 0 || x >= 2
		},
		HasHorizontalEdge: func(_, y int) bool {
			return y <= 0 || y >= 2
		},
	})

	lines := strings.Split(ansi.Strip(view), "\n")
	if got := []rune(lines[1])[4]; got != ' ' {
		t.Fatalf("interior vertical bridge rune = %q, want %q", got, ' ')
	}
	if got := []rune(lines[2])[4]; got != ' ' {
		t.Fatalf("interior junction rune = %q, want %q", got, ' ')
	}
}

func TestRenderDynamicGridDefaultsToZoneEdgesWithoutOverrides(t *testing.T) {
	view := RenderDynamicGrid(DynamicGridSpec{
		Width:  2,
		Height: 2,
		Cell: func(_, _ int) string {
			return "   "
		},
		ZoneAt: func(x, y int) int {
			return y*10 + x
		},
	})

	lines := strings.Split(ansi.Strip(view), "\n")
	if got := []rune(lines[1])[4]; got != '│' {
		t.Fatalf("interior vertical bridge rune = %q, want %q", got, '│')
	}
	if got := []rune(lines[2])[4]; got != '┼' {
		t.Fatalf("interior junction rune = %q, want %q", got, '┼')
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

func TestDynamicGridRenderBorderCharUsesTextOnTintedBackground(t *testing.T) {
	bg := theme.Current().Surface
	got := dynamicGridRenderBorderChar('│', DefaultBorderColors(), false, bg)
	want := lipgloss.NewStyle().
		Foreground(theme.TextOnBG(bg)).
		Background(bg).
		Render("│")
	if got != want {
		t.Fatal("expected tinted border char to use contrast-aware foreground")
	}
}

func TestBorderCharHighlightUsesTextOnCrosshairBackground(t *testing.T) {
	colors := DefaultBorderColors()
	got := BorderChar("│", colors, false, true)
	want := lipgloss.NewStyle().
		Foreground(theme.TextOnBG(colors.CrosshairBG)).
		Background(colors.CrosshairBG).
		Render("│")
	if got != want {
		t.Fatal("expected highlighted border char to use contrast-aware foreground")
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

func TestDynamicGridScreenToCellSupportsSparseBridges(t *testing.T) {
	metrics := DynamicGridMetrics{
		Width:     3,
		Height:    3,
		CellWidth: DynamicGridCellWidth,
		VerticalBridgeWidth: func(x int) int {
			if x == 2 {
				return 1
			}
			return 0
		},
		HorizontalBridgeHeight: func(y int) int {
			if y == 2 {
				return 1
			}
			return 0
		},
	}
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
		{"flush next cell", originX + DynamicGridCellWidth, originY, false, 1, 0, true},
		{"major vertical separator rejected", originX + 2*DynamicGridCellWidth, originY, false, 0, 0, false},
		{"major vertical separator snaps", originX + 2*DynamicGridCellWidth, originY, true, 2, 0, true},
		{"major horizontal separator rejected", originX, originY + 2, false, 0, 0, false},
		{"major horizontal separator snaps", originX, originY + 2, true, 0, 2, true},
		{"outside right", originX + 3*DynamicGridCellWidth + 1, originY, false, 0, 0, false},
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

func TestDynamicGridBridgeOnCrosshairAxis(t *testing.T) {
	cursor := Cursor{X: 2, Y: 1}

	tests := []struct {
		name   string
		bridge DynamicGridBridge
		want   bool
	}{
		{
			name: "vertical bridge matches cursor row via cells",
			bridge: DynamicGridBridge{
				Kind:  DynamicGridBridgeVertical,
				Cells: [4]image.Point{{X: 1, Y: 1}, {X: 2, Y: 1}},
				Count: 2,
			},
			want: true,
		},
		{
			name: "horizontal bridge matches cursor column via cells",
			bridge: DynamicGridBridge{
				Kind:  DynamicGridBridgeHorizontal,
				Cells: [4]image.Point{{X: 2, Y: 0}, {X: 2, Y: 1}},
				Count: 2,
			},
			want: true,
		},
		{
			name: "separator vertical bridge matches cursor row without cells",
			bridge: DynamicGridBridge{
				Kind: DynamicGridBridgeVertical,
				Y:    1,
			},
			want: true,
		},
		{
			name: "separator horizontal bridge matches cursor column without cells",
			bridge: DynamicGridBridge{
				Kind: DynamicGridBridgeHorizontal,
				X:    2,
			},
			want: true,
		},
		{
			name: "junction does not participate in axis highlighting",
			bridge: DynamicGridBridge{
				Kind:  DynamicGridBridgeJunction,
				Cells: [4]image.Point{{X: 2, Y: 1}},
				Count: 1,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DynamicGridBridgeOnCrosshairAxis(cursor, tt.bridge); got != tt.want {
				t.Fatalf("DynamicGridBridgeOnCrosshairAxis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSameColor(t *testing.T) {
	if !SameColor(color.NRGBA{R: 10, G: 20, B: 30, A: 40}, color.NRGBA{R: 10, G: 20, B: 30, A: 40}) {
		t.Fatal("expected identical colors to match")
	}
	if SameColor(color.NRGBA{R: 10, G: 20, B: 30, A: 40}, color.NRGBA{R: 10, G: 20, B: 30, A: 41}) {
		t.Fatal("expected different colors not to match")
	}
	if !SameColor(nil, nil) {
		t.Fatal("expected nil colors to match")
	}
	if SameColor(color.NRGBA{R: 10, G: 20, B: 30, A: 40}, nil) {
		t.Fatal("expected color and nil not to match")
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

func TestRenderDynamicGridVerticalBridgeText(t *testing.T) {
	view := RenderDynamicGrid(DynamicGridSpec{
		Width:  2,
		Height: 1,
		Cell: func(_, _ int) string {
			return "   "
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
		VerticalBridgeText: func(x, y int) string {
			if x == 1 && y == 0 {
				return "x"
			}
			return ""
		},
	})

	lines := strings.Split(ansi.Strip(view), "\n")
	if got := []rune(lines[1])[4]; got != 'x' {
		t.Fatalf("vertical bridge rune = %q, want %q", got, 'x')
	}
}

func TestRenderDynamicGridHorizontalBridgeText(t *testing.T) {
	view := RenderDynamicGrid(DynamicGridSpec{
		Width:  1,
		Height: 2,
		Cell: func(_, _ int) string {
			return "   "
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
		HorizontalBridgeText: func(x, y int) string {
			if x == 0 && y == 1 {
				return "="
			}
			return ""
		},
	})

	lines := strings.Split(ansi.Strip(view), "\n")
	if got := []rune(lines[2])[2]; got != '=' {
		t.Fatalf("horizontal bridge rune = %q, want %q", got, '=')
	}
}
