package shikaku

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 3

// rectColors returns the current theme's chromatic ANSI colors for
// rectangle backgrounds. Colors follow the active palette, so they
// adapt automatically when the user switches themes.
func rectColors() []color.Color { return theme.Current().CardColors() }

var (
	baseStyle = lipgloss.NewStyle()

	emptyDotColor = compat.AdaptiveColor{Dark: lipgloss.Color("242"), Light: lipgloss.Color("249")}
	clueColor     = compat.AdaptiveColor{Dark: lipgloss.Color("230"), Light: lipgloss.Color("236")}
	clueBgColor   = compat.AdaptiveColor{Dark: lipgloss.Color("238"), Light: lipgloss.Color("254")}

	// Cursor colors resolved at render time via game.CursorStyle().

	previewGoodBg = compat.AdaptiveColor{Dark: lipgloss.Color("28"), Light: lipgloss.Color("157")}
	previewBadBg  = compat.AdaptiveColor{Dark: lipgloss.Color("124"), Light: lipgloss.Color("217")}

	solvedBg     = compat.AdaptiveColor{Dark: lipgloss.Color("22"), Light: lipgloss.Color("151")}
	solvedBorder = compat.AdaptiveColor{Dark: lipgloss.Color("149"), Light: lipgloss.Color("22")}
	borderColor  = compat.AdaptiveColor{Dark: lipgloss.Color("240"), Light: lipgloss.Color("250")}

	infoSatisfied = compat.AdaptiveColor{Dark: lipgloss.Color("22"), Light: lipgloss.Color("149")}
	infoText      = compat.AdaptiveColor{Dark: lipgloss.Color("180"), Light: lipgloss.Color("95")}

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(solvedBorder)
)

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
		return gridBorderSolvedStyle.Render(grid)
	}
	return gridBorderStyle.Render(grid)
}

func cellView(m Model, x, y int, solved bool, preview *Rectangle, previewClue *Clue) string {
	isCursor := m.cursor.X == x && m.cursor.Y == y
	clue := m.puzzle.FindClueAt(x, y)
	owner := m.puzzle.CellOwner(x, y)

	// Determine display text.
	display := " · "
	if clue != nil {
		display = fmt.Sprintf("%2d ", clue.Value)
		if clue.Value < 10 {
			display = fmt.Sprintf(" %d ", clue.Value)
		}
	}

	// Base style.
	s := baseStyle.Foreground(emptyDotColor)

	if clue != nil {
		s = baseStyle.Foreground(clueColor).Background(clueBgColor).Bold(true)
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
			s = s.Background(previewGoodBg).Foreground(theme.TextOnBG(previewGoodBg))
		} else {
			s = s.Background(previewBadBg).Foreground(theme.TextOnBG(previewBadBg))
		}
	}

	// Selected clue highlight.
	if m.selectedClue != nil && clue != nil && clue.ID == *m.selectedClue && !inPreview {
		bg := theme.Current().Highlight
		s = s.Background(bg).Foreground(theme.TextOnBG(bg)).Bold(true)
	}

	// Solved styling.
	if solved {
		if owner >= 0 {
			bg := colors[owner%len(colors)]
			s = s.Background(bg).Foreground(theme.TextOnBG(bg))
		} else {
			s = s.Background(solvedBg)
		}
		if isCursor {
			s = game.CursorSolvedStyle()
		}
	} else if isCursor {
		s = game.CursorStyle()
	}

	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(display)
}

func infoView(p *Puzzle) string {
	placed := len(p.Rectangles)
	total := len(p.Clues)
	correct := 0
	for _, r := range p.Rectangles {
		clue := p.FindClueByID(r.ClueID)
		if clue != nil && r.Area() == clue.Value {
			correct++
		}
	}

	satisfiedStyle := lipgloss.NewStyle().Foreground(infoSatisfied)
	infoStyle := lipgloss.NewStyle().Foreground(infoText)

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
