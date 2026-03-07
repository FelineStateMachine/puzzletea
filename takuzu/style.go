package takuzu

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

	if isProvided && val != emptyCell {
		style = style.Bold(true)
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
			for i := 0; i < bridge.Count; i++ {
				cell := bridge.Cells[i]
				if cell.X == m.cursor.X || cell.Y == m.cursor.Y {
					return theme.Current().Surface
				}
			}
			return nil
		},
	})
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  mouse: click/cycle  z: ●  x: ○  bkspc: clear  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("mouse: click/cycle  z: ●  x: ○")
}
