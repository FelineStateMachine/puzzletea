package nonogram

import "github.com/FelineStateMachine/puzzletea/game"

func (m *Model) screenToGrid(screenX, screenY int, includeSeparators bool) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	board := buildBoardBlock(*m)
	return game.DynamicGridScreenToCell(
		board.Metrics,
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
	title := game.TitleBarView("Nonogram", m.modeTitle, m.solved)
	board := buildBoardBlock(*m)
	x, y = game.DynamicGridOrigin(m.termWidth, m.termHeight, m.View(), title, board.Block)
	return x + board.HintWidth, y + board.HintHeight
}
