package rippleeffect

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 3

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
	rows := make([]string, 0, m.height*2+1)
	activeCage := m.geo.cageGrid[m.cursor.Y][m.cursor.X]
	completed := completedCageBackgrounds(m)
	bridgeBG := bridgeBackgrounds(m, completed)
	rows = append(rows, boundaryRow(m, 0, bridgeBG))
	for y := range m.height {
		rows = append(rows, contentRow(m, y, activeCage, bridgeBG))
		rows = append(rows, boundaryRow(m, y+1, bridgeBG))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func contentRow(m Model, y, activeCage int, completed map[int]color.Color) string {
	var b strings.Builder
	for x := range m.width {
		b.WriteString(renderBorderChar(m, verticalEdge(m.geo, x, y), m.solved, verticalGapBackground(m.geo, completed, x, y)))
		b.WriteString(cellView(m, x, y, activeCage, completed[m.geo.cageGrid[y][x]]))
	}
	b.WriteString(renderBorderChar(m, verticalEdge(m.geo, m.width, y), m.solved, verticalGapBackground(m.geo, completed, m.width, y)))
	return b.String()
}

func boundaryRow(m Model, y int, completed map[int]color.Color) string {
	var b strings.Builder
	for x := 0; x <= m.width; x++ {
		b.WriteString(renderBorderChar(m, junctionRune(m.geo, x, y), m.solved, junctionGapBackground(m.geo, completed, x, y)))
		if x == m.width {
			continue
		}

		ch := ' '
		if horizontalEdge(m.geo, x, y) {
			ch = '─'
		}
		b.WriteString(renderBorderSegment(m, ch, m.solved, horizontalGapBackground(m.geo, completed, x, y)))
	}
	return b.String()
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

func renderBorderChar(m Model, ch rune, solved bool, bg color.Color) string {
	p := theme.Current()
	fg := p.Border
	if solved {
		fg = p.SuccessBorder
	}
	if bg == nil {
		bg = p.BG
		if solved {
			bg = p.SuccessBG
		}
	}
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Render(string(ch))
}

func renderBorderSegment(m Model, ch rune, solved bool, bg color.Color) string {
	return strings.Repeat(renderBorderChar(m, ch, solved, bg), cellWidth)
}

func horizontalEdge(geo *geometry, x, y int) bool {
	switch {
	case y <= 0:
		return geo.boundaries[0][x].has(boundaryTop)
	case y >= geo.height:
		return geo.boundaries[geo.height-1][x].has(boundaryBottom)
	default:
		return geo.boundaries[y][x].has(boundaryTop)
	}
}

func verticalEdge(geo *geometry, x, y int) rune {
	if hasVerticalEdge(geo, x, y) {
		return '│'
	}
	return ' '
}

func hasVerticalEdge(geo *geometry, x, y int) bool {
	switch {
	case x <= 0:
		return geo.boundaries[y][0].has(boundaryLeft)
	case x >= geo.width:
		return geo.boundaries[y][geo.width-1].has(boundaryRight)
	default:
		return geo.boundaries[y][x].has(boundaryLeft)
	}
}

func junctionRune(geo *geometry, x, y int) rune {
	north := y > 0 && hasVerticalEdge(geo, x, y-1)
	south := y < geo.height && hasVerticalEdge(geo, x, y)
	west := x > 0 && horizontalEdge(geo, x-1, y)
	east := x < geo.width && horizontalEdge(geo, x, y)

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

func verticalGapBackground(geo *geometry, completed map[int]color.Color, x, y int) color.Color {
	if hasVerticalEdge(geo, x, y) || x <= 0 || x >= geo.width {
		return nil
	}
	left := geo.cageGrid[y][x-1]
	right := geo.cageGrid[y][x]
	if left != right {
		return nil
	}
	return completed[left]
}

func horizontalGapBackground(geo *geometry, completed map[int]color.Color, x, y int) color.Color {
	if horizontalEdge(geo, x, y) || y <= 0 || y >= geo.height {
		return nil
	}
	top := geo.cageGrid[y-1][x]
	bottom := geo.cageGrid[y][x]
	if top != bottom {
		return nil
	}
	return completed[top]
}

func junctionGapBackground(geo *geometry, completed map[int]color.Color, x, y int) color.Color {
	if junctionRune(geo, x, y) != ' ' {
		return nil
	}

	cages := make([]int, 0, 4)
	if x > 0 && y > 0 {
		cages = append(cages, geo.cageGrid[y-1][x-1])
	}
	if x < geo.width && y > 0 {
		cages = append(cages, geo.cageGrid[y-1][x])
	}
	if x > 0 && y < geo.height {
		cages = append(cages, geo.cageGrid[y][x-1])
	}
	if x < geo.width && y < geo.height {
		cages = append(cages, geo.cageGrid[y][x])
	}
	if len(cages) == 0 {
		return nil
	}

	cageID := cages[0]
	for _, id := range cages[1:] {
		if id != cageID {
			return nil
		}
	}
	return completed[cageID]
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("1-9: place  bkspc: clear  arrows/wasd: move  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("1-9: place  bkspc: clear")
}
