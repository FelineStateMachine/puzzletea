package shikaku

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 3

// rectColors returns the current theme's chromatic ANSI colors for
// rectangle backgrounds. Colors follow the active palette, so they
// adapt automatically when the user switches themes.
func rectColors() []color.Color { return theme.Current().CardColors() }

func gridBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border)
}

func gridBorderSolvedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().SuccessBorder)
}

func gridView(m Model, solved bool) string {
	// Build preview rectangle for visual feedback.
	var preview *Rectangle
	var previewClue *Clue

	if m.mousePreview != nil {
		// Mouse drag in progress: use the free-form preview rect.
		preview = m.mousePreview
		clues := m.puzzle.CluesInRect(*preview)
		if len(clues) == 1 {
			previewClue = clues[0]
		}
	} else if m.selectedClue != nil {
		// Keyboard expansion mode.
		previewClue = m.puzzle.FindClueByID(*m.selectedClue)
		if previewClue != nil {
			r := m.expansion.rect(previewClue)
			preview = &r
		}
	}

	var rows []string
	for y := range m.puzzle.Height {
		var cells []string
		for x := range m.puzzle.Width {
			cells = append(cells, cellView(m, x, y, solved, preview, previewClue))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if solved {
		return gridBorderSolvedStyle().Render(grid)
	}
	return gridBorderStyle().Render(grid)
}

func cellView(m Model, x, y int, solved bool, preview *Rectangle, previewClue *Clue) string {
	p := theme.Current()
	isCursor := m.cursor.X == x && m.cursor.Y == y
	clue := m.puzzle.FindClueAt(x, y)
	owner := m.puzzle.CellOwner(x, y)

	// Determine display text.
	display := " \u00b7 "
	if clue != nil {
		display = fmt.Sprintf("%2d ", clue.Value)
		if clue.Value < 10 {
			display = fmt.Sprintf(" %d ", clue.Value)
		}
	}

	// Base style.
	s := lipgloss.NewStyle().Foreground(p.TextDim)

	if clue != nil {
		s = lipgloss.NewStyle().Foreground(p.Given).Background(p.Surface).Bold(true)
	}

	// Rectangle background color (from current theme palette).
	colors := rectColors()
	if owner >= 0 {
		bg := colors[owner%len(colors)]
		s = s.Background(bg).Foreground(theme.TextOnBG(bg))
		if clue != nil {
			s = s.Bold(true)
		}
	}

	// Preview overlay.
	inPreview := preview != nil && preview.Contains(x, y)
	if inPreview {
		good := previewClue != nil &&
			preview.Area() == previewClue.Value &&
			!m.puzzle.Overlaps(*preview, previewClue.ID)
		if good {
			s = s.Background(p.SuccessBG).Foreground(theme.TextOnBG(p.SuccessBG))
		} else {
			s = s.Background(p.ErrorBG).Foreground(theme.TextOnBG(p.ErrorBG))
		}
	}

	// Selected clue highlight.
	if m.selectedClue != nil && clue != nil && clue.ID == *m.selectedClue && !inPreview {
		s = s.Background(p.SelectionBG).Foreground(theme.TextOnBG(p.SelectionBG)).Bold(true)
	}

	// Solved styling.
	if solved {
		if owner >= 0 {
			bg := colors[owner%len(colors)]
			s = s.Background(bg).Foreground(theme.TextOnBG(bg))
		} else {
			s = s.Foreground(p.SolvedFG).Background(p.SuccessBG)
		}
		if isCursor {
			s = game.CursorSolvedStyle()
		}
	} else if isCursor {
		s = game.CursorStyle()
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
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
