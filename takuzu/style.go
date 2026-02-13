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

	borderFG = lipgloss.AdaptiveColor{Light: "250", Dark: "240"}

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
		s = s.Background(cursorStyle.GetBackground()).Bold(true)
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
	w := 0
	if len(g) > 0 {
		w = len(g[0])
	}

	var rows []string
	for y, row := range g {
		var rowCells []string
		inCursorRow := y == c.Y

		// Left border segment for this row.
		rowCells = append(rowCells, borderChar("│", solved, !solved && inCursorRow))

		for x, val := range row {
			isCursor := x == c.X && y == c.Y
			inCursorCol := x == c.X
			cell := cellView(val, provided[y][x], isCursor, inCursorRow, inCursorCol, solved)
			rowCells = append(rowCells, cell)
		}

		// Right border segment for this row.
		rowCells = append(rowCells, borderChar("│", solved, !solved && inCursorRow))

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}

	// Build top and bottom border rows with per-column crosshair highlighting.
	topRow := hBorderRow(w, c.X, "╭", "╮", solved)
	botRow := hBorderRow(w, c.X, "╰", "╯", solved)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, lipgloss.JoinVertical(lipgloss.Left, rows...), botRow)
}

// borderChar renders a single border character with optional crosshair highlighting.
func borderChar(ch string, solved, highlight bool) string {
	s := baseStyle.Foreground(borderFG).Background(backgroundColor)
	if solved {
		s = baseStyle.Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "149"}).Background(solvedBG)
	} else if highlight {
		s = s.Background(crosshairBG)
	}
	return s.Render(ch)
}

// hBorderRow builds a horizontal border row (top or bottom) with per-column crosshair.
func hBorderRow(w, cursorX int, left, right string, solved bool) string {
	var parts []string
	parts = append(parts, borderChar(left, solved, false))
	segment := "───" // cellWidth = 3
	for x := range w {
		highlight := !solved && x == cursorX
		s := baseStyle.Foreground(borderFG).Background(backgroundColor)
		if solved {
			s = baseStyle.Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "149"}).Background(solvedBG)
		} else if highlight {
			s = s.Background(crosshairBG)
		}
		parts = append(parts, s.Render(segment))
	}
	parts = append(parts, borderChar(right, solved, false))
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  z: ●  x: ○  bkspc: clear  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return statusBarStyle.Render("z: ●  x: ○  bkspc: clear")
}
