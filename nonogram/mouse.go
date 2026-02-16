package nonogram

import (
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

// screenToGrid converts terminal screen coordinates to grid cell coordinates.
// Returns (col, row, ok) where ok is false if the click landed outside the grid
// or on a separator.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.gridOrigin()
	lx := screenX - ox
	ly := screenY - oy
	if lx < 0 || ly < 0 {
		return 0, 0, false
	}

	col = localToCellX(lx, m.width)
	row = localToCellY(ly, m.height)
	if col < 0 || col >= m.width || row < 0 || row >= m.height {
		return 0, 0, false
	}

	return col, row, true
}

// localToCellX converts a local horizontal offset within the grid area to a
// column index, accounting for spacer separators every spacerEvery cells.
// Each cell is cellWidth (3) characters wide; separators are 1 char wide.
// Returns -1 if the position falls on a separator or out of bounds.
func localToCellX(pos, count int) int {
	return localToCell(pos, count, cellWidth)
}

// localToCellY converts a local vertical offset within the grid area to a
// row index. Each row is 1 terminal line tall; separators are 1 line tall.
// Returns -1 if the position falls on a separator or out of bounds.
func localToCellY(pos, count int) int {
	return localToCell(pos, count, 1)
}

// localToCell is the generic conversion for both axes.
// unitWidth is the visual width of one cell (cellWidth for X, 1 for Y).
func localToCell(pos, count, unitWidth int) int {
	if count <= spacerEvery {
		cell := pos / unitWidth
		if cell >= count {
			return -1
		}
		return cell
	}

	blockWidth := spacerEvery*unitWidth + 1
	block := pos / blockWidth
	rem := pos % blockWidth

	cellInBlock := rem / unitWidth
	if cellInBlock >= spacerEvery {
		return -1
	}

	cell := block*spacerEvery + cellInBlock
	if cell >= count {
		return -1
	}
	return cell
}

// gridOrigin computes the screen position of the top-left corner of the
// first grid cell by rendering the same components as View() and measuring
// their dimensions.
func (m *Model) gridOrigin() (x, y int) {
	// Render the same components View() renders to get exact measurements.
	maxWidth := m.rowHints.RequiredLen() * cellWidth
	maxHeight := m.colHints.RequiredLen()

	title := game.TitleBarView("Nonogram", m.modeTitle, m.solved)
	rowHints := rowHintView(m.rowHints, maxWidth, m.currentHints.rows)
	colHints := colHintView(m.colHints, maxHeight, m.currentHints.cols)
	spacer := baseStyle.Width(maxWidth).Height(maxHeight).Render("")
	g := gridView(m.grid, m.cursor, m.solved)
	status := statusBarView(m.showFullHelp)

	// Reproduce the exact layout from View().
	s1 := lipgloss.JoinHorizontal(lipgloss.Bottom, spacer, colHints)
	s2 := lipgloss.JoinHorizontal(lipgloss.Top, rowHints, g)
	gridBlock := lipgloss.JoinVertical(lipgloss.Center, s1, s2)
	fullContent := lipgloss.JoinVertical(lipgloss.Center, title, gridBlock, status)

	// Measure the full content dimensions for centering.
	contentWidth := lipgloss.Width(fullContent)
	contentHeight := lipgloss.Height(fullContent)

	// CenterView uses lipgloss.Place(termWidth, termHeight, Center, Center, ...).
	centerOffsetX := max((m.termWidth-contentWidth)/2, 0)
	centerOffsetY := max((m.termHeight-contentHeight)/2, 0)

	// Grid cell (0,0) is offset from the content's top-left by:
	// - X: the row hint area width
	// - Y: title height + column hint area height
	hintAreaWidth := lipgloss.Width(rowHints)
	titleAreaHeight := lipgloss.Height(title)
	colHintAreaHeight := lipgloss.Height(s1)

	// Within the full content, JoinVertical(Center) pads all lines to max
	// width. The grid block may be centered within that width.
	gridBlockWidth := lipgloss.Width(gridBlock)
	gridBlockPadLeft := max((contentWidth-gridBlockWidth)/2, 0)

	x = centerOffsetX + gridBlockPadLeft + hintAreaWidth
	y = centerOffsetY + titleAreaHeight + colHintAreaHeight
	return x, y
}
