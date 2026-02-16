package wordsearch

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

const cellWidth = 3

// screenToGrid converts terminal screen coordinates to grid cell coordinates.
// Returns (col, row, ok) where ok is false if the click landed outside the grid.
func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	lx := screenX - ox
	ly := screenY - oy
	if lx < 0 || ly < 0 {
		return 0, 0, false
	}

	// No separators: simple division by cell dimensions.
	col = lx / cellWidth
	row = ly // cellHeight is 1

	if col < 0 || col >= m.width || row < 0 || row >= m.height {
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
// first grid cell.
//
// The root view does:
//
//	lipgloss.Place(termWidth, termHeight, Center, Center, gameView)
//
// Within the game view the layout is:
//
//	title
//	┌─────────────┐  spacer  ┌──────────────┐
//	│ letter grid  │          │  word list    │
//	└─────────────┘          └──────────────┘
//	status
//
// The grid + spacer + word list form `mainView`, which is centered as a
// block within the overall view.
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
	title := game.TitleBarView("Word Search", m.modeTitle, m.solved)
	titleHeight := strings.Count(title, "\n") + 1

	// The mainView (grid + spacer + word list) is centered within the
	// overall view. Measure its width to compute left padding.
	gridContent := renderGrid(*m)
	gBorder := gridBorderStyle
	if m.solved {
		gBorder = gridBorderSolvedStyle
	}
	gridRendered := gBorder.Render(gridContent)

	wordListContent := renderWordList(*m)
	wBorder := wordListBorderStyle
	if m.solved {
		wBorder = wordListBorderSolvedStyle
	}
	gridHeight := lipgloss.Height(gridRendered)
	wBorder = wBorder.Height(gridHeight - 2)
	wordListRendered := wBorder.Render(wordListContent)

	spacer := strings.Repeat(" ", 2)
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, gridRendered, spacer, wordListRendered)
	mainViewWidth := lipgloss.Width(mainView)

	// The mainView block is centered within the full view width.
	mainViewPadLeft := max((viewWidth-mainViewWidth)/2, 0)

	// Border adds 1 character on the left and 1 line on top.
	const borderLeft = 1
	const borderTop = 1

	x = centerX + mainViewPadLeft + borderLeft
	y = centerY + titleHeight + borderTop
	return x, y
}
