package shikaku

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

// screenToGrid converts terminal screen coordinates to grid cell coordinates.
// Returns (col, row, ok) where ok is false if the click landed outside the grid.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	lx := screenX - ox
	ly := screenY - oy
	if lx < 0 || ly < 0 {
		return 0, 0, false
	}

	// Shikaku has no separators: simple division by cell dimensions.
	col = lx / cellWidth
	row = ly // cellHeight is 1

	if col < 0 || col >= m.puzzle.Width || row < 0 || row >= m.puzzle.Height {
		return 0, 0, false
	}

	return col, row, true
}

// cachedGridOrigin returns the screen position of the top-left corner of
// cell (0,0), using a cached value when available. The cache is invalidated
// on terminal resize and solve-state changes.
func (m *Model) cachedGridOrigin() (x, y int) {
	if m.originValid {
		return m.originX, m.originY
	}
	x, y = m.gridOrigin()
	m.originX, m.originY = x, y
	m.originValid = true
	return x, y
}

// gridOrigin computes the screen position of the top-left corner of the
// first grid cell. Mouse coordinates are terminal-absolute, so we need to
// account for how the root model centers the game's View() output.
//
// The root view does:
//
//	lipgloss.Place(termWidth, termHeight, Center, Center, gameView)
//
// So the centering offset is (termWidth - viewWidth) / 2 for X and
// (termHeight - viewHeight) / 2 for Y.
//
// Within the game view the layout is:
//
//	title
//	┌─────────────┐  <- border (1 line top)
//	│ cell grid    │  <- border adds 1 char left
//	└─────────────┘
//	info
//	status
func (m *Model) gridOrigin() (x, y int) {
	gameView := m.View()
	viewWidth := lipgloss.Width(gameView)
	viewHeight := lipgloss.Height(gameView)

	if viewWidth > m.termWidth {
		viewWidth = m.termWidth
	}

	// Centering offset applied by the root's lipgloss.Place().
	centerX := max((m.termWidth-viewWidth)/2, 0)
	centerY := max((m.termHeight-viewHeight)/2, 0)

	// Measure the title height.
	solved := m.puzzle.IsSolved()
	title := game.TitleBarView("Shikaku", m.modeTitle, solved)
	titleHeight := strings.Count(title, "\n") + 1

	// The grid is rendered inside a border. Measure its outer width to
	// compute horizontal padding from JoinVertical(Center, ...).
	grid := gridView(*m, solved)
	gridOuterWidth := lipgloss.Width(grid)
	gridPadLeft := max((viewWidth-gridOuterWidth)/2, 0)

	// Border adds 1 character on the left and 1 line on top.
	const borderLeft = 1
	const borderTop = 1

	x = centerX + gridPadLeft + borderLeft
	y = centerY + titleHeight + borderTop
	return x, y
}
