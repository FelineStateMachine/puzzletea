package hashiwokakero

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/FelineStateMachine/puzzletea/game"
)

const cellWidth = 3

// Color palette — warm earth-tone theme (ANSI 256, adaptive for light/dark terminals)
var (
	colorIslandFg      = compat.AdaptiveColor{Dark: lipgloss.Color("236"), Light: lipgloss.Color("187")}
	colorIslandBg      = compat.AdaptiveColor{Dark: lipgloss.Color("187"), Light: lipgloss.Color("58")}
	colorSatisfiedFg   = compat.AdaptiveColor{Dark: lipgloss.Color("22"), Light: lipgloss.Color("149")}
	colorSatisfiedBg   = compat.AdaptiveColor{Dark: lipgloss.Color("194"), Light: lipgloss.Color("236")}
	colorOverFg        = compat.AdaptiveColor{Dark: lipgloss.Color("160"), Light: lipgloss.Color("167")}
	colorOverBg        = compat.AdaptiveColor{Dark: lipgloss.Color("224"), Light: lipgloss.Color("52")}
	colorSelectedFg    = compat.AdaptiveColor{Dark: lipgloss.Color("255"), Light: lipgloss.Color("235")}
	colorSelectedBg    = compat.AdaptiveColor{Dark: lipgloss.Color("172"), Light: lipgloss.Color("179")}
	colorAdjacentBg    = compat.AdaptiveColor{Dark: lipgloss.Color("223"), Light: lipgloss.Color("58")}
	colorBridge        = compat.AdaptiveColor{Dark: lipgloss.Color("137"), Light: lipgloss.Color("137")}
	colorBridgeSolved  = compat.AdaptiveColor{Dark: lipgloss.Color("22"), Light: lipgloss.Color("149")}
	colorEmptyDot      = compat.AdaptiveColor{Dark: lipgloss.Color("252"), Light: lipgloss.Color("239")}
	colorEmptySolved   = compat.AdaptiveColor{Dark: lipgloss.Color("151"), Light: lipgloss.Color("107")}
	colorInfoSatisfied = compat.AdaptiveColor{Dark: lipgloss.Color("22"), Light: lipgloss.Color("149")}
	colorInfoText      = compat.AdaptiveColor{Dark: lipgloss.Color("137"), Light: lipgloss.Color("137")}
)

var (
	baseStyle = lipgloss.NewStyle()

	islandDefaultStyle = baseStyle.
				Foreground(colorIslandFg).
				Background(colorIslandBg).
				Bold(true)

	islandSatisfiedStyle = baseStyle.
				Foreground(colorSatisfiedFg).
				Background(colorSatisfiedBg).
				Bold(true)

	islandOverStyle = baseStyle.
			Foreground(colorOverFg).
			Background(colorOverBg).
			Bold(true)

	islandCursorStyle       = game.CursorStyle
	islandCursorSolvedStyle = game.CursorSolvedStyle

	islandSelectedStyle = baseStyle.
				Foreground(colorSelectedFg).
				Background(colorSelectedBg).
				Bold(true)

	islandAdjacentStyle = baseStyle.
				Foreground(colorIslandFg).
				Background(colorAdjacentBg).
				Bold(true)

	bridgeHStyle = baseStyle.
			Foreground(colorBridge)

	bridgeVStyle = baseStyle.
			Foreground(colorBridge)

	bridgeHSolvedStyle = baseStyle.
				Foreground(colorBridgeSolved)

	bridgeVSolvedStyle = baseStyle.
				Foreground(colorBridgeSolved)

	emptyStyle = baseStyle.
			Foreground(colorEmptyDot)

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			BorderForeground(colorEmptyDot)

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(1).
				BorderForeground(colorBridgeSolved)
)

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

// emptyCellView renders an empty grid cell as a subtle dot.
func emptyCellView(solved bool) string {
	s := emptyStyle
	if solved {
		s = s.Foreground(colorEmptySolved)
	}
	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render("·")
}

// gridView renders the puzzle using a display grid with gap cells between
// logical cells so that bridges between directly adjacent islands are visible.
// Display grid is (2*width-1) x (2*height-1): even coords are logical cells,
// odd coords are gap cells showing bridge connectors.
func gridView(m Model, solved bool) string {
	dispW := 2*m.puzzle.Width - 1
	dispH := 2*m.puzzle.Height - 1

	var rows []string
	for dy := range dispH {
		var cells []string
		for dx := range dispW {
			cells = append(cells, displayCellView(m, dx, dy, solved))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if solved {
		return gridBorderSolvedStyle.Render(grid)
	}
	return gridBorderStyle.Render(grid)
}

// displayCellView renders a single cell in the display grid.
func displayCellView(m Model, dx, dy int, solved bool) string {
	evenX := dx%2 == 0
	evenY := dy%2 == 0

	if evenX && evenY {
		// Logical cell position
		lx, ly := dx/2, dy/2
		ci := m.puzzle.CellContent(lx, ly)
		switch ci.Kind {
		case cellIsland:
			return islandView(m, ci.IslandID, solved)
		case cellBridgeH:
			return bridgeHView(ci.BridgeCount, solved)
		case cellBridgeV:
			return bridgeVView(ci.BridgeCount, solved)
		default:
			return emptyCellView(solved)
		}
	}

	if !evenX && evenY {
		// Horizontal gap between (dx/2, dy/2) and (dx/2+1, dy/2)
		lx, ly := dx/2, dy/2
		return hGapView(m, lx, ly, solved)
	}

	if evenX && !evenY {
		// Vertical gap between (dx/2, dy/2) and (dx/2, dy/2+1)
		lx, ly := dx/2, dy/2
		return vGapView(m, lx, ly, solved)
	}

	// Odd x and odd y — intersection of gaps, always empty
	return baseStyle.Width(cellWidth).Render(" ")
}

// hGapView renders the horizontal gap between logical cell (lx,ly) and (lx+1,ly).
// Shows a bridge connector if both neighbors are islands connected by a bridge.
func hGapView(m Model, lx, ly int, solved bool) string {
	left := m.puzzle.FindIslandAt(lx, ly)
	right := m.puzzle.FindIslandAt(lx+1, ly)
	if left != nil && right != nil {
		if b := m.puzzle.GetBridge(left.ID, right.ID); b != nil {
			return bridgeHView(b.Count, solved)
		}
	}
	// Also check if a bridge passes through this gap (non-adjacent islands)
	ci := m.puzzle.CellContent(lx, ly)
	if ci.Kind == cellBridgeH {
		return bridgeHView(ci.BridgeCount, solved)
	}
	ciR := m.puzzle.CellContent(lx+1, ly)
	if ciR.Kind == cellBridgeH {
		return bridgeHView(ciR.BridgeCount, solved)
	}
	return baseStyle.Width(cellWidth).Render(" ")
}

// vGapView renders the vertical gap between logical cell (lx,ly) and (lx,ly+1).
// Shows a bridge connector if both neighbors are islands connected by a bridge.
func vGapView(m Model, lx, ly int, solved bool) string {
	top := m.puzzle.FindIslandAt(lx, ly)
	bottom := m.puzzle.FindIslandAt(lx, ly+1)
	if top != nil && bottom != nil {
		if b := m.puzzle.GetBridge(top.ID, bottom.ID); b != nil {
			return bridgeVView(b.Count, solved)
		}
	}
	// Also check if a bridge passes through this gap (non-adjacent islands)
	ci := m.puzzle.CellContent(lx, ly)
	if ci.Kind == cellBridgeV {
		return bridgeVView(ci.BridgeCount, solved)
	}
	ciB := m.puzzle.CellContent(lx, ly+1)
	if ciB.Kind == cellBridgeV {
		return bridgeVView(ciB.BridgeCount, solved)
	}
	return baseStyle.Width(cellWidth).Render(" ")
}

func islandView(m Model, islandID int, solved bool) string {
	isl := m.puzzle.FindIslandByID(islandID)
	if isl == nil {
		return emptyStyle.Width(cellWidth).Render(" ")
	}

	style := islandDefaultStyle
	current := m.puzzle.BridgeCount(islandID)

	if solved && m.cursorIsland == islandID {
		style = islandCursorSolvedStyle
	} else if solved {
		style = islandSatisfiedStyle
	} else if m.selectedIsland != nil && *m.selectedIsland == islandID {
		style = islandSelectedStyle
	} else if m.cursorIsland == islandID {
		style = islandCursorStyle
	} else if current == isl.Required {
		style = islandSatisfiedStyle
	} else if current > isl.Required {
		style = islandOverStyle
	} else if isHighlightedNeighbor(m, islandID) {
		style = islandAdjacentStyle
	}

	label := fmt.Sprintf("%d", isl.Required)
	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(label)
}

func bridgeHView(count int, solved bool) string {
	r := "───"
	if count == 2 {
		r = "═══"
	}
	s := bridgeHStyle
	if solved {
		s = bridgeHSolvedStyle
	}
	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func bridgeVView(count int, solved bool) string {
	r := "│"
	if count == 2 {
		r = "║"
	}
	s := bridgeVStyle
	if solved {
		s = bridgeVSolvedStyle
	}
	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func statusBarView(selected, showFullHelp bool) string {
	if selected {
		if showFullHelp {
			return game.StatusBarStyle.Render("arrows/wasd: build bridge  enter/space/esc: cancel  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
		}
		return game.StatusBarStyle.Render("arrows/wasd: build bridge  enter/space/esc: cancel")
	}
	if showFullHelp {
		return game.StatusBarStyle.Render("arrows/wasd: move  enter/space: select island  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle.Render("arrows/wasd: move  enter/space: select island")
}

func infoView(p *Puzzle) string {
	satisfied := 0
	total := len(p.Islands)
	for _, isl := range p.Islands {
		if p.BridgeCount(isl.ID) == isl.Required {
			satisfied++
		}
	}

	satisfiedStyle := lipgloss.NewStyle().Foreground(colorInfoSatisfied)
	infoStyle := lipgloss.NewStyle().Foreground(colorInfoText)

	var sb strings.Builder
	sb.WriteString(infoStyle.Render("Islands: "))
	sb.WriteString(satisfiedStyle.Render(fmt.Sprintf("%d", satisfied)))
	sb.WriteString(infoStyle.Render(fmt.Sprintf("/%d satisfied  Bridges: %d", total, len(p.Bridges))))
	return sb.String()
}
