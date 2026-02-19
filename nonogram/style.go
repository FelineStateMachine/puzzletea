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
	cellWidth   = 3
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

// hSeparator builds a horizontal separator row using box-drawing characters.
// w is the number of grid columns. cursorX is the cursor column (-1 to disable crosshair).
// bg is the default background, crossBG is the crosshair-highlighted background.
func hSeparator(w, cursorX int, bg, crossBG color.Color) string {
	p := theme.Current()
	defStyle := lipgloss.NewStyle().Foreground(p.Border).Background(bg)
	highStyle := lipgloss.NewStyle().Foreground(p.Border).Background(crossBG)
	segment := strings.Repeat("\u2500", cellWidth)

	var parts []string
	for x := range w {
		style := defStyle
		if x == cursorX {
			style = highStyle
		}
		parts = append(parts, style.Render(segment))
		if needsSpacer(x, w) {
			parts = append(parts, defStyle.Render("\u253c"))
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

func gridView(g grid, c game.Cursor, solved bool) string {
	p := theme.Current()

	h := len(g)
	w := 0
	if h > 0 {
		w = len(g[0])
	}

	sepStyle := lipgloss.NewStyle().Foreground(p.Border)

	// Determine background for horizontal separators (matches grid background).
	gridBG := p.BG
	crosshairBG := p.Surface
	if solved {
		gridBG = p.SuccessBG
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
				rowBuilder = append(rowBuilder, sepStyle.Background(bg).Render("\u2502"))
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
		return game.StatusBarStyle().Render("arrows/wasd: move  z: fill (hold+move)  x: mark (hold+move)  bkspc: clear  LMB: fill  RMB: mark  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
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
