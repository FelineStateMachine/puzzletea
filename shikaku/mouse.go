package shikaku

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

// screenToGrid converts terminal screen coordinates to grid cell coordinates.
// Returns (col, row, ok) where ok is false if the click landed outside the grid.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	return m.screenToGridAt(screenX, screenY, false)
}

func (m *Model) screenToGridDrag(screenX, screenY int) (col, row int, ok bool) {
	return m.screenToGridAt(screenX, screenY, true)
}

func (m *Model) screenToGridAt(screenX, screenY int, includeSeparators bool) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	lx := screenX - ox
	ly := screenY - oy
	if lx < 0 || ly < 0 {
		return 0, 0, false
	}

	maxX := 0
	if m.puzzle.Width > 0 {
		maxX = (m.puzzle.Width-1)*(cellWidth+1) + (cellWidth - 1)
	}
	maxY := 0
	if m.puzzle.Height > 0 {
		maxY = (m.puzzle.Height - 1) * 2
	}
	if lx > maxX || ly > maxY {
		return 0, 0, false
	}

	if includeSeparators {
		col = min((lx+cellWidth/2)/(cellWidth+1), m.puzzle.Width-1)
		row = min((ly+1)/2, m.puzzle.Height-1)
	} else {
		col = lx / (cellWidth + 1)
		row = ly / 2
		if lx%(cellWidth+1) >= cellWidth || ly%2 != 0 {
			return 0, 0, false
		}
	}

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
//	dynamic grid
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

	// The grid is rendered directly inside the game view. Measure its width to
	// compute horizontal padding from JoinVertical(Center, ...).
	grid := gridView(*m, solved)
	gridOuterWidth := lipgloss.Width(grid)
	gridPadLeft := max((viewWidth-gridOuterWidth)/2, 0)

	// The dynamic grid includes an outer wall, so the first cell content starts
	// one column right and one row below the grid's top-left corner.
	const borderLeft = 1
	const borderTop = 1

	x = centerX + gridPadLeft + borderLeft
	y = centerY + titleHeight + borderTop
	return x, y
}
