package nonogram

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle()

	filledStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "222"}).
			Background(lipgloss.AdaptiveColor{Light: "223", Dark: "58"})

	markedStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "131", Dark: "173"}).
			Background(lipgloss.AdaptiveColor{Light: "224", Dark: "236"})

	emptyStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "235"})

	cursorStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "130", Dark: "214"})

	crosshairBG = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}
	solvedBG    = lipgloss.AdaptiveColor{Light: "151", Dark: "22"}

	hintStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "137", Dark: "137"})

	hintSatisfiedStyle = baseStyle.
				Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "149"})

	nonoStatusBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "244"}).
				MarginTop(1)
)

const cellWidth = 3

var (
	hintCellStyle = baseStyle.Width(cellWidth)

	renderStyleMap = map[rune]lipgloss.Style{
		filledTile: filledStyle,
		markedTile: markedStyle,
		emptyTile:  emptyStyle,
	}

	renderRuneMap = map[rune]string{
		filledTile: " ■ ",
		markedTile: " ✕ ",
		emptyTile:  " · ",
	}
)

func colHintView(c TomographyDefinition, height int, current ...TomographyDefinition) string {
	var hasCurrent bool
	var curr TomographyDefinition
	if len(current) > 0 {
		hasCurrent = true
		curr = current[0]
	}

	var renderedCols []string
	for i, hints := range c {
		var colHints []string
		for range height - len(hints) {
			spacer := hintCellStyle.Render(" ")
			colHints = append(colHints, spacer)
		}

		satisfied := hasCurrent && i < len(curr) && intSliceEqual(hints, curr[i])

		for _, hint := range hints {
			style := hintStyle
			if satisfied {
				style = hintSatisfiedStyle
			}
			hintCell := style.Width(cellWidth).
				Align(lipgloss.Center).
				Render(fmt.Sprintf("%d", hint))
			colHints = append(colHints, hintCell)
		}
		renderedCol := lipgloss.JoinVertical(lipgloss.Left, colHints...)
		renderedCols = append(renderedCols, renderedCol)
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

	var renderedRows []string
	for i, hints := range r {
		satisfied := hasCurrent && i < len(curr) && intSliceEqual(hints, curr[i])

		style := hintStyle
		if satisfied {
			style = hintSatisfiedStyle
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
	}
	s := lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
	return s
}

func gridView(g grid, c game.Cursor, solved bool) string {
	var rows []string
	for y, row := range g {
		var rowBuilder []string
		for x, cell := range row {
			isCursor := x == c.X && y == c.Y
			inCursorRow := y == c.Y
			inCursorCol := x == c.X
			cell := tileView(cell, isCursor, inCursorRow, inCursorCol, solved)
			rowBuilder = append(rowBuilder, cell)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, rowBuilder...)
		rows = append(rows, row)
	}
	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return grid
}

func tileView(val rune, isCursor, inCursorRow, inCursorCol, solved bool) string {
	s, ok := renderStyleMap[val]
	if !ok {
		s = renderStyleMap[emptyTile]
	}

	if isCursor && !solved {
		s = cursorStyle
	} else if !solved && (inCursorRow || inCursorCol) {
		// Apply crosshair background tint
		s = s.Background(crosshairBG)
	}

	if solved {
		// Brighten filled tiles when solved
		s = s.Background(solvedBG)
	}

	r, ok := renderRuneMap[val]
	if !ok {
		r = renderRuneMap[emptyTile]
	}
	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func nonoStatusBarView(showFullHelp bool) string {
	if showFullHelp {
		return nonoStatusBarStyle.Render("arrows/wasd: move  z: fill  x: mark  bkspc: clear  ctrl+n: menu  ctrl+h: help")
	}
	return nonoStatusBarStyle.Render("z: fill  x: mark  bkspc: clear")
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
