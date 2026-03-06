package fillomino

import "github.com/FelineStateMachine/puzzletea/game"

// screenToGrid converts terminal coordinates to fillomino grid coordinates.
// Returns false when the click lands on a border row/column or outside the grid.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	return game.DynamicGridScreenToCell(
		game.DynamicGridMetrics{Width: m.width, Height: m.height, CellWidth: cellWidth},
		ox,
		oy,
		screenX,
		screenY,
		false,
	)
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
	title := game.TitleBarView("Fillomino", m.modeTitle, m.solved)
	grid := gridView(*m)
	return game.DynamicGridOrigin(m.termWidth, m.termHeight, m.View(), title, grid)
}
