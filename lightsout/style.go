package lightsout

import (
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle()

	colorOn  = lipgloss.AdaptiveColor{Light: "222", Dark: "180"}
	colorOff = lipgloss.AdaptiveColor{Light: "254", Dark: "236"}

	styleOn = baseStyle.
		Background(colorOn)

	styleOff = baseStyle.
			Background(colorOff)

	cursorStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "130", Dark: "173"})

	solvedStyle = baseStyle.
			Background(lipgloss.AdaptiveColor{Light: "151", Dark: "22"})

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
			BorderBackground(colorOff)

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "22", Dark: "149"}).
				BorderBackground(lipgloss.AdaptiveColor{Light: "151", Dark: "22"})

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "244"}).
			MarginTop(1)
)

const (
	cellWidth  = 4
	cellHeight = 2
)

func cellView(isOn, isCursor, solved bool) string {
	s := styleOff
	if isOn {
		s = styleOn
	}

	if solved {
		s = solvedStyle
	} else if isCursor {
		s = cursorStyle
	}

	content := strings.Repeat(" ", cellWidth)
	return s.Width(cellWidth).Height(cellHeight).Render(content)
}

func gridView(g [][]bool, c game.Cursor, solved bool) string {
	var rows []string
	for y, row := range g {
		var rowCells []string
		for x, cell := range row {
			isCursor := x == c.X && y == c.Y
			rowCells = append(rowCells, cellView(cell, isCursor, solved))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}
	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if solved {
		return gridBorderSolvedStyle.Render(grid)
	}
	return gridBorderStyle.Render(grid)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  enter/space: toggle  ctrl+n: menu  ctrl+h: help")
	}
	return statusBarStyle.Render("enter/space: toggle")
}
