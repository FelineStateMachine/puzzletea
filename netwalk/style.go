package netwalk

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const (
	cellWidth  = 5
	cellHeight = 3
)

func gridView(m Model) string {
	colors := game.DefaultBorderColors()
	rows := make([]string, 0, m.puzzle.Size*cellHeight+2)
	rows = append(rows, game.HBorderRow(m.puzzle.Size, -1, cellWidth, "┌", "┐", colors, m.state.solved))
	for y := range m.puzzle.Size {
		for inner := range cellHeight {
			rows = append(rows, gridContentRow(m, y, inner, colors))
		}
	}
	rows = append(rows, game.HBorderRow(m.puzzle.Size, -1, cellWidth, "└", "┘", colors, m.state.solved))
	return strings.Join(rows, "\n")
}

func gridContentRow(m Model, y, inner int, colors game.GridBorderColors) string {
	var b strings.Builder
	b.WriteString(game.BorderChar("│", colors, m.state.solved, false))
	for x := range m.puzzle.Size {
		b.WriteString(cellRowView(m, x, y, inner))
	}
	b.WriteString(game.BorderChar("│", colors, m.state.solved, false))
	return b.String()
}

func cellRowView(m Model, x, y, inner int) string {
	rows := cellRows(m, x, y)
	t := m.puzzle.Tiles[y][x]
	if x == m.cursor.X && y == m.cursor.Y && !isActive(t) {
		rows = blankCursorRows()
	}
	style := lipgloss.NewStyle().
		Width(cellWidth).
		Background(cellBackground(m, x, y)).
		Foreground(cellForeground(m, x, y))

	if x == m.cursor.X && y == m.cursor.Y && isActive(t) {
		style = style.Bold(true)
	}
	if t.Kind == serverCell || degree(m.state.rotatedMasks[y][x]) == 1 || m.state.tileHasDangling[y][x] {
		style = style.Bold(true)
	}

	return style.Render(rows[inner])
}

func cellRows(m Model, x, y int) [cellHeight]string {
	t := m.puzzle.Tiles[y][x]
	if !isActive(t) {
		return [cellHeight]string{"     ", "     ", "     "}
	}

	mask := m.state.rotatedMasks[y][x]
	center := centerGlyph(m, x, y, t.Kind, mask)

	return [cellHeight]string{
		verticalCellRow(mask&north != 0),
		horizontalCellRow(mask&west != 0, center, mask&east != 0),
		verticalCellRow(mask&south != 0),
	}
}

func centerGlyph(m Model, x, y int, kind cellKind, mask directionMask) string {
	switch {
	case kind == serverCell:
		return "◆"
	case degree(mask) == 1:
		return "●"
	default:
		return maskGlyph(mask)
	}
}

func blankCursorRows() [cellHeight]string {
	return [cellHeight]string{
		"     ",
		game.CursorLeft + "   " + game.CursorRight,
		"     ",
	}
}

func verticalCellRow(on bool) string {
	if !on {
		return "     "
	}
	return "  │  "
}

func horizontalCellRow(left bool, center string, right bool) string {
	leftArm := "  "
	if left {
		leftArm = "──"
	}
	rightArm := "  "
	if right {
		rightArm = "──"
	}
	return leftArm + center + rightArm
}

func cellBackground(m Model, x, y int) color.Color {
	t := m.puzzle.Tiles[y][x]
	if m.state.solved {
		return theme.Current().SuccessBG
	}
	if x == m.cursor.X && y == m.cursor.Y {
		return theme.Blend(theme.Current().BG, theme.Current().Accent, 0.18)
	}
	if !isActive(t) {
		return theme.Current().BG
	}
	if t.Locked {
		return theme.Blend(theme.Current().BG, theme.Current().Surface, 0.60)
	}
	return theme.Current().BG
}

func cellForeground(m Model, x, y int) color.Color {
	return pipeForeground(m, point{X: x, Y: y})
}

func pipeForeground(m Model, cells ...point) color.Color {
	if m.state.solved {
		return theme.Current().SolvedFG
	}

	hasConnected := false
	for _, cell := range cells {
		if !m.puzzle.activeAt(cell.X, cell.Y) {
			continue
		}
		tile := m.puzzle.Tiles[cell.Y][cell.X]
		if tile.Kind == serverCell {
			return theme.Current().AccentSoft
		}
		if stateBoolAt(m.state.tileHasDangling, cell.X, cell.Y) {
			return theme.Current().Error
		}
		if stateBoolAt(m.state.connectedToRoot, cell.X, cell.Y) {
			hasConnected = true
		}
	}
	if hasConnected {
		return theme.Current().Secondary
	}
	return theme.Current().FG
}

func hasStateCell(cells [][]bool, x, y int) bool {
	return y >= 0 && y < len(cells) && x >= 0 && x < len(cells[y])
}

func stateBoolAt(cells [][]bool, x, y int) bool {
	return hasStateCell(cells, x, y) && cells[y][x]
}

func statusBarView(m Model, full bool) string {
	info := "connected " + strconv.Itoa(m.state.connected) + "/" + strconv.Itoa(m.state.nonEmpty) +
		"  dangling " + strconv.Itoa(m.state.dangling) +
		"  locks " + strconv.Itoa(m.state.locked)
	if !full {
		return game.StatusBarStyle().Render(info + "\nspace: rotate  enter: lock")
	}
	return game.StatusBarStyle().Render(info + "\nspace: rotate  backspace: reverse\nenter: toggle lock  ctrl+r: reset")
}
