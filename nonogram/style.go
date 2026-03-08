package nonogram

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const (
	cellWidth   = game.DynamicGridCellWidth
	spacerEvery = 5
)

var renderRuneMap = map[rune]string{
	filledTile: " \u25a0 ",
	markedTile: " \u2715 ",
	emptyTile:  " \u00b7 ",
}

func filledStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.AccentText).
		Background(p.AccentBG)
}

func markedStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.AccentSoft).
		Background(p.BG)
}

func emptyStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.TextDim).
		Background(p.BG)
}

func hintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Info)
}

func hintSatisfiedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().SuccessBorder)
}

func renderStyleMap() map[rune]lipgloss.Style {
	return map[rune]lipgloss.Style{
		filledTile: filledStyle(),
		markedTile: markedStyle(),
		emptyTile:  emptyStyle(),
	}
}

// needsSpacer reports whether a separator should be inserted after index i
// in a dimension of size n (i.e. after every spacerEvery cells, except the last).
func needsSpacer(i, n int) bool {
	return n > spacerEvery && (i+1)%spacerEvery == 0 && i < n-1
}

func colHintView(c TomographyDefinition, height int, current ...TomographyDefinition) string {
	var hasCurrent bool
	var curr TomographyDefinition
	if len(current) > 0 {
		hasCurrent = true
		curr = current[0]
	}

	hintCellStyle := lipgloss.NewStyle().Width(cellWidth)

	n := len(c)
	var renderedCols []string
	for i, hints := range c {
		var colHints []string
		for range height - len(hints) {
			pad := hintCellStyle.Render(" ")
			colHints = append(colHints, pad)
		}

		satisfied := hasCurrent && i < len(curr) && intSliceEqual(hints, curr[i])

		for _, hint := range hints {
			style := hintStyle()
			if satisfied {
				style = hintSatisfiedStyle()
			}
			hintCell := style.Width(cellWidth).
				Align(lipgloss.Center).
				Render(fmt.Sprintf("%d", hint))
			colHints = append(colHints, hintCell)
		}
		renderedCol := lipgloss.JoinVertical(lipgloss.Left, colHints...)
		renderedCols = append(renderedCols, renderedCol)

		if needsSpacer(i, n) {
			sepStyle := lipgloss.NewStyle().Foreground(theme.Current().Border)
			var lines []string
			for range height - 1 {
				lines = append(lines, " ")
			}
			lines = append(lines, sepStyle.Render("\u2502"))
			renderedCols = append(renderedCols, lipgloss.JoinVertical(lipgloss.Left, lines...))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
}

func rowHintView(r TomographyDefinition, width int, current ...TomographyDefinition) string {
	var hasCurrent bool
	var curr TomographyDefinition
	if len(current) > 0 {
		hasCurrent = true
		curr = current[0]
	}

	n := len(r)
	var renderedRows []string
	for i, hints := range r {
		satisfied := hasCurrent && i < len(curr) && intSliceEqual(hints, curr[i])

		style := hintStyle()
		if satisfied {
			style = hintSatisfiedStyle()
		}

		var rowHints []string
		for _, hint := range hints {
			hintCell := fmt.Sprintf("%2d", hint)
			rowHints = append(rowHints, hintCell)
		}
		renderedRow := style.Width(width).
			Align(lipgloss.Right).
			Render(strings.Join(rowHints, " "))
		renderedRows = append(renderedRows, renderedRow)

		if needsSpacer(i, n) {
			sep := lipgloss.NewStyle().Foreground(theme.Current().Border).
				Width(width).
				Align(lipgloss.Right).
				Render("\u2500")
			renderedRows = append(renderedRows, sep)
		}
	}
	s := lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
	return s
}

type boardBlockLayout struct {
	Block      string
	Grid       string
	HintWidth  int
	HintHeight int
	Metrics    game.DynamicGridMetrics
}

func nonogramMetrics(width, height int) game.DynamicGridMetrics {
	return game.DynamicGridMetrics{
		Width:     width,
		Height:    height,
		CellWidth: cellWidth,
		VerticalBridgeWidth: func(x int) int {
			return nonogramVerticalBridgeWidth(width, x)
		},
		HorizontalBridgeHeight: func(y int) int {
			return nonogramHorizontalBridgeHeight(height, y)
		},
	}
}

func nonogramVerticalBridgeWidth(width, x int) int {
	switch {
	case x <= 0, x >= width:
		return 1
	case needsSpacer(x-1, width):
		return 1
	default:
		return 0
	}
}

func nonogramHorizontalBridgeHeight(height, y int) int {
	switch {
	case y <= 0, y >= height:
		return 1
	case needsSpacer(y-1, height):
		return 1
	default:
		return 0
	}
}

func buildBoardBlock(m Model) boardBlockLayout {
	maxWidth := m.rowHints.RequiredLen() * cellWidth
	maxHeight := m.colHints.RequiredLen()
	metrics := nonogramMetrics(m.width, m.height)

	grid := gridView(m.grid, m.cursor, m.solved)
	rowHints := rowHintView(m.rowHints, maxWidth, m.currentHints.rows)
	colHints := colHintView(m.colHints, maxHeight, m.currentHints.cols)
	spacer := lipgloss.NewStyle().Width(maxWidth).Height(maxHeight).Render("")
	topBand := lipgloss.JoinHorizontal(lipgloss.Bottom, spacer, colHints)
	block := lipgloss.JoinVertical(
		lipgloss.Center,
		topBand,
		lipgloss.JoinHorizontal(lipgloss.Top, rowHints, grid),
	)

	return boardBlockLayout{
		Block:      block,
		Grid:       grid,
		HintWidth:  lipgloss.Width(rowHints),
		HintHeight: lipgloss.Height(topBand),
		Metrics:    metrics,
	}
}

func gridView(g grid, c game.Cursor, solved bool) string {
	height := len(g)
	width := 0
	if height > 0 {
		width = len(g[0])
	}
	metrics := nonogramMetrics(width, height)

	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:                  width,
		Height:                 height,
		CellWidth:              cellWidth,
		Solved:                 solved,
		VerticalBridgeWidth:    metrics.VerticalBridgeWidth,
		HorizontalBridgeHeight: metrics.HorizontalBridgeHeight,
		Cell: func(x, y int) string {
			return tileView(g[y][x], x == c.X && y == c.Y, y == c.Y, x == c.X, solved)
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
		HasVerticalEdge: func(x, _ int) bool {
			return x <= 0 || x >= width || needsSpacer(x-1, width)
		},
		HasHorizontalEdge: func(_, y int) bool {
			return y <= 0 || y >= height || needsSpacer(y-1, height)
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return separatorBridgeFill(c, solved, width, height, bridge)
		},
	})
}

func separatorBridgeFill(cursor game.Cursor, solved bool, width, height int, bridge game.DynamicGridBridge) color.Color {
	if solved {
		return nil
	}

	switch bridge.Kind {
	case game.DynamicGridBridgeVertical:
		if bridge.X <= 0 || bridge.X >= width || !needsSpacer(bridge.X-1, width) {
			return nil
		}
	case game.DynamicGridBridgeHorizontal:
		if bridge.Y <= 0 || bridge.Y >= height || !needsSpacer(bridge.Y-1, height) {
			return nil
		}
	default:
		return nil
	}

	if game.DynamicGridBridgeOnCrosshairAxis(cursor, bridge) {
		return theme.Current().Surface
	}
	return nil
}

func tileView(val rune, isCursor, inCursorRow, inCursorCol, solved bool) string {
	p := theme.Current()
	styles := renderStyleMap()
	s, ok := styles[val]
	if !ok {
		s = styles[emptyTile]
	}

	r, ok := renderRuneMap[val]
	if !ok {
		r = renderRuneMap[emptyTile]
	}

	if isCursor && solved {
		s = game.CursorSolvedStyle()
		r = game.CursorLeft + string([]rune(r)[1]) + game.CursorRight
	} else if isCursor {
		s = game.CursorStyle()
		r = game.CursorLeft + string([]rune(r)[1]) + game.CursorRight
	} else if solved {
		s = s.Foreground(p.SolvedFG).Background(p.SuccessBG)
	} else if inCursorRow || inCursorCol {
		// Apply crosshair background tint — filled cells get a more active color
		if val == filledTile {
			s = s.Background(theme.MidTone(p.Surface, p.AccentBG))
		} else {
			s = s.Background(p.Surface)
		}
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  z: fill (hold+move)  x: mark (hold+move)  bkspc: clear  LMB: fill  RMB: mark  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("z: fill  x: mark  bkspc: clear  mouse: click/drag")
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
