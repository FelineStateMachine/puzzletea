package hitori

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 3

const shadedCellDisplay = " █ "

func cellView(num rune, mark cellMark, isCursor, inCursorRow, inCursorCol, solved, conflict bool) string {
	p := theme.Current()
	s, display := hitoriCellBase(mark, num)

	// Priority: cursor+solved > cursor+conflict > cursor > solved > conflict > crosshair > base.
	if isCursor {
		s = hitoriCursorStyle(mark, solved, conflict)
	} else if solved {
		s = s.Foreground(p.SolvedFG).Background(p.SuccessBG)
	} else if conflict {
		s = s.Foreground(game.ConflictFG()).Background(game.ConflictBG())
	} else if inCursorRow || inCursorCol {
		s = s.Background(p.Surface)
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
}

func hitoriCellBase(mark cellMark, num rune) (lipgloss.Style, string) {
	p := theme.Current()

	switch mark {
	case shaded:
		return lipgloss.NewStyle().
				Foreground(p.Surface).
				Background(p.Surface),
			shadedCellDisplay
	case circled:
		return lipgloss.NewStyle().
				Foreground(p.Info).
				Background(p.BG),
			fmt.Sprintf(" %c ", num)
	default:
		return lipgloss.NewStyle().
				Foreground(p.FG).
				Background(p.BG),
			fmt.Sprintf(" %c ", num)
	}
}

func hitoriCursorStyle(mark cellMark, solved, conflict bool) lipgloss.Style {
	p := theme.Current()
	style := lipgloss.NewStyle().Bold(true)

	switch {
	case solved:
		style = style.Background(p.SuccessBG)
		switch mark {
		case circled:
			style = style.Foreground(p.Info)
		case shaded:
			style = style.Foreground(p.SolvedFG)
		default:
			style = style.Foreground(game.CursorFG())
		}
	case conflict:
		style = style.Background(game.ConflictBG()).Underline(true)
		switch mark {
		case circled:
			style = style.Foreground(p.Info)
		case shaded:
			style = style.Foreground(game.CursorFG())
		default:
			style = style.Foreground(game.CursorFG())
		}
	default:
		style = style.Background(game.CursorBG())
		switch mark {
		case circled:
			style = style.Foreground(p.Info)
		case shaded:
			style = style.Foreground(game.CursorFG())
		default:
			style = style.Foreground(game.CursorFG())
		}
	}

	return style
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
		return game.StatusBarStyle().Render("arrows/wasd: move  x: shade  z: circle  bkspc: clear  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("x: shade  z: circle  bkspc: clear")
}
