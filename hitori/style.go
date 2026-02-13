package hitori

import (
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle()

	clueStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "250"})

	shadedStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "16", Dark: "16"})

	emptyStyle = baseStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "235"})

	cursorStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "130", Dark: "214"})

	crosshairBG       = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}
	crosshairShadedBG = lipgloss.AdaptiveColor{Light: "250", Dark: "100"}
	solvedBG          = lipgloss.AdaptiveColor{Light: "151", Dark: "22"}

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "244"}).
			MarginTop(1)

	separatorFG = lipgloss.AdaptiveColor{Light: "250", Dark: "240"}
)

const (
	cellWidth   = 3
	spacerEvery = 5
)

func gridView(g grid, provided [][]bool, c game.Cursor, solved bool) string {
	size := len(g)
	sepStyle := baseStyle.Foreground(separatorFG)

	gridBG := emptyStyle.GetBackground()
	if solved {
		gridBG = solvedBG
	}

	var rows []string
	for y := 0; y < size; y++ {
		inCursorRow := y == c.Y

		var rowBuilder []string
		for x := 0; x < size; x++ {
			isCursor := x == c.X && y == c.Y
			inCursorCol := x == c.X
			cell := cellView(g[y][x], provided[y][x], isCursor, inCursorRow, inCursorCol, solved)
			rowBuilder = append(rowBuilder, cell)

			if needsSpacer(x, size) {
				bg := gridBG
				if !solved && inCursorRow {
					bg = crosshairBG
				}
				rowBuilder = append(rowBuilder, sepStyle.Background(bg).Render("│"))
			}
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, rowBuilder...)
		rows = append(rows, row)

		if needsSpacer(y, size) {
			cursorX := -1
			if !solved {
				cursorX = c.X
			}
			rows = append(rows, hSeparator(size, cursorX, gridBG))
		}
	}
	gridStr := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return gridStr
}

func cellView(val rune, isProvided, isCursor, inCursorRow, inCursorCol, solved bool) string {
	var s lipgloss.Style

	switch {
	case val == shadedCell:
		s = shadedStyle
	case val == emptyCell:
		s = emptyStyle
	default:
		s = clueStyle
	}

	if isCursor && !solved {
		s = cursorStyle
	} else if !solved && (inCursorRow || inCursorCol) {
		if val == shadedCell {
			s = s.Background(crosshairShadedBG)
		} else {
			s = s.Background(crosshairBG)
		}
	}

	if solved && val != shadedCell {
		s = s.Background(solvedBG)
	}

	var display string
	switch val {
	case shadedCell:
		display = " ■ "
	case emptyCell:
		display = " · "
	default:
		display = " " + string(val) + " "
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
}

func hSeparator(size, cursorX int, bg lipgloss.TerminalColor) string {
	defStyle := baseStyle.Foreground(separatorFG).Background(bg)
	highStyle := baseStyle.Foreground(separatorFG).Background(crosshairBG)
	segment := strings.Repeat("─", cellWidth)

	var parts []string
	for x := 0; x < size; x++ {
		style := defStyle
		if x == cursorX {
			style = highStyle
		}
		parts = append(parts, style.Render(segment))
		if needsSpacer(x, size) {
			parts = append(parts, defStyle.Render("┼"))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func needsSpacer(i, n int) bool {
	return n > spacerEvery && (i+1)%spacerEvery == 0 && i < n-1
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  z: shade  bkspc: clear  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return statusBarStyle.Render("z: shade  bkspc: clear")
}
