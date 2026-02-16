package nonogram

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
// first grid cell. This depends on the hint areas, title bar, centering
// offsets, and the status bar.
func (m *Model) gridOrigin() (x, y int) {
	rowHintWidth := m.rowHints.RequiredLen() * cellWidth
	colHintHeight := m.colHints.RequiredLen()

	// Grid content dimensions (the entire View output of the nonogram).
	gridRenderWidth := rowHintWidth + gridPixelWidth(m.width)
	gridRenderHeight := colHintHeight + gridPixelHeight(m.height)

	// Title bar is 1 line of text. The overall View joins:
	//   title (1 line) + grid block + status bar (2 lines with MarginTop 1).
	titleHeight := 1
	statusHeight := 2 // 1 line text + 1 line margin
	totalContentHeight := titleHeight + gridRenderHeight + statusHeight

	// The content width is the max of the title width and grid block width.
	// Title is short, so grid block dominates.
	totalContentWidth := gridRenderWidth

	// CenterView uses lipgloss.Place which centers within termWidth x termHeight.
	centerOffsetX := max((m.termWidth-totalContentWidth)/2, 0)
	centerOffsetY := max((m.termHeight-totalContentHeight)/2, 0)

	x = centerOffsetX + rowHintWidth
	y = centerOffsetY + titleHeight + colHintHeight
	return x, y
}

// gridPixelWidth returns the terminal character width of the grid area.
func gridPixelWidth(w int) int {
	pw := w * cellWidth
	if w > spacerEvery {
		pw += (w - 1) / spacerEvery
	}
	return pw
}

// gridPixelHeight returns the terminal line height of the grid area.
func gridPixelHeight(h int) int {
	ph := h
	if h > spacerEvery {
		ph += (h - 1) / spacerEvery
	}
	return ph
}
