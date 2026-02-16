package hitori

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"

	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle       = lipgloss.NewStyle()
	backgroundColor = lipgloss.AdaptiveColor{Light: "254", Dark: "235"}

	numberStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "250"}).
			Background(backgroundColor)

	shadedStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "238"}).
			Background(lipgloss.AdaptiveColor{Light: "240", Dark: "238"})

	circledStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "25", Dark: "75"}).
			Background(backgroundColor)

	cursorStyle       = game.CursorWarmStyle
	cursorSolvedStyle = game.CursorSolvedStyle

	conflictStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "167"}).
			Background(lipgloss.AdaptiveColor{Light: "224", Dark: "52"})

	crosshairBG = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}
	solvedBG    = lipgloss.AdaptiveColor{Light: "151", Dark: "22"}

	borderFG = lipgloss.AdaptiveColor{Light: "250", Dark: "240"}
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
		s = s.Background(cursorStyle.GetBackground()).
			Foreground(cursorStyle.GetForeground()).
			Bold(true)
	} else if solved {
		s = s.Background(solvedBG)
	} else if conflict {
		s = s.Background(conflictStyle.GetBackground()).
			Foreground(conflictStyle.GetForeground())
	} else if inCursorRow || inCursorCol {
		s = s.Background(crosshairBG)
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
}

var gridBorderColors = game.GridBorderColors{
	BorderFG:       borderFG,
	BackgroundBG:   backgroundColor,
	CrosshairBG:    crosshairBG,
	SolvedBorderFG: lipgloss.AdaptiveColor{Light: "22", Dark: "149"},
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
