package fillomino

import (
	"image/color"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = game.DynamicGridCellWidth

type renderGridState struct {
	zones      [][]int
	activeZone int
	completed  map[int]color.Color
}

func cellBaseStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.FG).
		Background(p.BG)
}

func emptyCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.TextDim).
		Background(p.BG)
}

func cellView(
	value int,
	provided, cursor, rowHighlight, colHighlight, regionHighlight, solved, conflict bool,
	completedBG color.Color,
) string {
	p := theme.Current()
	style := cellBaseStyle()
	text := " · "
	if value == 0 {
		style = emptyCellStyle()
	} else {
		text = lipgloss.NewStyle().Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(strconv.Itoa(value))
	}

	if provided && value != 0 {
		style = style.Bold(true)
	}
	if conflict && cursor {
		style = conflictedCursorStyle()
	} else if conflict {
		style = style.Foreground(game.ConflictFG()).Background(game.ConflictBG())
	} else if solved {
		style = style.Foreground(game.SolvedFG()).Background(p.SuccessBG)
	} else if cursor {
		style = game.CursorStyle()
	} else if completedBG != nil {
		style = style.Background(completedBG).Foreground(theme.TextOnBG(completedBG))
	} else if regionHighlight {
		style = style.Background(p.HighlightBG)
	} else if rowHighlight || colHighlight {
		style = style.Background(p.Surface)
	}

	if cursor {
		if value == 0 {
			text = game.CursorLeft + "·" + game.CursorRight
		} else {
			text = game.CursorLeft + strconv.Itoa(value) + game.CursorRight
		}
	}

	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(text)
}

func conflictedCursorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(game.CursorFG()).
		Background(game.ConflictBG())
}

func gridView(m Model) string {
	renderState := buildRenderGridState(m)

	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  m.width,
		Height: m.height,
		Solved: m.solved,
		Cell: func(x, y int) string {
			zone := renderState.zones[y][x]
			completedBG := renderState.completed[zone]
			return cellView(
				m.grid[y][x],
				m.provided[y][x],
				x == m.cursor.X && y == m.cursor.Y,
				y == m.cursor.Y,
				x == m.cursor.X,
				renderState.activeZone >= 0 && zone == renderState.activeZone,
				m.solved,
				m.conflicts[y][x],
				completedBG,
			)
		},
		ZoneAt: func(x, y int) int {
			return renderState.zones[y][x]
		},
		ZoneFill: func(zone int) color.Color {
			p := theme.Current()
			switch {
			case m.solved:
				return p.SuccessBG
			case renderState.completed[zone] != nil:
				return renderState.completed[zone]
			case renderState.activeZone >= 0 && zone == renderState.activeZone:
				return p.HighlightBG
			default:
				return nil
			}
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeFill(m, renderState, bridge)
		},
	})
}

func bridgeFill(m Model, renderState renderGridState, bridge game.DynamicGridBridge) color.Color {
	for i := 0; i < bridge.Count; i++ {
		cell := bridge.Cells[i]
		if m.conflicts[cell.Y][cell.X] {
			return game.ConflictBG()
		}
	}

	if m.solved {
		return nil
	}

	if bridge.Uniform {
		switch {
		case renderState.completed[bridge.Zone] != nil:
			return nil
		case renderState.activeZone >= 0 && bridge.Zone == renderState.activeZone:
			return nil
		}
	}

	if bridge.Count > 0 && !bridgeTouchesBorder(renderState, bridge, m.width, m.height) {
		return nil
	}

	if game.DynamicGridBridgeOnCrosshairAxis(m.cursor, bridge) {
		return theme.Current().Surface
	}

	return nil
}

func bridgeTouchesBorder(renderState renderGridState, bridge game.DynamicGridBridge, width, height int) bool {
	switch bridge.Kind {
	case game.DynamicGridBridgeVertical:
		return zoneJunctionRune(renderState.zones, width, height, bridge.X, bridge.Y) != ' ' ||
			zoneJunctionRune(renderState.zones, width, height, bridge.X, bridge.Y+1) != ' '
	case game.DynamicGridBridgeHorizontal:
		return zoneJunctionRune(renderState.zones, width, height, bridge.X, bridge.Y) != ' ' ||
			zoneJunctionRune(renderState.zones, width, height, bridge.X+1, bridge.Y) != ' '
	default:
		return false
	}
}

func zoneHorizontalEdge(zones [][]int, height, x, y int) bool {
	switch {
	case y <= 0, y >= height:
		return true
	default:
		return zones[y-1][x] != zones[y][x]
	}
}

func zoneVerticalEdge(zones [][]int, width, height, x, y int) bool {
	_ = height
	switch {
	case x <= 0, x >= width:
		return true
	default:
		return zones[y][x-1] != zones[y][x]
	}
}

func zoneJunctionRune(zones [][]int, width, height, x, y int) rune {
	north := y > 0 && zoneVerticalEdge(zones, width, height, x, y-1)
	south := y < height && zoneVerticalEdge(zones, width, height, x, y)
	west := x > 0 && zoneHorizontalEdge(zones, height, x-1, y)
	east := x < width && zoneHorizontalEdge(zones, height, x, y)

	switch {
	case north && south && west && east:
		return '┼'
	case north && south && west:
		return '┤'
	case north && south && east:
		return '├'
	case west && east && north:
		return '┴'
	case west && east && south:
		return '┬'
	case south && east:
		return '┌'
	case south && west:
		return '┐'
	case north && east:
		return '└'
	case north && west:
		return '┘'
	case north || south:
		return '│'
	case west || east:
		return '─'
	default:
		return ' '
	}
}

func buildRenderGridState(m Model) renderGridState {
	height := len(m.grid)
	if height == 0 {
		return renderGridState{}
	}
	width := len(m.grid[0])
	zones := make([][]int, height)
	for y := range height {
		zones[y] = make([]int, width)
	}

	visited := make([][]bool, height)
	for y := range height {
		visited[y] = make([]bool, width)
	}

	palette := theme.Current()
	colors := palette.ThemeColors()
	completed := make(map[int]color.Color)
	nextZone := 0
	activeZone := -1
	emptyZone := nextZone
	nextZone++

	for y := range height {
		for x := range width {
			if m.grid[y][x] == 0 {
				zones[y][x] = emptyZone
				continue
			}
			if visited[y][x] {
				continue
			}

			comp := buildComponent(m.grid, point{x: x, y: y}, visited)
			zone := nextZone
			nextZone++
			for _, cell := range comp.cells {
				zones[cell.y][cell.x] = zone
			}
			if len(colors) > 0 && len(comp.cells) == comp.value && !componentHasConflict(comp, m.conflicts) {
				completed[zone] = completedRegionColor(comp, colors, palette.Surface)
			}
		}
	}

	if m.cursor.Y >= 0 && m.cursor.Y < height && m.cursor.X >= 0 && m.cursor.X < width && m.grid[m.cursor.Y][m.cursor.X] != 0 {
		activeZone = zones[m.cursor.Y][m.cursor.X]
	}

	return renderGridState{
		zones:      zones,
		activeZone: activeZone,
		completed:  completed,
	}
}

func componentHasConflict(comp component, conflicts [][]bool) bool {
	for _, cell := range comp.cells {
		if conflicts[cell.y][cell.x] {
			return true
		}
	}
	return false
}

func completedRegionColor(comp component, colors []color.Color, base color.Color) color.Color {
	anchor := comp.cells[0]
	index := (anchor.y*37 + anchor.x*17 + comp.value*13) % len(colors)
	return theme.Blend(base, colors[index], 0.52)
}

func cursorRegionInfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Info)
}

func cursorRegionInfoView(m Model) string {
	current, target := cursorRegionInfoData(m)
	return cursorRegionInfoStyle().Render("region size: " + strconv.Itoa(current) + "/" + target)
}

func cursorRegionInfoVariants(m Model) []string {
	area := m.width * m.height
	maxTarget := m.maxCellValue
	maxTarget = max(maxTarget, 1)

	return []string{
		cursorRegionInfoStyle().Render("region size: " + strconv.Itoa(area) + "/-"),
		cursorRegionInfoStyle().Render("region size: " + strconv.Itoa(area) + "/" + strconv.Itoa(maxTarget)),
	}
}

func cursorRegionInfoData(m Model) (current int, target string) {
	if len(m.grid) == 0 || len(m.grid[0]) == 0 {
		return 0, "-"
	}

	cursor := point{x: m.cursor.X, y: m.cursor.Y}
	value := m.grid[cursor.y][cursor.x]
	visited := make([][]bool, len(m.grid))
	for y := range len(m.grid) {
		visited[y] = make([]bool, len(m.grid[y]))
	}

	comp := buildComponent(m.grid, cursor, visited)
	if value == 0 {
		return len(comp.cells), "-"
	}
	return len(comp.cells), strconv.Itoa(value)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("1-9: place  bkspc: clear  arrows/wasd: move  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("1-9: place  bkspc: clear")
}
