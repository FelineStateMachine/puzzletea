package shikaku

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/lipgloss"
)

const cellWidth = 3

// 16-color palette for placed rectangles (ANSI 256).
var rectColors = []lipgloss.AdaptiveColor{
	{Dark: "60", Light: "60"},   // muted purple
	{Dark: "66", Light: "66"},   // teal
	{Dark: "95", Light: "95"},   // mauve
	{Dark: "130", Light: "130"}, // brown/amber
	{Dark: "71", Light: "71"},   // olive green
	{Dark: "132", Light: "132"}, // dusty rose
	{Dark: "37", Light: "37"},   // cyan
	{Dark: "136", Light: "136"}, // dark gold
	{Dark: "97", Light: "97"},   // plum
	{Dark: "29", Light: "29"},   // forest green
	{Dark: "167", Light: "167"}, // salmon
	{Dark: "68", Light: "68"},   // steel blue
	{Dark: "94", Light: "94"},   // rust
	{Dark: "139", Light: "139"}, // lavender
	{Dark: "28", Light: "28"},   // deep green
	{Dark: "133", Light: "133"}, // orchid
}

var (
	baseStyle = lipgloss.NewStyle()

	emptyDotColor = lipgloss.AdaptiveColor{Dark: "242", Light: "249"}
	clueColor     = lipgloss.AdaptiveColor{Dark: "230", Light: "236"}
	clueBgColor   = lipgloss.AdaptiveColor{Dark: "238", Light: "254"}

	cursorBgColor = lipgloss.AdaptiveColor{Dark: "214", Light: "173"}
	cursorFgColor = lipgloss.AdaptiveColor{Dark: "235", Light: "255"}

	previewGoodBg = lipgloss.AdaptiveColor{Dark: "28", Light: "157"}
	previewBadBg  = lipgloss.AdaptiveColor{Dark: "124", Light: "217"}

	solvedBg     = lipgloss.AdaptiveColor{Dark: "22", Light: "151"}
	solvedBorder = lipgloss.AdaptiveColor{Dark: "149", Light: "22"}
	borderColor  = lipgloss.AdaptiveColor{Dark: "240", Light: "250"}

	selectedClueBg = lipgloss.AdaptiveColor{Dark: "172", Light: "179"}

	infoSatisfied = lipgloss.AdaptiveColor{Dark: "22", Light: "149"}
	infoText      = lipgloss.AdaptiveColor{Dark: "137", Light: "137"}

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "244"}).
			MarginTop(1)

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(solvedBorder)
)

func gridView(m Model, solved bool) string {
	// Build preview rectangle if in expansion mode.
	var preview *Rectangle
	var previewClue *Clue
	if m.selectedClue != nil {
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

	// Rectangle background color.
	if owner >= 0 {
		colorIdx := owner % len(rectColors)
		s = s.Background(rectColors[colorIdx]).Foreground(lipgloss.AdaptiveColor{Dark: "255", Light: "255"})
		if clue != nil {
			s = s.Foreground(lipgloss.AdaptiveColor{Dark: "255", Light: "255"}).Bold(true)
		}
	}

	// Preview overlay.
	inPreview := preview != nil && preview.Contains(x, y)
	if inPreview && previewClue != nil {
		if preview.Area() == previewClue.Value && !m.puzzle.Overlaps(*preview, previewClue.ID) {
			s = s.Background(previewGoodBg).Foreground(lipgloss.AdaptiveColor{Dark: "255", Light: "255"})
		} else {
			s = s.Background(previewBadBg).Foreground(lipgloss.AdaptiveColor{Dark: "255", Light: "255"})
		}
	}

	// Selected clue highlight.
	if m.selectedClue != nil && clue != nil && clue.ID == *m.selectedClue && !inPreview {
		s = s.Background(selectedClueBg).Foreground(lipgloss.AdaptiveColor{Dark: "255", Light: "235"}).Bold(true)
	}

	// Solved styling.
	if solved {
		if owner >= 0 {
			colorIdx := owner % len(rectColors)
			s = s.Background(rectColors[colorIdx]).Foreground(lipgloss.AdaptiveColor{Dark: "255", Light: "255"})
		} else {
			s = s.Background(solvedBg)
		}
		if isCursor {
			s = game.CursorSolvedStyle
		}
	} else if isCursor {
		s = s.Background(cursorBgColor).Foreground(cursorFgColor).Bold(true)
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
			return statusBarStyle.Render("arrows: expand  shift+arrows: shrink  enter: confirm  esc: cancel  bkspc: delete  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
		}
		return statusBarStyle.Render("arrows: expand  shift+arrows: shrink  enter: confirm  esc: cancel")
	}
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  enter/space: select clue  bkspc: delete  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return statusBarStyle.Render("enter/space: select clue  bkspc: delete")
}
