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
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#4a4a4a"))

	markedStyle = baseStyle.
			Foreground(lipgloss.Color("#ff6b6b")).
			Background(lipgloss.Color("#2a1a1a"))

	emptyStyle = baseStyle.
			Foreground(lipgloss.Color("#333333")).
			Background(lipgloss.Color("#1a1a1a"))

	cursorStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#ff00ff"))

	crosshairBG = lipgloss.Color("#252525")

	hintStyle = baseStyle.
			Foreground(lipgloss.Color("#888888"))

	hintSatisfiedStyle = baseStyle.
				Foreground(lipgloss.Color("#00ff00"))

	nonoStatusBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				MarginTop(1)

	cellWidth = 3

	hintCellStyle = baseStyle.Width(cellWidth)

	renderStyleMap = map[rune]lipgloss.Style{
		filledTile: filledStyle,
		markedTile: markedStyle,
		emptyTile:  emptyStyle,
	}

	renderRuneMap = map[rune]string{
		filledTile: "▐█",
		markedTile: "✕",
		emptyTile:  " ",
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
		s = s.Background(lipgloss.Color("#2a6a2a"))
	}

	r, ok := renderRuneMap[val]
	if !ok {
		r = renderRuneMap[emptyTile]
	}
	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func nonoTitleBarView(modeName string, solved bool) string {
	return game.TitleBarView("Nonogram", modeName, solved)
}

func nonoStatusBarView(_ KeyMap) string {
	return nonoStatusBarStyle.Render("arrows/wasd: move  z: fill  x: mark  bkspc: clear  ctrl+n: menu  ctrl+e: debug")
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
