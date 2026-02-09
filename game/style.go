package game

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#7B2FBE")).
			Padding(0, 1)

	solvedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00ff00"))
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
