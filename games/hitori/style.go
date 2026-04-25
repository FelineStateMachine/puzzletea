package hitori

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = game.DynamicGridCellWidth

const shadedCellDisplay = " █ "

type cellVisualState int

const (
	cellVisualStateBase cellVisualState = iota
	cellVisualStateCrosshair
	cellVisualStateCursor
	cellVisualStateSolved
	cellVisualStateConflict
)

type cellVisual struct {
	text     string
	fg       color.Color
	bg       color.Color
	bridgeBG color.Color
	bold     bool
	state    cellVisualState
}

func cellView(num rune, mark cellMark, isCursor, inCursorRow, inCursorCol, solved, conflict bool) string {
	return renderCellVisual(resolveCellVisual(num, mark, isCursor, inCursorRow, inCursorCol, solved, conflict))
}

func resolveCellVisual(
	num rune,
	mark cellMark,
	isCursor, inCursorRow, inCursorCol, solved, conflict bool,
) cellVisual {
	p := theme.Current()
	visual := hitoriBaseVisual(mark, num)

	// Priority: cursor+solved > cursor+conflict > cursor > solved > conflict > crosshair > base.
	switch {
	case isCursor:
		return hitoriCursorVisual(num, mark, solved, conflict)
	case solved:
		visual.fg = p.SolvedFG
		visual.bg = p.SuccessBG
		visual.bridgeBG = p.SuccessBG
		visual.state = cellVisualStateSolved
	case conflict:
		visual.text = conflictDisplay(mark, num)
		visual.fg = game.ConflictFG()
		visual.bg = game.ConflictBG()
		visual.bridgeBG = game.ConflictBG()
		visual.state = cellVisualStateConflict
	case inCursorRow || inCursorCol:
		visual.bg = p.Surface
		visual.bridgeBG = p.Surface
		visual.state = cellVisualStateCrosshair
	}

	return visual
}

func hitoriBaseVisual(mark cellMark, num rune) cellVisual {
	p := theme.Current()

	switch mark {
	case shaded:
		return cellVisual{
			text:     shadedCellDisplay,
			fg:       p.Surface,
			bg:       p.Surface,
			bridgeBG: p.Surface,
			state:    cellVisualStateBase,
		}
	case circled:
		return cellVisual{
			text:     fmt.Sprintf("(%c)", num),
			fg:       p.Info,
			bg:       p.BG,
			bridgeBG: p.BG,
			state:    cellVisualStateBase,
		}
	default:
		return cellVisual{
			text:     fmt.Sprintf(" %c ", num),
			fg:       p.FG,
			bg:       p.BG,
			bridgeBG: p.BG,
			state:    cellVisualStateBase,
		}
	}
}

func hitoriCursorVisual(num rune, mark cellMark, solved, conflict bool) cellVisual {
	p := theme.Current()
	visual := hitoriBaseVisual(mark, num)
	visual.bold = true

	switch mark {
	case circled:
		visual.fg = p.Info
	case shaded:
		visual.fg = game.CursorFG()
	default:
		visual.fg = game.CursorFG()
	}

	switch {
	case solved:
		visual.bg = p.SuccessBG
		visual.bridgeBG = p.SuccessBG
		visual.state = cellVisualStateSolved
		if mark == shaded {
			visual.fg = p.SolvedFG
		}
	case conflict:
		visual.text = conflictDisplay(mark, num)
		visual.bg = game.ConflictBG()
		visual.bridgeBG = game.ConflictBG()
		visual.state = cellVisualStateConflict
	default:
		visual.bg = game.CursorBG()
		visual.bridgeBG = p.Surface
		visual.state = cellVisualStateCursor
	}

	return visual
}

func conflictDisplay(mark cellMark, num rune) string {
	switch mark {
	case shaded:
		return "!█!"
	default:
		return fmt.Sprintf("!%c!", num)
	}
}

func renderCellVisual(visual cellVisual) string {
	style := lipgloss.NewStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Foreground(visual.fg).
		Background(visual.bg)
	if visual.bold {
		style = style.Bold(true)
	}
	return style.Render(visual.text)
}

func bridgeFill(numbers grid, marks [][]cellMark, c game.Cursor, solved bool, conflicts [][]bool, bridge game.DynamicGridBridge) color.Color {
	_, _, _ = numbers, marks, conflicts
	if solved {
		return nil
	}

	if game.DynamicGridBridgeOnCrosshairAxis(c, bridge) {
		return theme.Current().Surface
	}

	return nil
}

func gridView(numbers grid, marks [][]cellMark, c game.Cursor, solved bool, conflicts [][]bool) string {
	height := len(numbers)
	width := 0
	if height > 0 {
		width = len(numbers[0])
	}

	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:     width,
		Height:    height,
		CellWidth: cellWidth,
		Solved:    solved,
		Cell: func(x, y int) string {
			return renderCellVisual(resolveCellVisual(
				numbers[y][x],
				marks[y][x],
				x == c.X && y == c.Y,
				y == c.Y,
				x == c.X,
				solved,
				conflicts[y][x],
			))
		},
		HasVerticalEdge: func(x, _ int) bool {
			return x <= 0 || x >= width
		},
		HasHorizontalEdge: func(_, y int) bool {
			return y <= 0 || y >= height
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeFill(numbers, marks, c, solved, conflicts, bridge)
		},
	})
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  mouse: click move  x: shade  z: circle  bkspc: clear  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("mouse: click move  x: shade  z: circle  bkspc: clear")
}
