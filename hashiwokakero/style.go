package hashiwokakero

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const cellWidth = 3

var (
	baseStyle = lipgloss.NewStyle()

	islandDefaultStyle = baseStyle.
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#4a4a4a")).
				Bold(true)

	islandSatisfiedStyle = baseStyle.
				Foreground(lipgloss.Color("#00ff00")).
				Background(lipgloss.Color("#1a3a1a")).
				Bold(true)

	islandOverStyle = baseStyle.
			Foreground(lipgloss.Color("#ff0000")).
			Background(lipgloss.Color("#3a1a1a")).
			Bold(true)

	islandCursorStyle = baseStyle.
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#ff00ff")).
				Bold(true)

	islandSelectedStyle = baseStyle.
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#ff8800")).
				Bold(true)

	bridgeHStyle = baseStyle.
			Foreground(lipgloss.Color("#666666"))

	bridgeVStyle = baseStyle.
			Foreground(lipgloss.Color("#666666"))

	emptyStyle = baseStyle.
			Foreground(lipgloss.Color("#333333"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#7B2FBE")).
			Padding(0, 1)

	solvedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00ff00"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginTop(1)
)

// gridView renders the puzzle using a display grid with gap cells between
// logical cells so that bridges between directly adjacent islands are visible.
// Display grid is (2*width-1) x (2*height-1): even coords are logical cells,
// odd coords are gap cells showing bridge connectors.
func gridView(m Model) string {
	solved := m.puzzle.IsSolved()
	dispW := 2*m.puzzle.Width - 1
	dispH := 2*m.puzzle.Height - 1

	var rows []string
	for dy := 0; dy < dispH; dy++ {
		var cells []string
		for dx := 0; dx < dispW; dx++ {
			cells = append(cells, displayCellView(m, dx, dy, solved))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
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
			return bridgeHView(ci.BridgeCount)
		case cellBridgeV:
			return bridgeVView(ci.BridgeCount)
		default:
			return emptyStyle.Width(cellWidth).Render(" ")
		}
	}

	if !evenX && evenY {
		// Horizontal gap between (dx/2, dy/2) and (dx/2+1, dy/2)
		lx, ly := dx/2, dy/2
		return hGapView(m, lx, ly)
	}

	if evenX && !evenY {
		// Vertical gap between (dx/2, dy/2) and (dx/2, dy/2+1)
		lx, ly := dx/2, dy/2
		return vGapView(m, lx, ly)
	}

	// Odd x and odd y — intersection of gaps, always empty
	return emptyStyle.Width(cellWidth).Render(" ")
}

// hGapView renders the horizontal gap between logical cell (lx,ly) and (lx+1,ly).
// Shows a bridge connector if both neighbors are islands connected by a bridge.
func hGapView(m Model, lx, ly int) string {
	left := m.puzzle.FindIslandAt(lx, ly)
	right := m.puzzle.FindIslandAt(lx+1, ly)
	if left != nil && right != nil {
		if b := m.puzzle.GetBridge(left.ID, right.ID); b != nil {
			return bridgeHView(b.Count)
		}
	}
	// Also check if a bridge passes through this gap (non-adjacent islands)
	ci := m.puzzle.CellContent(lx, ly)
	if ci.Kind == cellBridgeH {
		return bridgeHView(ci.BridgeCount)
	}
	ciR := m.puzzle.CellContent(lx+1, ly)
	if ciR.Kind == cellBridgeH {
		return bridgeHView(ciR.BridgeCount)
	}
	return emptyStyle.Width(cellWidth).Render(" ")
}

// vGapView renders the vertical gap between logical cell (lx,ly) and (lx,ly+1).
// Shows a bridge connector if both neighbors are islands connected by a bridge.
func vGapView(m Model, lx, ly int) string {
	top := m.puzzle.FindIslandAt(lx, ly)
	bottom := m.puzzle.FindIslandAt(lx, ly+1)
	if top != nil && bottom != nil {
		if b := m.puzzle.GetBridge(top.ID, bottom.ID); b != nil {
			return bridgeVView(b.Count)
		}
	}
	// Also check if a bridge passes through this gap (non-adjacent islands)
	ci := m.puzzle.CellContent(lx, ly)
	if ci.Kind == cellBridgeV {
		return bridgeVView(ci.BridgeCount)
	}
	ciB := m.puzzle.CellContent(lx, ly+1)
	if ciB.Kind == cellBridgeV {
		return bridgeVView(ciB.BridgeCount)
	}
	return emptyStyle.Width(cellWidth).Render(" ")
}

func islandView(m Model, islandID int, solved bool) string {
	isl := m.puzzle.FindIslandByID(islandID)
	if isl == nil {
		return emptyStyle.Width(cellWidth).Render(" ")
	}

	style := islandDefaultStyle
	current := m.puzzle.BridgeCount(islandID)

	if solved {
		style = islandSatisfiedStyle
	} else if m.selectedIsland != nil && *m.selectedIsland == islandID {
		style = islandSelectedStyle
	} else if m.cursorIsland == islandID {
		style = islandCursorStyle
	} else if current == isl.Required {
		style = islandSatisfiedStyle
	} else if current > isl.Required {
		style = islandOverStyle
	}

	label := fmt.Sprintf("%d", isl.Required)
	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(label)
}

func bridgeHView(count int) string {
	r := "───"
	if count == 2 {
		r = "═══"
	}
	return bridgeHStyle.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func bridgeVView(count int) string {
	r := "│"
	if count == 2 {
		r = "║"
	}
	return bridgeVStyle.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}

func titleBarView(modeName string, solved bool) string {
	title := titleStyle.Render("Hashiwokakero  " + modeName)
	if solved {
		badge := solvedBadgeStyle.Render("  SOLVED")
		return title + badge
	}
	return title
}

func statusBarView(selected bool) string {
	if selected {
		return statusBarStyle.Render("arrows/wasd: build bridge  enter/space/esc: deselect  ctrl+n: menu  ctrl+e: debug")
	}
	return statusBarStyle.Render("arrows/wasd: move  enter/space: select island  ctrl+n: menu  ctrl+e: debug")
}

func infoView(p *Puzzle) string {
	satisfied := 0
	total := len(p.Islands)
	for _, isl := range p.Islands {
		if p.BridgeCount(isl.ID) == isl.Required {
			satisfied++
		}
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(
		fmt.Sprintf("Islands: %d/%d satisfied  Bridges: %d", satisfied, total, len(p.Bridges)),
	))
	return sb.String()
}
