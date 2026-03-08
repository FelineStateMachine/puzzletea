package sudokurgb

import "github.com/FelineStateMachine/puzzletea/game"

func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	return game.DynamicGridScreenToCell(
		game.DynamicGridMetrics{Width: gridSize, Height: gridSize, CellWidth: cellWidth},
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
	solved := m.isSolved()
	title := game.TitleBarView("Sudoku RGB", m.modeTitle, solved)
	grid := renderGrid(*m, solved)
	return game.DynamicGridOrigin(m.termWidth, m.termHeight, m.View(), title, grid)
}
