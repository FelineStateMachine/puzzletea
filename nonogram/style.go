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

	crosshairBG       = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}
	crosshairFilledBG = lipgloss.AdaptiveColor{Light: "223", Dark: "100"}
	solvedBG          = lipgloss.AdaptiveColor{Light: "151", Dark: "22"}

	hintStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "137", Dark: "137"})

	hintSatisfiedStyle = baseStyle.
				Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "149"})

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "244"}).
			MarginTop(1)

	separatorFG = lipgloss.AdaptiveColor{Light: "250", Dark: "240"}
)

const (
	cellWidth   = 3
	spacerEvery = 5
)

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

// needsSpacer reports whether a separator should be inserted after index i
// in a dimension of size n (i.e. after every spacerEvery cells, except the last).
func needsSpacer(i, n int) bool {
	return n > spacerEvery && (i+1)%spacerEvery == 0 && i < n-1
}

// hSeparator builds a horizontal separator row using box-drawing characters.
// w is the number of grid columns. cursorX is the cursor column (-1 to disable crosshair).
// bg is the default background, crossBG is the crosshair-highlighted background.
func hSeparator(w, cursorX int, bg, crossBG lipgloss.TerminalColor) string {
	defStyle := baseStyle.Foreground(separatorFG).Background(bg)
	highStyle := baseStyle.Foreground(separatorFG).Background(crossBG)
	segment := strings.Repeat("─", cellWidth)

	var parts []string
	for x := range w {
		style := defStyle
		if x == cursorX {
			style = highStyle
		}
		parts = append(parts, style.Render(segment))
		if needsSpacer(x, w) {
			parts = append(parts, defStyle.Render("┼"))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func colHintView(c TomographyDefinition, height int, current ...TomographyDefinition) string {
	var hasCurrent bool
	var curr TomographyDefinition
	if len(current) > 0 {
		hasCurrent = true
		curr = current[0]
	}

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

		if needsSpacer(i, n) {
			sepStyle := baseStyle.Foreground(separatorFG)
			var lines []string
			for range height - 1 {
				lines = append(lines, " ")
			}
			lines = append(lines, sepStyle.Render("│"))
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

		if needsSpacer(i, n) {
			sep := baseStyle.Foreground(separatorFG).
				Width(width).
				Align(lipgloss.Right).
				Render("─")
			renderedRows = append(renderedRows, sep)
		}
	}
	s := lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
	return s
}

func gridView(g grid, c game.Cursor, solved bool) string {
	h := len(g)
	w := 0
	if h > 0 {
		w = len(g[0])
	}

	sepStyle := baseStyle.Foreground(separatorFG)

	// Determine background for horizontal separators (matches grid background).
	gridBG := emptyStyle.GetBackground()
	if solved {
		gridBG = solvedBG
	}

	var rows []string
	for y, row := range g {
		inCursorRow := y == c.Y

		var rowBuilder []string
		for x, cell := range row {
			isCursor := x == c.X && y == c.Y
			inCursorCol := x == c.X
			cell := tileView(cell, isCursor, inCursorRow, inCursorCol, solved)
			rowBuilder = append(rowBuilder, cell)

			if needsSpacer(x, w) {
				bg := gridBG
				if !solved && inCursorRow {
					bg = crosshairBG
				}
				rowBuilder = append(rowBuilder, sepStyle.Background(bg).Render("│"))
			}
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, rowBuilder...)
		rows = append(rows, row)

		if needsSpacer(y, h) {
			cursorX := -1
			if !solved {
				cursorX = c.X
			}
			rows = append(rows, hSeparator(w, cursorX, gridBG, crosshairBG))
		}
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
		// Apply crosshair background tint — filled cells get a more active color
		if val == filledTile {
			s = s.Background(crosshairFilledBG)
		} else {
			s = s.Background(crosshairBG)
		}
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

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  z: fill  x: mark  bkspc: clear  ctrl+n: menu  ctrl+h: help")
	}
	return statusBarStyle.Render("z: fill  x: mark  bkspc: clear")
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
