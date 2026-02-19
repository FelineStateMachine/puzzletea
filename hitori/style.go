package hitori

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 3

func cellView(num rune, mark cellMark, isCursor, inCursorRow, inCursorCol, solved, conflict bool) string {
	p := theme.Current()
	var s lipgloss.Style
	var display string

	switch mark {
	case shaded:
		s = lipgloss.NewStyle().
			Foreground(p.Surface).
			Background(p.Surface)
		display = "   "
	case circled:
		s = lipgloss.NewStyle().
			Foreground(p.Info).
			Background(p.BG)
		display = fmt.Sprintf(" %c ", num)
	default:
		s = lipgloss.NewStyle().
			Foreground(p.FG).
			Background(p.BG)
		display = fmt.Sprintf(" %c ", num)
	}

	// Priority: cursor+solved > cursor > solved > conflict > crosshair > base.
	if isCursor && solved {
		s = game.CursorSolvedStyle()
		display = game.CursorLeft + fmt.Sprintf("%c", num) + game.CursorRight
	} else if isCursor {
		s = game.CursorStyle()
		display = game.CursorLeft + fmt.Sprintf("%c", num) + game.CursorRight
	} else if solved {
		s = s.Foreground(p.SolvedFG).Background(p.SuccessBG)
	} else if conflict {
		s = s.Foreground(game.ConflictFG()).Background(game.ConflictBG())
	} else if inCursorRow || inCursorCol {
		s = s.Background(p.Surface)
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
}

func gridView(numbers grid, marks [][]cellMark, c game.Cursor, solved bool, conflicts [][]bool) string {
	colors := game.DefaultBorderColors()

	w := 0
	if len(numbers) > 0 {
		w = len(numbers[0])
	}

	var rows []string
	for y, row := range numbers {
		var rowCells []string
		inCursorRow := y == c.Y

		rowCells = append(rowCells, game.BorderChar("\u2502", colors, solved, !solved && inCursorRow))

		for x, num := range row {
			isCursor := x == c.X && y == c.Y
			inCursorCol := x == c.X
			cell := cellView(num, marks[y][x], isCursor, inCursorRow, inCursorCol, solved, conflicts[y][x])
			rowCells = append(rowCells, cell)
		}

		rowCells = append(rowCells, game.BorderChar("\u2502", colors, solved, !solved && inCursorRow))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}

	topRow := game.HBorderRow(w, c.X, cellWidth, "\u256d", "\u256e", colors, solved)
	botRow := game.HBorderRow(w, c.X, cellWidth, "\u2570", "\u256f", colors, solved)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, lipgloss.JoinVertical(lipgloss.Left, rows...), botRow)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  x: shade  z: circle  bkspc: clear  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("x: shade  z: circle  bkspc: clear")
}
