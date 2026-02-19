package lightsout

import (
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/FelineStateMachine/puzzletea/game"
)

// screenToGrid converts terminal screen coordinates to grid cell coordinates.
// Returns (col, row, ok) where ok is false if the click landed outside the grid.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	gameView := m.View()
	viewWidth := lipgloss.Width(gameView)
	viewHeight := lipgloss.Height(gameView)

	if viewWidth > m.termWidth {
		viewWidth = m.termWidth
	}

	centerX := max((m.termWidth-viewWidth)/2, 0)
	centerY := max((m.termHeight-viewHeight)/2, 0)

	// Title bar is always 2 lines (title + blank/subtitle line).
	titleHeight := strings.Count(game.TitleBarView("Lights Out", m.modeTitle, m.IsSolved()), "\n") + 1

	// The grid border adds 1 character on each side.
	const borderSize = 1

	// Grid origin within the centered view.
	gridInnerWidth := m.width * cellWidth
	gridOuterWidth := gridInnerWidth + 2*borderSize
	gridPadLeft := max((viewWidth-gridOuterWidth)/2, 0)

	gridX := centerX + gridPadLeft + borderSize
	gridY := centerY + titleHeight + borderSize

	lx := screenX - gridX
	ly := screenY - gridY
	if lx < 0 || ly < 0 {
		return 0, 0, false
	}

	col = lx / cellWidth
	row = ly / cellHeight
	if col >= m.width || row >= m.height {
		return 0, 0, false
	}

	return col, row, true
}

var (
	baseStyle = lipgloss.NewStyle()

	colorOn  = compat.AdaptiveColor{Light: lipgloss.Color("222"), Dark: lipgloss.Color("180")}
	colorOff = compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("236")}

	styleOn = baseStyle.
		Background(colorOn)

	styleOff = baseStyle.
			Background(colorOff)

	// cursorStyle and cursorSolvedStyle resolved at render time
	// via game.CursorStyle() and game.CursorSolvedStyle().

	solvedStyle = baseStyle.
			Background(compat.AdaptiveColor{Light: lipgloss.Color("151"), Dark: lipgloss.Color("22")})

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(compat.AdaptiveColor{Light: lipgloss.Color("250"), Dark: lipgloss.Color("240")}).
			BorderBackground(colorOff)

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(compat.AdaptiveColor{Light: lipgloss.Color("22"), Dark: lipgloss.Color("149")}).
				BorderBackground(compat.AdaptiveColor{Light: lipgloss.Color("151"), Dark: lipgloss.Color("22")})
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

	if isCursor && solved {
		s = game.CursorSolvedStyle()
	} else if solved {
		s = solvedStyle
	} else if isCursor {
		s = game.CursorStyle()
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
		return game.StatusBarStyle().Render("arrows/wasd: move  enter/space/click: toggle  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("enter/space/click: toggle")
}
