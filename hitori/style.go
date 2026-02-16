package hitori

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

var (
	baseStyle       = lipgloss.NewStyle()
	backgroundColor = compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("235")}

	numberStyle = baseStyle.
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("250")}).
			Background(backgroundColor)

	shadedStyle = baseStyle.
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("238")}).
			Background(compat.AdaptiveColor{Light: lipgloss.Color("240"), Dark: lipgloss.Color("238")})

	circledStyle = baseStyle.
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("25"), Dark: lipgloss.Color("75")}).
			Background(backgroundColor)

	cursorSolvedStyle = game.CursorSolvedStyle

	crosshairBG = compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("237")}
	solvedBG    = compat.AdaptiveColor{Light: lipgloss.Color("151"), Dark: lipgloss.Color("22")}

	borderFG = compat.AdaptiveColor{Light: lipgloss.Color("250"), Dark: lipgloss.Color("240")}
)

const cellWidth = 3

func cellView(num rune, mark cellMark, isCursor, inCursorRow, inCursorCol, solved, conflict bool) string {
	var s lipgloss.Style
	var display string

	switch mark {
	case shaded:
		s = shadedStyle
		display = " ▒ "
	case circled:
		s = circledStyle
		display = fmt.Sprintf(" %c ", num)
	default:
		s = numberStyle
		display = fmt.Sprintf(" %c ", num)
	}

	// Priority: cursor+solved > cursor > conflict > crosshair > solved > base.
	if isCursor && solved {
		s = cursorSolvedStyle
	} else if isCursor {
		s = s.Background(game.CursorWarmBG).
			Foreground(game.CursorFG).
			Bold(true)
	} else if solved {
		s = s.Background(solvedBG)
	} else if conflict {
		s = s.Background(game.ConflictBG).
			Foreground(game.ConflictFG)
	} else if inCursorRow || inCursorCol {
		s = s.Background(crosshairBG)
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
}

var gridBorderColors = game.GridBorderColors{
	BorderFG:       borderFG,
	BackgroundBG:   backgroundColor,
	CrosshairBG:    crosshairBG,
	SolvedBorderFG: compat.AdaptiveColor{Light: lipgloss.Color("22"), Dark: lipgloss.Color("149")},
	SolvedBG:       solvedBG,
}

func gridView(numbers grid, marks [][]cellMark, c game.Cursor, solved bool, conflicts [][]bool) string {
	w := 0
	if len(numbers) > 0 {
		w = len(numbers[0])
	}

	var rows []string
	for y, row := range numbers {
		var rowCells []string
		inCursorRow := y == c.Y

		rowCells = append(rowCells, game.BorderChar("│", gridBorderColors, solved, !solved && inCursorRow))

		for x, num := range row {
			isCursor := x == c.X && y == c.Y
			inCursorCol := x == c.X
			cell := cellView(num, marks[y][x], isCursor, inCursorRow, inCursorCol, solved, conflicts[y][x])
			rowCells = append(rowCells, cell)
		}

		rowCells = append(rowCells, game.BorderChar("│", gridBorderColors, solved, !solved && inCursorRow))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}

	topRow := game.HBorderRow(w, c.X, cellWidth, "╭", "╮", gridBorderColors, solved)
	botRow := game.HBorderRow(w, c.X, cellWidth, "╰", "╯", gridBorderColors, solved)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, lipgloss.JoinVertical(lipgloss.Left, rows...), botRow)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle.Render("arrows/wasd: move  x: shade  z: circle  bkspc: clear  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle.Render("x: shade  z: circle  bkspc: clear")
}
