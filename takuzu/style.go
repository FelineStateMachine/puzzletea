package takuzu

import (
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 3

var renderRuneMap = map[rune]string{
	zeroCell:  " \u25cf ",
	oneCell:   " \u25cb ",
	emptyCell: " \u00b7 ",
}

func zeroStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.Accent).
		Background(p.BG)
}

func oneStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.Secondary).
		Background(p.BG)
}

func emptyStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.TextDim).
		Background(p.BG)
}

func renderStyleMap() map[rune]lipgloss.Style {
	return map[rune]lipgloss.Style{
		zeroCell:  zeroStyle(),
		oneCell:   oneStyle(),
		emptyCell: emptyStyle(),
	}
}

func cellView(val rune, isProvided, isCursor, inCursorRow, inCursorCol, solved bool) string {
	p := theme.Current()
	styles := renderStyleMap()
	s, ok := styles[val]
	if !ok {
		s = emptyStyle()
	}

	if isProvided && val != emptyCell {
		s = s.Bold(true)
	}

	r, ok := renderRuneMap[val]
	if !ok {
		r = renderRuneMap[emptyCell]
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
		s = s.Background(p.Surface)
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func gridView(g grid, provided [][]bool, c game.Cursor, solved bool) string {
	colors := game.DefaultBorderColors()

	w := 0
	if len(g) > 0 {
		w = len(g[0])
	}

	var rows []string
	for y, row := range g {
		var rowCells []string
		inCursorRow := y == c.Y

		// Left border segment for this row.
		rowCells = append(rowCells, game.BorderChar("\u2502", colors, solved, !solved && inCursorRow))

		for x, val := range row {
			isCursor := x == c.X && y == c.Y
			inCursorCol := x == c.X
			cell := cellView(val, provided[y][x], isCursor, inCursorRow, inCursorCol, solved)
			rowCells = append(rowCells, cell)
		}

		// Right border segment for this row.
		rowCells = append(rowCells, game.BorderChar("\u2502", colors, solved, !solved && inCursorRow))

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}

	// Build top and bottom border rows with per-column crosshair highlighting.
	topRow := game.HBorderRow(w, c.X, cellWidth, "\u256d", "\u256e", colors, solved)
	botRow := game.HBorderRow(w, c.X, cellWidth, "\u2570", "\u256f", colors, solved)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, lipgloss.JoinVertical(lipgloss.Left, rows...), botRow)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  z: \u25cf  x: \u25cb  bkspc: clear  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("z: \u25cf  x: \u25cb  bkspc: clear")
}
