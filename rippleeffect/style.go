package rippleeffect

import (
	"image/color"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = game.DynamicGridCellWidth

type visualKind int

const (
	visualNormal visualKind = iota
	visualCrosshair
	visualCage
	visualCompleted
	visualConflictCursor
	visualCursor
	visualSolved
	visualConflict
)

func chooseVisualKind(cursor, solved, conflict, completed, cageHighlight, crosshair bool) visualKind {
	switch {
	case cursor && conflict:
		return visualConflictCursor
	case conflict:
		return visualConflict
	case solved:
		return visualSolved
	case cursor:
		return visualCursor
	case completed:
		return visualCompleted
	case cageHighlight:
		return visualCage
	case crosshair:
		return visualCrosshair
	default:
		return visualNormal
	}
}

func gridView(m Model) string {
	activeCage := m.geo.cageGrid[m.cursor.Y][m.cursor.X]
	completed := completedCageBackgrounds(m)
	bridgeBG := bridgeBackgrounds(m, completed)
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  m.width,
		Height: m.height,
		Solved: m.solved,
		Cell: func(x, y int) string {
			return cellView(m, x, y, activeCage, completed[m.geo.cageGrid[y][x]])
		},
		ZoneAt: func(x, y int) int {
			return m.geo.cageGrid[y][x]
		},
		ZoneFill: func(zone int) color.Color {
			return bridgeBG[zone]
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeFill(m, bridgeBG, bridge)
		},
	})
}

func cellView(m Model, x, y, activeCage int, completedBG color.Color) string {
	p := theme.Current()
	kind := chooseVisualKind(
		x == m.cursor.X && y == m.cursor.Y,
		m.solved,
		m.conflicts[y][x],
		completedBG != nil,
		m.geo.cageGrid[y][x] == activeCage,
		y == m.cursor.Y || x == m.cursor.X,
	)

	style := lipgloss.NewStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center)

	fg := p.FG
	bg := p.BG
	text := " · "
	if value := m.grid[y][x]; value != 0 {
		text = " " + strconv.Itoa(value) + " "
	}

	switch kind {
	case visualConflictCursor:
		fg = game.CursorFG()
		bg = game.ConflictBG()
		text = cursorText(m.grid[y][x])
	case visualConflict:
		fg = game.ConflictFG()
		bg = game.ConflictBG()
	case visualSolved:
		fg = p.SolvedFG
		bg = p.SuccessBG
	case visualCursor:
		fg = p.AccentText
		bg = p.AccentBG
		text = cursorText(m.grid[y][x])
	case visualCompleted:
		bg = completedBG
		fg = theme.TextOnBG(bg)
	case visualCage:
		bg = p.SelectionBG
		fg = theme.TextOnBG(bg)
	case visualCrosshair:
		bg = p.Surface
	}

	if m.givens[y][x] != 0 {
		style = style.Bold(true)
	}

	return style.Foreground(fg).Background(bg).Render(text)
}

func cursorText(value int) string {
	if value == 0 {
		return game.CursorLeft + "·" + game.CursorRight
	}
	return game.CursorLeft + strconv.Itoa(value) + game.CursorRight
}

func completedCageBackgrounds(m Model) map[int]color.Color {
	palette := theme.Current()
	colors := palette.ThemeColors()
	if len(colors) == 0 {
		return nil
	}

	backgrounds := make(map[int]color.Color)
	for cageIdx, cells := range m.geo.cageCells {
		if !cageCompleted(m, cageIdx, cells) {
			continue
		}
		backgrounds[cageIdx] = completedCageColor(m.geo.cages[cageIdx], colors, palette.Surface)
	}
	return backgrounds
}

func solvedBridgeBackgrounds(m Model) map[int]color.Color {
	p := theme.Current()
	backgrounds := make(map[int]color.Color, len(m.geo.cages))
	for cageIdx := range m.geo.cages {
		backgrounds[cageIdx] = p.SuccessBG
	}
	return backgrounds
}

func conflictBridgeBackgrounds(m Model) map[int]color.Color {
	backgrounds := make(map[int]color.Color)
	for cageIdx, cells := range m.geo.cageCells {
		for _, cell := range cells {
			if !m.conflicts[cell.y][cell.x] {
				continue
			}
			backgrounds[cageIdx] = game.ConflictBG()
			break
		}
	}
	return backgrounds
}

func bridgeBackgrounds(m Model, completed map[int]color.Color) map[int]color.Color {
	if m.solved {
		return solvedBridgeBackgrounds(m)
	}

	backgrounds := activeCageBridgeBackgrounds(m)
	for cageIdx, bg := range completed {
		backgrounds[cageIdx] = bg
	}
	for cageIdx, bg := range conflictBridgeBackgrounds(m) {
		backgrounds[cageIdx] = bg
	}
	return backgrounds
}

func bridgeFill(m Model, bridgeBG map[int]color.Color, bridge game.DynamicGridBridge) color.Color {
	if m.solved {
		return nil
	}
	if bridge.Uniform && bridgeBG[bridge.Zone] != nil {
		return nil
	}
	if bridge.Count > 0 && !bridgeTouchesBorder(m.geo, bridge) {
		return nil
	}
	if bridgeOnCrosshairAxis(m.cursor, bridge) {
		return theme.Current().Surface
	}
	return nil
}

func bridgeOnCrosshairAxis(cursor game.Cursor, bridge game.DynamicGridBridge) bool {
	switch bridge.Kind {
	case game.DynamicGridBridgeVertical:
		if bridge.Count == 0 {
			return bridge.Y == cursor.Y
		}
		for i := 0; i < bridge.Count; i++ {
			if bridge.Cells[i].Y == cursor.Y {
				return true
			}
		}
	case game.DynamicGridBridgeHorizontal:
		if bridge.Count == 0 {
			return bridge.X == cursor.X
		}
		for i := 0; i < bridge.Count; i++ {
			if bridge.Cells[i].X == cursor.X {
				return true
			}
		}
	}
	return false
}

func bridgeTouchesBorder(geo *geometry, bridge game.DynamicGridBridge) bool {
	if geo == nil {
		return false
	}

	switch bridge.Kind {
	case game.DynamicGridBridgeVertical:
		return cageJunctionRune(geo.cageGrid, geo.width, geo.height, bridge.X, bridge.Y) != ' ' ||
			cageJunctionRune(geo.cageGrid, geo.width, geo.height, bridge.X, bridge.Y+1) != ' '
	case game.DynamicGridBridgeHorizontal:
		return cageJunctionRune(geo.cageGrid, geo.width, geo.height, bridge.X, bridge.Y) != ' ' ||
			cageJunctionRune(geo.cageGrid, geo.width, geo.height, bridge.X+1, bridge.Y) != ' '
	default:
		return false
	}
}

func cageHorizontalEdge(cageGrid [][]int, height, x, y int) bool {
	switch {
	case y <= 0, y >= height:
		return true
	default:
		return cageGrid[y-1][x] != cageGrid[y][x]
	}
}

func cageVerticalEdge(cageGrid [][]int, width, x, y int) bool {
	switch {
	case x <= 0, x >= width:
		return true
	default:
		return cageGrid[y][x-1] != cageGrid[y][x]
	}
}

func cageJunctionRune(cageGrid [][]int, width, height, x, y int) rune {
	north := y > 0 && cageVerticalEdge(cageGrid, width, x, y-1)
	south := y < height && cageVerticalEdge(cageGrid, width, x, y)
	west := x > 0 && cageHorizontalEdge(cageGrid, height, x-1, y)
	east := x < width && cageHorizontalEdge(cageGrid, height, x, y)

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

func activeCageBridgeBackgrounds(m Model) map[int]color.Color {
	backgrounds := make(map[int]color.Color, 1)
	if m.geo == nil || m.width == 0 || m.height == 0 {
		return backgrounds
	}

	activeCage := m.geo.cageGrid[m.cursor.Y][m.cursor.X]
	backgrounds[activeCage] = theme.Current().SelectionBG
	return backgrounds
}

func cageCompleted(m Model, cageIdx int, cells []point) bool {
	size := m.geo.cageSizes[cageIdx]
	seen := make(map[int]struct{}, size)
	for _, cell := range cells {
		if m.conflicts[cell.y][cell.x] {
			return false
		}
		value := m.grid[cell.y][cell.x]
		if value < 1 || value > size {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return len(seen) == size
}

func completedCageColor(cage Cage, colors []color.Color, base color.Color) color.Color {
	if len(colors) == 0 {
		return nil
	}
	first := cage.Cells[0]
	index := (first.Y*37 + first.X*17 + cage.Size*13 + cage.ID*7) % len(colors)
	return theme.Blend(base, colors[index], 0.52)
}

func verticalGapBackground(geo *geometry, backgrounds map[int]color.Color, x, y int) color.Color {
	if geo == nil || x <= 0 || x >= geo.width || y < 0 || y >= geo.height {
		return nil
	}
	left := geo.cageGrid[y][x-1]
	right := geo.cageGrid[y][x]
	if left != right {
		return nil
	}
	return backgrounds[left]
}

func horizontalGapBackground(geo *geometry, backgrounds map[int]color.Color, x, y int) color.Color {
	if geo == nil || y <= 0 || y >= geo.height || x < 0 || x >= geo.width {
		return nil
	}
	top := geo.cageGrid[y-1][x]
	bottom := geo.cageGrid[y][x]
	if top != bottom {
		return nil
	}
	return backgrounds[top]
}

func junctionGapBackground(geo *geometry, backgrounds map[int]color.Color, x, y int) color.Color {
	if geo == nil || x <= 0 || x >= geo.width || y <= 0 || y >= geo.height {
		return nil
	}

	cageID := geo.cageGrid[y-1][x-1]
	if geo.cageGrid[y-1][x] != cageID || geo.cageGrid[y][x-1] != cageID || geo.cageGrid[y][x] != cageID {
		return nil
	}
	return backgrounds[cageID]
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("1-9: place  bkspc: clear  arrows/wasd: move  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("1-9: place  bkspc: clear")
}
