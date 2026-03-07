package sudokurgb

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = game.DynamicGridCellWidth

func emptyCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.TextDim).
		Background(game.DefaultBorderColors().BackgroundBG)
}

func providedCellStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Background(game.DefaultBorderColors().BackgroundBG)
}

func userCellStyle(value int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(symbolColor(value))
}

func conflictCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.Error).
		Underline(true)
}

func valueCursorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true)
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
			text = game.CursorLeft + cellContentValue(c.v) + game.CursorRight
		}
	}

	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(text)
}

func cellStyle(m Model, c cell, x, y int, conflict, solved bool) lipgloss.Style {
	isCursor := m.cursor.X == x && m.cursor.Y == y
	var base lipgloss.Style

	switch c.v {
	case 0:
		base = emptyCellStyle()
	default:
		base = userCellStyle(c.v)
	}

	if m.providedGrid[y][x] {
		base = providedCellStyle()
		if c.v != 0 {
			base = base.Foreground(symbolColor(c.v))
		}
	}

	switch {
	case conflict:
		base = base.Inherit(conflictCellStyle())
	case solved:
		base = base.Bold(true)
	}

	if isCursor {
		base = base.Inherit(valueCursorStyle())
	}

	return base
}

func bridgeFill(m Model, solved bool, bridge game.DynamicGridBridge) color.Color {
	return nil
}

func activeBoxZoneFill(cursor game.Cursor, solved bool, zone int) color.Color {
	return nil
}

func sudokuBoxIndex(x, y int) int {
	return (y / 3 * 3) + (x / 3)
}

func cellContent(c cell) string {
	if c.v == 0 {
		return "·"
	}
	return cellContentValue(c.v)
}

func cellContentValue(value int) string {
	switch value {
	case 1:
		return "▲"
	case 2:
		return "■"
	case 3:
		return "●"
	default:
		return "·"
	}
}

func symbolColor(value int) color.Color {
	p := theme.Current()
	switch value {
	case 1:
		return p.Error
	case 2:
		return p.Success
	case 3:
		return p.Secondary
	default:
		return p.TextDim
	}
}

func computeConflicts(g grid) [gridSize][gridSize]bool {
	var conflicts [gridSize][gridSize]bool

	for y := range gridSize {
		var seen [valueCount + 1][]int
		for x := range gridSize {
			value := g[y][x].v
			if value != 0 {
				seen[value] = append(seen[value], x)
			}
		}
		for value := 1; value <= valueCount; value++ {
			if len(seen[value]) > houseQuota {
				for _, x := range seen[value] {
					conflicts[y][x] = true
				}
			}
		}
	}

	for x := range gridSize {
		var seen [valueCount + 1][]int
		for y := range gridSize {
			value := g[y][x].v
			if value != 0 {
				seen[value] = append(seen[value], y)
			}
		}
		for value := 1; value <= valueCount; value++ {
			if len(seen[value]) > houseQuota {
				for _, y := range seen[value] {
					conflicts[y][x] = true
				}
			}
		}
	}

	for boxY := range 3 {
		for boxX := range 3 {
			type pos struct{ x, y int }
			var seen [valueCount + 1][]pos
			for dy := range 3 {
				for dx := range 3 {
					x, y := boxX*3+dx, boxY*3+dy
					value := g[y][x].v
					if value != 0 {
						seen[value] = append(seen[value], pos{x: x, y: y})
					}
				}
			}
			for value := 1; value <= valueCount; value++ {
				if len(seen[value]) > houseQuota {
					for _, p := range seen[value] {
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
		return game.StatusBarStyle().Render("mouse: click focus  arrows/wasd: move  1/2/3: ▲■●  bkspc: clear  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("mouse: click focus  1/2/3: ▲■●  bkspc: clear")
}
