package takuzu

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/FelineStateMachine/puzzletea/game"
)

var (
	baseStyle       = lipgloss.NewStyle()
	backgroundColor = compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("235")}
	zeroStyle       = baseStyle.
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("130"), Dark: lipgloss.Color("222")}).
			Background(backgroundColor)

	oneStyle = baseStyle.
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("68"), Dark: lipgloss.Color("111")}).
			Background(backgroundColor)

	emptyStyle = baseStyle.
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("250"), Dark: lipgloss.Color("240")}).
			Background(backgroundColor)

	// cursorSolvedStyle resolved at render time via game.CursorSolvedStyle().

	crosshairBG = compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("237")}
	solvedBG    = compat.AdaptiveColor{Light: lipgloss.Color("151"), Dark: lipgloss.Color("22")}

	borderFG = compat.AdaptiveColor{Light: lipgloss.Color("250"), Dark: lipgloss.Color("240")}
)

const cellWidth = 3

var (
	renderStyleMap = map[rune]lipgloss.Style{
		zeroCell:  zeroStyle,
		oneCell:   oneStyle,
		emptyCell: emptyStyle,
	}

	renderRuneMap = map[rune]string{
		zeroCell:  " ● ",
		oneCell:   " ○ ",
		emptyCell: " · ",
	}
)

func cellView(val rune, isProvided, isCursor, inCursorRow, inCursorCol, solved bool) string {
	s, ok := renderStyleMap[val]
	if !ok {
		s = emptyStyle
	}

	if isProvided && val != emptyCell {
		s = s.Bold(true)
	}

	if isCursor && solved {
		s = game.CursorSolvedStyle()
	} else if isCursor {
		s = s.Background(game.CursorWarmBG()).Bold(true)
	} else if solved {
		s = s.Background(solvedBG)
	} else if inCursorRow || inCursorCol {
		s = s.Background(crosshairBG)
	}

	r, ok := renderRuneMap[val]
	if !ok {
		r = renderRuneMap[emptyCell]
	}
	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func gridView(g grid, provided [][]bool, c game.Cursor, solved bool) string {
	w := 0
	if len(g) > 0 {
		w = len(g[0])
	}

	var rows []string
	for y, row := range g {
		var rowCells []string
		inCursorRow := y == c.Y

		// Left border segment for this row.
		rowCells = append(rowCells, game.BorderChar("│", gridBorderColors, solved, !solved && inCursorRow))

		for x, val := range row {
			isCursor := x == c.X && y == c.Y
			inCursorCol := x == c.X
			cell := cellView(val, provided[y][x], isCursor, inCursorRow, inCursorCol, solved)
			rowCells = append(rowCells, cell)
		}

		// Right border segment for this row.
		rowCells = append(rowCells, game.BorderChar("│", gridBorderColors, solved, !solved && inCursorRow))

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}

	// Build top and bottom border rows with per-column crosshair highlighting.
	topRow := game.HBorderRow(w, c.X, cellWidth, "╭", "╮", gridBorderColors, solved)
	botRow := game.HBorderRow(w, c.X, cellWidth, "╰", "╯", gridBorderColors, solved)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, lipgloss.JoinVertical(lipgloss.Left, rows...), botRow)
}

var gridBorderColors = game.GridBorderColors{
	BorderFG:       borderFG,
	BackgroundBG:   backgroundColor,
	CrosshairBG:    crosshairBG,
	SolvedBorderFG: compat.AdaptiveColor{Light: lipgloss.Color("22"), Dark: lipgloss.Color("149")},
	SolvedBG:       solvedBG,
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  z: ●  x: ○  bkspc: clear  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("z: ●  x: ○  bkspc: clear")
}
