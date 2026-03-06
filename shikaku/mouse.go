package shikaku

import "github.com/FelineStateMachine/puzzletea/game"

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
	return game.DynamicGridScreenToCell(
		game.DynamicGridMetrics{
			Width:     m.puzzle.Width,
			Height:    m.puzzle.Height,
			CellWidth: cellWidth,
		},
		ox,
		oy,
		screenX,
		screenY,
		includeSeparators,
	)
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
	solved := m.puzzle.IsSolved()
	title := game.TitleBarView("Shikaku", m.modeTitle, solved)
	grid := gridView(*m, solved)
	return game.DynamicGridOrigin(m.termWidth, m.termHeight, m.View(), title, grid)
}
