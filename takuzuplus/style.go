package takuzuplus

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = game.DynamicGridCellWidth

var renderRuneMap = map[rune]string{
	zeroCell:  " ● ",
	oneCell:   " ○ ",
	emptyCell: " · ",
}

func zeroStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().Foreground(p.Accent).Background(p.BG)
}

func oneStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().Foreground(p.Secondary).Background(p.BG)
}

func emptyStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().Foreground(p.TextDim).Background(p.BG)
}

func renderStyleMap() map[rune]lipgloss.Style {
	return map[rune]lipgloss.Style{
		zeroCell:  zeroStyle(),
		oneCell:   oneStyle(),
		emptyCell: emptyStyle(),
	}
}

func cellView(val rune, isProvided, isCursor, inCursorRow, inCursorCol, solved bool) string {
	p := theme.Current()
	styles := renderStyleMap()
	style, ok := styles[val]
	if !ok {
		style = emptyStyle()
	}

	text, ok := renderRuneMap[val]
	if !ok {
		text = renderRuneMap[emptyCell]
	}

	switch {
	case isCursor && solved:
		style = game.CursorSolvedStyle()
		text = game.CursorLeft + string([]rune(text)[1]) + game.CursorRight
	case isCursor:
		style = game.CursorStyle()
		text = game.CursorLeft + string([]rune(text)[1]) + game.CursorRight
	case solved:
		style = style.Foreground(p.SolvedFG).Background(p.SuccessBG)
	case inCursorRow || inCursorCol:
		style = style.Background(p.Surface)
	}

	if isProvided && val != emptyCell && !isCursor && !solved {
		bg := p.BG
		if inCursorRow || inCursorCol {
			bg = p.Surface
		}
		style = style.Bold(true).Background(theme.GivenTint(bg))
	}

	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(text)
}

func gridView(m Model) string {
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  m.size,
		Height: m.size,
		Solved: m.solved,
		Cell: func(x, y int) string {
			return cellView(
				m.grid[y][x],
				m.provided[y][x],
				x == m.cursor.X && y == m.cursor.Y,
				y == m.cursor.Y,
				x == m.cursor.X,
				m.solved,
			)
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			if m.solved {
				return theme.Current().SuccessBG
			}
			if bg := relationBridgeBackground(m, bridge); bg != nil {
				return bg
			}
			for i := 0; i < bridge.Count; i++ {
				cell := bridge.Cells[i]
				if cell.X == m.cursor.X || cell.Y == m.cursor.Y {
					return theme.Current().Surface
				}
			}
			return nil
		},
		VerticalBridgeText: func(x, y int) string {
			if x <= 0 || x >= m.size {
				return ""
			}
			rel := m.relations.horizontal[y][x-1]
			if rel == relationNone {
				return ""
			}
			return string(rel)
		},
		HorizontalBridgeText: func(x, y int) string {
			if y <= 0 || y >= m.size {
				return ""
			}
			rel := m.relations.vertical[y-1][x]
			if rel == relationNone {
				return ""
			}
			return string(rel)
		},
	})
}

func relationBridgeBackground(m Model, bridge game.DynamicGridBridge) color.Color {
	switch bridge.Kind {
	case game.DynamicGridBridgeVertical:
		if bridge.X <= 0 || bridge.X >= m.size || bridge.Y < 0 || bridge.Y >= m.size {
			return nil
		}
		return relationStateBackground(relationState(m.relations.horizontal[bridge.Y][bridge.X-1], m.grid[bridge.Y][bridge.X-1], m.grid[bridge.Y][bridge.X]))
	case game.DynamicGridBridgeHorizontal:
		if bridge.Y <= 0 || bridge.Y >= m.size || bridge.X < 0 || bridge.X >= m.size {
			return nil
		}
		return relationStateBackground(relationState(m.relations.vertical[bridge.Y-1][bridge.X], m.grid[bridge.Y-1][bridge.X], m.grid[bridge.Y][bridge.X]))
	default:
		return nil
	}
}

func relationState(rel, left, right rune) int {
	if rel == relationNone || left == emptyCell || right == emptyCell {
		return 0
	}
	if relationSatisfied(rel, left, right) {
		return 1
	}
	return -1
}

func relationStateBackground(state int) color.Color {
	switch state {
	case 1:
		return theme.Current().SuccessBG
	case -1:
		return game.ConflictBG()
	default:
		return nil
	}
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  mouse: click/cycle  z: ●  x: ○  bkspc: clear  =: same clue  x: opposite clue  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("mouse: click/cycle  z: ●  x: ○  fixed clues: = and x")
}
