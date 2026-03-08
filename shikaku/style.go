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
			return regionBackground(zone, previewClue, previewValid)
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

	if regionBG := regionBackground(token, previewClue, previewValid); regionBG != nil {
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

func regionBackground(token int, previewClue *Clue, previewValid bool) color.Color {
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
			return game.StatusBarStyle().Render("arrows: expand  shift+arrows: shrink  enter: confirm  bkspc: cancel  mouse: drag rect  esc: menu  ctrl+r: reset  ctrl+h: help")
		}
		return game.StatusBarStyle().Render("arrows: expand  shift+arrows: shrink  enter: confirm  bkspc: cancel  mouse: drag")
	}
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  enter/space: select clue  bkspc: cancel/delete  mouse: click clue & drag  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("enter/space: select clue  bkspc: cancel/delete  mouse: click & drag")
}

func statusBarVariants() []string {
	return []string{
		statusBarView(false, false),
		statusBarView(false, true),
		statusBarView(true, false),
		statusBarView(true, true),
	}
}
