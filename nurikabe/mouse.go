package nurikabe

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

// screenToGrid converts terminal screen coordinates to grid cell coordinates.
// Returns (col, row, ok) where ok is false if click landed outside a cell.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	lx := screenX - ox
	ly := screenY - oy
	if lx < 0 || ly < 0 {
		return 0, 0, false
	}

	col = lx / cellWidth
	row = ly

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

	centerX := max((m.termWidth-viewWidth)/2, 0)
	centerY := max((m.termHeight-viewHeight)/2, 0)

	title := game.TitleBarView("Nurikabe", m.modeTitle, m.solved)
	titleHeight := strings.Count(title, "\n") + 1

	gridOuterWidth := lipgloss.Width(gridView(*m))
	gridPadLeft := max((viewWidth-gridOuterWidth)/2, 0)

	const borderLeft = 1
	const borderTop = 1

	x = centerX + gridPadLeft + borderLeft
	y = centerY + titleHeight + borderTop
	return x, y
}
