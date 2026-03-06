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
			for i := 0; i < bridge.Count; i++ {
				cell := bridge.Cells[i]
				if m.conflicts[cell.Y][cell.X] {
					return game.ConflictBG()
				}
			}
			return nil
		},
	})
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

func horizontalEdge(g grid, x, y int) bool {
	switch {
	case y <= 0, y >= len(g):
		return true
	default:
		top := point{x: x, y: y - 1}
		bottom := point{x: x, y: y}
		return !sameRegion(g, top, bottom)
	}
}

func verticalEdge(g grid, x, y int) rune {
	if hasVerticalEdge(g, x, y) {
		return '│'
	}
	return ' '
}

func hasVerticalEdge(g grid, x, y int) bool {
	width := len(g[0])
	switch {
	case x <= 0, x >= width:
		return true
	default:
		left := point{x: x - 1, y: y}
		right := point{x: x, y: y}
		return !sameRegion(g, left, right)
	}
}

func junctionRune(g grid, x, y int) rune {
	height := len(g)
	width := len(g[0])
	north := y > 0 && hasVerticalEdge(g, x, y-1)
	south := y < height && hasVerticalEdge(g, x, y)
	west := x > 0 && horizontalEdge(g, x-1, y)
	east := x < width && horizontalEdge(g, x, y)

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

func verticalGapBackground(m Model, regionSet map[point]struct{}, completed map[point]color.Color, x, y int) color.Color {
	if hasVerticalEdge(m.grid, x, y) || x <= 0 || x >= m.width {
		return nil
	}
	left := point{x: x - 1, y: y}
	right := point{x: x, y: y}
	return gapBackground(m, regionSet, completed, left, right)
}

func horizontalGapBackground(m Model, regionSet map[point]struct{}, completed map[point]color.Color, x, y int) color.Color {
	if horizontalEdge(m.grid, x, y) || y <= 0 || y >= m.height {
		return nil
	}
	top := point{x: x, y: y - 1}
	bottom := point{x: x, y: y}
	return gapBackground(m, regionSet, completed, top, bottom)
}

func junctionGapBackground(m Model, regionSet map[point]struct{}, completed map[point]color.Color, x, y int) color.Color {
	if junctionRune(m.grid, x, y) != ' ' {
		return nil
	}

	cells := make([]point, 0, 4)
	if x > 0 && y > 0 {
		cells = append(cells, point{x: x - 1, y: y - 1})
	}
	if x < m.width && y > 0 {
		cells = append(cells, point{x: x, y: y - 1})
	}
	if x > 0 && y < m.height {
		cells = append(cells, point{x: x - 1, y: y})
	}
	if x < m.width && y < m.height {
		cells = append(cells, point{x: x, y: y})
	}
	if len(cells) != 4 {
		return nil
	}
	for i := 1; i < len(cells); i++ {
		if !sameRegion(m.grid, cells[0], cells[i]) {
			return nil
		}
	}
	return gapBackground(m, regionSet, completed, cells...)
}

func gapBackground(m Model, regionSet map[point]struct{}, completed map[point]color.Color, cells ...point) color.Color {
	p := theme.Current()
	if len(cells) == 0 {
		return nil
	}
	if anyCellConflict(m.conflicts, cells...) {
		return game.ConflictBG()
	}
	if m.solved {
		return p.SuccessBG
	}
	if bg := completed[cells[0]]; bg != nil {
		for _, cell := range cells[1:] {
			if completed[cell] == nil {
				return nil
			}
		}
		return bg
	}
	for _, cell := range cells {
		if _, ok := regionSet[cell]; !ok {
			return nil
		}
	}
	if len(regionSet) == 0 {
		return nil
	}
	return p.HighlightBG
}

func anyCellConflict(conflicts [][]bool, cells ...point) bool {
	for _, cell := range cells {
		if conflicts[cell.y][cell.x] {
			return true
		}
	}
	return false
}

func sameRegion(g grid, a, b point) bool {
	return g[a.y][a.x] != 0 && g[a.y][a.x] == g[b.y][b.x]
}

func completedRegionBackgrounds(g grid, conflicts [][]bool) map[point]color.Color {
	height := len(g)
	if height == 0 {
		return nil
	}
	width := len(g[0])
	visited := make([][]bool, height)
	for y := range height {
		visited[y] = make([]bool, width)
	}

	palette := theme.Current()
	colors := palette.ThemeColors()
	if len(colors) == 0 {
		return nil
	}

	backgrounds := make(map[point]color.Color)
	for y := range height {
		for x := range width {
			if g[y][x] == 0 || visited[y][x] {
				continue
			}

			comp := buildComponent(g, point{x: x, y: y}, visited)
			if len(comp.cells) != comp.value || componentHasConflict(comp, conflicts) {
				continue
			}

			bg := completedRegionColor(comp, colors, palette.Surface)
			for _, cell := range comp.cells {
				backgrounds[cell] = bg
			}
		}
	}

	return backgrounds
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

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("1-9: place  bkspc: clear  arrows/wasd: move  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("1-9: place  bkspc: clear")
}
