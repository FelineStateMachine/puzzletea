package game

import "github.com/charmbracelet/lipgloss"

// GridBorderColors holds the color configuration for a grid border with
// crosshair and solved-state highlighting.
type GridBorderColors struct {
	BorderFG       lipgloss.AdaptiveColor
	BackgroundBG   lipgloss.AdaptiveColor
	CrosshairBG    lipgloss.AdaptiveColor
	SolvedBorderFG lipgloss.AdaptiveColor
	SolvedBG       lipgloss.AdaptiveColor
}

// BorderChar renders a single border character with optional crosshair highlighting.
func BorderChar(ch string, colors GridBorderColors, solved, highlight bool) string {
	s := lipgloss.NewStyle().Foreground(colors.BorderFG).Background(colors.BackgroundBG)
	if solved {
		s = lipgloss.NewStyle().Foreground(colors.SolvedBorderFG).Background(colors.SolvedBG)
	} else if highlight {
		s = s.Background(colors.CrosshairBG)
	}
	return s.Render(ch)
}

// HBorderRow builds a horizontal border row (top or bottom) with per-column
// crosshair highlighting. cellWidth is the visual width of each cell.
func HBorderRow(w, cursorX, cellWidth int, left, right string, colors GridBorderColors, solved bool) string {
	var parts []string
	parts = append(parts, BorderChar(left, colors, solved, false))
	segment := ""
	for range cellWidth {
		segment += "─"
	}
	for x := range w {
		highlight := !solved && x == cursorX
		s := lipgloss.NewStyle().Foreground(colors.BorderFG).Background(colors.BackgroundBG)
		if solved {
			s = lipgloss.NewStyle().Foreground(colors.SolvedBorderFG).Background(colors.SolvedBG)
		} else if highlight {
			s = s.Background(colors.CrosshairBG)
		}
		parts = append(parts, s.Render(segment))
	}
	parts = append(parts, BorderChar(right, colors, solved, false))
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
