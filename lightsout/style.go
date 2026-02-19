package lightsout

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
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

const (
	cellWidth  = 4
	cellHeight = 2
)

func lightOnStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Background(theme.Blend(p.BG, p.Accent, 0.85))
}

func lightOffStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().BG)
}

func solvedCellStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().SuccessBG)
}

func gridBorderStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		BorderBackground(p.BG)
}

func gridBorderSolvedStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.SuccessBorder).
		BorderBackground(p.SuccessBG)
}

func cursorOnStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Background(theme.Blend(p.Accent, p.AccentText, 0.25))
}

func cursorOffStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Background(theme.Blend(p.BG, p.Accent, 0.45))
}

func cellView(isOn, isCursor, solved bool) string {
	s := lightOffStyle()
	if isOn {
		s = lightOnStyle()
	}

	if isCursor && solved {
		s = game.CursorSolvedStyle()
	} else if solved {
		s = solvedCellStyle()
	} else if isCursor && isOn {
		s = cursorOnStyle()
	} else if isCursor {
		s = cursorOffStyle()
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
		return gridBorderSolvedStyle().Render(grid)
	}
	return gridBorderStyle().Render(grid)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  enter/space/click: toggle  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("enter/space/click: toggle")
}
