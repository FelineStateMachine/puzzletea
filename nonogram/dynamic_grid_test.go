package nonogram

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestGridViewUsesOnlyOuterBordersAndFiveCellSeparators(t *testing.T) {
	g := newGrid(createEmptyState(10, 10))

	lines := strings.Split(ansi.Strip(gridView(g, game.Cursor{X: 4, Y: 4}, false)), "\n")
	if got, want := len(lines), 13; got != want {
		t.Fatalf("line count = %d, want %d", got, want)
	}
	if got, want := lipgloss.Width(lines[0]), 33; got != want {
		t.Fatalf("top row width = %d, want %d", got, want)
	}
	wantTop := "┌" + strings.Repeat("─", cellWidth*5) + "┬" + strings.Repeat("─", cellWidth*5) + "┐"
	if got := lines[0]; got != wantTop {
		t.Fatalf("top row = %q, want %q", got, wantTop)
	}

	contentRow := []rune(lines[1])
	if got := contentRow[0]; got != '│' {
		t.Fatalf("left border = %q, want %q", got, '│')
	}
	if got := contentRow[16]; got != '│' {
		t.Fatalf("five-cell separator = %q, want %q", got, '│')
	}
	if got := contentRow[32]; got != '│' {
		t.Fatalf("right border = %q, want %q", got, '│')
	}
	if got := strings.Count(lines[1], "│"); got != 3 {
		t.Fatalf("content row border count = %d, want 3", got)
	}

	wantDivider := "├" + strings.Repeat("─", cellWidth*5) + "┼" + strings.Repeat("─", cellWidth*5) + "┤"
	if got := lines[6]; got != wantDivider {
		t.Fatalf("major divider = %q, want %q", got, wantDivider)
	}
}

func TestSeparatorBridgeFillHighlightsOnlySpacerEdgesOnAxis(t *testing.T) {
	cursor := game.Cursor{X: 4, Y: 4}
	surface := theme.Current().Surface

	tests := []struct {
		name   string
		bridge game.DynamicGridBridge
		want   color.Color
	}{
		{
			name: "vertical spacer on cursor row",
			bridge: game.DynamicGridBridge{
				Kind: game.DynamicGridBridgeVertical,
				X:    5,
				Y:    4,
			},
			want: surface,
		},
		{
			name: "horizontal spacer on cursor column",
			bridge: game.DynamicGridBridge{
				Kind: game.DynamicGridBridgeHorizontal,
				X:    4,
				Y:    5,
			},
			want: surface,
		},
		{
			name: "normal interior bridge stays empty",
			bridge: game.DynamicGridBridge{
				Kind: game.DynamicGridBridgeVertical,
				X:    1,
				Y:    4,
			},
			want: nil,
		},
		{
			name: "outer border stays empty",
			bridge: game.DynamicGridBridge{
				Kind: game.DynamicGridBridgeHorizontal,
				X:    4,
				Y:    0,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := separatorBridgeFill(cursor, false, 10, 10, tt.bridge); !game.SameColor(got, tt.want) {
				t.Fatalf("separatorBridgeFill() = %v, want %v", got, tt.want)
			}
		})
	}

	if got := separatorBridgeFill(cursor, true, 10, 10, game.DynamicGridBridge{
		Kind: game.DynamicGridBridgeVertical,
		X:    5,
		Y:    4,
	}); got != nil {
		t.Fatalf("solved separatorBridgeFill() = %v, want nil", got)
	}
}

func TestBuildBoardBlockPreservesHintBandLayout(t *testing.T) {
	m := Model{
		width:    5,
		height:   5,
		rowHints: TomographyDefinition{{1}, {2}, {3}, {4}, {5}},
		colHints: TomographyDefinition{{1}, {2}, {3}, {4}, {5}},
		grid:     newGrid(createEmptyState(5, 5)),
		cursor:   game.Cursor{X: 0, Y: 0},
	}

	layout := buildBoardBlock(m)
	lines := strings.Split(ansi.Strip(layout.Block), "\n")
	if len(lines) <= layout.HintHeight {
		t.Fatalf("block height = %d, want > hint height %d", len(lines), layout.HintHeight)
	}
	if got := lipgloss.Width(lines[layout.HintHeight]); got <= layout.HintWidth {
		t.Fatalf("content row width = %d, want > hint width %d", got, layout.HintWidth)
	}
	if got := []rune(lines[layout.HintHeight])[0]; got != ' ' {
		t.Fatalf("row hint should render outside the grid border, got %q", got)
	}
	if got := []rune(lines[layout.HintHeight])[layout.HintWidth]; got != '┌' {
		t.Fatalf("grid top border should start after hint band, got %q", got)
	}
	if got := []rune(lines[layout.HintHeight+1])[layout.HintWidth]; got != '│' {
		t.Fatalf("grid content should remain offset after hint band, got %q", got)
	}
}

func TestBuildBoardBlockAlignsHintSeparatorsWithGrid(t *testing.T) {
	m := Model{
		width:        10,
		height:       10,
		rowHints:     repeatedHints(10, []int{1}),
		colHints:     repeatedHints(10, []int{1}),
		grid:         newGrid(createEmptyState(10, 10)),
		currentHints: Hints{rows: repeatedHints(10, []int{1}), cols: repeatedHints(10, []int{1})},
	}

	layout := buildBoardBlock(m)
	lines := strings.Split(ansi.Strip(layout.Block), "\n")
	separatorColumn := layout.HintWidth + 1 + cellWidth*5
	if got := []rune(lines[layout.HintHeight-1])[separatorColumn]; got != '│' {
		t.Fatalf("column hint separator = %q, want %q", got, '│')
	}
	if got := []rune(lines[layout.HintHeight])[separatorColumn]; got != '┬' {
		t.Fatalf("top border separator = %q, want %q", got, '┬')
	}
	if got := []rune(lines[layout.HintHeight+6])[separatorColumn]; got != '┼' {
		t.Fatalf("row divider separator = %q, want %q", got, '┼')
	}
}

func TestBuildBoardBlockDimensionsAcrossModes(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{name: "5x5", size: 5},
		{name: "10x10", size: 10},
		{name: "15x15", size: 15},
		{name: "20x20", size: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := repeatedHints(tt.size, []int{1})
			m := Model{
				width:        tt.size,
				height:       tt.size,
				rowHints:     hints,
				colHints:     hints,
				grid:         newGrid(createEmptyState(tt.size, tt.size)),
				currentHints: Hints{rows: hints, cols: hints},
			}

			layout := buildBoardBlock(m)
			if layout.HintWidth <= 0 || layout.HintHeight <= 0 {
				t.Fatalf("hint bands should be non-zero, got (%d,%d)", layout.HintWidth, layout.HintHeight)
			}
			if lipgloss.Width(layout.Block) < layout.HintWidth+lipgloss.Width(layout.Grid) {
				t.Fatalf("block width = %d, want at least %d", lipgloss.Width(layout.Block), layout.HintWidth+lipgloss.Width(layout.Grid))
			}
			if lipgloss.Height(layout.Block) != layout.HintHeight+lipgloss.Height(layout.Grid) {
				t.Fatalf("block height = %d, want %d", lipgloss.Height(layout.Block), layout.HintHeight+lipgloss.Height(layout.Grid))
			}
		})
	}
}

func TestHintViewsUseSatisfiedStyle(t *testing.T) {
	rowHints := TomographyDefinition{{3}, {1, 1}}
	rowView := rowHintView(rowHints, 5, rowHints)
	wantRow := hintSatisfiedStyle().Width(5).Align(lipgloss.Right).Render(" 3")
	if !strings.Contains(rowView, wantRow) {
		t.Fatalf("rowHintView() = %q, want satisfied row %q", rowView, wantRow)
	}

	colHints := TomographyDefinition{{2}, {1}}
	colView := colHintView(colHints, 1, colHints)
	wantCol := hintSatisfiedStyle().Width(cellWidth).Align(lipgloss.Center).Render("2")
	if !strings.Contains(colView, wantCol) {
		t.Fatalf("colHintView() = %q, want satisfied column %q", colView, wantCol)
	}
}

func repeatedHints(size int, hint []int) TomographyDefinition {
	hints := make(TomographyDefinition, size)
	for i := range size {
		hints[i] = append([]int(nil), hint...)
	}
	return hints
}
