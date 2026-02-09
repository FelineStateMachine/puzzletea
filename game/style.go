package game

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "255"}).
			Background(lipgloss.AdaptiveColor{Light: "55", Dark: "134"}).
			Padding(0, 1)

	solvedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "78"})
)

// TitleBarView renders a title bar with the game name, mode name, and optional solved badge.
func TitleBarView(gameName, modeName string, solved bool) string {
	title := titleStyle.Render(gameName + "  " + modeName)
	if solved {
		badge := solvedBadgeStyle.Render("  SOLVED")
		return title + badge
	}
	return title
}
