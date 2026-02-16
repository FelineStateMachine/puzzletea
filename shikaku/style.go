package shikaku

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/FelineStateMachine/puzzletea/game"
)

const cellWidth = 3

// 16-color palette for placed rectangles (ANSI 256).
var rectColors = []compat.AdaptiveColor{
	{Dark: lipgloss.Color("60"), Light: lipgloss.Color("60")},   // muted purple
	{Dark: lipgloss.Color("66"), Light: lipgloss.Color("66")},   // teal
	{Dark: lipgloss.Color("95"), Light: lipgloss.Color("95")},   // mauve
	{Dark: lipgloss.Color("130"), Light: lipgloss.Color("130")}, // brown/amber
	{Dark: lipgloss.Color("71"), Light: lipgloss.Color("71")},   // olive green
	{Dark: lipgloss.Color("132"), Light: lipgloss.Color("132")}, // dusty rose
	{Dark: lipgloss.Color("37"), Light: lipgloss.Color("37")},   // cyan
	{Dark: lipgloss.Color("136"), Light: lipgloss.Color("136")}, // dark gold
	{Dark: lipgloss.Color("97"), Light: lipgloss.Color("97")},   // plum
	{Dark: lipgloss.Color("29"), Light: lipgloss.Color("29")},   // forest green
	{Dark: lipgloss.Color("167"), Light: lipgloss.Color("167")}, // salmon
	{Dark: lipgloss.Color("68"), Light: lipgloss.Color("68")},   // steel blue
	{Dark: lipgloss.Color("94"), Light: lipgloss.Color("94")},   // rust
	{Dark: lipgloss.Color("139"), Light: lipgloss.Color("139")}, // lavender
	{Dark: lipgloss.Color("28"), Light: lipgloss.Color("28")},   // deep green
	{Dark: lipgloss.Color("133"), Light: lipgloss.Color("133")}, // orchid
}

var (
	baseStyle = lipgloss.NewStyle()

	emptyDotColor = compat.AdaptiveColor{Dark: lipgloss.Color("242"), Light: lipgloss.Color("249")}
	clueColor     = compat.AdaptiveColor{Dark: lipgloss.Color("230"), Light: lipgloss.Color("236")}
	clueBgColor   = compat.AdaptiveColor{Dark: lipgloss.Color("238"), Light: lipgloss.Color("254")}

	cursorBgColor = compat.AdaptiveColor{Dark: lipgloss.Color("214"), Light: lipgloss.Color("173")}
	cursorFgColor = compat.AdaptiveColor{Dark: lipgloss.Color("235"), Light: lipgloss.Color("255")}

	previewGoodBg = compat.AdaptiveColor{Dark: lipgloss.Color("28"), Light: lipgloss.Color("157")}
	previewBadBg  = compat.AdaptiveColor{Dark: lipgloss.Color("124"), Light: lipgloss.Color("217")}

	solvedBg     = compat.AdaptiveColor{Dark: lipgloss.Color("22"), Light: lipgloss.Color("151")}
	solvedBorder = compat.AdaptiveColor{Dark: lipgloss.Color("149"), Light: lipgloss.Color("22")}
	borderColor  = compat.AdaptiveColor{Dark: lipgloss.Color("240"), Light: lipgloss.Color("250")}

	selectedClueBg = compat.AdaptiveColor{Dark: lipgloss.Color("172"), Light: lipgloss.Color("179")}

	infoSatisfied = compat.AdaptiveColor{Dark: lipgloss.Color("22"), Light: lipgloss.Color("149")}
	infoText      = compat.AdaptiveColor{Dark: lipgloss.Color("137"), Light: lipgloss.Color("137")}

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
		s = s.Background(rectColors[colorIdx]).Foreground(compat.AdaptiveColor{Dark: lipgloss.Color("255"), Light: lipgloss.Color("255")})
		if clue != nil {
			s = s.Foreground(compat.AdaptiveColor{Dark: lipgloss.Color("255"), Light: lipgloss.Color("255")}).Bold(true)
		}
	}

	// Preview overlay.
	inPreview := preview != nil && preview.Contains(x, y)
	if inPreview && previewClue != nil {
		if preview.Area() == previewClue.Value && !m.puzzle.Overlaps(*preview, previewClue.ID) {
			s = s.Background(previewGoodBg).Foreground(compat.AdaptiveColor{Dark: lipgloss.Color("255"), Light: lipgloss.Color("255")})
		} else {
			s = s.Background(previewBadBg).Foreground(compat.AdaptiveColor{Dark: lipgloss.Color("255"), Light: lipgloss.Color("255")})
		}
	}

	// Selected clue highlight.
	if m.selectedClue != nil && clue != nil && clue.ID == *m.selectedClue && !inPreview {
		s = s.Background(selectedClueBg).Foreground(compat.AdaptiveColor{Dark: lipgloss.Color("255"), Light: lipgloss.Color("235")}).Bold(true)
	}

	// Solved styling.
	if solved {
		if owner >= 0 {
			colorIdx := owner % len(rectColors)
			s = s.Background(rectColors[colorIdx]).Foreground(compat.AdaptiveColor{Dark: lipgloss.Color("255"), Light: lipgloss.Color("255")})
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
			return game.StatusBarStyle.Render("arrows: expand  shift+arrows: shrink  enter: confirm  esc: cancel  bkspc: delete  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
		}
		return game.StatusBarStyle.Render("arrows: expand  shift+arrows: shrink  enter: confirm  esc: cancel")
	}
	if showFullHelp {
		return game.StatusBarStyle.Render("arrows/wasd: move  enter/space: select clue  bkspc: delete  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle.Render("enter/space: select clue  bkspc: delete")
}
