package netwalk

import (
	"image/color"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 5

func gridView(m Model) string {
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:     m.puzzle.Size,
		Height:    m.puzzle.Size,
		CellWidth: cellWidth,
		Solved:    m.state.solved,
		Cell: func(x, y int) string {
			return cellView(m, x, y)
		},
		ZoneAt: func(x, y int) int {
			return 0
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeFill(m, bridge)
		},
		BridgeForeground: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeForeground(m, bridge)
		},
		VerticalBridgeText: func(x, y int) string {
			return verticalBridgeText(m, x, y)
		},
		HorizontalBridgeText: func(x, y int) string {
			return horizontalBridgeText(m, x, y)
		},
	})
}

func cellView(m Model, x, y int) string {
	t := m.puzzle.Tiles[y][x]
	bg := cellBackground(m, x, y)
	style := lipgloss.NewStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Background(bg).
		Foreground(cellForeground(m, x, y))

	if x == m.cursor.X && y == m.cursor.Y && isActive(t) {
		style = style.Bold(true)
	}
	if t.Kind == serverCell || m.state.tileHasDangling[y][x] {
		style = style.Bold(true)
	}

	return style.Render(cellText(m, x, y))
}

func cellText(m Model, x, y int) string {
	t := m.puzzle.Tiles[y][x]
	if !isActive(t) {
		return "     "
	}
	mask := m.state.rotatedMasks[y][x]
	if t.Kind == serverCell {
		return directionalSymbolText(mask, '★')
	}
	if degree(mask) == 1 {
		return directionalSymbolText(mask, '•')
	}

	left := "  "
	if mask&west != 0 {
		left = "──"
	}
	right := "  "
	if mask&east != 0 {
		right = "──"
	}
	return left + maskGlyph(mask) + right
}

func directionalSymbolText(mask directionMask, symbol rune) string {
	switch {
	case mask&west != 0 && mask&east != 0:
		return "──" + string(symbol) + "──"
	case mask&west != 0:
		return "──" + string(symbol) + "  "
	case mask&east != 0:
		return "  " + string(symbol) + "──"
	default:
		return "  " + string(symbol) + "  "
	}
}

func cellBackground(m Model, x, y int) color.Color {
	t := m.puzzle.Tiles[y][x]
	if !isActive(t) {
		if m.state.solved {
			return theme.Current().SuccessBG
		}
		return theme.Current().BG
	}
	if m.state.solved {
		return theme.Current().SuccessBG
	}
	if x == m.cursor.X && y == m.cursor.Y {
		return theme.Blend(theme.Current().BG, theme.Current().Accent, 0.18)
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
		if m.state.tileHasDangling[cell.Y][cell.X] {
			return theme.Current().Error
		}
		if m.state.connectedToRoot[cell.Y][cell.X] {
			hasConnected = true
		}
	}
	if hasConnected {
		return theme.Current().Secondary
	}
	return theme.Current().FG
}

func bridgeFill(m Model, _ game.DynamicGridBridge) color.Color {
	if m.state.solved {
		return theme.Current().SuccessBG
	}
	return nil
}

func bridgeForeground(m Model, bridge game.DynamicGridBridge) color.Color {
	cells := make([]point, 0, bridge.Count)
	for i := 0; i < bridge.Count; i++ {
		cells = append(cells, point{X: bridge.Cells[i].X, Y: bridge.Cells[i].Y})
	}
	return pipeForeground(m, cells...)
}

func verticalBridgeText(m Model, x, y int) string {
	if x <= 0 || x >= m.puzzle.Size || y < 0 || y >= m.puzzle.Size {
		return ""
	}
	leftMask := m.state.rotatedMasks[y][x-1]
	rightMask := m.state.rotatedMasks[y][x]
	if leftMask&east == 0 || rightMask&west == 0 {
		return ""
	}
	return "─"
}

func horizontalBridgeText(m Model, x, y int) string {
	if x < 0 || x >= m.puzzle.Size || y <= 0 || y >= m.puzzle.Size {
		return ""
	}
	topMask := m.state.rotatedMasks[y-1][x]
	bottomMask := m.state.rotatedMasks[y][x]
	if topMask&south == 0 || bottomMask&north == 0 {
		return ""
	}
	return "  │  "
}

func statusBarView(m Model, full bool) string {
	info := "connected " + strconv.Itoa(m.state.connected) + "/" + strconv.Itoa(m.state.nonEmpty) +
		"  dangling " + strconv.Itoa(m.state.dangling) +
		"  locks " + strconv.Itoa(m.state.locked)
	if !full {
		return game.StatusBarStyle().Render(info + "  space rotate  l lock")
	}
	return game.StatusBarStyle().Render(info + "  enter/space rotate  backspace reverse  l toggle lock  ctrl+r reset")
}
