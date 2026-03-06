package shikaku

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const (
	cellWidth       = game.DynamicGridCellWidth
	previewRegionID = -2
)

func rectColors() []color.Color { return theme.Current().ThemeColors() }

func gridView(m Model, solved bool) string {
	preview, previewClue, previewValid := activePreview(m)
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  m.puzzle.Width,
		Height: m.puzzle.Height,
		Solved: solved,
		Cell: func(x, y int) string {
			return cellView(m, x, y, preview, previewClue, previewValid, solved)
		},
		ZoneAt: func(x, y int) int {
			return regionTokenAt(m, preview, x, y)
		},
		ZoneFill: func(zone int) color.Color {
			return regionBackground(m, zone, previewClue, previewValid)
		},
	})
}

func cellView(m Model, x, y int, preview *Rectangle, previewClue *Clue, previewValid, solved bool) string {
	p := theme.Current()
	isCursor := m.cursor.X == x && m.cursor.Y == y
	inPreview := preview != nil && preview.Contains(x, y)
	clue := m.puzzle.FindClueAt(x, y)
	token := regionTokenAt(m, preview, x, y)

	display := " · "
	if clue != nil {
		display = fmt.Sprintf("%2d ", clue.Value)
		if clue.Value < 10 {
			display = fmt.Sprintf(" %d ", clue.Value)
		}
	}

	style := lipgloss.NewStyle().
		Width(cellWidth).
		AlignHorizontal(lipgloss.Center)

	fg := p.TextDim
	bg := p.BG
	if clue != nil {
		fg = p.Given
		style = style.Bold(true)
	}

	if regionBG := regionBackground(m, token, previewClue, previewValid); regionBG != nil {
		bg = regionBG
		fg = theme.TextOnBG(regionBG)
	}

	if m.selectedClue != nil && clue != nil && clue.ID == *m.selectedClue && !inPreview {
		bg = p.SelectionBG
		fg = theme.TextOnBG(bg)
		style = style.Bold(true)
	}

	if solved && token == unownedRegionID {
		bg = p.SuccessBG
		fg = p.SolvedFG
	}

	if solved && isCursor {
		style = game.CursorSolvedStyle()
		if clue != nil {
			style = style.Bold(true)
		}
		return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
	}
	if isCursor {
		style = game.CursorStyle()
		if clue != nil {
			style = style.Bold(true)
		}
		return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
	}

	return style.Foreground(fg).Background(bg).Render(display)
}

func activePreview(m Model) (*Rectangle, *Clue, bool) {
	var preview *Rectangle
	var previewClue *Clue

	if m.mousePreview != nil {
		preview = m.mousePreview
		clues := m.puzzle.CluesInRect(*preview)
		if len(clues) == 1 {
			previewClue = clues[0]
		}
	} else if m.selectedClue != nil {
		previewClue = m.puzzle.FindClueByID(*m.selectedClue)
		if previewClue != nil {
			r := m.expansion.rect(previewClue)
			preview = &r
		}
	}

	return preview, previewClue, previewClue != nil && m.puzzle.ValidRectangleForClue(*preview, previewClue.ID)
}

func regionTokenAt(m Model, preview *Rectangle, x, y int) int {
	if preview != nil && preview.Contains(x, y) {
		return previewRegionID
	}

	owner := m.puzzle.CellOwner(x, y)
	if owner >= 0 {
		return owner
	}
	return unownedRegionID
}

func horizontalEdge(m Model, preview *Rectangle, x, y int) bool {
	switch {
	case y <= 0, y >= m.puzzle.Height:
		return true
	default:
		return regionTokenAt(m, preview, x, y-1) != regionTokenAt(m, preview, x, y)
	}
}

func verticalEdge(m Model, preview *Rectangle, x, y int) rune {
	if hasVerticalEdge(m, preview, x, y) {
		return '│'
	}
	return ' '
}

func hasVerticalEdge(m Model, preview *Rectangle, x, y int) bool {
	switch {
	case x <= 0, x >= m.puzzle.Width:
		return true
	default:
		return regionTokenAt(m, preview, x-1, y) != regionTokenAt(m, preview, x, y)
	}
}

func junctionRune(m Model, preview *Rectangle, x, y int) rune {
	north := y > 0 && hasVerticalEdge(m, preview, x, y-1)
	south := y < m.puzzle.Height && hasVerticalEdge(m, preview, x, y)
	west := x > 0 && horizontalEdge(m, preview, x-1, y)
	east := x < m.puzzle.Width && horizontalEdge(m, preview, x, y)

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

func verticalGapBackground(m Model, preview *Rectangle, previewClue *Clue, previewValid bool, x, y int) color.Color {
	if hasVerticalEdge(m, preview, x, y) || x <= 0 || x >= m.puzzle.Width {
		return nil
	}
	token := regionTokenAt(m, preview, x-1, y)
	return regionBackground(m, token, previewClue, previewValid)
}

func horizontalGapBackground(m Model, preview *Rectangle, previewClue *Clue, previewValid bool, x, y int) color.Color {
	if horizontalEdge(m, preview, x, y) || y <= 0 || y >= m.puzzle.Height {
		return nil
	}
	token := regionTokenAt(m, preview, x, y-1)
	return regionBackground(m, token, previewClue, previewValid)
}

func junctionGapBackground(m Model, preview *Rectangle, previewClue *Clue, previewValid bool, x, y int) color.Color {
	if junctionRune(m, preview, x, y) != ' ' {
		return nil
	}

	tokens := make([]int, 0, 4)
	if x > 0 && y > 0 {
		tokens = append(tokens, regionTokenAt(m, preview, x-1, y-1))
	}
	if x < m.puzzle.Width && y > 0 {
		tokens = append(tokens, regionTokenAt(m, preview, x, y-1))
	}
	if x > 0 && y < m.puzzle.Height {
		tokens = append(tokens, regionTokenAt(m, preview, x-1, y))
	}
	if x < m.puzzle.Width && y < m.puzzle.Height {
		tokens = append(tokens, regionTokenAt(m, preview, x, y))
	}
	if len(tokens) != 4 {
		return nil
	}
	for _, token := range tokens[1:] {
		if token != tokens[0] {
			return nil
		}
	}
	return regionBackground(m, tokens[0], previewClue, previewValid)
}

func regionBackground(m Model, token int, previewClue *Clue, previewValid bool) color.Color {
	p := theme.Current()
	switch token {
	case previewRegionID:
		if previewClue == nil {
			return p.ErrorBG
		}
		if previewValid {
			return p.SuccessBG
		}
		return p.ErrorBG
	case unownedRegionID:
		return nil
	default:
		colors := rectColors()
		if len(colors) == 0 {
			return nil
		}
		return colors[token%len(colors)]
	}
}

func infoView(puzzle *Puzzle) string {
	p := theme.Current()

	placed := len(puzzle.Rectangles)
	total := len(puzzle.Clues)
	correct := 0
	for _, r := range puzzle.Rectangles {
		clue := puzzle.FindClueByID(r.ClueID)
		if clue != nil && r.Area() == clue.Value {
			correct++
		}
	}

	satisfiedStyle := lipgloss.NewStyle().Foreground(p.SuccessBorder)
	infoStyle := lipgloss.NewStyle().Foreground(p.Info)

	var sb strings.Builder
	sb.WriteString(infoStyle.Render("Rectangles: "))
	sb.WriteString(satisfiedStyle.Render(fmt.Sprintf("%d", correct)))
	sb.WriteString(infoStyle.Render(fmt.Sprintf("/%d correct  Placed: %d", total, placed)))
	return sb.String()
}

func statusBarView(selected, showFullHelp bool) string {
	if selected {
		if showFullHelp {
			return game.StatusBarStyle().Render("arrows: expand  shift+arrows: shrink  enter: confirm  esc: cancel  bkspc: delete  mouse: drag rect  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
		}
		return game.StatusBarStyle().Render("arrows: expand  shift+arrows: shrink  enter: confirm  esc: cancel  mouse: drag")
	}
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  enter/space: select clue  bkspc: delete  mouse: click clue & drag  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("enter/space: select clue  bkspc: delete  mouse: click & drag")
}

func statusBarVariants() []string {
	return []string{
		statusBarView(false, false),
		statusBarView(false, true),
		statusBarView(true, false),
		statusBarView(true, true),
	}
}
