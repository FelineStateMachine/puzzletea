package nonogram

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

// screenToGrid converts terminal screen coordinates to grid cell coordinates.
// Returns (col, row, ok) where ok is false if the click landed outside the grid
// or on a separator.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
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
// Within the game view, the grid cell (0,0) is offset by the row hint
// area width (X) and the title + column hint area height (Y).
func (m *Model) gridOrigin() (x, y int) {
	// Render the game view to get its exact dimensions.
	gameView := m.View()
	viewWidth := lipgloss.Width(gameView)
	viewHeight := lipgloss.Height(gameView)

	// The root view caps the game view to terminal width before centering.
	if viewWidth > m.termWidth {
		viewWidth = m.termWidth
	}

	// Centering offset applied by the root's lipgloss.Place().
	centerX := max((m.termWidth-viewWidth)/2, 0)
	centerY := max((m.termHeight-viewHeight)/2, 0)

	// Within the game view, find where cell (0,0) starts.
	// Render the subcomponents to measure their sizes.
	maxWidth := m.rowHints.RequiredLen() * cellWidth
	maxHeight := m.colHints.RequiredLen()

	title := game.TitleBarView("Nonogram", m.modeTitle, m.solved)
	rowHints := rowHintView(m.rowHints, maxWidth, m.currentHints.rows)
	colHints := colHintView(m.colHints, maxHeight, m.currentHints.cols)
	spacer := lipgloss.NewStyle().Width(maxWidth).Height(maxHeight).Render("")

	s1 := lipgloss.JoinHorizontal(lipgloss.Bottom, spacer, colHints)
	gridBlock := lipgloss.JoinVertical(lipgloss.Center, s1, "placeholder")

	// The grid block may be narrower than the full view (title or status
	// could be wider). JoinVertical(Center) pads to the widest line.
	gridBlockWidth := lipgloss.Width(gridBlock)
	gridBlockPadLeft := max((viewWidth-gridBlockWidth)/2, 0)

	// The row hint width within the grid block.
	hintAreaWidth := lipgloss.Width(rowHints)

	// Vertical: title lines + column hint area lines.
	titleHeight := strings.Count(title, "\n") + 1
	colHintAreaHeight := strings.Count(s1, "\n") + 1

	x = centerX + gridBlockPadLeft + hintAreaWidth
	y = centerY + titleHeight + colHintAreaHeight
	return x, y
}
