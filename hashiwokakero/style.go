package hashiwokakero

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 5
const islandPillWidth = game.DynamicGridCellWidth

type cellVisual struct {
	text    string
	fg      color.Color
	bg      color.Color
	outerBG color.Color
	bold    bool
	pill    bool
}

func boardBackground() color.Color {
	return theme.Current().BG
}

func baseIslandBackground() color.Color {
	p := theme.Current()
	return theme.Blend(p.BG, p.Success, 0.45)
}

func adjacentIslandBackground() color.Color {
	return theme.Blend(baseIslandBackground(), theme.Current().Highlight, 0.12)
}

func solvedIslandBackground() color.Color {
	return theme.Blend(baseIslandBackground(), theme.Current().Success, 0.18)
}

func baseCellStyle(fg, bg color.Color, bold bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center).
		Foreground(fg).
		Background(bg)
	if bold {
		style = style.Bold(true)
	}
	return style
}

func selectedIslandStyle() lipgloss.Style {
	bg := theme.Current().SelectionBG
	return lipgloss.NewStyle().
		Width(islandPillWidth).
		AlignHorizontal(lipgloss.Center).
		Foreground(theme.TextOnBG(bg)).
		Background(bg).
		Bold(true)
}

func satisfiedIslandStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Width(islandPillWidth).
		AlignHorizontal(lipgloss.Center).
		Foreground(theme.TextOnBG(p.SuccessBG)).
		Background(p.SuccessBG).
		Bold(true)
}

func overfilledIslandStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Width(islandPillWidth).
		AlignHorizontal(lipgloss.Center).
		Foreground(theme.TextOnBG(p.ErrorBG)).
		Background(p.ErrorBG).
		Bold(true)
}

func adjacentIslandStyle() lipgloss.Style {
	bg := adjacentIslandBackground()
	return lipgloss.NewStyle().
		Width(islandPillWidth).
		AlignHorizontal(lipgloss.Center).
		Foreground(theme.TextOnBG(bg)).
		Background(bg).
		Bold(true)
}

func bridgeColors() []color.Color {
	return theme.Current().ThemeColors()
}

func bridgeColor(index int) color.Color {
	colors := bridgeColors()
	if index < 0 || len(colors) == 0 {
		return theme.TextOnBG(boardBackground())
	}
	return colors[index%len(colors)]
}

// isHighlightedNeighbor returns true if islandID is directly connectable
// to the cursor (or selected) island in any cardinal direction.
func isHighlightedNeighbor(m Model, islandID int) bool {
	sourceID := m.cursorIsland
	if m.selectedIsland != nil {
		sourceID = *m.selectedIsland
	}
	if sourceID == islandID {
		return false
	}
	for _, dir := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		adj := m.puzzle.FindAdjacentIsland(sourceID, dir[0], dir[1])
		if adj != nil && adj.ID == islandID {
			return true
		}
	}
	return false
}

func resolveCellVisual(m Model, x, y int, solved bool) cellVisual {
	boardBG := boardBackground()
	visual := cellVisual{
		text:    "   ",
		fg:      theme.TextOnBG(boardBG),
		bg:      boardBG,
		outerBG: boardBG,
	}

	ci := m.puzzle.CellContent(x, y)
	switch ci.Kind {
	case cellIsland:
		isl := m.puzzle.FindIslandByID(ci.IslandID)
		if isl == nil {
			return visual
		}

		label := fmt.Sprintf(" %d ", isl.Required)
		isCursor := m.cursorIsland == ci.IslandID
		current := m.puzzle.BridgeCount(ci.IslandID)
		islandBG := baseIslandBackground()
		visual.pill = true
		visual.outerBG = boardBG

		switch {
		case solved && isCursor:
			visual.fg = game.CursorFG()
			visual.bg = theme.Current().SuccessBG
			visual.bold = true
		case solved:
			islandBG = solvedIslandBackground()
			visual.fg = theme.TextOnBG(islandBG)
			visual.bg = islandBG
			visual.bold = true
		case m.selectedIsland != nil && *m.selectedIsland == ci.IslandID:
			visual.fg = theme.TextOnBG(theme.Current().SelectionBG)
			visual.bg = theme.Current().SelectionBG
			visual.bold = true
		case isCursor:
			visual.fg = game.CursorFG()
			visual.bg = game.CursorBG()
			visual.bold = true
		case current > isl.Required:
			visual.fg = theme.TextOnBG(theme.Current().ErrorBG)
			visual.bg = theme.Current().ErrorBG
			visual.bold = true
		case current == isl.Required:
			visual.fg = theme.TextOnBG(theme.Current().SuccessBG)
			visual.bg = theme.Current().SuccessBG
			visual.bold = true
		case isHighlightedNeighbor(m, ci.IslandID):
			adjacentBG := adjacentIslandBackground()
			visual.fg = theme.TextOnBG(adjacentBG)
			visual.bg = adjacentBG
			visual.bold = true
		default:
			visual.fg = theme.TextOnBG(islandBG)
			visual.bg = islandBG
			visual.bold = true
		}

		if isCursor && !solved {
			label = game.CursorLeft + fmt.Sprintf("%d", isl.Required) + game.CursorRight
		}
		visual.text = label

	case cellBridgeH:
		visual.text = horizontalBridgeGlyph(ci.BridgeCount)
		visual.fg = bridgeColor(ci.BridgeIdx)
		visual.bold = ci.BridgeCount == 2
	case cellBridgeV:
		visual.text = verticalBridgeGlyph(ci.BridgeCount)
		visual.fg = bridgeColor(ci.BridgeIdx)
		visual.bold = ci.BridgeCount == 2
	}

	return visual
}

func renderCellVisual(visual cellVisual) string {
	if visual.pill {
		sideWidth := max((cellWidth-islandPillWidth)/2, 0)
		left := lipgloss.NewStyle().Width(sideWidth).Background(visual.outerBG).Render("")
		right := lipgloss.NewStyle().Width(cellWidth - islandPillWidth - sideWidth).Background(visual.outerBG).Render("")
		center := lipgloss.NewStyle().
			Width(islandPillWidth).
			AlignHorizontal(lipgloss.Center).
			Foreground(visual.fg).
			Background(visual.bg)
		if visual.bold {
			center = center.Bold(true)
		}
		return left + center.Render(visual.text) + right
	}
	return baseCellStyle(visual.fg, visual.bg, visual.bold).Render(visual.text)
}

func gridView(m Model, solved bool) string {
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:     m.puzzle.Width,
		Height:    m.puzzle.Height,
		CellWidth: cellWidth,
		Solved:    solved,
		Cell: func(x, y int) string {
			return renderCellVisual(resolveCellVisual(m, x, y, solved))
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
		HasVerticalEdge: func(x, _ int) bool {
			return x <= 0 || x >= m.puzzle.Width
		},
		HasHorizontalEdge: func(_, y int) bool {
			return y <= 0 || y >= m.puzzle.Height
		},
		BridgeForeground: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeForeground(m, bridge)
		},
		BridgeBold: func(bridge game.DynamicGridBridge) bool {
			return bridgeBold(m, bridge)
		},
		VerticalBridgeText: func(x, y int) string {
			return horizontalSeparatorBridge(m, x, y)
		},
		HorizontalBridgeText: func(x, y int) string {
			return verticalSeparatorBridge(m, x, y)
		},
	})
}

func horizontalBridgeGlyph(count int) string {
	if count == 2 {
		return "═════"
	}
	return "━━━━━"
}

func verticalBridgeGlyph(count int) string {
	if count == 2 {
		return "  ║  "
	}
	return "  ┃  "
}

func horizontalSeparatorBridge(m Model, x, y int) string {
	count := horizontalBridgeCountAt(m.puzzle, x, y)
	if count == 0 {
		return ""
	}
	if count == 2 {
		return "═"
	}
	return "━"
}

func verticalSeparatorBridge(m Model, x, y int) string {
	count := verticalBridgeCountAt(m.puzzle, x, y)
	if count == 0 {
		return ""
	}
	if count == 2 {
		return "  ║  "
	}
	return "  ┃  "
}

func horizontalBridgeCountAt(p Puzzle, x, y int) int {
	if x <= 0 || x >= p.Width || y < 0 || y >= p.Height {
		return 0
	}

	left := p.CellContent(x-1, y)
	right := p.CellContent(x, y)

	switch {
	case left.Kind == cellBridgeH:
		return left.BridgeCount
	case right.Kind == cellBridgeH:
		return right.BridgeCount
	case left.Kind == cellIsland && right.Kind == cellIsland:
		if b := p.GetBridge(left.IslandID, right.IslandID); b != nil {
			return b.Count
		}
	}

	return 0
}

func verticalBridgeCountAt(p Puzzle, x, y int) int {
	if y <= 0 || y >= p.Height || x < 0 || x >= p.Width {
		return 0
	}

	top := p.CellContent(x, y-1)
	bottom := p.CellContent(x, y)

	switch {
	case top.Kind == cellBridgeV:
		return top.BridgeCount
	case bottom.Kind == cellBridgeV:
		return bottom.BridgeCount
	case top.Kind == cellIsland && bottom.Kind == cellIsland:
		if b := p.GetBridge(top.IslandID, bottom.IslandID); b != nil {
			return b.Count
		}
	}

	return 0
}

func bridgeForeground(m Model, bridge game.DynamicGridBridge) color.Color {
	return bridgeColor(bridgeIndexForDynamicBridge(&m.puzzle, bridge))
}

func bridgeBold(m Model, bridge game.DynamicGridBridge) bool {
	return bridgeCountForDynamicBridge(&m.puzzle, bridge) == 2
}

func bridgeIndexForDynamicBridge(p *Puzzle, bridge game.DynamicGridBridge) int {
	count, index := dynamicBridgeIdentity(p, bridge)
	_ = count
	return index
}

func bridgeCountForDynamicBridge(p *Puzzle, bridge game.DynamicGridBridge) int {
	count, _ := dynamicBridgeIdentity(p, bridge)
	return count
}

func dynamicBridgeIdentity(p *Puzzle, bridge game.DynamicGridBridge) (count int, index int) {
	for i := 0; i < bridge.Count; i++ {
		cell := p.CellContent(bridge.Cells[i].X, bridge.Cells[i].Y)
		switch cell.Kind {
		case cellBridgeH, cellBridgeV:
			return cell.BridgeCount, cell.BridgeIdx
		}
	}

	ids := make([]int, 0, 2)
	for i := 0; i < bridge.Count; i++ {
		cell := p.CellContent(bridge.Cells[i].X, bridge.Cells[i].Y)
		if cell.Kind != cellIsland {
			continue
		}
		duplicate := false
		for _, id := range ids {
			if id == cell.IslandID {
				duplicate = true
				break
			}
		}
		if !duplicate {
			ids = append(ids, cell.IslandID)
		}
	}

	if len(ids) == 2 {
		if p.bridgeIndex == nil {
			p.buildBridgeIndex()
		}
		if idx, ok := p.bridgeIndex[bridgeKey(ids[0], ids[1])]; ok {
			if idx >= 0 && idx < len(p.Bridges) {
				return p.Bridges[idx].Count, idx
			}
			return 0, idx
		}
	}

	return 0, -1
}

func statusBarView(selected, showFullHelp bool) string {
	if selected {
		if showFullHelp {
			return game.StatusBarStyle().Render("arrows/wasd: build bridge  enter/space: deselect  esc: menu  ctrl+r: reset  ctrl+h: help")
		}
		return game.StatusBarStyle().Render("bridge mode  arrows/wasd: build  enter/space: deselect")
	}
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  enter/space: select island  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("arrows/wasd: move  enter/space: select island")
}

func statusBarVariants() []string {
	return []string{
		statusBarView(false, false),
		statusBarView(false, true),
		statusBarView(true, false),
		statusBarView(true, true),
		game.StatusBarStyle().Render("arrows/wasd: build bridge\nenter/space: deselect"),
	}
}

func infoView(p *Puzzle) string {
	pal := theme.Current()

	satisfied := 0
	total := len(p.Islands)
	for _, isl := range p.Islands {
		if p.BridgeCount(isl.ID) == isl.Required {
			satisfied++
		}
	}

	satisfiedStyle := lipgloss.NewStyle().Foreground(pal.SuccessBorder)
	infoStyle := lipgloss.NewStyle().Foreground(pal.Info)

	var sb strings.Builder
	sb.WriteString(infoStyle.Render("Islands: "))
	sb.WriteString(satisfiedStyle.Render(fmt.Sprintf("%d", satisfied)))
	sb.WriteString(infoStyle.Render(fmt.Sprintf("/%d satisfied  Bridges: %d", total, len(p.Bridges))))
	return sb.String()
}
