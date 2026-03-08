package sudokurgb

import (
	"image"
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

func TestBridgeFillHighlightsCursorAxis(t *testing.T) {
	m := Model{cursor: game.Cursor{X: 1, Y: 1}}
	surface := theme.Current().Surface

	tests := []struct {
		name   string
		bridge game.DynamicGridBridge
		want   color.Color
	}{
		{
			name: "vertical bridge on cursor row",
			bridge: game.DynamicGridBridge{
				Kind:  game.DynamicGridBridgeVertical,
				Cells: [4]image.Point{image.Pt(0, 1), image.Pt(1, 1)},
				Count: 2,
			},
			want: surface,
		},
		{
			name: "horizontal bridge on cursor column",
			bridge: game.DynamicGridBridge{
				Kind:  game.DynamicGridBridgeHorizontal,
				Cells: [4]image.Point{image.Pt(1, 0), image.Pt(1, 1)},
				Count: 2,
			},
			want: surface,
		},
		{
			name: "junction stays neutral",
			bridge: game.DynamicGridBridge{
				Kind:  game.DynamicGridBridgeJunction,
				Cells: [4]image.Point{image.Pt(1, 1)},
				Count: 1,
			},
			want: nil,
		},
		{
			name: "solved board disables bridge highlight",
			bridge: game.DynamicGridBridge{
				Kind:  game.DynamicGridBridgeVertical,
				Cells: [4]image.Point{image.Pt(0, 1), image.Pt(1, 1)},
				Count: 2,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solved := tt.name == "solved board disables bridge highlight"
			if got := bridgeFill(m, solved, tt.bridge); !game.SameColor(got, tt.want) {
				t.Fatalf("bridgeFill() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActiveBoxZoneFillHighlightsOnlyActiveBox(t *testing.T) {
	cursor := game.Cursor{X: 1, Y: 1}
	surface := theme.Current().Surface

	if got := activeBoxZoneFill(cursor, false, 0); !game.SameColor(got, surface) {
		t.Fatalf("activeBoxZoneFill(in box) = %v, want %v", got, surface)
	}
	if got := activeBoxZoneFill(cursor, false, 1); got != nil {
		t.Fatalf("activeBoxZoneFill(out of box) = %v, want nil", got)
	}
	if got := activeBoxZoneFill(cursor, true, 0); got != nil {
		t.Fatalf("activeBoxZoneFill(solved) = %v, want nil", got)
	}
}

func TestCellViewHighlightsSameBoxAndCrosshair(t *testing.T) {
	g := newGrid(nil)
	g[2][2].v = 2
	g[1][4].v = 3
	m := Model{
		cursor: game.Cursor{X: 1, Y: 1},
		grid:   g,
		keys:   DefaultKeyMap,
	}

	surface := theme.Current().Surface

	gotSameBox := cellView(m, 2, 2, false)
	wantSameBox := userCellStyle(2).
		Background(surface).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("■")
	if gotSameBox != wantSameBox {
		t.Fatalf("same-box cellView() = %q, want %q", gotSameBox, wantSameBox)
	}

	gotSameRow := cellView(m, 4, 1, false)
	wantSameRow := userCellStyle(3).
		Background(surface).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render("●")
	if gotSameRow != wantSameRow {
		t.Fatalf("same-row cellView() = %q, want %q", gotSameRow, wantSameRow)
	}
}

func TestCellViewUsesCursorBackground(t *testing.T) {
	g := newGrid(nil)
	g[1][1].v = 2
	m := Model{
		cursor: game.Cursor{X: 1, Y: 1},
		grid:   g,
		keys:   DefaultKeyMap,
	}

	gotFilled := cellView(m, 1, 1, false)
	wantFilled := valueCursorStyle(2).
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render(game.CursorLeft + "■" + game.CursorRight)
	if gotFilled != wantFilled {
		t.Fatalf("filled cursor cellView() = %q, want %q", gotFilled, wantFilled)
	}

	m.grid[1][1].v = 0
	gotEmpty := cellView(m, 1, 1, false)
	wantEmpty := game.CursorStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Render(game.CursorLeft + "·" + game.CursorRight)
	if gotEmpty != wantEmpty {
		t.Fatalf("empty cursor cellView() = %q, want %q", gotEmpty, wantEmpty)
	}
}
