package takuzu

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle       = lipgloss.NewStyle()
	backgroundColor = lipgloss.AdaptiveColor{Light: "254", Dark: "235"}
	zeroStyle       = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "222"}).
			Background(backgroundColor)

	oneStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "68", Dark: "111"}).
			Background(backgroundColor)

	emptyStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
			Background(backgroundColor)

	cursorStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "130", Dark: "214"})

	crosshairBG = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}
	solvedBG    = lipgloss.AdaptiveColor{Light: "151", Dark: "22"}

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
			BorderBackground(backgroundColor)

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "22", Dark: "149"}).
				BorderBackground(backgroundColor)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "244"}).
			MarginTop(1)
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

	if isCursor && !solved {
		s = cursorStyle
	} else if !solved && (inCursorRow || inCursorCol) {
		s = s.Background(crosshairBG)
	}

	if solved {
		s = s.Background(solvedBG)
	}

	r, ok := renderRuneMap[val]
	if !ok {
		r = renderRuneMap[emptyCell]
	}
	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func gridView(g grid, provided [][]bool, c game.Cursor, solved bool) string {
	var rows []string
	for y, row := range g {
		var rowCells []string
		for x, val := range row {
			isCursor := x == c.X && y == c.Y
			inCursorRow := y == c.Y
			inCursorCol := x == c.X
			cell := cellView(val, provided[y][x], isCursor, inCursorRow, inCursorCol, solved)
			rowCells = append(rowCells, cell)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if solved {
		return gridBorderSolvedStyle.Render(content)
	}
	return gridBorderStyle.Render(content)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  z: ●  x: ○  bkspc: clear  ctrl+n: menu  ctrl+h: help")
	}
	return statusBarStyle.Render("z: ●  x: ○  bkspc: clear")
}
