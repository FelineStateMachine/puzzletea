package nurikabe

import (
	"github.com/FelineStateMachine/puzzletea/game"
)

// screenToGrid converts terminal screen coordinates to grid cell coordinates.
// Returns (col, row, ok) where ok is false if click landed outside a cell.
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
			Width:     m.width,
			Height:    m.height,
			CellWidth: cellWidth,
		},
		ox,
		oy,
		screenX,
		screenY,
		includeSeparators,
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
	title := game.TitleBarView("Nurikabe", m.modeTitle, m.solved)
	grid := gridView(*m)
	return game.DynamicGridOrigin(m.termWidth, m.termHeight, m.View(), title, grid)
}
