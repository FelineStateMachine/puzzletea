package fillomino

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

// screenToGrid converts terminal coordinates to fillomino grid coordinates.
// Returns false when the click lands on a border row/column or outside the grid.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	lx := screenX - ox
	ly := screenY - oy
	if lx < 0 || ly < 0 {
		return 0, 0, false
	}

	maxX := 0
	if m.width > 0 {
		maxX = (m.width-1)*(cellWidth+1) + (cellWidth - 1)
	}
	maxY := 0
	if m.height > 0 {
		maxY = (m.height - 1) * 2
	}
	if lx > maxX || ly > maxY {
		return 0, 0, false
	}

	col = lx / (cellWidth + 1)
	row = ly / 2
	if lx%(cellWidth+1) >= cellWidth || ly%2 != 0 {
		return 0, 0, false
	}
	if col < 0 || col >= m.width || row < 0 || row >= m.height {
		return 0, 0, false
	}

	return col, row, true
}

func (m *Model) cachedGridOrigin() (x, y int) {
	if m.originValid {
		return m.originX, m.originY
	}
	x, y = m.gridOrigin()
	m.originX, m.originY = x, y
	m.originValid = true
	return x, y
}

func (m *Model) gridOrigin() (x, y int) {
	gameView := m.View()
	viewWidth := lipgloss.Width(gameView)
	viewHeight := lipgloss.Height(gameView)

	if viewWidth > m.termWidth {
		viewWidth = m.termWidth
	}
	if viewHeight > m.termHeight {
		viewHeight = m.termHeight
	}

	centerX := max((m.termWidth-viewWidth)/2, 0)
	centerY := max((m.termHeight-viewHeight)/2, 0)

	title := game.TitleBarView("Fillomino", m.modeTitle, m.solved)
	titleHeight := strings.Count(title, "\n") + 1

	grid := gridView(*m)
	gridWidth := lipgloss.Width(grid)
	gridPadLeft := max((viewWidth-gridWidth)/2, 0)

	const borderLeft = 1
	const borderTop = 1

	x = centerX + gridPadLeft + borderLeft
	y = centerY + titleHeight + borderTop
	return x, y
}
