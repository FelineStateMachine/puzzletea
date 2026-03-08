package sudoku

import (
	"image/color"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = game.DynamicGridCellWidth

func emptyCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.TextDim).
		Background(p.BG)
}

func providedCellStyle(v int) lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(digitColor(v)).
		Background(p.BG)
}

func userCellStyle(v int) lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(digitColor(v)).
		Background(p.BG)
}

func conflictCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.Error).
		Background(p.ErrorBG)
}

func sameNumberStyle(v int) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(digitColor(v))
}

func digitCursorStyle(value int) lipgloss.Style {
	bg := digitColor(value)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextOnBG(bg)).
		Background(bg)
}

func renderGrid(m Model, solved bool) string {
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  gridSize,
		Height: gridSize,
		Solved: solved,
		Cell: func(x, y int) string {
			return cellView(m, x, y, solved)
		},
		ZoneAt: func(x, y int) int {
			return sudokuBoxIndex(x, y)
		},
		ZoneFill: func(zone int) color.Color {
			return activeBoxZoneFill(m.cursor, solved, zone)
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeFill(m, solved, bridge)
		},
	})
}

func cellView(m Model, x, y int, solved bool) string {
	c := m.grid[y][x]
	style := cellStyle(m, c, x, y, m.conflicts[y][x], solved)
	text := cellContent(c)

	if x == m.cursor.X && y == m.cursor.Y {
		if c.v == 0 {
			text = game.CursorLeft + "·" + game.CursorRight
		} else {
			text = game.CursorLeft + strconv.Itoa(c.v) + game.CursorRight
		}
	}

	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(text)
}

func cellStyle(m Model, c cell, x, y int, conflict, solved bool) lipgloss.Style {
	p := theme.Current()
	isCursor := m.cursor.X == x && m.cursor.Y == y
	cursorVal := m.grid[m.cursor.Y][m.cursor.X].v

	switch {
	case isCursor && solved:
		if c.v != 0 {
			return digitCursorStyle(c.v)
		}
		return game.CursorSolvedStyle()
	case isCursor:
		if c.v != 0 {
			return digitCursorStyle(c.v)
		}
		return game.CursorStyle()
	case solved:
		return lipgloss.NewStyle().
			Foreground(p.SolvedFG).
			Background(p.SuccessBG)
	case conflict:
		return conflictCellStyle()
	case cursorVal != 0 && c.v == cursorVal:
		return sameNumberStyle(c.v)
	}

	isProvided := m.providedGrid[y][x]
	if isProvided {
		style := providedCellStyle(c.v)
		if inCursorContext(m.cursor, x, y) {
			style = style.Background(p.Surface)
		}
		return style
	}

	if c.v != 0 {
		style := userCellStyle(c.v)
		if inCursorContext(m.cursor, x, y) {
			style = style.Background(p.Surface)
		}
		return style
	}

	style := emptyCellStyle()
	if inCursorContext(m.cursor, x, y) {
		style = style.Background(p.Surface)
	}
	return style
}

func bridgeFill(m Model, solved bool, bridge game.DynamicGridBridge) color.Color {
	if solved {
		return nil
	}

	if game.DynamicGridBridgeOnCrosshairAxis(m.cursor, bridge) {
		return theme.Current().Surface
	}
	return nil
}

func activeBoxZoneFill(cursor game.Cursor, solved bool, zone int) color.Color {
	if solved || zone != sudokuBoxIndex(cursor.X, cursor.Y) {
		return nil
	}
	return theme.Current().Surface
}

func inCursorContext(cursor game.Cursor, x, y int) bool {
	return cursor.X == x ||
		cursor.Y == y ||
		sudokuBoxIndex(cursor.X, cursor.Y) == sudokuBoxIndex(x, y)
}

func sudokuBoxIndex(x, y int) int {
	return (y / 3 * 3) + (x / 3)
}

func digitColor(value int) color.Color {
	if value <= 0 {
		return theme.Current().TextDim
	}

	colors := theme.Current().ThemeColors()
	if len(colors) == 0 {
		return theme.Current().FG
	}

	return colors[(value-1)%len(colors)]
}

func cellContent(c cell) string {
	if c.v == 0 {
		return "·"
	}
	return strconv.Itoa(c.v)
}

// computeConflicts returns a grid of booleans indicating which cells have conflicts.
// A cell has a conflict if its value appears more than once in its row, column, or 3x3 box.
func computeConflicts(g grid) [gridSize][gridSize]bool {
	var conflicts [gridSize][gridSize]bool

	for y := range gridSize {
		var seen [10][]int // value -> list of x positions
		for x := range gridSize {
			v := g[y][x].v
			if v != 0 {
				seen[v] = append(seen[v], x)
			}
		}
		for v := 1; v <= 9; v++ {
			if len(seen[v]) > 1 {
				for _, x := range seen[v] {
					conflicts[y][x] = true
				}
			}
		}
	}

	for x := range gridSize {
		var seen [10][]int
		for y := range gridSize {
			v := g[y][x].v
			if v != 0 {
				seen[v] = append(seen[v], y)
			}
		}
		for v := 1; v <= 9; v++ {
			if len(seen[v]) > 1 {
				for _, y := range seen[v] {
					conflicts[y][x] = true
				}
			}
		}
	}

	for boxY := range 3 {
		for boxX := range 3 {
			type pos struct{ y, x int }
			var seen [10][]pos
			for dy := range 3 {
				for dx := range 3 {
					y, x := boxY*3+dy, boxX*3+dx
					v := g[y][x].v
					if v != 0 {
						seen[v] = append(seen[v], pos{y, x})
					}
				}
			}
			for v := 1; v <= 9; v++ {
				if len(seen[v]) > 1 {
					for _, p := range seen[v] {
						conflicts[p.y][p.x] = true
					}
				}
			}
		}
	}

	return conflicts
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("mouse: click focus  arrows/wasd: move  1-9: fill  bkspc: clear  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("mouse: click focus  1-9: fill  bkspc: clear")
}
