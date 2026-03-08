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

func (m *Model) lockedDragTarget(rawCol, rawRow int) (col, row int, ok bool) {
	if m.dragging == 0 {
		return 0, 0, false
	}

	if m.dragAxis == dragAxisHorizontal {
		return rawCol, m.dragStartRow, true
	}
	if m.dragAxis == dragAxisVertical {
		return m.dragStartCol, rawRow, true
	}

	colDelta := abs(rawCol - m.dragStartCol)
	rowDelta := abs(rawRow - m.dragStartRow)

	switch {
	case colDelta == 0 && rowDelta == 0:
		return m.dragStartCol, m.dragStartRow, false
	case rowDelta == 0:
		m.dragAxis = dragAxisHorizontal
		return rawCol, m.dragStartRow, true
	case colDelta == 0:
		m.dragAxis = dragAxisVertical
		return m.dragStartCol, rawRow, true
	case colDelta > rowDelta:
		m.dragAxis = dragAxisHorizontal
		return rawCol, m.dragStartRow, true
	case rowDelta > colDelta:
		m.dragAxis = dragAxisVertical
		return m.dragStartCol, rawRow, true
	default:
		return m.dragStartCol, m.dragStartRow, false
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
